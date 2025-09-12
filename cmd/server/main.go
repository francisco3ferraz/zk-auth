package main

import (
	"log"

	"github.com/francisco3ferraz/zk-auth/internal/config"
	"github.com/francisco3ferraz/zk-auth/internal/database"
	"github.com/francisco3ferraz/zk-auth/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	db, err := database.New(&cfg.Database)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := database.RunMigrations(cfg.Database.URL, "migrations"); err != nil {
		log.Printf("Migration failed: %v", err)

		if cfg.Server.Environment == "production" {
			log.Println("Continuing without migrations in production")
		} else {
			db.Close()
			panic("failed to run migrations")
		}
	}

	srv, err := server.New(cfg, db)
	if err != nil {
		panic(err)
	}

	if err := srv.Start(); err != nil {
		panic(err)
	}
}
