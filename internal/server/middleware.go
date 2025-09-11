package server

import (
	"log"
	"net/http"

	"github.com/francisco3ferraz/zk-auth/internal/errors"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				errors.NewInternalError("internal server error").WriteResponse(w)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
