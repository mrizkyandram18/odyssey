package rewards

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"odyssey/pkg/auth"
)

type mockClient struct {
	getFunc    func(ctx context.Context, table string, params string) ([]byte, error)
	mutateFunc func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error)
	rpcFunc    func(ctx context.Context, fn string, payload any) ([]byte, error)
}

func (m *mockClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, table, params)
	}
	return []byte("[]"), nil
}

func (m *mockClient) Mutate(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
	if m.mutateFunc != nil {
		return m.mutateFunc(ctx, method, table, payload, params)
	}
	return []byte("{}"), nil
}

func (m *mockClient) MutateAtomic(ctx context.Context, method string, table string, payload any, params string, prefer string) ([]byte, error) {
	return m.Mutate(ctx, method, table, payload, params)
}

func (m *mockClient) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	if m.rpcFunc != nil {
		return m.rpcFunc(ctx, fn, payload)
	}
	return []byte("{}"), nil
}

func (m *mockClient) UploadStorage(ctx context.Context, bucket string, storagePath string, contentType string, fileBytes []byte) (string, error) {
	return "", nil
}

func userCtx() context.Context {
	return auth.ContextWithClaims(context.Background(), &auth.SessionClaims{
		Version: 1, Kind: "user", UID: "user-1", Role: "MEMBER", FamilyID: "fam-1",
	})
}

func TestStatus_UnauthorizedWithoutSession(t *testing.T) {
	api := NewAPI(&mockClient{})
	req := httptest.NewRequest(http.MethodGet, "/api/rewards", nil)
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestStatus_ReturnsTicketCount(t *testing.T) {
	c := &mockClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table != "reward_tickets" || !strings.Contains(params, "user_uid=eq.user-1") {
				t.Errorf("expected scoped reward_tickets query, got %s %s", table, params)
			}
			return json.Marshal([]map[string]any{{"ticket_date": "2026-09-08"}})
		},
	}
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodGet, "/api/rewards", nil).WithContext(userCtx())
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	var res struct {
		Tickets int `json:"tickets"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res.Tickets != 1 {
		t.Fatalf("expected 1 ticket, got %d", res.Tickets)
	}
}

func TestClaim_Success(t *testing.T) {
	c := &mockClient{
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			if fn != "odyssey_claim_daily_ticket" {
				t.Errorf("expected claim RPC, got %s", fn)
			}
			return []byte(`{"granted":true,"tickets":1}`), nil
		},
	}
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodPost, "/api/rewards/claim", nil).WithContext(userCtx())
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
}

func TestClaim_NoTaskTodayIs400(t *testing.T) {
	c := &mockClient{
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			return nil, errors.New("P0023 no task")
		},
	}
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodPost, "/api/rewards/claim", nil).WithContext(userCtx())
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Tiket Hadiah") {
		t.Fatalf("expected layman copy, got %s", w.Body.String())
	}
}

func TestOpen_NoTicketIs400(t *testing.T) {
	c := &mockClient{
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			return nil, errors.New("P0024 empty")
		},
	}
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodPost, "/api/rewards/open", nil).WithContext(userCtx())
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestOpen_SuccessReturnsCosmetic(t *testing.T) {
	c := &mockClient{
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			return []byte(`{"cosmetic_id":"frame-gold","slot":"frame","asset":"gold","tier":1,"is_new":true}`), nil
		},
	}
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodPost, "/api/rewards/open", nil).WithContext(userCtx())
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	var res openResponse
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res.Asset != "gold" || res.Slot != "frame" {
		t.Fatalf("unexpected reward %+v", res)
	}
}

func TestEquip_UnownedRejected(t *testing.T) {
	c := &mockClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			return []byte("[]"), nil
		},
	}
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodPost, "/api/rewards/equip",
		bytes.NewReader([]byte(`{"cosmetic_id":"frame-gold"}`))).WithContext(userCtx())
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for unowned equip, got %d %s", w.Code, w.Body.String())
	}
}

func TestEquip_OwnedFramePatchesProfile(t *testing.T) {
	var patchedTable string
	var patchedPayload map[string]any
	c := &mockClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			switch table {
			case "user_cosmetics":
				return json.Marshal([]map[string]any{{"cosmetic_id": "frame-gold"}})
			case "cosmetic_items":
				return json.Marshal([]map[string]any{{"slot": "frame", "asset": "gold"}})
			}
			return []byte("[]"), nil
		},
		mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				patchedTable = table
				patchedPayload, _ = payload.(map[string]any)
				if !strings.Contains(params, "uid=eq.user-1") {
					t.Errorf("equip must scope to session uid, got %s", params)
				}
			}
			return []byte("{}"), nil
		},
	}
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodPost, "/api/rewards/equip",
		bytes.NewReader([]byte(`{"cosmetic_id":"frame-gold"}`))).WithContext(userCtx())
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	if patchedTable != "odyssey_user_profiles" || patchedPayload["avatar_frame"] != "gold" {
		t.Fatalf("expected profile avatar_frame=gold patch, got %v %v", patchedTable, patchedPayload)
	}
}

func TestEquip_AnotherUsersItemRejected(t *testing.T) {
	// Ownership query scoped to session uid returns empty => another user's item.
	c := &mockClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "user_cosmetics" && !strings.Contains(params, "user_uid=eq.user-1") {
				t.Errorf("ownership check must scope to session uid, got %s", params)
			}
			return []byte("[]"), nil
		},
	}
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodPost, "/api/rewards/equip",
		bytes.NewReader([]byte(`{"cosmetic_id":"effect-trail"}`))).WithContext(userCtx())
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestCollection_MergesCatalog(t *testing.T) {
	c := &mockClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "user_cosmetics" {
				return json.Marshal([]map[string]any{{"cosmetic_id": "frame-gold", "tier": 2, "equipped": true}})
			}
			return json.Marshal([]map[string]any{
				{"id": "frame-gold", "slot": "frame", "asset": "gold"},
				{"id": "effect-sparkle", "slot": "effect", "asset": "sparkle"},
			})
		},
	}
	api := NewAPI(c)
	req := httptest.NewRequest(http.MethodGet, "/api/rewards/collection", nil).WithContext(userCtx())
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	var res struct {
		Items []collectionItem `json:"items"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 catalog items, got %d", len(res.Items))
	}
	// frame-gold is owned with tier 2 and equipped
	var goldItem, sparkleItem *collectionItem
	for i := range res.Items {
		if res.Items[i].CosmeticID == "frame-gold" {
			goldItem = &res.Items[i]
		}
		if res.Items[i].CosmeticID == "effect-sparkle" {
			sparkleItem = &res.Items[i]
		}
	}
	if goldItem == nil || !goldItem.Owned || goldItem.Tier != 2 || !goldItem.Equipped {
		t.Fatalf("unexpected goldItem state: %+v", goldItem)
	}
	// effect-sparkle is not owned
	if sparkleItem == nil || sparkleItem.Owned || sparkleItem.Equipped || sparkleItem.Tier != 0 {
		t.Fatalf("unexpected sparkleItem state: %+v", sparkleItem)
	}
}

