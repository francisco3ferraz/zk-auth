package auth

import (
	"github.com/francisco3ferraz/zk-auth/internal/config"
	"github.com/francisco3ferraz/zk-auth/internal/crypto"
	"github.com/francisco3ferraz/zk-auth/internal/model"
)

type Service struct {
	srp         *crypto.SRP
	userRepo    *model.UserRepository
	sessionRepo *model.SessionRepository
	config      *config.Config
	challenges  map[string]*AuthChallenge // In-memory challenge storage
}

func NewService(userRepo *model.UserRepository, sessionRepo *model.SessionRepository, cfg *config.Config) *Service {
	return &Service{
		srp:         crypto.NewSRP(),
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		config:      cfg,
		challenges:  make(map[string]*AuthChallenge),
	}
}
