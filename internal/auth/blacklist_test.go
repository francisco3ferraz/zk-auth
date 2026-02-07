package auth

import (
	"testing"
	"time"
)

func TestTokenBlacklist_RevokeAndCheck(t *testing.T) {
	bl := NewTokenBlacklist()

	token := "test-token-123"
	expiresAt := time.Now().Add(1 * time.Hour)

	// Token should not be revoked initially
	if bl.IsRevoked(token) {
		t.Error("token should not be revoked initially")
	}

	// Revoke the token
	bl.Revoke(token, expiresAt)

	// Token should now be revoked
	if !bl.IsRevoked(token) {
		t.Error("token should be revoked after calling Revoke")
	}
}

func TestTokenBlacklist_DifferentTokensIndependent(t *testing.T) {
	bl := NewTokenBlacklist()

	token1 := "token-1"
	token2 := "token-2"

	bl.Revoke(token1, time.Now().Add(1*time.Hour))

	if !bl.IsRevoked(token1) {
		t.Error("token1 should be revoked")
	}

	if bl.IsRevoked(token2) {
		t.Error("token2 should not be revoked")
	}
}
