package auth

import (
	"context"
	"database/sql"
	"encoding/hex"
	"math/big"
	"sync"
	"time"

	"github.com/francisco3ferraz/zk-auth/internal/config"
	"github.com/francisco3ferraz/zk-auth/internal/crypto"
	"github.com/francisco3ferraz/zk-auth/internal/errors"
	"github.com/francisco3ferraz/zk-auth/internal/model"
)

type Service struct {
	srp          *crypto.SRP
	userRepo     *model.UserRepository
	sessionRepo  *model.SessionRepository
	config       *config.Config
	challenges   map[string]*AuthChallenge // In-memory challenge storage
	challengesMu sync.RWMutex              // Protects challenges map
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

// StartCleanup starts a background goroutine that periodically removes expired challenges.
// It should be called when the server starts and will stop when the context is cancelled.
func (s *Service) StartCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				s.cleanupExpiredChallenges()
			}
		}
	}()
}

func (s *Service) cleanupExpiredChallenges() {
	s.challengesMu.Lock()
	defer s.challengesMu.Unlock()

	now := time.Now()
	for id, challenge := range s.challenges {
		if now.Sub(challenge.CreatedAt) > 5*time.Minute {
			delete(s.challenges, id)
		}
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

func (s *Service) StartChallenge(ctx context.Context, req *ChallengeRequest) (*ChallengeResponse, error) {
	clientABytes, err := hex.DecodeString(req.ClientA)
	if err != nil {
		return nil, errors.NewBadRequestError("invalid client_a format")
	}
	clientA := new(big.Int).SetBytes(clientABytes)

	if clientA.Sign() == 0 {
		return nil, errors.NewBadRequestError("invalid client_a value")
	}

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal whether user exists
			return nil, errors.NewAuthenticationError("invalid credentials")
		}
		return nil, errors.NewInternalError("failed to retrieve user")
	}

	verifier := new(big.Int).SetBytes(user.Verifier)
	serverSecret, serverB, err := s.srp.GenerateServerKeys(verifier)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate server keys")
	}

	session := &model.Session{
		UserID:       user.ID,
		Challenge:    clientA.Bytes(),
		ServerSecret: serverSecret.Bytes(),
		ExpiresAt:    time.Now().Add(5 * time.Minute),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, errors.NewInternalError("failed to create session")
	}

	s.challengesMu.Lock()
	s.challenges[session.ID] = &AuthChallenge{
		SessionID:    session.ID,
		Username:     user.Username,
		ClientA:      clientA,
		ServerB:      serverB,
		ServerSecret: serverSecret,
		Salt:         user.Salt,
		Verifier:     user.Verifier,
		CreatedAt:    time.Now(),
	}
	s.challengesMu.Unlock()

	return &ChallengeResponse{
		SessionID: session.ID,
		Salt:      hex.EncodeToString(user.Salt),
		ServerB:   hex.EncodeToString(serverB.Bytes()),
	}, nil
}

func (s *Service) VerifyChallenge(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	s.challengesMu.Lock()
	challenge, exists := s.challenges[req.SessionID]
	if !exists {
		s.challengesMu.Unlock()
		return nil, errors.NewAuthenticationError("invalid or expired session")
	}
	delete(s.challenges, req.SessionID)
	s.challengesMu.Unlock()

	if time.Since(challenge.CreatedAt) > 5*time.Minute {
		return nil, errors.NewSessionExpiredError()
	}

	clientProof, err := hex.DecodeString(req.ClientProof)
	if err != nil {
		return nil, errors.NewBadRequestError("invalid client_proof format")
	}

	// Compute u = H(A | B)
	u := s.srp.ComputeU(challenge.ClientA, challenge.ServerB)

	verifier := new(big.Int).SetBytes(challenge.Verifier)
	serverKey, err := s.srp.ComputeServerSessionKey(
		challenge.ClientA,
		challenge.ServerSecret,
		verifier,
		u,
	)
	if err != nil {
		return nil, errors.NewInternalError("failed to compute session key")
	}

	if !s.srp.VerifyClientProof(
		challenge.Username,
		challenge.Salt,
		challenge.ClientA,
		challenge.ServerB,
		serverKey,
		clientProof,
	) {
		return nil, errors.NewAuthenticationError("invalid credentials")
	}
	serverProof := s.srp.ComputeServerProof(challenge.ClientA, clientProof, serverKey)

	token, expiresAt, err := s.generateToken(challenge.SessionID, challenge.Username)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate token")
	}

	session, err := s.sessionRepo.GetByID(ctx, challenge.SessionID)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve session")
	}

	session.Token = token
	session.ExpiresAt = expiresAt
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, errors.NewInternalError("failed to update session")
	}

	return &VerifyResponse{
		Token:       token,
		ServerProof: hex.EncodeToString(serverProof),
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *Service) Logout(ctx context.Context, token string) (*LogoutResponse, error) {
	claims, err := s.verifyToken(token)
	if err != nil {
		return nil, errors.NewAuthenticationError("invalid token")
	}

	if err := s.sessionRepo.Delete(ctx, claims.SessionID); err != nil {
		return nil, errors.NewInternalError("failed to delete session")
	}

	return &LogoutResponse{
		Message: "Logged out successfully",
	}, nil
}

func (s *Service) ValidateToken(token string) (*TokenClaims, error) {
	return s.verifyToken(token)
}
