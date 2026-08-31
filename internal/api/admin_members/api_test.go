package admin_members

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestHandleCreateMember_ForbiddenForNonAdmin(t *testing.T) {
	mockClient := &mockSupabaseClient{}
	api := NewAPI(mockClient)

	reqBody := map[string]string{
		"username":      "newuser",
		"password":      "secret123",
		"explorer_name": "New Member",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members", bytes.NewReader(bodyBytes))

	// Attach MEMBER claim
	claims := &auth.SessionClaims{
		UID:      "member_1",
		Role:     "MEMBER",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.HandleCreateMember(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for MEMBER role, got %d", w.Code)
	}
}

func TestHandleCreateMember_SuccessAsMemberByDefault(t *testing.T) {
	createdProfile := false
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_local_users" {
				return []byte("[]"), nil
			}
			return []byte("[]"), nil
		},
		mutateAtomicFunc: func(ctx context.Context, method string, table string, payload any, params string, extraHeader string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				createdProfile = true
				payloadMap := payload.(map[string]any)
				if payloadMap["role"] != "MEMBER" {
					t.Errorf("expected new member role to be MEMBER, got %v", payloadMap["role"])
				}
				return json.Marshal([]map[string]any{payloadMap})
			}
			return []byte(`[{"id": 1}]`), nil
		},
	}
	api := NewAPI(mockClient)

	reqBody := map[string]string{
		"username":      "realuser1",
		"password":      "secret123",
		"explorer_name": "Real User One",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members", bytes.NewReader(bodyBytes))

	claims := &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.HandleCreateMember(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}
	if !createdProfile {
		t.Errorf("expected user profile to be created in DB")
	}
}

func TestLegacyRoleNormalization(t *testing.T) {
	if auth.NormalizeRole("GUIDE") != auth.RoleAdmin {
		t.Errorf("expected GUIDE to normalize to ADMIN")
	}
	if auth.NormalizeRole("BUILDER") != auth.RoleAdmin {
		t.Errorf("expected BUILDER to normalize to ADMIN")
	}
	if auth.NormalizeRole("SEEKER") != auth.RoleMember {
		t.Errorf("expected SEEKER to normalize to MEMBER")
	}
	if auth.NormalizeRole("ADMIN") != auth.RoleAdmin {
		t.Errorf("expected ADMIN to normalize to ADMIN")
	}
	if auth.NormalizeRole("MEMBER") != auth.RoleMember {
		t.Errorf("expected MEMBER to normalize to MEMBER")
	}
}

func TestHandleUpdateMember_DeactivateMember(t *testing.T) {
	updatedProfile := false
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				profiles := []db.UserProfile{
					{
						UID:          "usr_target",
						FamilyID:     "fam_1",
						ExplorerName: "Target User",
						Role:         "MEMBER",
						IsActive:     true,
					},
				}
				return json.Marshal(profiles)
			}
			return []byte("[]"), nil
		},
		mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				updatedProfile = true
			}
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
		UID:      "admin_1",
		Role:     "ADMIN",
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
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				profiles := []db.UserProfile{
					{
						UID:          "usr_target",
						FamilyID:     "fam_1",
						ExplorerName: "Target User",
						Role:         "MEMBER",
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
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for reset device, got %d. Body: %s", w.Code, w.Body.String())
	}
}
