package auth

import (
	"encoding/json"
	"net/http"

	"github.com/francisco3ferraz/zk-auth/internal/errors"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.NewBadRequestError("invalid request body").WriteResponse(w)
		return
	}

	if err := h.validateRegisterRequest(&req); err != nil {
		err.WriteResponse(w)
		return
	}

	resp, err := h.service.Register(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteResponse(w)
		} else {
			errors.NewInternalError("registration failed").WriteResponse(w)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleChallenge(w http.ResponseWriter, r *http.Request) {
	var req ChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.NewBadRequestError("invalid request body").WriteResponse(w)
		return
	}

	if req.Username == "" || req.ClientA == "" {
		errors.NewValidationError("username and client_a are required").WriteResponse(w)
		return
	}

	resp, err := h.service.StartChallenge(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteResponse(w)
		} else {
			errors.NewInternalError("challenge failed").WriteResponse(w)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleVerify(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.NewBadRequestError("invalid request body").WriteResponse(w)
		return
	}

	if req.SessionID == "" || req.ClientProof == "" {
		errors.NewValidationError("session_id and client_proof are required").WriteResponse(w)
		return
	}

	resp, err := h.service.VerifyChallenge(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteResponse(w)
		} else {
			errors.NewInternalError("verification failed").WriteResponse(w)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	token := h.extractToken(r)
	if token == "" {
		errors.NewAuthenticationError("missing authorization token").WriteResponse(w)
		return
	}

	resp, err := h.service.Logout(r.Context(), token)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			appErr.WriteResponse(w)
		} else {
			errors.NewInternalError("logout failed").WriteResponse(w)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
