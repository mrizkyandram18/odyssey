package family_tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
)

type mockSupabaseClient struct {
	getFunc    func(ctx context.Context, table string, params string) ([]byte, error)
	getResp    []byte
	getErr     error
	mutateResp []byte
	mutateErr  error
	rpcResp    []byte
	rpcErr     error
	uploadResp string
	uploadErr  error
}

func (m *mockSupabaseClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, table, params)
	}
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getResp, nil
}

func (m *mockSupabaseClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
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

func TestHandleGetTodayLinearProgression(t *testing.T) {
	tasks := []TaskRecord{
		{ID: 1, FamilyID: "fam-1", Title: "Step 1 Video", TaskType: "VIDEO", StepOrder: 1, RewardCoins: 50, RewardXP: 100, IsActive: true},
		{ID: 2, FamilyID: "fam-1", Title: "Step 2 Doc", TaskType: "DOCUMENT_UPLOAD", StepOrder: 2, RewardCoins: 75, RewardXP: 150, IsActive: true},
		{ID: 3, FamilyID: "fam-1", Title: "Step 3 Photo", TaskType: "PHOTO_UPLOAD", StepOrder: 3, RewardCoins: 100, RewardXP: 200, IsActive: true},
		{ID: 4, FamilyID: "fam-1", Title: "Step 4 Text", TaskType: "TEXT_RESPONSE", StepOrder: 4, RewardCoins: 60, RewardXP: 120, IsActive: true},
		{ID: 5, FamilyID: "fam-1", Title: "Step 5 Game", TaskType: "MINI_GAME", StepOrder: 5, RewardCoins: 80, RewardXP: 160, IsActive: true},
	}
	tasksBytes, _ := json.Marshal(tasks)

	client := &mockSupabaseClient{
		getResp: tasksBytes,
	}

	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
	claims := &auth.SessionClaims{UID: "user-123", FamilyID: "fam-1", Role: "SEEKER"}
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Tasks []TaskView `json:"tasks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Tasks) != 5 {
		t.Fatalf("expected 5 tasks, got %d", len(resp.Tasks))
	}

	// Step 1 should be UNLOCKED, Step 2..5 should be LOCKED
	if resp.Tasks[0].IsLocked || resp.Tasks[0].Status != "UNLOCKED" {
		t.Errorf("expected step 1 to be UNLOCKED, got %s (is_locked=%v)", resp.Tasks[0].Status, resp.Tasks[0].IsLocked)
	}
	for i := 1; i < 5; i++ {
		if !resp.Tasks[i].IsLocked || resp.Tasks[i].Status != "LOCKED" {
			t.Errorf("expected step %d to be LOCKED, got %s (is_locked=%v)", i+1, resp.Tasks[i].Status, resp.Tasks[i].IsLocked)
		}
	}
}

func TestHandleGetToday_SanitizesQuizAnswerKeys(t *testing.T) {
	tasks := []TaskRecord{
		{
			ID:          1,
			FamilyID:    "fam-1",
			Title:       "Quiz Task",
			TaskType:    "QUIZ",
			StepOrder:   1,
			RewardCoins: 50,
			RewardXP:    100,
			IsActive:    true,
			Config: map[string]any{
				"youtube_url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				"questions": []map[string]any{
					{
						"id":              "q1",
						"question":        "Berapa 1 + 1?",
						"options":         []string{"1", "2", "3", "4"},
						"correct_answer":  "2",
						"expected_answer": "2",
						"is_correct":      true,
						"answer_key":      "2",
					},
				},
			},
		},
	}
	tasksBytes, _ := json.Marshal(tasks)

	client := &mockSupabaseClient{
		getResp: tasksBytes,
	}

	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
	claims := &auth.SessionClaims{UID: "user-123", FamilyID: "fam-1", Role: "SEEKER"}
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rawResponse := rec.Body.String()

	// ASSERT: Answer key fields MUST NOT appear anywhere in the HTTP response body!
	forbiddenKeys := []string{"correct_answer", "expected_answer", "is_correct", "answer_key"}
	for _, fk := range forbiddenKeys {
		if strings.Contains(rawResponse, fk) {
			t.Errorf("SECURITY RISK: leaked quiz answer key field %q in response: %s", fk, rawResponse)
		}
	}
}

func TestHandleGetToday_FamilyIsolation(t *testing.T) {
	tasks := []TaskRecord{
		{ID: 1, FamilyID: "fam-A", Title: "Family A Task", TaskType: "QUIZ", StepOrder: 1, IsActive: true},
		{ID: 2, FamilyID: "fam-B", Title: "Family B Task", TaskType: "QUIZ", StepOrder: 1, IsActive: true},
	}
	tasksBytes, _ := json.Marshal(tasks)

	client := &mockSupabaseClient{
		getResp: tasksBytes,
	}

	api := NewAPI(client)

	// User belongs to Family A
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
	claims := &auth.SessionClaims{UID: "user-A", FamilyID: "fam-A", Role: "SEEKER"}
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	var resp struct {
		Tasks []TaskView `json:"tasks"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Tasks) != 1 {
		t.Fatalf("expected exactly 1 task for Family A, got %d", len(resp.Tasks))
	}
	if resp.Tasks[0].ID != 1 {
		t.Errorf("expected task ID 1 for Family A, got %d", resp.Tasks[0].ID)
	}
}

