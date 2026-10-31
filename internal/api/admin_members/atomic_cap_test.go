package admin_members

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
)

func TestHandleUpdateMember_AtomicCap_Success(t *testing.T) {
	var rpcPayload map[string]any
	var rpcCalled bool
	var mutateCalled bool

	capVal := 3500
	statusVal := "CONFIGURED"
	reasonVal := "Penyesuaian target performa kuartal 4"

	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				if table == "odyssey_user_profiles" {
					profiles := []db.UserProfile{{
						UID:                "usr_target",
						FamilyID:           "fam_1",
						ExplorerName:       "Target Member",
						Role:               "MEMBER",
						IsActive:           true,
						IncomeStatus:       statusVal,
						MonthlyEarningCap:  &capVal,
						MonthlyCoinTarget:  &capVal,
					}}
					return json.Marshal(profiles)
				}
				if table == "odyssey_system_config" {
					return json.Marshal([]map[string]string{{"value": "10000"}})
				}
				return []byte("[]"), nil
			},
			mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
				mutateCalled = true
				return []byte(`{"status":"updated"}`), nil
			},
		},
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			if fn == "odyssey_admin_update_member_income_config" {
				rpcCalled = true
				if p, ok := payload.(map[string]any); ok {
					rpcPayload = p
				}
				return json.Marshal(map[string]any{
					"success":           true,
					"changed":           true,
					"target_uid":        "usr_target",
					"old_income_status": "NOT_CONFIGURED",
					"new_income_status": statusVal,
					"old_cap":           nil,
					"new_cap":           capVal,
					"audit_id":          "11111111-2222-3333-4444-555555555555",
				})
			}
			return []byte("{}"), nil
		},
	}

	api := NewAPI(mock)
	body, _ := json.Marshal(UpdateMemberRequest{
		IncomeStatus:      &statusVal,
		MonthlyEarningCap: &capVal,
		Reason:            &reasonVal,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d body %s", w.Code, w.Body.String())
	}
	if !rpcCalled {
		t.Fatal("expected RPC odyssey_admin_update_member_income_config to be called")
	}
	if rpcPayload["p_target_uid"] != "usr_target" {
		t.Errorf("expected p_target_uid 'usr_target', got %v", rpcPayload["p_target_uid"])
	}
	if rpcPayload["p_admin_uid"] != "admin_1" {
		t.Errorf("expected p_admin_uid 'admin_1', got %v", rpcPayload["p_admin_uid"])
	}
	if rpcPayload["p_income_status"] != "CONFIGURED" {
		t.Errorf("expected p_income_status 'CONFIGURED', got %v", rpcPayload["p_income_status"])
	}
	if rpcPayload["p_monthly_cap"] != 3500 {
		t.Errorf("expected p_monthly_cap 3500, got %v", rpcPayload["p_monthly_cap"])
	}
	if rpcPayload["p_reason"] != reasonVal {
		t.Errorf("expected p_reason '%s', got %v", reasonVal, rpcPayload["p_reason"])
	}
	// Verify that redundant non-atomic mutate on user profiles was NOT invoked since only cap/status changed
	if mutateCalled {
		t.Error("redundant profile Mutate was called despite cap/status handled atomically by RPC")
	}
}

func TestHandleUpdateMember_AtomicCap_RPCFailureRollback(t *testing.T) {
	var mutateCalled bool
	capVal := 2000
	statusVal := "CONFIGURED"

	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				if table == "odyssey_user_profiles" {
					profiles := []db.UserProfile{{
						UID:          "usr_target",
						FamilyID:     "fam_1",
						ExplorerName: "Target Member",
						Role:         "MEMBER",
						IsActive:     true,
					}}
					return json.Marshal(profiles)
				}
				return []byte("[]"), nil
			},
			mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
				mutateCalled = true
				return []byte(`{"status":"updated"}`), nil
			},
		},
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			if fn == "odyssey_admin_update_member_income_config" {
				return nil, errors.New("P0012: audit log insert failed: disk full")
			}
			return []byte("{}"), nil
		},
	}

	api := NewAPI(mock)
	body, _ := json.Marshal(UpdateMemberRequest{
		IncomeStatus:      &statusVal,
		MonthlyEarningCap: &capVal,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on RPC error, got %d body %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("gagal update budget earning")) {
		t.Errorf("expected error message to contain 'gagal update budget earning', got %s", w.Body.String())
	}
	if mutateCalled {
		t.Fatal("profile Mutate must not be called when RPC fails (must fail-closed)")
	}
}

