package auth

import (
	"math/big"
	"time"
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
