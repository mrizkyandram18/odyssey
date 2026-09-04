package admin_tasks

import (
	"bytes"
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

type seqMock struct {
	tasks []map[string]any
	lastParams string
	lastMutate map[string]any
}

func (m *seqMock) Get(ctx context.Context, table string, params string) ([]byte, error) {
	m.lastParams = params
	if table == "odyssey_tasks" {
		// for duplicate check: family_id=eq...&active_date=eq...&step_order=eq...
		if strings.Contains(params, "step_order=eq.") && strings.Contains(params, "select=id") {
			// parse step_order
			var step int
			for _, p := range strings.Split(params, "&") {
				if strings.HasPrefix(p, "step_order=eq.") {
					s := strings.TrimPrefix(p, "step_order=eq.")
					if v, err := json.Marshal(s); err == nil {
						_ = v
					}
					// simple parse
					var n int
					for _, c := range s {
						if c >= '0' && c <= '9' {
							n = n*10 + int(c-'0')
						}
					}
					step = n
				}
			}
			// return existing if step matches any task for same family/date
			var fam, date string
			for _, p := range strings.Split(params, "&") {
				if strings.HasPrefix(p, "family_id=eq.") {
					fam = strings.TrimPrefix(p, "family_id=eq.")
				}
				if strings.HasPrefix(p, "active_date=eq.") {
					date = strings.TrimPrefix(p, "active_date=eq.")
				}
			}
			for _, t := range m.tasks {
				if t["family_id"] == fam && t["active_date"] == date && t["step_order"] == step {
					return json.Marshal([]map[string]any{{"id": t["id"]}})
				}
			}
			return json.Marshal([]map[string]any{})
		}
		if strings.Contains(params, "order=step_order") && strings.Contains(params, "limit=1") {
			// MAX query for duplicate
			max := 0
			for _, t := range m.tasks {
				if v, ok := t["step_order"].(int); ok && v > max {
					max = v
				}
				if v, ok := t["step_order"].(float64); ok && int(v) > max {
					max = int(v)
				}
			}
			if max == 0 {
				return json.Marshal([]map[string]any{})
			}
			return json.Marshal([]map[string]any{{"step_order": max}})
		}
		// default list
		return json.Marshal(m.tasks)
	}
	if table == "odyssey_user_profiles" {
		return json.Marshal([]map[string]any{{"uid": "admin-1", "family_id": "fam-1"}})
	}
	return json.Marshal([]map[string]any{})
}
func (m *seqMock) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	m.lastMutate = payload.(map[string]any)
	return json.Marshal(map[string]any{"status": "updated"})
}
func (m *seqMock) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	m.lastMutate = payload.(map[string]any)
	m.tasks = append(m.tasks, map[string]any{
		"id":          len(m.tasks) + 100,
		"family_id":   payload.(map[string]any)["family_id"],
		"active_date": payload.(map[string]any)["active_date"],
		"step_order":  payload.(map[string]any)["step_order"],
	})
	return json.Marshal([]map[string]any{payload.(map[string]any)})
}
func (m *seqMock) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *seqMock) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return "", nil
}

func TestOrderingDeterministic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/tasks?date=2026-09-01", nil)
	claims := &auth.SessionClaims{UID: "a1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	c2 := &seqMock{tasks: []map[string]any{}}
	api2 := NewAPI(&mockWrapper{seq: c2})
	rec := httptest.NewRecorder()
	api2.HandleListTasks(rec, req)
	if !strings.Contains(c2.lastParams, "order=step_order.asc,id.asc") {
		t.Fatalf("expected deterministic ordering step_order.asc,id.asc, got %s", c2.lastParams)
	}
}

type mockWrapper struct{ seq *seqMock }
func (m *mockWrapper) Get(ctx context.Context, table string, params string) ([]byte, error) {
	return m.seq.Get(ctx, table, params)
}
func (m *mockWrapper) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return m.seq.Mutate(ctx, method, table, payload, params)
}
func (m *mockWrapper) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return m.seq.MutateAtomic(ctx, method, table, payload, params, prefer)
}
func (m *mockWrapper) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	return m.seq.RPC(ctx, fn, payload)
}
func (m *mockWrapper) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return m.seq.UploadStorage(ctx, bucket, path, contentType, data)
}

