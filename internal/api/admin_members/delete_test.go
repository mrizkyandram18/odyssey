package admin_members

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
)

type mutateCall struct {
	method  string
	table   string
	payload any
	params  string
}

type deleteTestClient struct {
	mockSupabaseClient
	mutates []mutateCall
}

func newDeleteTestClient(profiles []db.UserProfile, hasCredential bool, rpcErr error) *deleteTestClient {
	c := &deleteTestClient{}
	c.getFunc = func(ctx context.Context, table string, params string) ([]byte, error) {
		if table == "odyssey_user_profiles" {
			return json.Marshal(profiles)
		}
		if table == "odyssey_local_users" {
			if hasCredential {
				return json.Marshal([]map[string]any{{"profile_uid": "usr_target"}})
			}
			return []byte("[]"), nil
		}
		return []byte("[]"), nil
	}
	c.mutateFunc = func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
		c.mutates = append(c.mutates, mutateCall{method: method, table: table, payload: payload, params: params})
		return []byte("{}"), nil
	}
	if rpcErr != nil {
		c.rpcFunc = func(ctx context.Context, fn string, payload any) ([]byte, error) {
			return nil, rpcErr
		}
	}
	return c
}

func runDeleteRequest(c *deleteTestClient, targetUID string, claims *auth.SessionClaims) *httptest.ResponseRecorder {
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/members/"+targetUID, nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	return w
}

func tablesMutated(c *deleteTestClient, method string) map[string]int {
	out := map[string]int{}
	for _, m := range c.mutates {
		if m.method == method {
			out[m.table]++
		}
	}
	return out
}

