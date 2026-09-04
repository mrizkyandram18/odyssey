package admin_tasks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"

	"odyssey/pkg/auth"
)

// verificationMock emulates PostgREST list + exact-count semantics for the
// submissions queue: profiles, paginated submissions (created_at DESC,
// id DESC), task enrichment, and GetWithCount totals.
type verificationMock struct {
	profiles []map[string]any
	subs     []map[string]any
	tasks    []map[string]any

	lastSubParams   string
	lastCountParams string
	countCalls      int
}

func verificationSeed() (profiles, subs, tasks []map[string]any) {
	profiles = []map[string]any{
		{"uid": "u1", "explorer_name": "Andi", "family_id": "fam-1"},
		{"uid": "u2", "explorer_name": "Budi", "family_id": "fam-1"},
	}
	subs = []map[string]any{
		{"id": float64(1), "task_id": float64(10), "user_uid": "u1", "submission_type": "MANUAL_VERIFY", "status": "PENDING", "payload": map[string]any{}, "created_at": "2026-09-01T10:00:00Z"},
		{"id": float64(2), "task_id": float64(10), "user_uid": "u2", "submission_type": "AUTO_QUIZ", "status": "APPROVED", "payload": map[string]any{}, "created_at": "2026-09-02T10:00:00Z"},
		{"id": float64(3), "task_id": float64(20), "user_uid": "u1", "submission_type": "MANUAL_VERIFY", "status": "PENDING", "payload": map[string]any{}, "created_at": "2026-09-03T10:00:00Z"},
		{"id": float64(4), "task_id": float64(20), "user_uid": "u2", "submission_type": "MANUAL_VERIFY", "status": "REJECTED", "payload": map[string]any{}, "created_at": "2026-09-04T09:00:00Z"},
		{"id": float64(5), "task_id": float64(10), "user_uid": "u1", "submission_type": "MANUAL_VERIFY", "status": "PENDING", "payload": map[string]any{}, "created_at": "2026-09-04T10:00:00Z"},
		{"id": float64(6), "task_id": float64(20), "user_uid": "u2", "submission_type": "MANUAL_VERIFY", "status": "PENDING", "payload": map[string]any{}, "created_at": "2026-09-04T10:00:00Z"},
		// Other family's submission — must never leak into results or totals.
		{"id": float64(7), "task_id": float64(10), "user_uid": "u3", "submission_type": "MANUAL_VERIFY", "status": "PENDING", "payload": map[string]any{}, "created_at": "2026-09-05T10:00:00Z"},
	}
	tasks = []map[string]any{
		{"id": float64(10), "title": "Tugas A", "task_type": "TEXT_RESPONSE", "reward_coins": float64(50), "reward_xp": float64(100)},
		{"id": float64(20), "title": "Tugas B", "task_type": "PHOTO_UPLOAD", "reward_coins": float64(30), "reward_xp": float64(60)},
	}
	return profiles, subs, tasks
}

func verificationStatusParam(params string) string {
	for _, part := range strings.Split(params, "&") {
		if strings.HasPrefix(part, "status=eq.") {
			return strings.TrimPrefix(part, "status=eq.")
		}
	}
	return ""
}

func verificationUIDs(params string) map[string]bool {
	out := map[string]bool{}
	idx := strings.Index(params, "user_uid=in.(")
	if idx < 0 {
		return out
	}
	seg := params[idx+len("user_uid=in.("):]
	if end := strings.Index(seg, ")"); end >= 0 {
		seg = seg[:end]
	}
	for _, s := range strings.Split(seg, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			out[s] = true
		}
	}
	return out
}

func verificationFilteredSubs(all []map[string]any, params string) []map[string]any {
	uids := verificationUIDs(params)
	status := verificationStatusParam(params)
	filtered := make([]map[string]any, 0, len(all))
	for _, s := range all {
		uid, _ := s["user_uid"].(string)
		if len(uids) > 0 && !uids[uid] {
			continue
		}
		if status != "" && s["status"] != status {
			continue
		}
		filtered = append(filtered, s)
	}
	// Deterministic DB order: created_at DESC, id DESC.
	sort.Slice(filtered, func(i, j int) bool {
		ci, _ := filtered[i]["created_at"].(string)
		cj, _ := filtered[j]["created_at"].(string)
		if ci != cj {
			return ci > cj
		}
		return idOf(filtered[i]) > idOf(filtered[j])
	})
	return filtered
}

