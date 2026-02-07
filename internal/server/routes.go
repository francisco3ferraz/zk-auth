package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/francisco3ferraz/zk-auth/internal/auth"
	"github.com/francisco3ferraz/zk-auth/internal/database"
	"github.com/gorilla/mux"
)

func SetupRoutes(r *mux.Router, db *database.DB, authService *auth.Service, authHandler *auth.Handler) {
	r.HandleFunc("/health", handleHealth(db)).Methods("GET")
	r.HandleFunc("/", handleAPIInfo).Methods("GET")

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/register", authHandler.HandleRegister).Methods("POST")
	api.HandleFunc("/auth/challenge", authHandler.HandleChallenge).Methods("POST")
	api.HandleFunc("/auth/verify", authHandler.HandleVerify).Methods("POST")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(AuthMiddleware(authService))

	protected.HandleFunc("/auth/logout", authHandler.HandleLogout).Methods("POST")
	protected.HandleFunc("/auth/refresh", authHandler.HandleRefresh).Methods("POST")
	protected.HandleFunc("/auth/password", authHandler.HandleChangePassword).Methods("PUT")
	protected.HandleFunc("/profile", authHandler.HandleProfile).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(handleNotFound)
}

func handleHealth(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "healthy"
		if err := db.Health(r.Context()); err != nil {
			dbStatus = "unhealthy"
		}

		health := map[string]interface{}{
			"status":    "healthy",
			"database":  dbStatus,
			"timestamp": time.Now().Format(time.RFC3339),
		}

		statusCode := http.StatusOK
		if dbStatus != "healthy" {
			health["status"] = "degraded"
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(health)
	}
}

func handleAPIInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"name":    "Zero-Knowledge Authentication Server",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"register":  "POST /api/v1/register",
			"challenge": "POST /api/v1/auth/challenge",
			"verify":    "POST /api/v1/auth/verify",
			"logout":    "POST /api/v1/auth/logout",
			"health":    "GET /health",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "endpoint not found",
		"path":  r.URL.Path,
	})
}
