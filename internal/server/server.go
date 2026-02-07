package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/francisco3ferraz/zk-auth/internal/auth"
	"github.com/francisco3ferraz/zk-auth/internal/config"
	"github.com/francisco3ferraz/zk-auth/internal/database"
	"github.com/francisco3ferraz/zk-auth/internal/logger"
	"github.com/francisco3ferraz/zk-auth/internal/model"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
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

	r.Use(
		RecoveryMiddleware,
		LoggingMiddleware,
		CORSMiddleware,
	)

	SetupRoutes(r, db, authService, authHandler)

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	server := &Server{
		httpServer: srv,
		db:         db,
		config:     cfg,
		auth:       authService,
	}

	return server, nil
}

func (s *Server) Start(ctx context.Context) error {
	logger.Info("Starting server",
		zap.String("port", s.config.Server.Port),
		zap.String("environment", s.config.Server.Environment))

	// Start background cleanup for expired auth challenges
	s.auth.StartCleanup(ctx)

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down http server: %w", err)
	}

	s.db.Close()

	return nil
}