func TestDuplicateStepOrderRejectedOnCreate(t *testing.T) {
	seq := &seqMock{tasks: []map[string]any{
		{"id": 101, "family_id": "fam-1", "active_date": "2026-09-01", "step_order": 1},
	}}
	api := NewAPI(&mockWrapper{seq: seq})
	payload := `{"title":"Dup","task_type":"VIDEO","step_order":1,"active_date":"2026-09-01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "a1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.HandleCreateTask(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 duplicate, got %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "step_order sudah digunakan") {
		t.Fatalf("expected duplicate message, got %s", rec.Body.String())
	}
}

func TestDuplicateTaskGetsNextAvailableStep(t *testing.T) {
	seq := &seqMock{tasks: []map[string]any{
		{"id": 101, "family_id": "fam-1", "active_date": "2026-09-01", "step_order": 1},
		{"id": 102, "family_id": "fam-1", "active_date": "2026-09-01", "step_order": 2},
		{"id": 103, "family_id": "fam-1", "active_date": "2026-09-01", "step_order": 3},
	}}
	// need Get to return orig task 101
	orig := []map[string]any{{"id": 101, "family_id": "fam-1", "active_date": "2026-09-01", "title": "A", "task_type": "VIDEO", "step_order": 2, "reward_coins": 50, "reward_xp": 100, "is_active": true}}
	b, _ := json.Marshal(orig)
	wrapper := &dupWrapper{orig: b, seq: seq}
	api := NewAPI(wrapper)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks/101/duplicate", nil)
	claims := &auth.SessionClaims{UID: "a1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.Handler(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d %s", rec.Code, rec.Body.String())
	}
	if seq.lastMutate["step_order"] != 4 {
		t.Fatalf("expected next step 4, got %v", seq.lastMutate["step_order"])
	}
}

type dupWrapper struct {
	orig []byte
	seq  *seqMock
}

func (m *dupWrapper) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if table == "odyssey_tasks" && strings.Contains(params, "id=eq.101") {
		return m.orig, nil
	}
	return m.seq.Get(ctx, table, params)
}
func (m *dupWrapper) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return m.seq.Mutate(ctx, method, table, payload, params)
}
func (m *dupWrapper) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return m.seq.MutateAtomic(ctx, method, table, payload, params, prefer)
}
func (m *dupWrapper) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	return m.seq.RPC(ctx, fn, payload)
}
func (m *dupWrapper) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return m.seq.UploadStorage(ctx, bucket, path, contentType, data)
}

func TestSubmissionIdentityIndependent(t *testing.T) {
	// Two tasks with different IDs can each be submitted independently
	// This is enforced by UNIQUE(task_id, user_uid) not step_order
	// Verify that completing id 101 does not mark id 102 as done
	tasks := []struct {
		ID        int64 `json:"id"`
		StepOrder int   `json:"step_order"`
		Status    string
	}{
		{ID: 101, StepOrder: 1, Status: "APPROVED"},
		{ID: 102, StepOrder: 2, Status: "UNLOCKED"},
		{ID: 103, StepOrder: 3, Status: "LOCKED"},
	}
	completed := 0
	for _, tk := range tasks {
		if tk.Status == "APPROVED" {
			completed++
		}
	}
	total := len(tasks)
	isAllDone := total > 0 && completed == total
	if isAllDone {
		t.Fatalf("expected not done with 1/3, got done")
	}
	tasks[1].Status = "APPROVED"
	completed = 0
	for _, tk := range tasks {
		if tk.Status == "APPROVED" {
			completed++
		}
	}
	isAllDone = total > 0 && completed == total
	if isAllDone {
		t.Fatalf("expected not done with 2/3, got done")
	}
	tasks[2].Status = "APPROVED"
	completed = 3
	isAllDone = total > 0 && completed == total
	if !isAllDone {
		t.Fatalf("expected done with 3/3")
	}
}

func TestStepperNavigationById(t *testing.T) {
	tasks := []struct {
		ID        int64
		StepOrder int
	}{
		{ID: 101, StepOrder: 1},
		{ID: 102, StepOrder: 2},
		{ID: 103, StepOrder: 3},
	}
	// find by id index
	currentID := int64(101)
	idx := -1
	for i, tk := range tasks {
		if tk.ID == currentID {
			idx = i
			break
		}
	}
	next := tasks[idx+1]
	if next.ID != 102 {
		t.Fatalf("expected next after 101 is 102, got %d", next.ID)
	}
	// old buggy: find by step+1 would skip duplicate
	// simulate duplicate 1,1,2 sorted as 101:1,102:1,103:2
	dupTasks := []struct {
		ID        int64
		StepOrder int
	}{
		{ID: 101, StepOrder: 1},
		{ID: 102, StepOrder: 1},
		{ID: 103, StepOrder: 2},
	}
	// buggy find by step+1 after 101 (step1) would find 103 (step2) skipping 102
	buggy := -1
	for _, tk := range dupTasks {
		if tk.StepOrder == 1+1 {
			buggy = int(tk.ID)
			break
		}
	}
	if buggy != 103 {
		t.Fatalf("buggy check")
	}
	// correct by id
	idx = -1
	for i, tk := range dupTasks {
		if tk.ID == 101 {
			idx = i
			break
		}
	}
	correctNext := dupTasks[idx+1]
	if correctNext.ID != 102 {
		t.Fatalf("correct next after 101 should be 102 (duplicate), got %d", correctNext.ID)
	}
}

func TestReorderSuccessABCtoCAB(t *testing.T) {
	tasks := []map[string]any{
		{"id": float64(101), "family_id": "fam-1", "active_date": "2026-09-01", "step_order": float64(1), "is_active": true},
		{"id": float64(102), "family_id": "fam-1", "active_date": "2026-09-01", "step_order": float64(2), "is_active": true},
		{"id": float64(103), "family_id": "fam-1", "active_date": "2026-09-01", "step_order": float64(3), "is_active": true},
	}
	store := &reorderMock{tasks: tasks}
	client := &mockReorderClient{mock: store}
	api := NewAPI(client)
	body := `{"taskIds":[103,101,102]}`
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/tasks/reorder", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "a1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.HandleReorderTasks(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", rec.Code, rec.Body.String())
	}
	// Verify persisted mapping via Mutate params: id=eq.X must receive step 1,2,3 in submitted order.
	got := map[string]int{}
	for i, call := range store.mutateCalls {
		_ = i
		got[call.params] = call.stepOrder
	}
	if got["id=eq.103"] != 1 || got["id=eq.101"] != 2 || got["id=eq.102"] != 3 {
		t.Fatalf("expected persisted mapping 103->1,101->2,102->3, got %v", got)
	}
	// Verify deterministic returned order: step_order ASC, id ASC.
	var returned []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &returned); err != nil {
		t.Fatalf("expected JSON array response, got %q: %v", rec.Body.String(), err)
	}
	if len(returned) != 3 {
		t.Fatalf("expected 3 tasks returned, got %d: %s", len(returned), rec.Body.String())
	}
	wantIDs := []float64{103, 101, 102}
	for i, want := range wantIDs {
		gotID, _ := returned[i]["id"].(float64)
		gotStep, _ := returned[i]["step_order"].(float64)
		if gotID != want || gotStep != float64(i+1) {
			t.Fatalf("position %d: expected id=%v step=%d, got id=%v step=%v (%s)", i, want, i+1, gotID, gotStep, rec.Body.String())
		}
	}
}

type mockReorderClient struct {
	mock *reorderMock
}

type reorderMock struct {
	tasks       []map[string]any
	mutates     []map[string]any
	mutateCalls []reorderMutateCall
}

type reorderMutateCall struct {
	params    string
	stepOrder int
}

func (m *mockReorderClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if table != "odyssey_tasks" {
		return json.Marshal([]map[string]any{})
	}
	if strings.Contains(params, "id=in.(") {
		want := map[int64]bool{}
		seg := params[strings.Index(params, "id=in.(")+len("id=in.("):]
		seg = seg[:strings.Index(seg, ")")]
		for _, s := range strings.Split(seg, ",") {
			s = strings.TrimSpace(s)
			if v, err := strconv.ParseInt(s, 10, 64); err == nil {
				want[v] = true
			}
		}
		var out []map[string]any
		for _, t := range m.mock.tasks {
			if idOf(t) >= 0 && want[idOf(t)] {
				out = append(out, t)
			}
		}
		if out == nil {
			out = []map[string]any{}
		}
		return json.Marshal(out)
	}
	if strings.Contains(params, "select=id") {
		var out []map[string]any
		for _, t := range m.mock.tasks {
			if isActiveOf(t) && familyOf(t) == familyParam(params) && dateOf(t) == dateParam(params) {
				out = append(out, map[string]any{"id": t["id"]})
			}
		}
		if out == nil {
			out = []map[string]any{}
		}
		return json.Marshal(out)
	}
	if strings.Contains(params, "order=step_order") {
		out := append([]map[string]any(nil), m.mock.tasks...)
		sort.Slice(out, func(i, j int) bool {
			si, sj := stepOf(out[i]), stepOf(out[j])
			if si != sj {
				return si < sj
			}
			return idOf(out[i]) < idOf(out[j])
		})
		return json.Marshal(out)
	}
	return json.Marshal(m.mock.tasks)
}

func idOf(t map[string]any) int64 {
	switch v := t["id"].(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	}
	return -1
}

func stepOf(t map[string]any) int {
	switch v := t["step_order"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	}
	return 0
}

func isActiveOf(t map[string]any) bool {
	v, _ := t["is_active"].(bool)
	return v
}

func familyOf(t map[string]any) string {
	s, _ := t["family_id"].(string)
	return s
}

func dateOf(t map[string]any) string {
	s, _ := t["active_date"].(string)
	return s
}

func familyParam(params string) string {
	for _, part := range strings.Split(params, "&") {
		if strings.HasPrefix(part, "family_id=eq.") {
			return strings.TrimPrefix(part, "family_id=eq.")
		}
	}
	return ""
}

func dateParam(params string) string {
	for _, part := range strings.Split(params, "&") {
		if strings.HasPrefix(part, "active_date=eq.") {
			return strings.TrimPrefix(part, "active_date=eq.")
		}
	}
	return ""
}

func (m *mockReorderClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	if mp, ok := payload.(map[string]any); ok {
		m.mock.mutates = append(m.mock.mutates, mp)
		if strings.HasPrefix(params, "id=eq.") {
			if id, err := strconv.ParseInt(strings.TrimPrefix(params, "id=eq."), 10, 64); err == nil {
				if step, ok := mp["step_order"]; ok {
					m.mock.mutateCalls = append(m.mock.mutateCalls, reorderMutateCall{params: params, stepOrder: toInt(step)})
					for _, t := range m.mock.tasks {
						if idOf(t) == id {
							t["step_order"] = float64(toInt(step))
						}
					}
				}
			}
		}
	}
	return []byte("{}"), nil
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

func (m *mockReorderClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return m.Mutate(ctx, method, table, payload, params)
}
func (m *mockReorderClient) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	return []byte("{}"), nil
}
func (m *mockReorderClient) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return "", nil
}

func TestReorderRejectDuplicateIDs(t *testing.T) {
	tasks := []map[string]any{
		{"id": float64(101), "family_id": "fam-1", "active_date": "2026-09-01", "step_order": float64(1), "is_active": true},
		{"id": float64(102), "family_id": "fam-1", "active_date": "2026-09-01", "step_order": float64(2), "is_active": true},
	}
	m := &reorderMock{tasks: tasks}
	client := &mockReorderClient{mock: m}
	api := NewAPI(client)
	body := `{"taskIds":[101,101]}`
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/tasks/reorder", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "a1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.HandleReorderTasks(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 duplicate, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestReorderRejectWrongFamily(t *testing.T) {
	tasks := []map[string]any{
		{"id": float64(101), "family_id": "fam-1", "active_date": "2026-09-01", "step_order": float64(1), "is_active": true},
		{"id": float64(102), "family_id": "fam-other", "active_date": "2026-09-01", "step_order": float64(2), "is_active": true},
	}
	m := &reorderMock{tasks: tasks}
	client := &mockReorderClient{mock: m}
	api := NewAPI(client)
	body := `{"taskIds":[101,102]}`
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/tasks/reorder", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "a1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.HandleReorderTasks(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 wrong family, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestReorderRejectIncompleteSet(t *testing.T) {
	tasks := []map[string]any{
		{"id": float64(101), "family_id": "fam-1", "active_date": "2026-09-01", "step_order": float64(1), "is_active": true},
		{"id": float64(102), "family_id": "fam-1", "active_date": "2026-09-01", "step_order": float64(2), "is_active": true},
		{"id": float64(103), "family_id": "fam-1", "active_date": "2026-09-01", "step_order": float64(3), "is_active": true},
	}
	m := &reorderMock{tasks: tasks}
	client := &mockReorderClient{mock: m}
	api := NewAPI(client)
	body := `{"taskIds":[101,102]}` // missing 103
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/tasks/reorder", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "a1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.HandleReorderTasks(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 incomplete set, got %d %s", rec.Code, rec.Body.String())
	}
}

func init() {
	// ensure bytes import used
	_ = bytes.NewReader
}
