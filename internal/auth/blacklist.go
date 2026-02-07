package auth

import (
	"sync"
	"time"
)

// TokenBlacklist manages revoked tokens in memory
// Tokens are stored with their expiration time for automatic cleanup
type TokenBlacklist struct {
	tokens map[string]time.Time // token -> expiration time
	mu     sync.RWMutex
}

// NewTokenBlacklist creates a new token blacklist
func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}
	go bl.cleanupLoop()
	return bl
}

// Revoke adds a token to the blacklist
func (bl *TokenBlacklist) Revoke(token string, expiresAt time.Time) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.tokens[token] = expiresAt
}

// IsRevoked checks if a token has been revoked
func (bl *TokenBlacklist) IsRevoked(token string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	_, exists := bl.tokens[token]
	return exists
}

// cleanupLoop periodically removes expired tokens from the blacklist
func (bl *TokenBlacklist) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bl.cleanup()
	}
}

func (bl *TokenBlacklist) cleanup() {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := time.Now()
	for token, expiresAt := range bl.tokens {
		if now.After(expiresAt) {
			delete(bl.tokens, token)
		}
	}
}
