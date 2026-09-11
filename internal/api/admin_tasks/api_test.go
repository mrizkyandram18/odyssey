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
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
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
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "ADMIN"}
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
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "ADMIN"}
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

func TestAdminCreateTask_VideoEssayEvaluationType(t *testing.T) {
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "ADMIN"}

	newCreateReq := func(client *mockSupabaseClient, payload string) *httptest.ResponseRecorder {
		api := NewAPI(client)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))
		rec := httptest.NewRecorder()
		api.Handler(rec, req)
		return rec
	}

	// AC2: VIDEO + essay prompt must default to ADMIN_REVIEW
	t.Run("VIDEO essay defaults to ADMIN_REVIEW", func(t *testing.T) {
		client := &mockSupabaseClient{mutateResp: []byte(`{"id":3,"title":"Video Essay"}`)}
		payload := `{"title":"Video Essay","task_type":"VIDEO","step_order":3,` +
			`"config":{"video_url":"https://www.youtube.com/watch?v=x","youtube_url":"https://www.youtube.com/watch?v=x",` +
			`"prompt":"Jelaskan apa yang kamu pahami.","minimum_characters":10,"maximum_characters":500}}`
		rec := newCreateReq(client, payload)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}
		mutateMap, ok := client.lastMutatePayload.(map[string]any)
		if !ok {
			t.Fatalf("expected map payload on task create")
		}
		if mutateMap["evaluation_type"] != "ADMIN_REVIEW" {
			t.Errorf("expected evaluation_type ADMIN_REVIEW for VIDEO essay, got %v", mutateMap["evaluation_type"])
		}
	})

	// AC1: VIDEO watch-only keeps existing AUTO behavior
	t.Run("VIDEO watch-only stays AUTO", func(t *testing.T) {
		client := &mockSupabaseClient{mutateResp: []byte(`{"id":4,"title":"Video Nonton"}`)}
		payload := `{"title":"Video Nonton","task_type":"VIDEO","step_order":4,` +
			`"config":{"video_url":"https://www.youtube.com/watch?v=x","youtube_url":"https://www.youtube.com/watch?v=x"}}`
		rec := newCreateReq(client, payload)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}
		mutateMap := client.lastMutatePayload.(map[string]any)
		if mutateMap["evaluation_type"] != "AUTO" {
			t.Errorf("expected evaluation_type AUTO for VIDEO watch-only, got %v", mutateMap["evaluation_type"])
		}
	})
}

