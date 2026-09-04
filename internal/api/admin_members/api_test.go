package admin_members

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/shared"
)

type mockSupabaseClient struct {
	getFunc          func(ctx context.Context, table string, params string) ([]byte, error)
	mutateAtomicFunc func(ctx context.Context, method string, table string, payload any, params string, extraHeader string) ([]byte, error)
	mutateFunc       func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error)
	rpcFunc          func(ctx context.Context, fnName string, payload any) ([]byte, error)
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
	if m.rpcFunc != nil {
		return m.rpcFunc(ctx, fn, payload)
	}
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

func TestHandleListMembers_TenantIsolation(t *testing.T) {
	var requestedProfileParams string
	var requestedLocalParams string

	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				requestedProfileParams = params
				profiles := []map[string]any{
					{
						"uid":           "usr_fam1_a",
						"family_id":     "fam_1",
						"explorer_name": "Family 1 Member",
						"role":          "MEMBER",
						"is_active":     true,
					},
				}
				return json.Marshal(profiles)
			}
			if table == "odyssey_local_users" {
				requestedLocalParams = params
				return json.Marshal([]map[string]any{
					{"username": "fam1_user", "profile_uid": "usr_fam1_a"},
				})
			}
			return []byte("[]"), nil
		},
	}
	api := NewAPI(mockClient)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/members?page=1&limit=50", nil)
	claims := &auth.SessionClaims{
		UID:      "admin_fam1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(requestedProfileParams, "family_id=eq.fam_1") {
		t.Errorf("expected profile query to be strictly scoped to fam_1, got %s", requestedProfileParams)
	}
	if !strings.Contains(requestedLocalParams, "profile_uid=in.(usr_fam1_a)") {
		t.Errorf("expected local_users query to be strictly scoped to profile_uid in (usr_fam1_a), got %s", requestedLocalParams)
	}

	var res struct {
		Items []MemberView `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response envelope: %v", err)
	}
	if len(res.Items) != 1 || res.Items[0].UID != "usr_fam1_a" {
		t.Errorf("expected 1 member usr_fam1_a, got %v", res.Items)
	}
}

func TestHandleListMembers_Pagination(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				// The paginated page query carries order=+limit/offset; the
				// lightweight pre-query (uid/is_active) does not.
				if strings.Contains(params, "order=created_at.desc") && !strings.Contains(params, "limit=2&offset=2") {
					t.Errorf("expected limit=2&offset=2 for page 2, got %s", params)
				}
				profiles := []map[string]any{
					{"uid": "usr_3", "family_id": "fam_1", "explorer_name": "Member 3"},
					{"uid": "usr_4", "family_id": "fam_1", "explorer_name": "Member 4"},
				}
				return json.Marshal(profiles)
			}
			return json.Marshal([]map[string]any{
				{"username": "m3", "profile_uid": "usr_3"},
				{"username": "m4", "profile_uid": "usr_4"},
			})
		},
	}
	api := NewAPI(mockClient)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/members?page=2&limit=2", nil)
	claims := &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	api.Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		Items      []MemberView          `json:"items"`
		Pagination shared.PaginationMeta `json:"pagination"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response envelope: %v", err)
	}

	if res.Pagination.Page != 2 || res.Pagination.Limit != 2 {
		t.Errorf("expected page 2 limit 2, got %+v", res.Pagination)
	}
	if len(res.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(res.Items))
	}
}

// --- Admin Reset Password Tests ---

func adminClaims(familyID string) *auth.SessionClaims {
	return &auth.SessionClaims{UID: "admin_1", Role: "ADMIN", FamilyID: familyID}
}

func memberClaims(familyID string) *auth.SessionClaims {
	return &auth.SessionClaims{UID: "member_1", Role: "MEMBER", FamilyID: familyID}
}

func profileJSON(uid, familyID string) []byte {
	p := []db.UserProfile{{UID: uid, FamilyID: familyID, ExplorerName: "Test Member", Role: "MEMBER", IsActive: true}}
	b, _ := json.Marshal(p)
	return b
}

func TestAdminResetPassword_Success(t *testing.T) {
	var capturedLocalHash string
	var capturedMustChange interface{}
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				return profileJSON("usr_target", "fam_1"), nil
			}
			return []byte("[]"), nil
		},
		mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
			m, ok := payload.(map[string]any)
			if ok {
				if table == "odyssey_local_users" {
					if v, exists := m["password_hash"]; exists {
						capturedLocalHash, _ = v.(string)
					}
				}
				if table == "odyssey_user_profiles" {
					capturedMustChange = m["must_change_password"]
				}
			}
			return []byte("{}"), nil
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/reset-password", bytes.NewReader([]byte("{}")))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d body %s", w.Code, w.Body.String())
	}
	var res map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	tmp := res["temporary_password"]
	if len(tmp) < 12 {
		t.Fatalf("expected temporary_password >=12 chars, got %q len %d", tmp, len(tmp))
	}
	if capturedLocalHash == "" {
		t.Fatal("expected password_hash to be stored")
	}
	if capturedMustChange != true {
		t.Fatalf("expected must_change_password true, got %v", capturedMustChange)
	}
	// Ensure response does not contain hash
	bodyStr := w.Body.String()
	if strings.Contains(bodyStr, "password_hash") {
		t.Fatal("response must not expose password_hash")
	}
}

