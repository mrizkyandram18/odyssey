package auth

import (
	"context"
	"testing"
)

type mockLocalStore struct {
	user *LocalUser
	err  error
}

func (m *mockLocalStore) GetLocalUserByUsername(ctx context.Context, username string) (*LocalUser, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.user == nil || m.user.Username != username {
		return nil, ErrLocalUserNotFound
	}
	return m.user, nil
}

func TestLocalAuthProvider_Verify(t *testing.T) {
	hasher := NewBcryptHasher()
	pwd, _ := hasher.Hash("secret")

	store := &mockLocalStore{
		user: &LocalUser{
			Username:     "demo1",
			PasswordHash: pwd,
			ProfileUID:   "demo-uid-1",
		},
	}
	provider := NewLocalAuthProvider(hasher, store)

	uid, newlyBound, err := provider.Verify(context.Background(), "demo1", "secret", DevicePayload{})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if uid != "demo-uid-1" {
		t.Fatalf("expected demo-uid-1, got %v", uid)
	}
	if newlyBound {
		t.Fatal("expected false for newly bound")
	}

	// Wrong password
	_, _, err = provider.Verify(context.Background(), "demo1", "wrong", DevicePayload{})
	if err != ErrCredentialInvalid {
		t.Fatalf("expected ErrCredentialInvalid, got %v", err)
	}

	// Unknown user
	_, _, err = provider.Verify(context.Background(), "unknown", "secret", DevicePayload{})
	if err != ErrCredentialInvalid {
		t.Fatalf("expected ErrCredentialInvalid, got %v", err)
	}
}