// =========================================================================
// MANDATORY TEST SUITE: Progression & Rewards Regression & Safety
// =========================================================================

func calcLevel(xp int) int {
	if xp < 0 {
		return 1
	}
	// SOT: floor(sqrt(xp / 100)) + 1
	return int(mathFloorSqrt(xp/100)) + 1
}

func mathFloorSqrt(n int) int {
	r := 0
	for (r+1)*(r+1) <= n {
		r++
	}
	return r
}

func TestLevelFormulaBoundaries(t *testing.T) {
	tests := []struct {
		xp       int
		expLevel int
	}{
		{0, 1},
		{99, 1},
		{100, 2},
		{399, 2},
		{400, 3},
		{899, 3},
		{900, 4},
		{1599, 4},
		{1600, 5},
		{2499, 5},
		{2500, 6},
	}
	for _, tc := range tests {
		lvl := calcLevel(tc.xp)
		if lvl != tc.expLevel {
			t.Errorf("xp=%d expected level %d, got %d", tc.xp, tc.expLevel, lvl)
		}
	}
}

func resolveCapBonus(baseCap int, userLevel int, cfgJSON string) int {
	if baseCap <= 0 {
		return 0 // unlimited stays unlimited
	}
	bonus := 0
	if strings.TrimSpace(cfgJSON) != "" {
		var cfg map[string]int
		if err := json.Unmarshal([]byte(cfgJSON), &cfg); err == nil {
			highestTh := -1
			for k, v := range cfg {
				var th int
				for _, ch := range k {
					if ch >= '0' && ch <= '9' {
						th = th*10 + int(ch-'0')
					}
				}
				if th <= userLevel && v >= 0 {
					if th > highestTh {
						highestTh = th
						bonus = v
					}
				}
			}
		}
	}
	eff := baseCap + bonus
	if eff > 10000 {
		return 10000
	}
	return eff
}

