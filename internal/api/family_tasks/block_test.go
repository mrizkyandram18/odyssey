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

func TestHandleSubmit_BlockedUserRejected(t *testing.T) {
	task := []TaskRecord{{ID: 1, FamilyID: "fam-1", Title: "Task", TaskType: "QUIZ", StepOrder: 1, IsActive: true}}
	taskBytes, _ := json.Marshal(task)
	client := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				return []byte(`[{"is_active":false}]`), nil
			}
			if table == "odyssey_tasks" {
				return taskBytes, nil
			}
			return []byte("[]"), nil
		},
	}
	api := NewAPI(client)
	body := `{"answers":{"q1":"A"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/1/submit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "blocked_user", FamilyID: "fam-1", Role: "MEMBER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.Handler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for blocked user, got %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "diblokir") {
		t.Fatalf("expected block message got %s", rec.Body.String())
	}
}

func TestHandleUpload_BlockedUserRejected(t *testing.T) {
	client := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				return []byte(`[{"is_active":false}]`), nil
			}
			return []byte("[]"), nil
		},
	}
	api := NewAPI(client)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/upload", nil)
	claims := &auth.SessionClaims{UID: "blocked_user", FamilyID: "fam-1", Role: "MEMBER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.Handler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 blocked upload got %d", rec.Code)
	}
}
