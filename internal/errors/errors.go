package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrCodeInternal   ErrorCode = "INTERNAL_ERROR"
	ErrCodeNotFound   ErrorCode = "NOT_FOUND"
	ErrCodeBadRequest ErrorCode = "BAD_REQUEST"
	ErrCodeConflict   ErrorCode = "CONFLICT"
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
)

type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) WriteResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	json.NewEncoder(w).Encode(e)
}

func NewInternalError(details string) *AppError {
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    "An internal error occurred",
		Details:    details,
		StatusCode: http.StatusInternalServerError,
	}
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}