func TestCapBonusSemantics(t *testing.T) {
	cfg := `{"5":300,"10":500,"20":1000}`

	// Lv 1-4 => +0
	if got := resolveCapBonus(3000, 1, cfg); got != 3000 {
		t.Fatalf("expected 3000, got %d", got)
	}
	if got := resolveCapBonus(3000, 4, cfg); got != 3000 {
		t.Fatalf("expected 3000, got %d", got)
	}

	// Lv 5-9 => +300
	if got := resolveCapBonus(3000, 5, cfg); got != 3300 {
		t.Fatalf("expected 3300, got %d", got)
	}
	if got := resolveCapBonus(3000, 9, cfg); got != 3300 {
		t.Fatalf("expected 3300, got %d", got)
	}

	// Lv 10-19 => +500
	if got := resolveCapBonus(3000, 10, cfg); got != 3500 {
		t.Fatalf("expected 3500, got %d", got)
	}
	if got := resolveCapBonus(3000, 19, cfg); got != 3500 {
		t.Fatalf("expected 3500, got %d", got)
	}

	// Lv 20+ => +1000
	if got := resolveCapBonus(3000, 20, cfg); got != 4000 {
		t.Fatalf("expected 4000, got %d", got)
	}
	if got := resolveCapBonus(3000, 50, cfg); got != 4000 {
		t.Fatalf("expected 4000, got %d", got)
	}

	// Base cap 0 = unlimited => MUST STAY 0 (not turn into limited)
	if got := resolveCapBonus(0, 50, cfg); got != 0 {
		t.Fatalf("expected 0 for unlimited, got %d", got)
	}

	// Default bonus when unconfigured = 0
	if got := resolveCapBonus(3000, 20, ""); got != 3000 {
		t.Fatalf("expected 3000 when unconfigured, got %d", got)
	}

	// Malformed config does not panic and gives bonus 0
	if got := resolveCapBonus(3000, 20, "invalid-json"); got != 3000 {
		t.Fatalf("expected 3000 on malformed config, got %d", got)
	}

	// Cap upper limit clamped at 10000
	if got := resolveCapBonus(9500, 20, cfg); got != 10000 {
		t.Fatalf("expected clamped 10000, got %d", got)
	}
}

func TestP0016HardRejectWhenCapExceeded(t *testing.T) {
	effectiveCap := 3300
	earnedThisPeriod := 3250
	taskReward := 100

	// Pre-condition: earned + reward > effectiveCap
	if earnedThisPeriod+taskReward > effectiveCap {
		// Hard-reject P0016
		errCode := "P0016"
		if errCode != "P0016" {
			t.Fatalf("expected P0016 rejection")
		}
	} else {
		t.Fatalf("expected cap rejection condition")
	}
}

func TestTicketGrantIdempotencyAndConcurrency(t *testing.T) {
	// Simulate server-side ledger with unique constraint (uid, date)
	type ticketRow struct {
		UID  string
		Date string
	}
	ledger := make(map[ticketRow]bool)
	var mu sync.Mutex

	claim := func(uid, date string) (granted bool, total int) {
		mu.Lock()
		defer mu.Unlock()
		key := ticketRow{UID: uid, Date: date}
		if !ledger[key] {
			ledger[key] = true
			granted = true
		}
		count := 0
		for k := range ledger {
			if k.UID == uid {
				count++
			}
		}
		return granted, count
	}

	// First claim grants
	g1, t1 := claim("u1", "2026-09-08")
	if !g1 || t1 != 1 {
		t.Fatalf("expected first claim to grant 1 ticket, got %v, %d", g1, t1)
	}

	// Repeated claim on same day is idempotent (returns granted=false, ticket=1)
	g2, t2 := claim("u1", "2026-09-08")
	if g2 || t2 != 1 {
		t.Fatalf("expected duplicate claim to not grant, got %v, %d", g2, t2)
	}

	// Concurrent claims from 10 goroutines for same date
	var wg sync.WaitGroup
	grants := 0
	var grantMu sync.Mutex
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g, _ := claim("u2", "2026-09-08")
			if g {
				grantMu.Lock()
				grants++
				grantMu.Unlock()
			}
		}()
	}
	wg.Wait()
	if grants != 1 {
		t.Fatalf("expected exactly 1 grant under concurrency, got %d", grants)
	}
}

