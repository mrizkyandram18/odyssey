package auth

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrLocalUserNotFound = errors.New("local user not found")
)

type LocalUser struct {
	ID           string
	Username     string
	PasswordHash string
	ProfileUID   string
}

type LocalUserStore interface {
	GetLocalUserByUsername(ctx context.Context, username string) (*LocalUser, error)
}

type LocalAuthProvider struct {
	hasher PasswordHasher
	store  LocalUserStore
}

func NewLocalAuthProvider(hasher PasswordHasher, store LocalUserStore) *LocalAuthProvider {
	return &LocalAuthProvider{
		hasher: hasher,
		store:  store,
	}
}

func (p *LocalAuthProvider) Verify(ctx context.Context, identifier, credential string, device DevicePayload) (string, bool, error) {
	if identifier == "" {
		return "", false, ErrUIDRequired
	}
	if credential == "" {
		return "", false, ErrCredentialRequired
	}

	user, err := p.store.GetLocalUserByUsername(ctx, identifier)
	if err != nil {
		if errors.Is(err, ErrLocalUserNotFound) {
			return "", false, ErrCredentialInvalid // don't leak user existence
		}
		return "", false, fmt.Errorf("local auth provider: %w", err)
	}

	if err := p.hasher.Verify(user.PasswordHash, credential); err != nil {
		return "", false, ErrCredentialInvalid
	}

	// Local provider doesn't track new device bindings in the same way as gatekeeper, always false for newlyBound
	return user.ProfileUID, false, nil
}