func TestHandleUpdateMember_AtomicCap_UnauthorizedActor(t *testing.T) {
	mock := &rpcMockClient{}
	api := NewAPI(mock)

	capVal := 1500
	body, _ := json.Marshal(UpdateMemberRequest{
		MonthlyEarningCap: &capVal,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	// Calling with MEMBER role instead of ADMIN
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:      "member_1",
		Role:     "MEMBER",
		FamilyID: "fam_1",
	}))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-admin, got %d body %s", w.Code, w.Body.String())
	}
}

func TestHandleUpdateMember_AtomicCap_FamilyIsolation_DifferentFamilies(t *testing.T) {
	var rpcCalled bool
	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				if table == "odyssey_user_profiles" {
					// Target belongs to fam_2
					profiles := []db.UserProfile{{
						UID:          "usr_target",
						FamilyID:     "fam_2",
						ExplorerName: "Target Member",
						Role:         "MEMBER",
						IsActive:     true,
					}}
					return json.Marshal(profiles)
				}
				return []byte("[]"), nil
			},
		},
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			rpcCalled = true
			return []byte("{}"), nil
		},
	}

	api := NewAPI(mock)
	capVal := 2000
	body, _ := json.Marshal(UpdateMemberRequest{
		MonthlyEarningCap: &capVal,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	// Admin belongs to fam_1
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for cross-family update, got %d body %s", w.Code, w.Body.String())
	}
	if rpcCalled {
		t.Fatal("RPC must never be invoked when cross-family check fails")
	}
}

func TestHandleUpdateMember_AtomicCap_FamilyIsolation_AdminFamilyEmpty(t *testing.T) {
	var rpcCalled bool
	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				if table == "odyssey_user_profiles" {
					profiles := []db.UserProfile{{
						UID:          "usr_target",
						FamilyID:     "fam_1",
						ExplorerName: "Target Member",
						Role:         "MEMBER",
						IsActive:     true,
					}}
					return json.Marshal(profiles)
				}
				return []byte("[]"), nil
			},
		},
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			rpcCalled = true
			return []byte("{}"), nil
		},
	}

	api := NewAPI(mock)
	capVal := 2000
	body, _ := json.Marshal(UpdateMemberRequest{
		MonthlyEarningCap: &capVal,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	// Admin has EMPTY family_id
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "",
	}))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when admin has empty family_id, got %d body %s", w.Code, w.Body.String())
	}
	if rpcCalled {
		t.Fatal("RPC must never be invoked when admin family_id is empty")
	}
}

func TestHandleUpdateMember_AtomicCap_FamilyIsolation_MemberFamilyEmpty(t *testing.T) {
	var rpcCalled bool
	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				if table == "odyssey_user_profiles" {
					// Target has EMPTY family_id
					profiles := []db.UserProfile{{
						UID:          "usr_target",
						FamilyID:     "",
						ExplorerName: "Target Member",
						Role:         "MEMBER",
						IsActive:     true,
					}}
					return json.Marshal(profiles)
				}
				return []byte("[]"), nil
			},
		},
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			rpcCalled = true
			return []byte("{}"), nil
		},
	}

	api := NewAPI(mock)
	capVal := 2000
	body, _ := json.Marshal(UpdateMemberRequest{
		MonthlyEarningCap: &capVal,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	// Admin has fam_1
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when member has empty family_id, got %d body %s", w.Code, w.Body.String())
	}
	if rpcCalled {
		t.Fatal("RPC must never be invoked when member family_id is empty")
	}
}

