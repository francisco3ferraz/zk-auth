package server

import (
	"encoding/json"
	"net/http"

	"github.com/francisco3ferraz/zk-auth/internal/auth"
	"github.com/francisco3ferraz/zk-auth/internal/database"
	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router, db *database.DB, authService *auth.Service) {
	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/health", handleHealth(db)).Methods("GET")
	api.HandleFunc("/", handleAPIInfo).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(handleNotFound)
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
			"timestamp": http.TimeFormat,
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
			"health": "GET /health",
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