func TestAdminUpdateTask_VideoEssaySyncsEvaluationType(t *testing.T) {
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}

	newUpdateReq := func(client *mockSupabaseClient, patch string) *httptest.ResponseRecorder {
		api := NewAPI(client)
		req := httptest.NewRequest(http.MethodPatch, "/api/admin/tasks/9", strings.NewReader(patch))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))
		rec := httptest.NewRecorder()
		api.Handler(rec, req)
		return rec
	}

	existingVideo := func(eval string) []byte {
		b, _ := json.Marshal([]map[string]any{
			{"id": 9, "family_id": "fam-1", "task_type": "VIDEO", "evaluation_type": eval,
				"reward_coins": float64(50), "reward_xp": float64(100),
				"config": map[string]any{"video_url": "https://www.youtube.com/watch?v=x"}},
		})
		return b
	}
	essayConfig := `{"config":{"video_url":"https://www.youtube.com/watch?v=x",` +
		`"prompt":"Jelaskan apa yang kamu pahami.","minimum_characters":10,"maximum_characters":500}}`

	// AC2/backward compat: adding an essay prompt flips stored AUTO to ADMIN_REVIEW
	t.Run("adding essay prompt sets ADMIN_REVIEW", func(t *testing.T) {
		client := &mockSupabaseClient{getResp: existingVideo("AUTO"), mutateResp: []byte(`{"status":"updated","id":9}`)}
		rec := newUpdateReq(client, essayConfig)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		mutateMap, ok := client.lastMutatePayload.(map[string]any)
		if !ok {
			t.Fatalf("expected map payload on task update")
		}
		if mutateMap["evaluation_type"] != "ADMIN_REVIEW" {
			t.Errorf("expected evaluation_type ADMIN_REVIEW after adding prompt, got %v", mutateMap["evaluation_type"])
		}
	})

	// AC1: removing the prompt flips back to AUTO
	t.Run("removing essay prompt restores AUTO", func(t *testing.T) {
		b, _ := json.Marshal([]map[string]any{
			{"id": 9, "family_id": "fam-1", "task_type": "VIDEO", "evaluation_type": "ADMIN_REVIEW",
				"reward_coins": float64(50), "reward_xp": float64(100),
				"config": map[string]any{"video_url": "https://www.youtube.com/watch?v=x", "prompt": "Old"}},
		})
		client := &mockSupabaseClient{getResp: b, mutateResp: []byte(`{"status":"updated","id":9}`)}
		rec := newUpdateReq(client, `{"config":{"video_url":"https://www.youtube.com/watch?v=x"}}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		mutateMap := client.lastMutatePayload.(map[string]any)
		if mutateMap["evaluation_type"] != "AUTO" {
			t.Errorf("expected evaluation_type AUTO after removing prompt, got %v", mutateMap["evaluation_type"])
		}
	})

	// Non-config edits must not touch evaluation_type
	t.Run("title-only edit preserves evaluation_type", func(t *testing.T) {
		client := &mockSupabaseClient{getResp: existingVideo("AUTO"), mutateResp: []byte(`{"status":"updated","id":9}`)}
		rec := newUpdateReq(client, `{"title":"Judul Baru"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		mutateMap := client.lastMutatePayload.(map[string]any)
		if _, present := mutateMap["evaluation_type"]; present {
			t.Errorf("expected evaluation_type untouched on title-only edit, got %v", mutateMap["evaluation_type"])
		}
	})
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
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "ADMIN"}
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
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "ADMIN"}

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
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
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
	guideClaims := &auth.SessionClaims{UID: "admin-1", Role: "ADMIN"}

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

