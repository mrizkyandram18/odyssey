package admin_tasks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
)

type mockSupabaseClient struct {
	getResp    []byte
	getErr     error
	mutateResp []byte
	mutateErr  error
	rpcResp    []byte
	rpcErr     error
	uploadResp string
	uploadErr  error

	lastMutatePayload any
}

func (m *mockSupabaseClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getResp, nil
}

func (m *mockSupabaseClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	m.lastMutatePayload = payload
	if m.mutateErr != nil {
		return nil, m.mutateErr
	}
	return m.mutateResp, nil
}

func (m *mockSupabaseClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return m.Mutate(ctx, method, table, payload, params)
}

func (m *mockSupabaseClient) RPC(ctx context.Context, fnName string, payload any) ([]byte, error) {
	if m.rpcErr != nil {
		return nil, m.rpcErr
	}
	return m.rpcResp, nil
}

func (m *mockSupabaseClient) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	if m.uploadErr != nil {
		return "", m.uploadErr
	}
	return m.uploadResp, nil
}

func TestAdminVerifySubmissionAuth(t *testing.T) {
	client := &mockSupabaseClient{}
	api := NewAPI(client)

	// Non-guide user should get 403 Forbidden
	req := httptest.NewRequest(http.MethodPost, "/api/admin/submissions/1/verify", strings.NewReader(`{"status":"APPROVED"}`))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "user-123", Role: "SEEKER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for non-guide user, got %d", rec.Code)
	}

	// Guide user should proceed to verify
	client.rpcResp = []byte(`{"success":true,"status":"APPROVED","coins_earned":50,"xp_earned":100}`)
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "GUIDE"}
	req2 := httptest.NewRequest(http.MethodPost, "/api/admin/submissions/1/verify", strings.NewReader(`{"status":"APPROVED"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2 = req2.WithContext(auth.ContextWithClaims(req2.Context(), guideClaims))

	rec2 := httptest.NewRecorder()
	api.Handler(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected status 200 for guide user, got %d: %s", rec2.Code, rec2.Body.String())
	}
}

func TestAdminCreateTask_AssignsFamilyIDAndDefaults(t *testing.T) {
	client := &mockSupabaseClient{
		mutateResp: []byte(`{"id":1,"title":"Tugas Baru"}`),
	}
	api := NewAPI(client)

	payload := `{"title":"Tugas Baru","task_type":"DOCUMENT_UPLOAD","step_order":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "GUIDE"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	mutateMap, ok := client.lastMutatePayload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload on task create")
	}
	if mutateMap["family_id"] != "fam-alpha" {
		t.Errorf("expected family_id to be 'fam-alpha', got %v", mutateMap["family_id"])
	}
	if mutateMap["created_by"] != "admin-1" {
		t.Errorf("expected created_by to be 'admin-1', got %v", mutateMap["created_by"])
	}
	if mutateMap["evaluation_type"] != "ADMIN_REVIEW" {
		t.Errorf("expected evaluation_type to default to ADMIN_REVIEW for DOCUMENT_UPLOAD, got %v", mutateMap["evaluation_type"])
	}
}

func TestAdminCreateTask_MiniGameEvaluationType(t *testing.T) {
	client := &mockSupabaseClient{
		mutateResp: []byte(`{"id":2,"title":"Memory Game"}`),
	}
	api := NewAPI(client)

	payload := `{"title":"Memory Game","task_type":"MINI_GAME","step_order":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "GUIDE"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", rec.Code)
	}

	mutateMap := client.lastMutatePayload.(map[string]any)
	if mutateMap["evaluation_type"] != "AUTO" {
		t.Errorf("expected evaluation_type to default to AUTO for MINI_GAME, got %v", mutateMap["evaluation_type"])
	}
}

func TestAdminUpdateTask_FamilyMismatchForbidden(t *testing.T) {
	// Task belongs to fam-beta
	taskData, _ := json.Marshal([]map[string]any{
		{"id": 10, "family_id": "fam-beta"},
	})
	client := &mockSupabaseClient{
		getResp: taskData,
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/tasks/10", strings.NewReader(`{"title":"Hacked"}`))
	req.Header.Set("Content-Type", "application/json")
	// Admin belongs to fam-alpha
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "GUIDE"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for cross-family task update, got %d", rec.Code)
	}
}

func TestAdminCreateTask_ValidationRules(t *testing.T) {
	client := &mockSupabaseClient{
		mutateResp: []byte(`{"id":3,"title":"Task"}`),
	}
	api := NewAPI(client)
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "GUIDE"}

	testCases := []struct {
		name        string
		payload     string
		expectCode  int
		errContains string
	}{
		{
			name:        "Empty title",
			payload:     `{"title":"","task_type":"VIDEO"}`,
			expectCode:  http.StatusBadRequest,
			errContains: "judul tugas tidak boleh kosong",
		},
		{
			name:        "Invalid task type",
			payload:     `{"title":"Tugas","task_type":"NON_EXISTENT_TYPE"}`,
			expectCode:  http.StatusBadRequest,
			errContains: "tipe tugas tidak valid",
		},
		{
			name:        "Quiz without questions",
			payload:     `{"title":"Kuis","task_type":"QUIZ","config":{"questions":[]}}`,
			expectCode:  http.StatusBadRequest,
			errContains: "minimal 1 pertanyaan",
		},
		{
			name:        "Quiz with missing answer key",
			payload:     `{"title":"Kuis","task_type":"QUIZ","config":{"questions":[{"id":"1","question":"Q1","options":["A","B"],"correct_answer":""}]}}`,
			expectCode:  http.StatusBadRequest,
			errContains: "kunci jawaban",
		},
		{
			name:        "Quiz with less than 2 options",
			payload:     `{"title":"Kuis","task_type":"QUIZ","config":{"questions":[{"id":"1","question":"Q1","options":["A"],"correct_answer":"A"}]}}`,
			expectCode:  http.StatusBadRequest,
			errContains: "minimal 2 pilihan jawaban",
		},
		{
			name:        "Text response with invalid character limits",
			payload:     `{"title":"Teks","task_type":"TEXT_RESPONSE","config":{"minimum_characters":500,"maximum_characters":10}}`,
			expectCode:  http.StatusBadRequest,
			errContains: "batasan karakter tidak valid",
		},
		{
			name:        "Invalid reward coins (negative/zero)",
			payload:     `{"title":"Tugas","task_type":"VIDEO","reward_coins":-10}`,
			expectCode:  http.StatusBadRequest,
			errContains: "reward coins tidak valid",
		},
		{
			name:        "Valid video task",
			payload:     `{"title":"Video Edukasi","task_type":"VIDEO","reward_coins":50,"config":{"video_url":"https://youtube.com/watch?v=123"}}`,
			expectCode:  http.StatusCreated,
			errContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", strings.NewReader(tc.payload))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

			rec := httptest.NewRecorder()
			api.Handler(rec, req)

			if rec.Code != tc.expectCode {
				t.Fatalf("expected code %d, got %d: %s", tc.expectCode, rec.Code, rec.Body.String())
			}
			if tc.errContains != "" && !strings.Contains(rec.Body.String(), tc.errContains) {
				t.Errorf("expected error containing %q, got %s", tc.errContains, rec.Body.String())
			}
		})
	}
}

func TestAdminConfig_Get(t *testing.T) {
	configData, _ := json.Marshal([]map[string]any{
		{"key": "redemption_start_day", "value": "21"},
		{"key": "redemption_end_day", "value": "26"},
	})
	client := &mockSupabaseClient{
		getResp: configData,
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "GUIDE"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if res["redemption_start_day"] != float64(21) || res["redemption_end_day"] != float64(26) {
		t.Fatalf("expected 21-26, got %v", res)
	}
}

func TestAdminConfig_NonGuideForbidden(t *testing.T) {
	client := &mockSupabaseClient{}
	api := NewAPI(client)

	// GET as non-guide
	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	seekerClaims := &auth.SessionClaims{UID: "user-1", Role: "SEEKER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), seekerClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", rec.Code)
	}

	// POST as non-guide
	reqPost := httptest.NewRequest(http.MethodPost, "/api/admin/config", strings.NewReader(`{"start_day":10,"end_day":15}`))
	reqPost.Header.Set("Content-Type", "application/json")
	reqPost = reqPost.WithContext(auth.ContextWithClaims(reqPost.Context(), seekerClaims))

	recPost := httptest.NewRecorder()
	api.Handler(recPost, reqPost)

	if recPost.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden on update, got %d", recPost.Code)
	}
}

func TestAdminConfig_UpdateSuccessAndValidation(t *testing.T) {
	client := &mockSupabaseClient{}
	api := NewAPI(client)
	guideClaims := &auth.SessionClaims{UID: "admin-1", Role: "GUIDE"}

	cases := []struct {
		name       string
		payload    string
		expectCode int
	}{
		{
			name:       "Valid range update",
			payload:    `{"start_day":10,"end_day":15}`,
			expectCode: http.StatusOK,
		},
		{
			name:       "Invalid range: start < 1",
			payload:    `{"start_day":0,"end_day":15}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Invalid range: end > 31",
			payload:    `{"start_day":10,"end_day":32}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Invalid range: start > end",
			payload:    `{"start_day":20,"end_day":10}`,
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/admin/config", strings.NewReader(tc.payload))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

			rec := httptest.NewRecorder()
			api.Handler(rec, req)

			if rec.Code != tc.expectCode {
				t.Fatalf("expected code %d, got %d: %s", tc.expectCode, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAdminDuplicateTask_Success(t *testing.T) {
	origTask, _ := json.Marshal([]map[string]any{
		{
			"id":              10,
			"family_id":       "fam-alpha",
			"title":           "Original Task",
			"description":     "Task Desc",
			"task_type":       "VIDEO",
			"evaluation_type": "AUTO",
			"step_order":      1,
			"reward_coins":    50,
			"reward_xp":       100,
			"target_scope":    "ALL",
			"is_active":       true,
		},
	})

	client := &mockSupabaseClient{
		getResp:    origTask,
		mutateResp: []byte(`{"id":11,"title":"Original Task (Salinan)"}`),
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks/10/duplicate", nil)
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "GUIDE"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on task duplicate, got %d: %s", rec.Code, rec.Body.String())
	}

	mutateMap, ok := client.lastMutatePayload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload on duplicate task mutate")
	}

	if mutateMap["title"] != "Original Task (Salinan)" {
		t.Errorf("expected duplicated title 'Original Task (Salinan)', got %v", mutateMap["title"])
	}
	if mutateMap["family_id"] != "fam-alpha" {
		t.Errorf("expected family_id 'fam-alpha', got %v", mutateMap["family_id"])
	}
}

func TestAdminDuplicateTask_FamilyMismatchForbidden(t *testing.T) {
	origTask, _ := json.Marshal([]map[string]any{
		{
			"id":        10,
			"family_id": "fam-other",
			"title":     "Other Family Task",
		},
	})

	client := &mockSupabaseClient{
		getResp: origTask,
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks/10/duplicate", nil)
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "GUIDE"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden on duplicating cross-family task, got %d", rec.Code)
	}
}
