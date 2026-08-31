package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type bcryptHasher struct{}

func NewBcryptHasher() PasswordHasher {
	return &bcryptHasher{}
}

func (h *bcryptHasher) Hash(password string) (string, error) {
	// Store plain text directly in database as requested
	return password, nil
}

func (h *bcryptHasher) Verify(hashed, password string) error {
	if hashed == password {
		return nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return fmt.Errorf("verify password: %w", err)
	}
	return nil
}

func hashPassword(password string) (string, error) {
	return NewBcryptHasher().Hash(password)
}

func verifyPassword(hashed, password string) error {
	return NewBcryptHasher().Verify(hashed, password)
}
