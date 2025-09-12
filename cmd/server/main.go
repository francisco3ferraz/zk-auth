package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/francisco3ferraz/zk-auth/internal/config"
	"github.com/francisco3ferraz/zk-auth/internal/database"
	"github.com/francisco3ferraz/zk-auth/internal/logger"
	"github.com/francisco3ferraz/zk-auth/internal/server"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	if err := run(ctx, signalChan, cancel); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, signalChan chan os.Signal, cancel context.CancelFunc) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	if err := logger.Initialize(cfg.Server.Environment); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	// Set up signal handling
	go func() {
		sig := <-signalChan
		logger.Info("Received shutdown signal",
			zap.String("signal", sig.String()))
		cancel()
	}()

	db, err := initializeDatabase(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	if err := runMigrationsWithRetry(ctx, cfg); err != nil {
		logger.Error("Migration failed", zap.Error(err))
		if cfg.Server.Environment != "production" {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
		logger.Warn("Continuing without migrations in production")
	}

	srv, err := server.New(cfg, db)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start()
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("Shutting down server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func initializeDatabase(ctx context.Context, cfg *config.Config) (*database.DB, error) {
	var db *database.DB
	var err error

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			db, err = database.New(&cfg.Database)
			if err == nil {
				return db, nil
			}
			logger.Warn("Failed to connect to database",
				zap.Int("attempt", i+1),
				zap.Int("maxRetries", maxRetries),
				zap.Error(err))
			if i < maxRetries-1 {
				time.Sleep(time.Second * time.Duration(i+1))
			}
		}
	}
	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

func runMigrationsWithRetry(ctx context.Context, cfg *config.Config) error {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := database.RunMigrations(cfg.Database.URL, "migrations")
			if err == nil {
				logger.Info("Database migrations completed successfully")
				return nil
			}
			logger.Warn("Migration attempt failed",
				zap.Int("attempt", i+1),
				zap.Int("maxRetries", maxRetries),
				zap.Error(err))
			if i < maxRetries-1 {
				time.Sleep(time.Second * 2)
			}
		}
	}
	return fmt.Errorf("migrations failed after %d attempts", maxRetries)
}