func TestAdminConfig_MonthlyDefaults_MissingReturnsZero(t *testing.T) {
	// Missing default_monthly_coin_target resolves to 0.
	// Missing default_monthly_earning_cap resolves to -1 (unconfigured error state; never 0=unlimited).
	client := &mockSupabaseClient{
		getResp: []byte(`[]`),
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
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
	if v, ok := res["default_monthly_coin_target"]; !ok || v != float64(0) {
		t.Fatalf("expected default_monthly_coin_target 0 for missing row, got %v", res)
	}
	if v, ok := res["default_monthly_earning_cap"]; !ok || v != float64(-1) {
		t.Fatalf("expected default_monthly_earning_cap -1 (unconfigured), got %v", res)
	}
}

func TestAdminConfig_UpdateMonthlyDefaults_Validation(t *testing.T) {
	client := &mockSupabaseClient{
		getResp: []byte(`[{"key":"max_monthly_earning_cap_ceiling","value":"10000"},{"key":"default_monthly_earning_cap","value":"3320"}]`),
	}
	api := NewAPI(client)
	guideClaims := &auth.SessionClaims{UID: "admin-1", Role: "ADMIN"}

	cases := []struct {
		name       string
		payload    string
		expectCode int
	}{
		{
			name:       "Reject negative monthly coin target",
			payload:    `{"default_monthly_coin_target":-1}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Reject monthly coin target above 10000",
			payload:    `{"default_monthly_coin_target":10001}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Reject negative monthly earning cap",
			payload:    `{"default_monthly_earning_cap":-1}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Reject monthly earning cap above dynamic ceiling 10000",
			payload:    `{"default_monthly_earning_cap":10001}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Reject negative max monthly earning cap ceiling",
			payload:    `{"max_monthly_earning_cap_ceiling":-1}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Reject invalid level_cap_bonus key",
			payload:    `{"level_cap_bonus":{"non_int":500}}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Reject negative level_cap_bonus value",
			payload:    `{"level_cap_bonus":{"5":-100}}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "Accept explicit zero for both member defaults",
			payload:    `{"default_monthly_coin_target":0,"default_monthly_earning_cap":0}`,
			expectCode: http.StatusOK,
		},
		{
			name:       "Accept valid dynamic ceiling and level_cap_bonus",
			payload:    `{"max_monthly_earning_cap_ceiling":8000,"default_monthly_earning_cap":3500,"level_cap_bonus":{"5":300,"10":500}}`,
			expectCode: http.StatusOK,
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

func TestAdminCosmetics_CRUD_And_Auth(t *testing.T) {
	client := &mockSupabaseClient{
		getResp: []byte(`[{"id":"frame-gold","slot":"frame","asset":"gold","tier":1,"is_active":true}]`),
	}
	api := NewAPI(client)
	userClaims := &auth.SessionClaims{UID: "user-1", Role: "USER"}
	adminClaims := &auth.SessionClaims{UID: "admin-1", Role: "ADMIN"}

	// 1. Non-admin forbidden
	req := httptest.NewRequest(http.MethodGet, "/api/admin/cosmetics", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), userClaims))
	rec := httptest.NewRecorder()
	api.Handler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-admin, got %d", rec.Code)
	}

	// 2. Admin list cosmetics
	req = httptest.NewRequest(http.MethodGet, "/api/admin/cosmetics", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims))
	rec = httptest.NewRecorder()
	api.Handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	// 3. Create cosmetic invalid slot
	badReq := httptest.NewRequest(http.MethodPost, "/api/admin/cosmetics", strings.NewReader(`{"id":"c1","slot":"invalid","asset":"test"}`))
	badReq = badReq.WithContext(auth.ContextWithClaims(badReq.Context(), adminClaims))
	rec = httptest.NewRecorder()
	api.Handler(rec, badReq)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}

	// 4. Create cosmetic valid
	goodReq := httptest.NewRequest(http.MethodPost, "/api/admin/cosmetics", strings.NewReader(`{"id":"c1","name":"Golden Crown","slot":"frame","asset":"gold","tier":1,"is_active":true}`))
	goodReq = goodReq.WithContext(auth.ContextWithClaims(goodReq.Context(), adminClaims))
	rec = httptest.NewRecorder()
	api.Handler(rec, goodReq)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	// 5. Update cosmetic valid
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/admin/cosmetics/c1", strings.NewReader(`{"is_active":false}`))
	patchReq = patchReq.WithContext(auth.ContextWithClaims(patchReq.Context(), adminClaims))
	rec = httptest.NewRecorder()
	api.Handler(rec, patchReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
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
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "ADMIN"}
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
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-alpha", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden on duplicating cross-family task, got %d", rec.Code)
	}
}

func TestAdminVerifySubmission_RejectWithPenalty(t *testing.T) {
	client := &mockSupabaseClient{
		rpcResp: []byte(`{"success":true,"status":"REJECTED","coins_deducted":20,"new_balance":80}`),
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/submissions/1/verify", strings.NewReader(`{"status":"REJECTED","notes":"Salah jawaban","penalty_coins":20}`))
	req.Header.Set("Content-Type", "application/json")
	adminClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminVerifySubmission_ValidationErrors(t *testing.T) {
	client := &mockSupabaseClient{}
	api := NewAPI(client)
	adminClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}

	// 1. Negative penalty
	req1 := httptest.NewRequest(http.MethodPost, "/api/admin/submissions/1/verify", strings.NewReader(`{"status":"REJECTED","penalty_coins":-5}`))
	req1.Header.Set("Content-Type", "application/json")
	req1 = req1.WithContext(auth.ContextWithClaims(req1.Context(), adminClaims))
	rec1 := httptest.NewRecorder()
	api.Handler(rec1, req1)
	if rec1.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for negative penalty, got %d", rec1.Code)
	}

	// 2. Penalty on APPROVED
	req2 := httptest.NewRequest(http.MethodPost, "/api/admin/submissions/1/verify", strings.NewReader(`{"status":"APPROVED","penalty_coins":10}`))
	req2.Header.Set("Content-Type", "application/json")
	req2 = req2.WithContext(auth.ContextWithClaims(req2.Context(), adminClaims))
	rec2 := httptest.NewRecorder()
	api.Handler(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for penalty on approved, got %d", rec2.Code)
	}
}

func TestAdminEditSubmission_Success(t *testing.T) {
	client := &mockSupabaseClient{
		rpcResp: []byte(`{"success":true,"submission_id":1,"status":"PENDING","payload":{"text":"Jawaban yang sudah diperbaiki"}}`),
	}
	api := NewAPI(client)
	adminClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/submissions/1", strings.NewReader(`{"payload":{"text":"Jawaban yang sudah diperbaiki"},"notes":"Koreksi ejaan oleh admin"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminEditSubmission_Validation(t *testing.T) {
	client := &mockSupabaseClient{}
	api := NewAPI(client)
	adminClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}

	// 1. Missing payload
	req1 := httptest.NewRequest(http.MethodPatch, "/api/admin/submissions/1", strings.NewReader(`{"notes":"no payload"}`))
	req1.Header.Set("Content-Type", "application/json")
	req1 = req1.WithContext(auth.ContextWithClaims(req1.Context(), adminClaims))
	rec1 := httptest.NewRecorder()
	api.Handler(rec1, req1)
	if rec1.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing payload, got %d", rec1.Code)
	}

	// 2. Score out of range
	req2 := httptest.NewRequest(http.MethodPatch, "/api/admin/submissions/1", strings.NewReader(`{"payload":{"score":-10}}`))
	req2.Header.Set("Content-Type", "application/json")
	req2 = req2.WithContext(auth.ContextWithClaims(req2.Context(), adminClaims))
	rec2 := httptest.NewRecorder()
	api.Handler(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative score, got %d", rec2.Code)
	}

	// 3. Non-admin forbidden
	seekerClaims := &auth.SessionClaims{UID: "seeker-1", FamilyID: "fam-1", Role: "MEMBER"}
	req3 := httptest.NewRequest(http.MethodPatch, "/api/admin/submissions/1", strings.NewReader(`{"payload":{"text":"test"}}`))
	req3.Header.Set("Content-Type", "application/json")
	req3 = req3.WithContext(auth.ContextWithClaims(req3.Context(), seekerClaims))
	rec3 := httptest.NewRecorder()
	api.Handler(rec3, req3)
	if rec3.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-admin edit, got %d", rec3.Code)
	}
}

type mockDynamicClient struct {
	getFn func(ctx context.Context, table string, params string) ([]byte, error)
}

func (m *mockDynamicClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if m.getFn != nil {
		return m.getFn(ctx, table, params)
	}
	return []byte("[]"), nil
}
func (m *mockDynamicClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *mockDynamicClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *mockDynamicClient) RPC(ctx context.Context, fnName string, payload any) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *mockDynamicClient) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return "", nil
}

