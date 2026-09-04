package db

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/auth"
)

type mockRPCClient struct {
	getData      []byte
	getErr       error
	rpcData      []byte
	rpcErr       error
	mutateData   []byte
	mutateErr    error
	lastPayload  map[string]any
	lastParams   string
	getCalls     []string
}

func (m *mockRPCClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	m.getCalls = append(m.getCalls, table+"?"+params)
	if table == "odyssey_user_profiles" && strings.Contains(params, "device_id") {
		// Simulate no bound device
		data, _ := json.Marshal([]map[string]string{{"device_id": ""}})
		return data, nil
	}
	return m.getData, m.getErr
}
func (m *mockRPCClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	if mp, ok := payload.(map[string]any); ok {
		m.lastPayload = mp
	}
	m.lastParams = params
	return m.mutateData, m.mutateErr
}
func (m *mockRPCClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	if mp, ok := payload.(map[string]any); ok {
		m.lastPayload = mp
	}
	m.lastParams = params
	return m.mutateData, m.mutateErr
}
func (m *mockRPCClient) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	return m.rpcData, m.rpcErr
}
func (m *mockRPCClient) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return "", nil
}

func TestBindOrVerifyDevice_FallbackUsesRealTimestamp(t *testing.T) {
	client := &mockRPCClient{
		rpcErr:  auth.ErrDeviceRequired, // force fallback path by returning error that is not P0022/P0021 but will trigger fallback
		getData: []byte("[]"),
	}
	// Make RPC return generic error to trigger fallback branch that does GetBoundDeviceID
	client.rpcErr = &mockError{msg: "rpc unavailable"}
	store := NewProfileStore(client)
	_, err := store.BindOrVerifyDevice(context.Background(), "user-1", "web_test_device_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.lastPayload == nil {
		t.Fatal("expected mutate payload")
	}
	val, ok := client.lastPayload["device_bound_at"]
	if !ok {
		t.Fatal("device_bound_at not set")
	}
	str, ok := val.(string)
	if !ok {
		t.Fatalf("device_bound_at not string: %v", val)
	}
	if str == "now()" {
		t.Fatalf("device_bound_at should be real timestamp, got literal now()")
	}
	if _, err := time.Parse(time.RFC3339, str); err != nil {
		t.Fatalf("device_bound_at not RFC3339: %v err %v", str, err)
	}
	if client.lastPayload["device_id"] != "web_test_device_123" {
		t.Fatalf("device_id mismatch")
	}
}

type mockError struct{ msg string }
func (m *mockError) Error() string { return m.msg }

func TestBindOrVerifyDevice_BlockedError(t *testing.T) {
	client := &mockRPCClient{
		rpcData: []byte("[]"),
		rpcErr:  &mockError{msg: "Akun sudah terhubung ke perangkat lain P0022"},
	}
	store := NewProfileStore(client)
	_, err := store.BindOrVerifyDevice(context.Background(), "user-1", "web_other")
	if err != auth.ErrDeviceBlocked {
		t.Fatalf("expected ErrDeviceBlocked, got %v", err)
	}
}