func TestHandleDeleteMember_Success(t *testing.T) {
	profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", ExplorerName: "Target", Role: "MEMBER", IsActive: true}}
	c := newDeleteTestClient(profiles, true, nil)
	w := runDeleteRequest(c, "usr_target", adminClaims("fam_1"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body %s", w.Code, w.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["success"] != true || res["deleted"] != true || res["is_active"] != false {
		t.Fatalf("expected deleted soft-delete response, got %v", res)
	}
	// Exactly one credential DELETE; profile row itself is never hard-deleted.
	dels := tablesMutated(c, http.MethodDelete)
	if len(dels) != 1 || dels["odyssey_local_users"] != 1 {
		t.Fatalf("expected single odyssey_local_users DELETE, got %v (all mutates %v)", dels, c.mutates)
	}
	for _, m := range c.mutates {
		if m.table == "odyssey_task_submissions" || m.table == "odyssey_claims" ||
			m.table == "odyssey_coin_transactions" || m.table == "odyssey_tasks" ||
			m.table == "odyssey_user_profiles" {
			t.Fatalf("history/profile must never be deleted, got %s %s", m.method, m.table)
		}
	}
}

func TestHandleDeleteMember_FallbackPatch(t *testing.T) {
	profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", ExplorerName: "Target", Role: "MEMBER", IsActive: true}}
	c := newDeleteTestClient(profiles, true, errors.New("rpc unavailable"))
	w := runDeleteRequest(c, "usr_target", adminClaims("fam_1"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body %s", w.Code, w.Body.String())
	}
	var profilePatch, credDelete bool
	for _, m := range c.mutates {
		if m.method == http.MethodPatch && m.table == "odyssey_user_profiles" {
			profilePatch = true
			pm, _ := m.payload.(map[string]any)
			if pm["is_active"] != false {
				t.Fatalf("expected is_active=false patch, got %v", pm)
			}
			if !strings.Contains(m.params, "uid=eq.usr_target") || !strings.Contains(m.params, "is_active=eq.true") {
				t.Fatalf("expected conditional deactivation params, got %s", m.params)
			}
		}
		if m.method == http.MethodDelete && m.table == "odyssey_local_users" {
			credDelete = true
		}
	}
	if !profilePatch || !credDelete {
		t.Fatalf("expected profile PATCH + credential DELETE, got %v", c.mutates)
	}
}

func TestHandleDeleteMember_AlreadyDeleted(t *testing.T) {
	profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", ExplorerName: "Target", Role: "MEMBER", IsActive: false}}
	c := newDeleteTestClient(profiles, false, nil)
	w := runDeleteRequest(c, "usr_target", adminClaims("fam_1"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "already_deleted") {
		t.Fatalf("expected already_deleted in %s", w.Body.String())
	}
	if len(c.mutates) != 0 {
		t.Fatalf("idempotent re-delete must not write, got %v", c.mutates)
	}
}

func TestHandleDeleteMember_BlockedWithCredentialUpgrades(t *testing.T) {
	// Blocked (inactive) but credential still present: delete proceeds.
	profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", ExplorerName: "Target", Role: "MEMBER", IsActive: false}}
	c := newDeleteTestClient(profiles, true, nil)
	w := runDeleteRequest(c, "usr_target", adminClaims("fam_1"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body %s", w.Code, w.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["deleted"] != true {
		t.Fatalf("expected deleted:true, got %v", res)
	}
	if tablesMutated(c, http.MethodDelete)["odyssey_local_users"] != 1 {
		t.Fatalf("expected credential DELETE, got %v", c.mutates)
	}
}

func TestHandleDeleteMember_NotFound(t *testing.T) {
	c := newDeleteTestClient(nil, false, nil)
	c.getFunc = func(ctx context.Context, table string, params string) ([]byte, error) {
		return []byte("[]"), nil
	}
	w := runDeleteRequest(c, "usr_missing", adminClaims("fam_1"))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d body %s", w.Code, w.Body.String())
	}
	if len(c.mutates) != 0 {
		t.Fatalf("not-found delete must not write, got %v", c.mutates)
	}
}

func TestHandleDeleteMember_CrossFamily(t *testing.T) {
	profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_2", ExplorerName: "Target", Role: "MEMBER", IsActive: true}}
	c := newDeleteTestClient(profiles, true, nil)
	w := runDeleteRequest(c, "usr_target", adminClaims("fam_1"))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-family (no existence leak) got %d body %s", w.Code, w.Body.String())
	}
	if len(c.mutates) != 0 {
		t.Fatalf("cross-family delete must not write, got %v", c.mutates)
	}
}

func TestHandleDeleteMember_Unauthorized(t *testing.T) {
	c := newDeleteTestClient(nil, false, nil)
	w := runDeleteRequest(c, "usr_target", memberClaims("fam_1"))
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", w.Code)
	}
}

func TestHandleDeleteMember_AdminTarget(t *testing.T) {
	profiles := []db.UserProfile{{UID: "admin_target", FamilyID: "fam_1", ExplorerName: "Admin", Role: "ADMIN", IsActive: true}}
	c := newDeleteTestClient(profiles, true, nil)
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/members/admin_target", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for admin target got %d body %s", w.Code, w.Body.String())
	}
	if len(c.mutates) != 0 {
		t.Fatalf("admin-target delete must not write, got %v", c.mutates)
	}
}

func TestHandleDeleteMember_SelfDelete(t *testing.T) {
	profiles := []db.UserProfile{{UID: "admin_1", FamilyID: "fam_1", ExplorerName: "Self", Role: "MEMBER", IsActive: true}}
	c := newDeleteTestClient(profiles, true, nil)
	w := runDeleteRequest(c, "admin_1", adminClaims("fam_1"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for self delete got %d body %s", w.Code, w.Body.String())
	}
	if len(c.mutates) != 0 {
		t.Fatalf("self delete must not write, got %v", c.mutates)
	}
}

func TestHandleUnblockMember_DeletedRejected(t *testing.T) {
	// Inactive member with revoked credential cannot be resurrected.
	profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", ExplorerName: "Target", Role: "MEMBER", IsActive: false}}
	c := newDeleteTestClient(profiles, false, nil)
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/unblock", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 unblocking a deleted member got %d body %s", w.Code, w.Body.String())
	}
	if len(c.mutates) != 0 {
		t.Fatalf("rejected unblock must not write, got %v", c.mutates)
	}
}

func TestHandleUnblockMember_BlockedWithCredentialAllowed(t *testing.T) {
	// Ordinary blocked member (credential intact) still unblocks normally.
	profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", ExplorerName: "Target", Role: "MEMBER", IsActive: false}}
	c := newDeleteTestClient(profiles, true, nil)
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/unblock", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body %s", w.Code, w.Body.String())
	}
}
