package main

import "github.com/francisco3ferraz/zk-auth/internal/crypto"

func main() {
	/*
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
	*/

	srp := crypto.NewSRP()

	salt, err := srp.GenerateSalt()
	if err != nil {
		panic(err)
	}

	v, err := srp.ComputeVerifier("username", "password", salt)
	if err != nil {
		panic(err)
	}

	println("Salt:", salt)
	println("Verifier:", v)

}