func verificationLimitOffset(params string) (limit, offset int) {
	limit = 50
	offset = 0
	for _, part := range strings.Split(params, "&") {
		if strings.HasPrefix(part, "limit=") {
			if v, err := strconv.Atoi(strings.TrimPrefix(part, "limit=")); err == nil {
				limit = v
			}
		}
		if strings.HasPrefix(part, "offset=") {
			if v, err := strconv.Atoi(strings.TrimPrefix(part, "offset=")); err == nil {
				offset = v
			}
		}
	}
	return limit, offset
}

func (m *verificationMock) Get(ctx context.Context, table string, params string) ([]byte, error) {
	switch table {
	case "odyssey_user_profiles":
		return json.Marshal(m.profiles)
	case "odyssey_task_submissions":
		m.lastSubParams = params
		filtered := verificationFilteredSubs(m.subs, params)
		limit, offset := verificationLimitOffset(params)
		if offset >= len(filtered) {
			return json.Marshal([]map[string]any{})
		}
		end := offset + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		return json.Marshal(filtered[offset:end])
	case "odyssey_tasks":
		return json.Marshal(m.tasks)
	}
	return json.Marshal([]map[string]any{})
}

func (m *verificationMock) GetWithCount(ctx context.Context, table string, params string) ([]byte, int64, error) {
	m.countCalls++
	m.lastCountParams = params
	filtered := verificationFilteredSubs(m.subs, params)
	return []byte("[]"), int64(len(filtered)), nil
}

func (m *verificationMock) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *verificationMock) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *verificationMock) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *verificationMock) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return "", nil
}

func runVerificationRequest(m *verificationMock, target string) *httptest.ResponseRecorder {
	api := NewAPI(m)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	claims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.HandleListPendingSubmissions(rec, req)
	return rec
}

type verificationPage struct {
	Items      []map[string]any `json:"items"`
	Pagination struct {
		Page    int  `json:"page"`
		Limit   int  `json:"limit"`
		Total   int  `json:"total"`
		HasNext bool `json:"has_next"`
	} `json:"pagination"`
}

func decodeVerificationPage(t *testing.T, rec *httptest.ResponseRecorder) verificationPage {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", rec.Code, rec.Body.String())
	}
	var page verificationPage
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("expected paginated JSON, got %q: %v", rec.Body.String(), err)
	}
	return page
}

func verificationIDs(page verificationPage) []int64 {
	ids := make([]int64, len(page.Items))
	for i, it := range page.Items {
		if v, ok := it["id"].(float64); ok {
			ids[i] = int64(v)
		}
	}
	return ids
}

