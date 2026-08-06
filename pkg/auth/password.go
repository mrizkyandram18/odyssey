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
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func (h *bcryptHasher) Verify(hashed, password string) error {
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