func TestHandleUpdateMember_AtomicCap_FamilyIsolation_BothFamiliesEmpty(t *testing.T) {
	var rpcCalled bool
	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				if table == "odyssey_user_profiles" {
					// Target has EMPTY family_id
					profiles := []db.UserProfile{{
						UID:          "usr_target",
						FamilyID:     "",
						ExplorerName: "Target Member",
						Role:         "MEMBER",
						IsActive:     true,
					}}
					return json.Marshal(profiles)
				}
				return []byte("[]"), nil
			},
		},
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			rpcCalled = true
			return []byte("{}"), nil
		},
	}

	api := NewAPI(mock)
	capVal := 2000
	body, _ := json.Marshal(UpdateMemberRequest{
		MonthlyEarningCap: &capVal,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	// Both admin and member have empty family_id
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "",
	}))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when both family_ids are empty, got %d body %s", w.Code, w.Body.String())
	}
	if rpcCalled {
		t.Fatal("RPC must never be invoked when both family_ids are empty")
	}
}

func TestHandleUpdateMember_AtomicCap_ValidationInvalidStatus(t *testing.T) {
	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				if table == "odyssey_user_profiles" {
					profiles := []db.UserProfile{{
						UID:          "usr_target",
						FamilyID:     "fam_1",
						ExplorerName: "Target Member",
						Role:         "MEMBER",
						IsActive:     true,
					}}
					return json.Marshal(profiles)
				}
				return []byte("[]"), nil
			},
		},
	}

	api := NewAPI(mock)
	invalidStatus := "ILLEGAL_STATUS"
	body, _ := json.Marshal(UpdateMemberRequest{
		IncomeStatus: &invalidStatus,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for invalid income_status, got %d body %s", w.Code, w.Body.String())
	}
}

func TestHandleUpdateMember_AtomicCap_ValidationInvalidCapBounds(t *testing.T) {
	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				if table == "odyssey_user_profiles" {
					profiles := []db.UserProfile{{
						UID:          "usr_target",
						FamilyID:     "fam_1",
						ExplorerName: "Target Member",
						Role:         "MEMBER",
						IsActive:     true,
					}}
					return json.Marshal(profiles)
				}
				return []byte("[]"), nil
			},
		},
	}

	api := NewAPI(mock)
	negativeCap := -100
	body, _ := json.Marshal(UpdateMemberRequest{
		MonthlyEarningCap: &negativeCap,
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for negative cap, got %d body %s", w.Code, w.Body.String())
	}
}

func TestHandleUpdateMember_AtomicCap_DefaultReasonWhenOmitted(t *testing.T) {
	var capturedReason string
	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				if table == "odyssey_user_profiles" {
					profiles := []db.UserProfile{{
						UID:          "usr_target",
						FamilyID:     "fam_1",
						ExplorerName: "Target Member",
						Role:         "MEMBER",
						IsActive:     true,
					}}
					return json.Marshal(profiles)
				}
				return []byte("[]"), nil
			},
		},
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			if fn == "odyssey_admin_update_member_income_config" {
				if p, ok := payload.(map[string]any); ok {
					if r, ok := p["p_reason"].(string); ok {
						capturedReason = r
					}
				}
				return []byte(`{"success":true,"changed":true}`), nil
			}
			return []byte("{}"), nil
		},
	}

	api := NewAPI(mock)
	capVal := 2000
	body, _ := json.Marshal(UpdateMemberRequest{
		MonthlyEarningCap: &capVal,
		// Reason intentionally omitted
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:      "admin_1",
		Role:     "ADMIN",
		FamilyID: "fam_1",
	}))

	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d body %s", w.Code, w.Body.String())
	}
	expectedDefault := "Perubahan konfigurasi budget bulanan oleh admin"
	if capturedReason != expectedDefault {
		t.Errorf("expected default reason '%s', got '%s'", expectedDefault, capturedReason)
	}
}