func TestHandleSubmit_FamilyIsolationForbidden(t *testing.T) {
	// Task belongs to Family B
	task := []TaskRecord{
		{ID: 99, FamilyID: "fam-B", Title: "Family B Task", TaskType: "QUIZ", StepOrder: 1, IsActive: true},
	}
	taskBytes, _ := json.Marshal(task)

	client := &mockSupabaseClient{
		getResp: taskBytes,
	}

	api := NewAPI(client)

	// User belongs to Family A
	body := `{"answers":{"q1":"A"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/99/submit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "user-A", FamilyID: "fam-A", Role: "SEEKER"}
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403 Forbidden for cross-family task submission, got %d", rec.Code)
	}
}

func TestHandleSubmitAutoTask_Success(t *testing.T) {
	task := []TaskRecord{
		{ID: 1, FamilyID: "fam-1", Title: "Step 1", TaskType: "MINI_GAME", StepOrder: 1, IsActive: true},
	}
	taskBytes, _ := json.Marshal(task)

	client := &mockSupabaseClient{
		getResp: taskBytes,
		rpcResp: []byte(`{"success":true,"coins_earned":50,"xp_earned":100,"new_balance":150}`),
	}

	api := NewAPI(client)

	body := `{"answers":{"score":90,"moves":10,"game":"MEMORY"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/1/submit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "user-123", FamilyID: "fam-1", Role: "SEEKER"}
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleSubmitManualTask_TextResponse(t *testing.T) {
	task := []TaskRecord{
		{ID: 2, FamilyID: "fam-1", Title: "Text Task", TaskType: "TEXT_RESPONSE", StepOrder: 2, IsActive: true},
	}
	taskBytes, _ := json.Marshal(task)

	client := &mockSupabaseClient{
		getResp: taskBytes,
		rpcResp: []byte(`{"success":true,"submission_id":102,"status":"PENDING"}`),
	}

	api := NewAPI(client)

	body := `{"payload":{"text":"Hari ini saya belajar tentang pentingnya menabung untuk masa depan."}}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/2/submit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "user-123", FamilyID: "fam-1", Role: "SEEKER"}
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleSubmitAutoTask_AntiDoubleClaimRejection(t *testing.T) {
	task := []TaskRecord{
		{ID: 1, FamilyID: "fam-1", Title: "Step 1", TaskType: "QUIZ", StepOrder: 1, IsActive: true},
	}
	taskBytes, _ := json.Marshal(task)

	// Simulate RPC error on duplicate submission
	client := &mockSupabaseClient{
		getResp: taskBytes,
		rpcErr:  errors.New("Tugas ini sudah diselesaikan dan reward sudah diterima"),
	}

	api := NewAPI(client)

	body := `{"answers":{"1":"A"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/1/submit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "user-123", FamilyID: "fam-1", Role: "SEEKER"}
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request on duplicate reward claim, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleUploadProof_DisallowsExecutableExtensions(t *testing.T) {
	client := &mockSupabaseClient{}
	api := NewAPI(client)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "malicious_script.exe")
	fw.Write([]byte("fake-binary-content"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	claims := &auth.SessionClaims{UID: "user-123", FamilyID: "fam-1", Role: "SEEKER"}
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request on .exe upload, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "tidak diizinkan") {
		t.Errorf("expected error message to mention rejected extension, got: %s", rec.Body.String())
	}
}

func TestHandleGetTask_SuccessAndSanitization(t *testing.T) {
	task := TaskRecord{
		ID:          42,
		FamilyID:    "fam-1",
		Title:       "Test Quiz Task",
		TaskType:    "QUIZ",
		StepOrder:   1,
		RewardCoins: 50,
		RewardXP:    100,
		IsActive:    true,
		Config: map[string]any{
			"questions": []any{
				map[string]any{
					"id":             "1",
					"question":       "What is the capital?",
					"options":        []any{"A", "B", "C"},
					"correct_answer": "SECRET_A",
				},
			},
		},
	}
	taskBytes, _ := json.Marshal([]TaskRecord{task})

	client := &mockSupabaseClient{
		getResp: taskBytes,
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/42", nil)
	claims := &auth.SessionClaims{UID: "user-123", FamilyID: "fam-1", Role: "SEEKER"}
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	bodyStr := rec.Body.String()
	if strings.Contains(bodyStr, "SECRET_A") || strings.Contains(bodyStr, "correct_answer") {
		t.Fatalf("CRITICAL SECURITY LEAK: Task response exposed secret answer key: %s", bodyStr)
	}

	var view TaskView
	if err := json.Unmarshal(rec.Body.Bytes(), &view); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if view.ID != 42 || view.Title != "Test Quiz Task" {
		t.Errorf("unexpected task view: %+v", view)
	}
}

func TestHandleGetTask_CrossFamilyAccessForbidden(t *testing.T) {
	task := TaskRecord{
		ID:       42,
		FamilyID: "fam-2", // Belongs to Family 2
		Title:    "Other Family Task",
		TaskType: "PHOTO_UPLOAD",
		IsActive: true,
	}
	taskBytes, _ := json.Marshal([]TaskRecord{task})

	client := &mockSupabaseClient{
		getResp: taskBytes,
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/42", nil)
	claims := &auth.SessionClaims{UID: "user-123", FamilyID: "fam-1", Role: "SEEKER"} // User in Family 1
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden on cross-family task request, got %d", rec.Code)
	}
}

func TestHandleGetTask_NotFound(t *testing.T) {
	client := &mockSupabaseClient{
		getResp: []byte(`[]`),
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/999", nil)
	claims := &auth.SessionClaims{UID: "user-123", FamilyID: "fam-1", Role: "SEEKER"}
	ctx := auth.ContextWithClaims(req.Context(), claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found on missing task, got %d", rec.Code)
	}
}

func TestHandleGetToday_PersonalTaskTargeting(t *testing.T) {
	tasks := []TaskRecord{
		{ID: 1, FamilyID: "fam-1", Title: "Shared Task", TaskType: "VIDEO", StepOrder: 1, TargetScope: "ALL", IsActive: true},
		{ID: 2, FamilyID: "fam-1", Title: "User A Personal Task", TaskType: "QUIZ", StepOrder: 2, TargetScope: "USER", TargetUserUID: "user-A", IsActive: true},
		{ID: 3, FamilyID: "fam-1", Title: "User B Personal Task", TaskType: "PHOTO_UPLOAD", StepOrder: 3, TargetScope: "USER", TargetUserUID: "user-B", IsActive: true},
	}
	tasksBytes, _ := json.Marshal(tasks)

	client := &mockSupabaseClient{
		getResp: tasksBytes,
	}
	api := NewAPI(client)

	// User A request
	reqA := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
	claimsA := &auth.SessionClaims{UID: "user-A", FamilyID: "fam-1", Role: "SEEKER"}
	reqA = reqA.WithContext(auth.ContextWithClaims(reqA.Context(), claimsA))
	recA := httptest.NewRecorder()
	api.Handler(recA, reqA)

	var respA struct {
		Tasks []TaskView `json:"tasks"`
	}
	_ = json.Unmarshal(recA.Body.Bytes(), &respA)

	if len(respA.Tasks) != 2 {
		t.Fatalf("expected User A to see 2 tasks (Shared & User A), got %d", len(respA.Tasks))
	}
	if respA.Tasks[0].ID != 1 || respA.Tasks[1].ID != 2 {
		t.Errorf("unexpected tasks for User A: %+v", respA.Tasks)
	}

	// User B request
	reqB := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
	claimsB := &auth.SessionClaims{UID: "user-B", FamilyID: "fam-1", Role: "SEEKER"}
	reqB = reqB.WithContext(auth.ContextWithClaims(reqB.Context(), claimsB))
	recB := httptest.NewRecorder()
	api.Handler(recB, reqB)

	var respB struct {
		Tasks []TaskView `json:"tasks"`
	}
	_ = json.Unmarshal(recB.Body.Bytes(), &respB)

	if len(respB.Tasks) != 2 {
		t.Fatalf("expected User B to see 2 tasks (Shared & User B), got %d", len(respB.Tasks))
	}
	if respB.Tasks[0].ID != 1 || respB.Tasks[1].ID != 3 {
		t.Errorf("unexpected tasks for User B: %+v", respB.Tasks)
	}
}

func TestHandleSubmit_PersonalTaskForbidden(t *testing.T) {
	// Task 2 belongs to user-A
	task := []TaskRecord{
		{ID: 2, FamilyID: "fam-1", Title: "User A Personal Task", TaskType: "QUIZ", StepOrder: 1, TargetScope: "USER", TargetUserUID: "user-A", IsActive: true},
	}
	taskBytes, _ := json.Marshal(task)

	client := &mockSupabaseClient{
		getResp: taskBytes,
	}
	api := NewAPI(client)

	// User B tries to submit User A's personal task
	body := `{"answers":{"q1":"A"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/2/submit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "user-B", FamilyID: "fam-1", Role: "SEEKER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when user-B submits user-A's personal task, got %d", rec.Code)
	}
}

func TestHandleGetToday_NewUserJoinDateBacklogGuard(t *testing.T) {
	tasks := []TaskRecord{
		{ID: 10, FamilyID: "fam-1", Title: "Past Task", TaskType: "VIDEO", StepOrder: 1, ActiveDate: "2026-08-01", TargetScope: "ALL", IsActive: true},
		{ID: 11, FamilyID: "fam-1", Title: "Join Date Task", TaskType: "VIDEO", StepOrder: 2, ActiveDate: "2026-08-15", TargetScope: "ALL", IsActive: true},
	}
	tasksBytes, _ := json.Marshal(tasks)
	profileBytes := []byte(`[{"uid":"user-new","family_id":"fam-1","explorer_name":"New User","role":"SEEKER","is_active":true,"created_at":"2026-08-15T10:00:00Z"}]`)

	client := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				return profileBytes, nil
			}
			return tasksBytes, nil
		},
	}
	api := NewAPI(client)

	// User created on 2026-08-15
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
	claims := &auth.SessionClaims{UID: "user-new", FamilyID: "fam-1", Role: "SEEKER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	var resp struct {
		Tasks []TaskView `json:"tasks"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	// User created on 2026-08-15 will see tasks starting from 2026-08-15 (Task 11), NOT historical Task 10 (2026-08-01)
	for _, task := range resp.Tasks {
		if task.ID == 10 {
			t.Errorf("backlog guard failure: new user received historical task prior to join date")
		}
	}
}
