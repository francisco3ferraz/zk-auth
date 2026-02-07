package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/francisco3ferraz/zk-auth/internal/auth"
	"github.com/francisco3ferraz/zk-auth/internal/errors"
	"github.com/francisco3ferraz/zk-auth/internal/logger"
	"go.uber.org/zap"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)
				errors.NewInternalError("internal server error").WriteResponse(w)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		logger.Info("HTTP Request",
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration),
			zap.String("user_agent", r.UserAgent()),
		)
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.NewAuthenticationError("missing authorization header").WriteResponse(w)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				errors.NewAuthenticationError("invalid authorization header format").WriteResponse(w)
				return
			}

			token := parts[1]

			claims, err := authService.ValidateToken(token)
			if err != nil {
				errors.NewAuthenticationError("invalid or expired token").WriteResponse(w)
				return
			}

			ctx := context.WithValue(r.Context(), auth.ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	type contextKey string
	const requestIDKey contextKey = "request_id"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
