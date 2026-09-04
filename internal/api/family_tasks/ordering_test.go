package family_tasks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
)

type mockClient struct {
	lastParams string
	taskParams string
}

func (m *mockClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	m.lastParams = params
	if table == "odyssey_user_profiles" {
		return json.Marshal([]map[string]any{{"uid": "user-1", "family_id": "fam-1", "is_active": true, "created_at": "2026-09-01T00:00:00Z"}})
	}
	if table == "odyssey_tasks" {
		m.taskParams = params
		return json.Marshal([]map[string]any{})
	}
	if table == "odyssey_task_submissions" {
		return json.Marshal([]map[string]any{})
	}
	if table == "odyssey_system_config" {
		return json.Marshal([]map[string]any{{"value": "Asia/Jakarta"}})
	}
	return []byte("[]"), nil
}
func (m *mockClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *mockClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *mockClient) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *mockClient) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return "", nil
}

func TestFamilyTasksOrderingDeterministic(t *testing.T) {
	client := &mockClient{}
	api := NewAPI(client)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
	claims := &auth.SessionClaims{UID: "user-1", FamilyID: "fam-1", Role: "MEMBER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.HandleGetToday(rec, req)
	if !strings.Contains(client.taskParams, "order=step_order.asc,id.asc") {
		t.Fatalf("expected deterministic ordering step_order.asc,id.asc, got %s", client.taskParams)
	}
}
