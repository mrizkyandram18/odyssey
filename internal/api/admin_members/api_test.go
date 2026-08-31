package admin_members

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
)

type mockSupabaseClient struct {
	getFunc          func(ctx context.Context, table string, params string) ([]byte, error)
	mutateAtomicFunc func(ctx context.Context, method string, table string, payload any, params string, extraHeader string) ([]byte, error)
	mutateFunc       func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error)
}

func (m *mockSupabaseClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, table, params)
	}
	return []byte("[]"), nil
}

func (m *mockSupabaseClient) MutateAtomic(ctx context.Context, method string, table string, payload any, params string, extraHeader string) ([]byte, error) {
	if m.mutateAtomicFunc != nil {
		return m.mutateAtomicFunc(ctx, method, table, payload, params, extraHeader)
	}
	return []byte("{}"), nil
}

func (m *mockSupabaseClient) Mutate(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
	if m.mutateFunc != nil {
		return m.mutateFunc(ctx, method, table, payload, params)
	}
	return []byte("{}"), nil
}

func (m *mockSupabaseClient) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	return []byte("{}"), nil
}

func (m *mockSupabaseClient) UploadStorage(ctx context.Context, bucket string, storagePath string, contentType string, fileBytes []byte) (string, error) {
	return "http://localhost/storage/" + storagePath, nil
}

func TestHandleCreateMember_ForbiddenForNonGuide(t *testing.T) {
	mockClient := &mockSupabaseClient{}
	api := NewAPI(mockClient)

	reqBody := map[string]string{
		"username":      "newuser",
		"password":      "secret123",
		"explorer_name": "New Explorer",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members", bytes.NewReader(bodyBytes))

	// Attach non-GUIDE claim (SEEKER)
	claims := &auth.SessionClaims{
		UID:      "seeker_1",
		Role:     "SEEKER",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.HandleCreateMember(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-guide role, got %d", w.Code)
	}
}

func TestHandleCreateMember_Success(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_local_users" {
				return []byte("[]"), nil // Username does not exist
			}
			return []byte("[]"), nil
		},
		mutateAtomicFunc: func(ctx context.Context, method string, table string, payload any, params string, extraHeader string) ([]byte, error) {
			return []byte(`{"status":"created"}`), nil
		},
	}
	api := NewAPI(mockClient)

	reqBody := map[string]string{
		"username":      "budi_new",
		"password":      "password123",
		"explorer_name": "Budi New User",
		"role":          "SEEKER",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members", bytes.NewReader(bodyBytes))

	claims := &auth.SessionClaims{
		UID:      "guide_1",
		Role:     "GUIDE",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.HandleCreateMember(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var res MemberView
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse created member response: %v", err)
	}

	if res.Username != "budi_new" {
		t.Errorf("expected username budi_new, got %s", res.Username)
	}
	if res.ExplorerName != "Budi New User" {
		t.Errorf("expected explorer name Budi New User, got %s", res.ExplorerName)
	}
	if res.FamilyID != "fam_1" {
		t.Errorf("expected family_id fam_1, got %s", res.FamilyID)
	}
	if !res.IsActive {
		t.Errorf("expected new member to be active by default")
	}
}

func TestHandleCreateMember_DuplicateUsername(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_local_users" {
				// Return existing user with same username
				return []byte(`[{"username":"budi_existing","profile_uid":"usr_123"}]`), nil
			}
			return []byte("[]"), nil
		},
	}
	api := NewAPI(mockClient)

	reqBody := map[string]string{
		"username":      "budi_existing",
		"password":      "password123",
		"explorer_name": "Budi Duplicate",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members", bytes.NewReader(bodyBytes))

	claims := &auth.SessionClaims{
		UID:      "guide_1",
		Role:     "GUIDE",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.HandleCreateMember(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for duplicate username, got %d", w.Code)
	}
}

func TestHandleListMembers_Success(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				profiles := []db.UserProfile{
					{
						UID:          "usr_1",
						FamilyID:     "fam_1",
						ExplorerName: "User One",
						Role:         "SEEKER",
						IsActive:     true,
						Level:        1,
						XP:           100,
						Coins:        50,
						CreatedAt:    time.Now().UTC(),
					},
				}
				return json.Marshal(profiles)
			}
			if table == "odyssey_local_users" {
				return []byte(`[{"username":"userone","profile_uid":"usr_1"}]`), nil
			}
			return []byte("[]"), nil
		},
	}
	api := NewAPI(mockClient)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/members", nil)
	claims := &auth.SessionClaims{
		UID:      "guide_1",
		Role:     "GUIDE",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.HandleListMembers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var members []MemberView
	if err := json.Unmarshal(w.Body.Bytes(), &members); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}

	if len(members) != 1 || members[0].Username != "userone" {
		t.Errorf("unexpected list response: %+v", members)
	}
}

func TestHandleUpdateMember_Deactivate(t *testing.T) {
	updatedProfile := false
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				profiles := []db.UserProfile{
					{
						UID:          "usr_target",
						FamilyID:     "fam_1",
						ExplorerName: "Target User",
						Role:         "SEEKER",
						IsActive:     !updatedProfile,
					},
				}
				return json.Marshal(profiles)
			}
			if table == "odyssey_local_users" {
				return []byte(`[{"username":"targetuser"}]`), nil
			}
			return []byte("[]"), nil
		},
		mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
			updatedProfile = true
			return []byte(`{"status":"updated"}`), nil
		},
	}
	api := NewAPI(mockClient)

	patchBody := map[string]any{
		"is_active": false,
	}
	bodyBytes, _ := json.Marshal(patchBody)
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(bodyBytes))

	claims := &auth.SessionClaims{
		UID:      "guide_1",
		Role:     "GUIDE",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}
	if !updatedProfile {
		t.Errorf("expected mutateFunc to be called for profile deactivation")
	}
}

func TestHandleUpdateMember_ResetDevice(t *testing.T) {
	rpcCalled := false
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				profiles := []db.UserProfile{
					{
						UID:          "usr_target",
						FamilyID:     "fam_1",
						ExplorerName: "Target User",
						Role:         "SEEKER",
						IsActive:     true,
					},
				}
				return json.Marshal(profiles)
			}
			return []byte("[]"), nil
		},
		mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
			return []byte(`{"status":"updated"}`), nil
		},
	}
	mockClient.mutateAtomicFunc = func(ctx context.Context, method, table string, payload any, params, extraHeader string) ([]byte, error) {
		return []byte(`{"status":"ok"}`), nil
	}
	api := NewAPI(mockClient)

	patchBody := map[string]any{
		"reset_device": true,
	}
	bodyBytes, _ := json.Marshal(patchBody)
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(bodyBytes))

	claims := &auth.SessionClaims{
		UID:      "guide_1",
		Role:     "GUIDE",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for reset device, got %d. Body: %s", w.Code, w.Body.String())
	}
	_ = rpcCalled
}
