package server

import (
	"fmt"
	"net/http"

	"github.com/francisco3ferraz/zk-auth/internal/auth"
	"github.com/francisco3ferraz/zk-auth/internal/config"
	"github.com/francisco3ferraz/zk-auth/internal/database"
	"github.com/francisco3ferraz/zk-auth/internal/model"
	"github.com/gorilla/mux"
)

type Server struct {
	httpServer *http.Server
	db         *database.DB
	config     *config.Config
	auth       *auth.Service
}

func New(cfg *config.Config, db *database.DB) (*Server, error) {
	userRepo := model.NewUserRepository(db.Pool())
	sessionRepo := model.NewSessionRepository(db.Pool())

	authService := auth.NewService(userRepo, sessionRepo, cfg)
	authHandler := auth.NewHandler(authService)

	r := mux.NewRouter()

	r.Use(RecoveryMiddleware)
	SetupRoutes(r, db, authService, authHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: r,
	}

	server := &Server{
		httpServer: srv,
		db:         db,
		config:     cfg,
		auth:       authService,
	}

	return server, nil
}

func (s *Server) Start() error {
	fmt.Printf("Starting server on port %s...\n", s.config.Server.Port)
	return s.httpServer.ListenAndServe()
}
