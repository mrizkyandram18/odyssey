package admin_members

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
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

// listFilterMock honors family scoping, uid=not.in exclusion, and
// limit/offset like PostgREST so list-filtering behavior is testable.
type listFilterMock struct {
	profiles    []map[string]any
	credentials map[string]bool
	pageParams  []string
}

func newListFilterMock() *listFilterMock {
	profiles := []map[string]any{
		{"uid": "u1", "family_id": "fam_1", "explorer_name": "Active", "role": "MEMBER", "is_active": true, "level": 1, "xp": 0, "coins": 0, "created_at": "2026-09-03T00:00:00Z"},
		{"uid": "u2", "family_id": "fam_1", "explorer_name": "Blocked", "role": "MEMBER", "is_active": false, "level": 1, "xp": 0, "coins": 0, "created_at": "2026-09-02T00:00:00Z"},
		{"uid": "u3", "family_id": "fam_1", "explorer_name": "Deleted", "role": "MEMBER", "is_active": false, "level": 1, "xp": 0, "coins": 0, "created_at": "2026-09-05T00:00:00Z"},
		{"uid": "u4", "family_id": "fam_1", "explorer_name": "Admin", "role": "ADMIN", "is_active": true, "level": 1, "xp": 0, "coins": 0, "created_at": "2026-09-04T00:00:00Z"},
		{"uid": "u5", "family_id": "fam_1", "explorer_name": "Guide", "role": "GUIDE", "is_active": true, "level": 1, "xp": 0, "coins": 0, "created_at": "2026-09-01T00:00:00Z"},
		{"uid": "u9", "family_id": "fam_2", "explorer_name": "Other", "role": "MEMBER", "is_active": true, "level": 1, "xp": 0, "coins": 0, "created_at": "2026-09-06T00:00:00Z"},
	}
	return &listFilterMock{
		profiles:    profiles,
		credentials: map[string]bool{"u1": true, "u2": true, "u4": true, "u5": true, "u9": true},
	}
}

func listFamilyParam(params string) string {
	for _, part := range strings.Split(params, "&") {
		if strings.HasPrefix(part, "family_id=eq.") {
			return strings.TrimPrefix(part, "family_id=eq.")
		}
	}
	return ""
}

func listExcludedUIDs(params string) map[string]bool {
	out := map[string]bool{}
	idx := strings.Index(params, "uid=not.in.(")
	if idx < 0 {
		return out
	}
	seg := params[idx+len("uid=not.in.("):]
	if end := strings.Index(seg, ")"); end >= 0 {
		seg = seg[:end]
	}
	for _, s := range strings.Split(seg, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out[s] = true
		}
	}
	return out
}

func listCredentialUIDs(params string) map[string]bool {
	out := map[string]bool{}
	idx := strings.Index(params, "profile_uid=in.(")
	if idx < 0 {
		return out
	}
	seg := params[idx+len("profile_uid=in.("):]
	if end := strings.Index(seg, ")"); end >= 0 {
		seg = seg[:end]
	}
	for _, s := range strings.Split(seg, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out[s] = true
		}
	}
	return out
}

func (m *listFilterMock) Get(ctx context.Context, table string, params string) ([]byte, error) {
	switch table {
	case "odyssey_user_profiles":
		fam := listFamilyParam(params)
		excluded := listExcludedUIDs(params)
		filtered := make([]map[string]any, 0)
		for _, p := range m.profiles {
			if fam != "" && p["family_id"] != fam {
				continue
			}
			if excluded[p["uid"].(string)] {
				continue
			}
			filtered = append(filtered, p)
		}
		sort.Slice(filtered, func(i, j int) bool {
			ci, _ := filtered[i]["created_at"].(string)
			cj, _ := filtered[j]["created_at"].(string)
			if ci != cj {
				return ci > cj
			}
			return filtered[i]["uid"].(string) > filtered[j]["uid"].(string)
		})
		if strings.Contains(params, "order=created_at.desc") {
			m.pageParams = append(m.pageParams, params)
			limit, offset := 50, 0
			for _, part := range strings.Split(params, "&") {
				if strings.HasPrefix(part, "limit=") {
					limit, _ = strconv.Atoi(strings.TrimPrefix(part, "limit="))
				}
				if strings.HasPrefix(part, "offset=") {
					offset, _ = strconv.Atoi(strings.TrimPrefix(part, "offset="))
				}
			}
			if offset >= len(filtered) {
				return json.Marshal([]map[string]any{})
			}
			end := offset + limit
			if end > len(filtered) {
				end = len(filtered)
			}
			return json.Marshal(filtered[offset:end])
		}
		return json.Marshal(filtered)
	case "odyssey_local_users":
		uids := listCredentialUIDs(params)
		rows := make([]map[string]any, 0)
		for uid := range uids {
			if m.credentials[uid] {
				rows = append(rows, map[string]any{"username": "user_" + uid, "profile_uid": uid})
			}
		}
		// verificationUIDs-style fallback: plain profile_uid=eq.X
		for _, part := range strings.Split(params, "&") {
			if strings.HasPrefix(part, "profile_uid=eq.") {
				uid := strings.TrimPrefix(part, "profile_uid=eq.")
				if m.credentials[uid] {
					rows = append(rows, map[string]any{"username": "user_" + uid, "profile_uid": uid})
				}
			}
		}
		return json.Marshal(rows)
	}
	return json.Marshal([]map[string]any{})
}

