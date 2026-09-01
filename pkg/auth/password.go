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
	// Support legacy plaintext hashes during migration period.
	// If stored hash is plaintext (no $2a$ prefix), fall back to direct comparison.
	if !isBcryptHash(hashed) {
		if hashed == password {
			return nil
		}
		return fmt.Errorf("verify password: credential invalid")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return fmt.Errorf("verify password: %w", err)
	}
	return nil
}

func isBcryptHash(s string) bool {
	return len(s) >= 4 && s[0] == '$' && s[1] == '2' && (s[2] == 'a' || s[2] == 'b' || s[2] == 'y') && s[3] == '$'
}

func hashPassword(password string) (string, error) {
	return NewBcryptHasher().Hash(password)
}

func verifyPassword(hashed, password string) error {
	return NewBcryptHasher().Verify(hashed, password)
}