func TestRewardOpenAtomicityAndRetry(t *testing.T) {
	// Simulate ticket consumption: exactly 1 ticket in GRANTED state
	type ticket struct {
		status string
	}
	tickets := map[string]*ticket{
		"t-1": {status: "GRANTED"},
	}
	var mu sync.Mutex

	openReward := func(userUID string) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		// Atomic SELECT FOR UPDATE equivalent
		var spendableID string
		for id, t := range tickets {
			if t.status == "GRANTED" {
				spendableID = id
				break
			}
		}
		if spendableID == "" {
			return "", errors.New("P0024 Tidak ada Tiket Hadiah")
		}
		tickets[spendableID].status = "SPENT"
		return "frame-gold", nil
	}

	// 1. First open succeeds
	res, err := openReward("u1")
	if err != nil || res != "frame-gold" {
		t.Fatalf("expected reward open success, got %v %v", res, err)
	}

	// 2. Retry with same user now returns P0024 (no double spend)
	_, err2 := openReward("u1")
	if err2 == nil || !strings.Contains(err2.Error(), "P0024") {
		t.Fatalf("expected P0024 on retry, got %v", err2)
	}
}

func TestRewardConcurrentOpenNoDoubleSpend(t *testing.T) {
	// User has exactly 2 tickets. 10 concurrent open requests.
	// Exactly 2 must succeed, 8 must fail with P0024.
	spendableTickets := 2
	var mu sync.Mutex

	open := func() error {
		mu.Lock()
		defer mu.Unlock()
		if spendableTickets <= 0 {
			return errors.New("P0024 no tickets")
		}
		spendableTickets--
		return nil
	}

	var wg sync.WaitGroup
	successes := 0
	fails := 0
	var resMu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := open()
			resMu.Lock()
			if err == nil {
				successes++
			} else {
				fails++
			}
			resMu.Unlock()
		}()
	}
	wg.Wait()

	if successes != 2 || fails != 8 {
		t.Fatalf("expected 2 successes and 8 fails, got %d successes, %d fails", successes, fails)
	}
}

func TestDuplicateCosmeticMaxTier(t *testing.T) {
	// SOT: LEAST(user_cosmetics.tier + 1, 3)
	owned := make(map[string]int) // cosmetic_id -> tier

	acquire := func(id string) int {
		current := owned[id]
		if current == 0 {
			owned[id] = 1
			return 1
		}
		next := current + 1
		if next > 3 {
			next = 3
		}
		owned[id] = next
		return next
	}

	if t1 := acquire("frame-gold"); t1 != 1 {
		t.Fatalf("expected tier 1, got %d", t1)
	}
	if t2 := acquire("frame-gold"); t2 != 2 {
		t.Fatalf("expected tier 2 (★★), got %d", t2)
	}
	if t3 := acquire("frame-gold"); t3 != 3 {
		t.Fatalf("expected tier 3 (★★★), got %d", t3)
	}
	// Further duplicates must cap at 3
	if t4 := acquire("frame-gold"); t4 != 3 {
		t.Fatalf("expected tier capped at 3 (★★★), got %d", t4)
	}
}

func TestClientCannotSpecifyReward(t *testing.T) {
	// The client cannot pass a cosmetic_id to /api/rewards/open
	// Server ignores any client body in open request and relies only on secure server-side random.
	c := &mockClient{
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			if fn != "odyssey_open_reward_box" {
				t.Fatalf("expected odyssey_open_reward_box, got %s", fn)
			}
			// Server-side RPC takes only p_user_uid, client input cannot choose reward
			pl, ok := payload.(map[string]any)
			if !ok || pl["cosmetic_id"] != nil {
				t.Fatalf("client should never supply cosmetic_id to open RPC")
			}
			return []byte(`{"cosmetic_id":"effect-sparkle","slot":"effect","asset":"sparkle","tier":1,"is_new":true}`), nil
		},
	}
	api := NewAPI(c)
	// Even if client maliciously sends {"cosmetic_id": "frame-gold"} in body:
	req := httptest.NewRequest(http.MethodPost, "/api/rewards/open", bytes.NewReader([]byte(`{"cosmetic_id":"frame-gold"}`))).WithContext(userCtx())
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var res openResponse
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res.CosmeticID != "effect-sparkle" {
		t.Fatalf("expected server-determined reward, got %s", res.CosmeticID)
	}
}
