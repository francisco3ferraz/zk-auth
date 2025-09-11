package main

import (
	"github.com/francisco3ferraz/zk-auth/internal/config"
	"github.com/francisco3ferraz/zk-auth/internal/database"
	"github.com/francisco3ferraz/zk-auth/internal/model"
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

	userRepo := model.NewUserRepository(db.Pool())
	sessionRepo := model.NewSessionRepository(db.Pool())

	_ = userRepo
	_ = sessionRepo
}