func equalIDs(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestVerificationPage1ExactTotal(t *testing.T) {
	profs, subs, tasks := verificationSeed()
	m := &verificationMock{profiles: profs, subs: subs, tasks: tasks}
	rec := runVerificationRequest(m, "/api/admin/submissions?page=1&limit=2")
	page := decodeVerificationPage(t, rec)
	if !equalIDs(verificationIDs(page), []int64{6, 5}) {
		t.Fatalf("expected deterministic order [6,5] (created_at DESC, id DESC), got %v", verificationIDs(page))
	}
	if page.Pagination.Total != 6 {
		t.Fatalf("expected exact total 6 (u3 excluded), got %d", page.Pagination.Total)
	}
	if !page.Pagination.HasNext {
		t.Fatalf("expected has_next=true on page 1 of 3")
	}
	if !strings.Contains(m.lastSubParams, "order=created_at.desc,id.desc") {
		t.Fatalf("expected deterministic ordering in query, got %s", m.lastSubParams)
	}
	if !strings.Contains(m.lastSubParams, "limit=2") || !strings.Contains(m.lastSubParams, "offset=0") {
		t.Fatalf("expected limit/offset pagination in query, got %s", m.lastSubParams)
	}
	if m.countCalls != 1 {
		t.Fatalf("expected exactly 1 lightweight count query, got %d", m.countCalls)
	}
}

func TestVerificationPage2And3(t *testing.T) {
	profs, subs, tasks := verificationSeed()
	m := &verificationMock{profiles: profs, subs: subs, tasks: tasks}
	page2 := decodeVerificationPage(t, runVerificationRequest(m, "/api/admin/submissions?page=2&limit=2"))
	if !equalIDs(verificationIDs(page2), []int64{4, 3}) {
		t.Fatalf("expected page 2 [4,3], got %v", verificationIDs(page2))
	}
	if page2.Pagination.Total != 6 || !page2.Pagination.HasNext {
		t.Fatalf("expected total 6 has_next=true on page 2, got %+v", page2.Pagination)
	}
	page3 := decodeVerificationPage(t, runVerificationRequest(m, "/api/admin/submissions?page=3&limit=2"))
	if !equalIDs(verificationIDs(page3), []int64{2, 1}) {
		t.Fatalf("expected page 3 [2,1], got %v", verificationIDs(page3))
	}
	if page3.Pagination.Total != 6 || page3.Pagination.HasNext {
		t.Fatalf("expected total 6 has_next=false on last page, got %+v", page3.Pagination)
	}
}

func TestVerificationStatusFilterCounts(t *testing.T) {
	profs, subs, tasks := verificationSeed()
	m := &verificationMock{profiles: profs, subs: subs, tasks: tasks}

	pending := decodeVerificationPage(t, runVerificationRequest(m, "/api/admin/submissions?status=PENDING&page=1&limit=2"))
	if !equalIDs(verificationIDs(pending), []int64{6, 5}) {
		t.Fatalf("expected pending page 1 [6,5], got %v", verificationIDs(pending))
	}
	if pending.Pagination.Total != 4 || !pending.Pagination.HasNext {
		t.Fatalf("expected pending total 4 has_next=true, got %+v", pending.Pagination)
	}
	if !strings.Contains(m.lastCountParams, "status=eq.PENDING") {
		t.Fatalf("count must apply the same status filter, got %s", m.lastCountParams)
	}

	approved := decodeVerificationPage(t, runVerificationRequest(m, "/api/admin/submissions?status=APPROVED"))
	if approved.Pagination.Total != 1 || !equalIDs(verificationIDs(approved), []int64{2}) {
		t.Fatalf("expected approved total 1 [2], got %v %+v", verificationIDs(approved), approved.Pagination)
	}
	if approved.Pagination.HasNext {
		t.Fatalf("expected has_next=false for single approved row")
	}

	rejected := decodeVerificationPage(t, runVerificationRequest(m, "/api/admin/submissions?status=REJECTED"))
	if rejected.Pagination.Total != 1 || !equalIDs(verificationIDs(rejected), []int64{4}) {
		t.Fatalf("expected rejected total 1 [4], got %v %+v", verificationIDs(rejected), rejected.Pagination)
	}

	all := decodeVerificationPage(t, runVerificationRequest(m, "/api/admin/submissions?status=ALL"))
	if all.Pagination.Total != 6 {
		t.Fatalf("expected ALL total 6, got %+v", all.Pagination)
	}
	if strings.Contains(m.lastCountParams, "status=eq.") {
		t.Fatalf("ALL count must not carry a status clause, got %s", m.lastCountParams)
	}
}

func TestVerificationEmptyPageKeepsExactTotal(t *testing.T) {
	profs, subs, tasks := verificationSeed()
	m := &verificationMock{profiles: profs, subs: subs, tasks: tasks}
	page := decodeVerificationPage(t, runVerificationRequest(m, "/api/admin/submissions?page=99&limit=2"))
	if len(page.Items) != 0 {
		t.Fatalf("expected empty items on out-of-range page, got %d", len(page.Items))
	}
	if page.Pagination.Total != 6 || page.Pagination.HasNext {
		t.Fatalf("expected exact total 6 has_next=false on empty page, got %+v", page.Pagination)
	}
}

// plainGetClient exposes only db.SupabaseClient (no GetWithCount method),
// so the handler must fall back to the legacy estimation instead of failing.
func TestVerificationFallbackWithoutCount(t *testing.T) {
	profs, subs, tasks := verificationSeed()
	plain := &plainGetClient{profiles: profs, subs: subs, tasks: tasks}
	api := NewAPI(plain)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/submissions?page=1&limit=2", nil)
	claims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.HandleListPendingSubmissions(rec, req)
	page := decodeVerificationPage(t, rec)
	if !equalIDs(verificationIDs(page), []int64{6, 5}) {
		t.Fatalf("expected page 1 [6,5] even in fallback, got %v", verificationIDs(page))
	}
	// Legacy estimate: offset(0) + len(2) + 1 (full page) = 3.
	if page.Pagination.Total != 3 || !page.Pagination.HasNext {
		t.Fatalf("expected fallback estimate total 3 has_next=true, got %+v", page.Pagination)
	}
}

type plainGetClient struct {
	profiles []map[string]any
	subs     []map[string]any
	tasks    []map[string]any
}

func (m *plainGetClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	inner := &verificationMock{profiles: m.profiles, subs: m.subs, tasks: m.tasks}
	return inner.Get(ctx, table, params)
}
func (m *plainGetClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *plainGetClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *plainGetClient) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *plainGetClient) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return "", nil
}
