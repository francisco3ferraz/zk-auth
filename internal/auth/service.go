package auth

import (
	"context"

	"github.com/francisco3ferraz/zk-auth/internal/config"
	"github.com/francisco3ferraz/zk-auth/internal/crypto"
	"github.com/francisco3ferraz/zk-auth/internal/errors"
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

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.NewInternalError("failed to check user existence")
	}
	if exists {
		return nil, errors.NewConflictError("username already exists")
	}

	salt, err := s.srp.GenerateSalt()
	if err != nil {
		return nil, errors.NewInternalError("failed to generate salt")
	}

	verifier, err := s.srp.ComputeVerifier(req.Username, req.Password, salt)
	if err != nil {
		return nil, errors.NewInternalError("failed to compute verifier")
	}

	user := &model.User{
		Username: req.Username,
		Salt:     salt,
		Verifier: verifier.Bytes(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.NewInternalError("failed to create user")
	}

	return &RegisterResponse{
		UserID:   user.ID,
		Username: user.Username,
		Message:  "User registered successfully",
	}, nil
}
