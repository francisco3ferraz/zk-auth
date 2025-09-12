package auth

import (
	"github.com/francisco3ferraz/zk-auth/internal/errors"
)

func (h *Handler) validateRegisterRequest(req *RegisterRequest) *errors.AppError {
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return errors.NewValidationError("username must be between 3 and 50 characters")
	}

	if len(req.Password) < 8 {
		return errors.NewValidationError("password must be at least 8 characters")
	}

	for _, char := range req.Username {
		if !isAlphanumeric(char) && char != '_' {
			return errors.NewValidationError("username can only contain letters, numbers, and underscores")
		}
	}

	return nil
}

func isAlphanumeric(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}
