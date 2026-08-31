package auth

import (
	"context"
	"sync"
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

type mockDeviceBinder struct {
	mu    sync.Mutex
	bound map[string]string
}

func (m *mockDeviceBinder) BindOrVerifyDevice(ctx context.Context, uid, deviceID string) (bool, error) {
	if deviceID == "" {
		return false, ErrDeviceRequired
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	cur, exists := m.bound[uid]
	if !exists || cur == "" {
		m.bound[uid] = deviceID
		return true, nil
	}
	if cur == deviceID {
		return false, nil
	}
	return false, ErrDeviceBlocked
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

	uid, newlyBound, err := provider.Verify(context.Background(), "demo1", "secret", DevicePayload{DeviceID: "dev-1"})
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
	_, _, err = provider.Verify(context.Background(), "demo1", "wrong", DevicePayload{DeviceID: "dev-1"})
	if err != ErrCredentialInvalid {
		t.Fatalf("expected ErrCredentialInvalid, got %v", err)
	}

	// Unknown user
	_, _, err = provider.Verify(context.Background(), "unknown", "secret", DevicePayload{DeviceID: "dev-1"})
	if err != ErrCredentialInvalid {
		t.Fatalf("expected ErrCredentialInvalid, got %v", err)
	}
}

func TestLocalAuthProvider_DeviceBindingMatrix(t *testing.T) {
	hasher := NewBcryptHasher()
	pwd, _ := hasher.Hash("secret")
	store := &mockLocalStore{
		user: &LocalUser{
			Username:     "userA",
			PasswordHash: pwd,
			ProfileUID:   "uid-A",
		},
	}
	binder := &mockDeviceBinder{bound: make(map[string]string)}
	provider := NewLocalAuthProviderWithBinder(hasher, store, binder)

	// 1. Missing Device ID rejected
	_, _, err := provider.Verify(context.Background(), "userA", "secret", DevicePayload{DeviceID: ""})
	if err != ErrDeviceRequired {
		t.Fatalf("expected ErrDeviceRequired, got %v", err)
	}

	// 2. First Login Device 1 -> PASS (Bound)
	uid, newlyBound, err := provider.Verify(context.Background(), "userA", "secret", DevicePayload{DeviceID: "device-1"})
	if err != nil {
		t.Fatalf("first login failed: %v", err)
	}
	if uid != "uid-A" || !newlyBound {
		t.Fatalf("expected uid-A and newlyBound=true, got uid=%s, newlyBound=%v", uid, newlyBound)
	}

	// 3. Same Device 1 Login Again -> PASS
	uid, newlyBound, err = provider.Verify(context.Background(), "userA", "secret", DevicePayload{DeviceID: "device-1"})
	if err != nil {
		t.Fatalf("same device relogin failed: %v", err)
	}
	if newlyBound {
		t.Fatal("expected newlyBound=false for same device relogin")
	}

	// 4. Different Device 2 Login -> BLOCK (403 / ErrDeviceBlocked)
	_, _, err = provider.Verify(context.Background(), "userA", "secret", DevicePayload{DeviceID: "device-2"})
	if err != ErrDeviceBlocked {
		t.Fatalf("expected ErrDeviceBlocked for different device, got %v", err)
	}

	// 5. Concurrent First Login Race Test: Device 1 & Device 2 concurrently
	// Reset binder
	binder.mu.Lock()
	binder.bound = make(map[string]string)
	binder.mu.Unlock()

	var wg sync.WaitGroup
	results := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		devID := "dev-race-1"
		if i%2 == 1 {
			devID = "dev-race-2"
		}
		go func(id string) {
			defer wg.Done()
			_, _, e := provider.Verify(context.Background(), "userA", "secret", DevicePayload{DeviceID: id})
			results <- e
		}(devID)
	}
	wg.Wait()
	close(results)

	successes := 0
	blocks := 0
	for e := range results {
		if e == nil {
			successes++
		} else if e == ErrDeviceBlocked {
			blocks++
		}
	}
	if successes == 0 || blocks == 0 {
		t.Fatalf("expected concurrent race to allow exactly 1 winning device family and block others, got successes=%d, blocks=%d", successes, blocks)
	}

	// 6. Admin Reset Device -> Allow Device 2
	binder.mu.Lock()
	delete(binder.bound, "uid-A")
	binder.mu.Unlock()

	uid, newlyBound, err = provider.Verify(context.Background(), "userA", "secret", DevicePayload{DeviceID: "device-2"})
	if err != nil {
		t.Fatalf("login after device reset failed: %v", err)
	}
	if !newlyBound {
		t.Fatal("expected newlyBound=true for new device after reset")
	}

	// 7. Device 1 now blocked
	_, _, err = provider.Verify(context.Background(), "userA", "secret", DevicePayload{DeviceID: "device-1"})
	if err != ErrDeviceBlocked {
		t.Fatalf("expected ErrDeviceBlocked for old device after reset, got %v", err)
	}
}
