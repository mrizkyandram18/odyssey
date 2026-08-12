package db

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"odyssey/pkg/game/audit"
)

type mockSupabaseClientForAudit struct {
	mutateCalled  bool
	mutateTable   string
	mutateMethod  string
	mutatePayload map[string]any
	mutateParams  string
	mutateResult  []byte
	mutateErr     error
}

func (m *mockSupabaseClientForAudit) Get(ctx context.Context, table, params string) ([]byte, error) {
	return nil, nil
}

func (m *mockSupabaseClientForAudit) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	m.mutateCalled = true
	m.mutateMethod = method
	m.mutateTable = table
	m.mutatePayload = payload.(map[string]any)
	m.mutateParams = params
	return m.mutateResult, m.mutateErr
}

func (m *mockSupabaseClientForAudit) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	m.mutateCalled = true
	m.mutateMethod = method
	m.mutateTable = table
	m.mutatePayload = payload.(map[string]any)
	m.mutateParams = params
	return m.mutateResult, m.mutateErr
}

func TestAuditStore_Write(t *testing.T) {
	mockClient := &mockSupabaseClientForAudit{
		mutateResult: []byte(`[]`),
	}
	store := NewAuditStore(mockClient)

	entry := audit.AuditEntry{
		Resource:   "realms",
		ResourceID: "forest",
		Operation:  "create",
		AdminUID:   "admin1",
		RequestID:  "req-abc-123",
		OldValue:   json.RawMessage(`{"slug":"forest"}`),
		NewValue:   json.RawMessage(`{"slug":"forest-updated"}`),
	}

	err := store.Write(context.Background(), entry)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !mockClient.mutateCalled {
		t.Error("expected Mutate to be called")
	}
	if mockClient.mutateTable != "odyssey_audit_logs" {
		t.Errorf("expected table odyssey_audit_logs, got %s", mockClient.mutateTable)
	}
	if mockClient.mutateMethod != "POST" {
		t.Errorf("expected POST, got %s", mockClient.mutateMethod)
	}
	if mockClient.mutatePayload["resource"] != "realms" {
		t.Errorf("expected resource=realms, got %v", mockClient.mutatePayload["resource"])
	}
	if mockClient.mutatePayload["admin_uid"] != "admin1" {
		t.Errorf("expected admin_uid=admin1, got %v", mockClient.mutatePayload["admin_uid"])
	}
	if mockClient.mutatePayload["operation"] != "create" {
		t.Errorf("expected operation=create, got %v", mockClient.mutatePayload["operation"])
	}
	if mockClient.mutatePayload["request_id"] != "req-abc-123" {
		t.Errorf("expected request_id=req-abc-123, got %v", mockClient.mutatePayload["request_id"])
	}
}

func TestAuditStore_Write_NilValues(t *testing.T) {
	mockClient := &mockSupabaseClientForAudit{
		mutateResult: []byte(`[]`),
	}
	store := NewAuditStore(mockClient)

	entry := audit.AuditEntry{
		Resource:  "missions",
		Operation: "delete",
		AdminUID:  "admin1",
	}

	err := store.Write(context.Background(), entry)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if mockClient.mutatePayload["resource_id"] != "" {
		t.Errorf("expected empty resource_id, got %v", mockClient.mutatePayload["resource_id"])
	}
}

func TestAuditLogger_Log(t *testing.T) {
	mockClient := &mockSupabaseClientForAudit{
		mutateResult: []byte(`[]`),
	}
	store := NewAuditStore(mockClient)
	logger := audit.NewLogger(store)

	err := logger.Log(context.Background(), "realms", "forest", "update", "admin1",
		map[string]any{"slug": "forest"},
		map[string]any{"slug": "forest-updated"})
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}
	if !mockClient.mutateCalled {
		t.Error("expected Mutate to be called")
	}
}

func TestAuditLogger_LogError(t *testing.T) {
	mockClient := &mockSupabaseClientForAudit{
		mutateResult: []byte(`[]`),
	}
	store := NewAuditStore(mockClient)
	logger := audit.NewLogger(store)

	logger.LogError(context.Background(), "missions", "quest1", "update", "admin1", fmt.Errorf("something went wrong"))

	if mockClient.mutatePayload["resource"] != "missions" {
		t.Errorf("expected resource=missions, got %v", mockClient.mutatePayload["resource"])
	}
}
