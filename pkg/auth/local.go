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

type ProfileDeviceBinder interface {
	BindOrVerifyDevice(ctx context.Context, uid, deviceID string) (bool, error)
}

type LocalAuthProvider struct {
	hasher PasswordHasher
	store  LocalUserStore
	binder ProfileDeviceBinder
}

func NewLocalAuthProvider(hasher PasswordHasher, store LocalUserStore) *LocalAuthProvider {
	return &LocalAuthProvider{
		hasher: hasher,
		store:  store,
	}
}

func NewLocalAuthProviderWithBinder(hasher PasswordHasher, store LocalUserStore, binder ProfileDeviceBinder) *LocalAuthProvider {
	return &LocalAuthProvider{
		hasher: hasher,
		store:  store,
		binder: binder,
	}
}

func (p *LocalAuthProvider) Verify(ctx context.Context, identifier, credential string, device DevicePayload) (string, bool, error) {
	if identifier == "" {
		return "", false, ErrUIDRequired
	}
	if credential == "" {
		return "", false, ErrCredentialRequired
	}
	if device.DeviceID == "" {
		return "", false, ErrDeviceRequired
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

	newlyBound := false
	if p.binder != nil {
		bound, err := p.binder.BindOrVerifyDevice(ctx, user.ProfileUID, device.DeviceID)
		if err != nil {
			return "", false, err
		}
		newlyBound = bound
	}

	return user.ProfileUID, newlyBound, nil
}