func TestAdminResetPassword_Unauthorized(t *testing.T) {
	mockClient := &mockSupabaseClient{}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/reset-password", bytes.NewReader([]byte("{}")))
	// no claims
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", w.Code)
	}
}

func TestAdminResetPassword_NonAdmin(t *testing.T) {
	mockClient := &mockSupabaseClient{}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/reset-password", bytes.NewReader([]byte("{}")))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), memberClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-admin, got %d", w.Code)
	}
}

func TestAdminResetPassword_CrossTenantDenied(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			return profileJSON("usr_target", "fam_other"), nil
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/reset-password", bytes.NewReader([]byte("{}")))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusNotFound && w.Code != http.StatusForbidden {
		t.Fatalf("expected 404 or 403 for cross-tenant, got %d", w.Code)
	}
	// Ensure we don't leak whether UID exists via different message
	body := w.Body.String()
	if strings.Contains(body, "password") {
		t.Fatalf("error response must not leak password info, got %s", body)
	}
}

func TestAdminResetPassword_UnknownMember(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			return []byte("[]"), nil
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/unknown_uid/reset-password", bytes.NewReader([]byte("{}")))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown member, got %d", w.Code)
	}
}

func TestAdminResetPassword_ForcesPasswordChange(t *testing.T) {
	var mustChangeSeen bool
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			return profileJSON("usr_target", "fam_1"), nil
		},
		mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				if m, ok := payload.(map[string]any); ok {
					if v, exists := m["must_change_password"]; exists && v == true {
						mustChangeSeen = true
					}
				}
			}
			return []byte("{}"), nil
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/reset-password", bytes.NewReader([]byte("{}")))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !mustChangeSeen {
		t.Fatal("expected must_change_password=true to be set")
	}
}

func TestAdminResetPassword_PasswordIsNotStoredPlaintext(t *testing.T) {
	var storedHash string
	var tmpPassword string
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			return profileJSON("usr_target", "fam_1"), nil
		},
		mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
			if table == "odyssey_local_users" {
				if m, ok := payload.(map[string]any); ok {
					if v, exists := m["password_hash"]; exists {
						storedHash, _ = v.(string)
					}
				}
			}
			return []byte("{}"), nil
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/reset-password", bytes.NewReader([]byte("{}")))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	tmpPassword = res["temporary_password"]
	if tmpPassword == "" {
		t.Fatal("temporary_password missing")
	}
	if storedHash == "" {
		t.Fatal("stored hash missing")
	}
	if storedHash == tmpPassword {
		t.Fatal("password is stored as plaintext")
	}
	// bcrypt hashes start with $2a$
	if !strings.HasPrefix(storedHash, "$2a$") && !strings.HasPrefix(storedHash, "$2b$") {
		t.Fatalf("expected bcrypt hash, got %q", storedHash)
	}
	// Verify hash matches temp password via hasher
	h := auth.NewBcryptHasher()
	if err := h.Verify(storedHash, tmpPassword); err != nil {
		t.Fatalf("stored hash does not verify against temporary password: %v", err)
	}
}

func TestAdminResetPassword_TemporaryPasswordIsRandom(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 5; i++ {
		mockClient := &mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				return profileJSON("usr_target", "fam_1"), nil
			},
			mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
				return []byte("{}"), nil
			},
		}
		api := NewAPI(mockClient)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/reset-password", bytes.NewReader([]byte("{}")))
		req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
		w := httptest.NewRecorder()
		api.Handler(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("iter %d expected 200, got %d", i, w.Code)
		}
		var res map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		tmp := res["temporary_password"]
		if len(tmp) < 12 {
			t.Fatalf("expected len >=12, got %d", len(tmp))
		}
		if seen[tmp] {
			t.Fatalf("temporary password repeated: %q - not random", tmp)
		}
		seen[tmp] = true
		// Ensure not predictable like uid/username/timestamp/simple
		if tmp == "usr_target" || tmp == "123456" || strings.Contains(tmp, "usr_target") {
			t.Fatalf("temporary password is predictable: %q", tmp)
		}
	}
	if len(seen) != 5 {
		t.Fatalf("expected 5 unique passwords, got %d", len(seen))
	}
}