func (m *listFilterMock) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *listFilterMock) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *listFilterMock) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *listFilterMock) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return "", nil
}

func runListRequest(m *listFilterMock, target string) (int, []MemberView, map[string]any) {
	api := NewAPI(m)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	var body struct {
		Items      []MemberView   `json:"items"`
		Pagination map[string]any `json:"pagination"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body.Items, body.Pagination
}

func listUIDs(items []MemberView) []string {
	uids := make([]string, len(items))
	for i, m := range items {
		uids[i] = m.UID
	}
	return uids
}

func TestListMembers_HidesDeletedKeepsBlocked(t *testing.T) {
	m := newListFilterMock()
	code, items, _ := runListRequest(m, "/api/admin/members?page=1&limit=10")
	if code != http.StatusOK {
		t.Fatalf("expected 200 got %d", code)
	}
	got := listUIDs(items)
	// Deleted u3 excluded (despite newest created_at); active, blocked,
	// admin, guide visible in deterministic order; cross-family u9 never.
	want := []string{"u4", "u1", "u2", "u5"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
	if len(m.pageParams) == 0 {
		t.Fatalf("expected paginated page query to be issued")
	}
	pageQ := m.pageParams[len(m.pageParams)-1]
	if !strings.Contains(pageQ, "uid=not.in.(u3)") {
		t.Fatalf("expected DB-level deleted exclusion, got %s", pageQ)
	}
	if !strings.Contains(pageQ, "family_id=eq.fam_1") || !strings.Contains(pageQ, "order=created_at.desc,uid.desc") {
		t.Fatalf("expected family scope + deterministic order, got %s", pageQ)
	}
}

func TestListMembers_NoExclusionClauseWhenNothingDeleted(t *testing.T) {
	m := newListFilterMock()
	m.credentials["u3"] = true // u3 becomes ordinary blocked, not deleted
	code, items, _ := runListRequest(m, "/api/admin/members?page=1&limit=10")
	if code != http.StatusOK {
		t.Fatalf("expected 200 got %d", code)
	}
	if len(items) != 5 {
		t.Fatalf("expected all 5 family members, got %v", listUIDs(items))
	}
	pageQ := m.pageParams[len(m.pageParams)-1]
	if strings.Contains(pageQ, "not.in") {
		t.Fatalf("expected no exclusion clause when nothing deleted, got %s", pageQ)
	}
}

func TestListMembers_DeletedExcludedFromPagination(t *testing.T) {
	m := newListFilterMock()
	_, page1, pag1 := runListRequest(m, "/api/admin/members?page=1&limit=2")
	if len(page1) != 2 || page1[0].UID != "u4" || page1[1].UID != "u1" {
		t.Fatalf("expected page 1 [u4,u1], got %v", listUIDs(page1))
	}
	if pag1["has_next"] != true {
		t.Fatalf("expected has_next=true on page 1, got %v", pag1)
	}
	_, page2, pag2 := runListRequest(m, "/api/admin/members?page=2&limit=2")
	if len(page2) != 2 || page2[0].UID != "u2" || page2[1].UID != "u5" {
		t.Fatalf("expected page 2 [u2,u5] (deleted u3 excluded), got %v", listUIDs(page2))
	}
	// Members list uses the established estimated-total pattern: a full page
	// implies has_next. The deleted member never surfaces on any page.
	if pag2["has_next"] != true {
		t.Fatalf("expected has_next=true on full page 2 (estimate pattern), got %v", pag2)
	}
}
