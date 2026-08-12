package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type mockSupabaseClient struct {
	data              []byte
	err               error
	getCalls          []string
	lastMutatePayload any
	lastMutateTable   string
	lastMutatePrefer  string
	lastMutateParams  string
	lastMutateMethod  string
}

func (m *mockSupabaseClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	m.getCalls = append(m.getCalls, table+"?"+params)
	return m.data, m.err
}

func (m *mockSupabaseClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	m.lastMutateMethod = method
	m.lastMutateTable = table
	m.lastMutatePayload = payload
	m.lastMutateParams = params
	return m.data, m.err
}

func (m *mockSupabaseClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	m.lastMutateMethod = method
	m.lastMutateTable = table
	m.lastMutatePayload = payload
	m.lastMutateParams = params
	m.lastMutatePrefer = prefer
	return m.data, m.err
}

func TestProfileStore_GetUserProfile_Found(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]UserProfile{
		{
			UID:          "user-1",
			FamilyID:       "crew-1",
			ExplorerName: "Alice",
			Role:         "SEEKER",
			Level:        1,
			XP:           100,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	})
	store := NewProfileStore(&mockSupabaseClient{data: data})
	profile, err := store.GetUserProfile(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.UID != "user-1" {
		t.Errorf("expected UID user-1, got %s", profile.UID)
	}
	if profile.FamilyID != "crew-1" {
		t.Errorf("expected FamilyID crew-1, got %s", profile.FamilyID)
	}
	if profile.ExplorerName != "Alice" {
		t.Errorf("expected ExplorerName Alice, got %s", profile.ExplorerName)
	}
	if profile.Role != "SEEKER" {
		t.Errorf("expected Role SEEKER, got %s", profile.Role)
	}
}

func TestProfileStore_GetUserProfile_NotFound(t *testing.T) {
	data, _ := json.Marshal([]UserProfile{})
	store := NewProfileStore(&mockSupabaseClient{data: data})
	_, err := store.GetUserProfile(context.Background(), "user-1")
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestProfileStore_GetUserProfile_Error(t *testing.T) {
	store := NewProfileStore(&mockSupabaseClient{err: errors.New("network")})
	_, err := store.GetUserProfile(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProfileStore_GetUserProfile_UidFilter(t *testing.T) {
	data, _ := json.Marshal([]UserProfile{})
	client := &mockSupabaseClient{data: data}
	store := NewProfileStore(client)
	_, _ = store.GetUserProfile(context.Background(), "user-1")
	if len(client.getCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(client.getCalls))
	}
	call := client.getCalls[0]
	expected := "odyssey_user_profiles?uid=eq.user-1"
	if call != expected {
		t.Errorf("expected query %q, got %q", expected, call)
	}
}

func TestProfileStore_GetPasswordHash_Found(t *testing.T) {
	data, _ := json.Marshal([]struct {
		PasswordHash string `json:"password_hash"`
	}{
		{PasswordHash: "$2a$10$hashvalue"},
	})
	store := NewProfileStore(&mockSupabaseClient{data: data})
	hash, err := store.GetPasswordHash(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "$2a$10$hashvalue" {
		t.Errorf("expected hash, got %s", hash)
	}
}

func TestProfileStore_GetPasswordHash_NotFound(t *testing.T) {
	data, _ := json.Marshal([]struct {
		PasswordHash string `json:"password_hash"`
	}{})
	store := NewProfileStore(&mockSupabaseClient{data: data})
	hash, err := store.GetPasswordHash(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "" {
		t.Errorf("expected empty hash, got %s", hash)
	}
}

func TestProfileStore_GetBoundDeviceID_Found(t *testing.T) {
	data, _ := json.Marshal([]struct {
		DeviceID string `json:"device_id"`
	}{
		{DeviceID: "device-123"},
	})
	store := NewProfileStore(&mockSupabaseClient{data: data})
	id, err := store.GetBoundDeviceID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "device-123" {
		t.Errorf("expected device-123, got %s", id)
	}
}

func TestProfileStore_GetBoundDeviceID_NotFound(t *testing.T) {
	data, _ := json.Marshal([]struct {
		DeviceID string `json:"device_id"`
	}{})
	store := NewProfileStore(&mockSupabaseClient{data: data})
	id, err := store.GetBoundDeviceID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "" {
		t.Errorf("expected empty device ID, got %s", id)
	}
}

func TestProfileStore_ImplementsProfileStore(t *testing.T) {
	var _ ProfileStore = NewProfileStore(&mockSupabaseClient{})
}