func TestHandleListPendingSubmissions_TenantIsolationAndPagination(t *testing.T) {
	var requestedProfileParams string
	var requestedSubmissionParams string
	var requestedTaskParams string

	client := &mockDynamicClient{
		getFn: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				requestedProfileParams = params
				profiles := []map[string]any{
					{"uid": "usr_alpha_1", "explorer_name": "Alpha One"},
					{"uid": "usr_alpha_2", "explorer_name": "Alpha Two"},
				}
				return json.Marshal(profiles)
			}
			if table == "odyssey_task_submissions" {
				requestedSubmissionParams = params
				subs := []map[string]any{
					{
						"id":              int64(101),
						"task_id":         int64(55),
						"user_uid":        "usr_alpha_1",
						"submission_type": "MANUAL_VERIFY",
						"status":          "PENDING",
						"created_at":      "2026-09-01T12:00:00Z",
					},
				}
				return json.Marshal(subs)
			}
			if table == "odyssey_tasks" {
				requestedTaskParams = params
				tasks := []map[string]any{
					{
						"id":           int64(55),
						"title":        "Target Task 55",
						"task_type":    "PHOTO_UPLOAD",
						"reward_coins": 50,
						"reward_xp":    100,
					},
				}
				return json.Marshal(tasks)
			}
			return []byte("[]"), nil
		},
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/submissions?status=PENDING&page=1&limit=10", nil)
	adminClaims := &auth.SessionClaims{
		UID:      "admin-1",
		FamilyID: "fam-alpha",
		Role:     "ADMIN",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(requestedProfileParams, "family_id=eq.fam-alpha") {
		t.Errorf("expected profile query scoped to fam-alpha, got %s", requestedProfileParams)
	}
	if !strings.Contains(requestedSubmissionParams, "user_uid=in.(usr_alpha_1,usr_alpha_2)") {
		t.Errorf("expected submissions scoped to family members, got %s", requestedSubmissionParams)
	}
	if !strings.Contains(requestedSubmissionParams, "limit=10&offset=0") {
		t.Errorf("expected pagination limit=10 offset=0, got %s", requestedSubmissionParams)
	}
	if !strings.Contains(requestedTaskParams, "id=in.(55)") {
		t.Errorf("expected tasks scoped strictly to id=in.(55), got %s", requestedTaskParams)
	}

	var res struct {
		Items []PendingSubmissionView `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response envelope: %v", err)
	}
	if len(res.Items) != 1 || res.Items[0].TaskTitle != "Target Task 55" {
		t.Errorf("expected enriched item with task title, got %+v", res.Items)
	}
}

func TestHandleUpdateAdminConfig_AnnouncementValidation(t *testing.T) {
	adminClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}

	// Enabling with blank title+body must be rejected (no hardcoded content).
	client := &mockSupabaseClient{getResp: []byte(`[]`), mutateResp: []byte(`[]`)}
	api := NewAPI(client)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/config",
		strings.NewReader(`{"announcement_enabled":true,"announcement_title":"","announcement_body":""}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims))
	rec := httptest.NewRecorder()
	api.Handler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when enabling announcement with blank content, got %d: %s", rec.Code, rec.Body.String())
	}

	// Invalid audience must be rejected.
	req2 := httptest.NewRequest(http.MethodPost, "/api/admin/config",
		strings.NewReader(`{"announcement_audience":"EVERYONE"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2 = req2.WithContext(auth.ContextWithClaims(req2.Context(), adminClaims))
	rec2 := httptest.NewRecorder()
	api.Handler(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid audience, got %d: %s", rec2.Code, rec2.Body.String())
	}
}

func TestHandleGetAdminConfig_IncludesAnnouncement(t *testing.T) {
	adminClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
	rows := `[
		{"key":"announcement_enabled","value":"true"},
		{"key":"announcement_title","value":"Judul Uji"},
		{"key":"announcement_body","value":"Isi uji dari konfigurasi."},
		{"key":"announcement_audience","value":"ALL"},
		{"key":"announcement_priority","value":"normal"}
	]`
	client := &mockSupabaseClient{getResp: []byte(rows), mutateResp: []byte(`[]`)}
	api := NewAPI(client)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims))
	rec := httptest.NewRecorder()
	api.Handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res struct {
		Announcement *struct {
			Enabled bool   `json:"enabled"`
			Title   string `json:"title"`
			Body    string `json:"body"`
			Visible bool   `json:"visible"`
		} `json:"announcement"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if res.Announcement == nil || !res.Announcement.Enabled || !res.Announcement.Visible {
		t.Fatalf("expected visible enabled announcement from config, got %+v", res.Announcement)
	}
	if res.Announcement.Title != "Judul Uji" {
		t.Fatalf("expected title from DB config, got %q", res.Announcement.Title)
	}
}
