package db

import (
	"context"
	"encoding/json"
	"testing"
)

type mockAdminClient struct {
	getResult    []byte
	getErr       error
	mutateResult []byte
	mutateErr    error
	mutateMethod string
	mutateTable  string
	mutateParams string
}

func (m *mockAdminClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	return m.getResult, m.getErr
}

func (m *mockAdminClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	m.mutateMethod = method
	m.mutateTable = table
	m.mutateParams = params
	return m.mutateResult, m.mutateErr
}

func (m *mockAdminClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	m.mutateMethod = method
	m.mutateTable = table
	m.mutateParams = params
	return m.mutateResult, m.mutateErr
}

func TestAdminStore_GetBySlug(t *testing.T) {
	rows := []map[string]any{
		{"id": float64(1), "slug": "forest", "name": "Forest", "published": true},
	}
	data, _ := json.Marshal(rows)
	mockClient := &mockAdminClient{
		getResult: data,
	}
	store := NewDefinitionStore(mockClient)

	result, err := store.GetBySlug(context.Background(), "odyssey_journey_definitions", "forest")
	if err != nil {
		t.Fatalf("GetBySlug failed: %v", err)
	}
	if result["slug"] != "forest" {
		t.Errorf("expected slug=forest, got %v", result["slug"])
	}
}

func TestAdminStore_GetBySlug_NotFound(t *testing.T) {
	data, _ := json.Marshal([]map[string]any{})
	mockClient := &mockAdminClient{
		getResult: data,
	}
	store := NewDefinitionStore(mockClient)

	result, err := store.GetBySlug(context.Background(), "odyssey_journey_definitions", "unknown")
	if err != nil {
		t.Fatalf("GetBySlug should not error for not found: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestAdminStore_Create(t *testing.T) {
	returnedRow := []map[string]any{
		{"id": float64(1), "slug": "new-journey", "name": "New Journey", "published": false, "version": 1},
	}
	data, _ := json.Marshal(returnedRow)
	mockClient := &mockAdminClient{
		mutateResult: data,
	}
	store := NewDefinitionStore(mockClient)

	result, err := store.Create(context.Background(), "odyssey_journey_definitions", map[string]any{
		"slug": "new-journey",
		"name": "New Journey",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if result["slug"] != "new-journey" {
		t.Errorf("expected slug=new-journey, got %v", result["slug"])
	}
	if mockClient.mutateMethod != "POST" {
		t.Errorf("expected POST, got %s", mockClient.mutateMethod)
	}
}

func TestAdminStore_UpdateDraft(t *testing.T) {
	rows := []map[string]any{
		{"id": float64(1), "slug": "forest", "name": "Forest", "version": float64(1)},
	}
	getData, _ := json.Marshal(rows)

	mockClient := &mockAdminClient{
		getResult:    getData,
		mutateResult: []byte(`[]`),
	}
	store := NewDefinitionStore(mockClient)

	err := store.UpdateDraft(context.Background(), "odyssey_journey_definitions", "forest",
		map[string]any{"name": "Updated Forest"}, "admin1")
	if err != nil {
		t.Fatalf("UpdateDraft failed: %v", err)
	}
	if mockClient.mutateMethod != "PATCH" {
		t.Errorf("expected PATCH, got %s", mockClient.mutateMethod)
	}
}

func TestAdminStore_Publish(t *testing.T) {
	rows := []map[string]any{
		{"id": float64(1), "slug": "forest", "name": "Forest", "published": false, "version": float64(1)},
	}
	getData, _ := json.Marshal(rows)

	mockClient := &mockAdminClient{
		getResult:    getData,
		mutateResult: []byte(`[]`),
	}
	store := NewDefinitionStore(mockClient)

	err := store.Publish(context.Background(), "odyssey_journey_definitions", "forest", "admin1")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	if mockClient.mutateMethod != "PATCH" {
		t.Errorf("expected PATCH, got %s", mockClient.mutateMethod)
	}
}

func TestAdminStore_SoftDelete(t *testing.T) {
	rows := []map[string]any{
		{"id": float64(1), "slug": "quest1", "title": "Mission 1", "published": true, "version": float64(1)},
	}
	getData, _ := json.Marshal(rows)

	mockClient := &mockAdminClient{
		getResult:    getData,
		mutateResult: []byte(`[]`),
	}
	store := NewDefinitionStore(mockClient)

	err := store.SoftDelete(context.Background(), "odyssey_quest_definitions", "quest1")
	if err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}
	if mockClient.mutateMethod != "PATCH" {
		t.Errorf("expected PATCH, got %s", mockClient.mutateMethod)
	}
}

func TestAdminStore_Restore(t *testing.T) {
	rows := []map[string]any{
		{"id": float64(1), "slug": "quest1", "title": "Mission 1", "published": true, "version": float64(1), "deleted_at": "2025-01-01T00:00:00Z"},
	}
	getData, _ := json.Marshal(rows)

	mockClient := &mockAdminClient{
		getResult:    getData,
		mutateResult: []byte(`[]`),
	}
	store := NewDefinitionStore(mockClient)

	err := store.Restore(context.Background(), "odyssey_quest_definitions", "quest1")
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if mockClient.mutateMethod != "PATCH" {
		t.Errorf("expected PATCH, got %s", mockClient.mutateMethod)
	}
}

func TestAdminStore_ListAll(t *testing.T) {
	rows := []map[string]any{
		{"id": float64(1), "slug": "forest", "published": true},
		{"id": float64(2), "slug": "desert", "published": true},
	}
	data, _ := json.Marshal(rows)
	mockClient := &mockAdminClient{
		getResult: data,
	}
	store := NewDefinitionStore(mockClient)

	result, err := store.ListAll(context.Background(), "odyssey_journey_definitions", false)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result))
	}
}

func TestAdminStore_Delete_NotFound(t *testing.T) {
	data, _ := json.Marshal([]map[string]any{})
	mockClient := &mockAdminClient{
		getResult: data,
	}
	store := NewDefinitionStore(mockClient)

	err := store.SoftDelete(context.Background(), "odyssey_quest_definitions", "nonexistent")
	if err == nil {
		t.Error("expected error for deleting nonexistent definition")
	}
}
