package audit

import (
	"context"
	"testing"

	"odyssey/pkg/observability"
)

type mockStore struct {
	writtenEntry AuditEntry
	writtenErr   error
}

func (m *mockStore) Write(ctx context.Context, entry AuditEntry) error {
	m.writtenEntry = entry
	return m.writtenErr
}

func (m *mockStore) List(ctx context.Context, filter ListFilter) ([]AuditEntry, error) {
	return nil, nil
}

func TestLogger_Log_InjectsRequestID(t *testing.T) {
	store := &mockStore{}
	logger := NewLogger(store)

	ctx := observability.WithRequestID(context.Background(), "req-trace-123")
	err := logger.Log(ctx, "realms", "forest", "create", "admin1", nil, nil)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	if store.writtenEntry.RequestID != "req-trace-123" {
		t.Errorf("expected request_id 'req-trace-123', got '%s'", store.writtenEntry.RequestID)
	}
	if store.writtenEntry.AdminUID != "admin1" {
		t.Errorf("expected admin_uid 'admin1', got '%s'", store.writtenEntry.AdminUID)
	}
}

func TestLogger_Log_NoRequestID(t *testing.T) {
	store := &mockStore{}
	logger := NewLogger(store)

	err := logger.Log(context.Background(), "realms", "forest", "create", "admin1", nil, nil)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	if store.writtenEntry.RequestID != "" {
		t.Errorf("expected empty request_id, got '%s'", store.writtenEntry.RequestID)
	}
}

func TestLogger_LogError_InjectsRequestID(t *testing.T) {
	store := &mockStore{}
	logger := NewLogger(store)

	ctx := observability.WithRequestID(context.Background(), "req-err-456")
	logger.LogError(ctx, "realms", "forest", "delete", "admin1", errTest)

	if store.writtenEntry.RequestID != "req-err-456" {
		t.Errorf("expected request_id 'req-err-456', got '%s'", store.writtenEntry.RequestID)
	}
}

var errTest = newTestError("test error")

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

func newTestError(msg string) error { return &testError{msg: msg} }
