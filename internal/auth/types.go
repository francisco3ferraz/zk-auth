package auth

import (
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthChallenge struct {
	SessionID    string
	Username     string
	ClientA      *big.Int
	ServerB      *big.Int
	ServerSecret *big.Int
	Salt         []byte
	Verifier     []byte
	CreatedAt    time.Time
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

type ChallengeRequest struct {
	Username string `json:"username" validate:"required"`
	ClientA  string `json:"client_a" validate:"required"`
}

type ChallengeResponse struct {
	SessionID string `json:"session_id"`
	Salt      string `json:"salt"`
	ServerB   string `json:"server_b"`
}

type VerifyRequest struct {
	SessionID   string `json:"session_id" validate:"required"`
	ClientProof string `json:"client_proof" validate:"required"`
}

type VerifyResponse struct {
	Token       string    `json:"token"`
	ServerProof string    `json:"server_proof"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type TokenClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}
