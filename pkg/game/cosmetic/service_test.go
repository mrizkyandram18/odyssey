package cosmetic

import (
	"context"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type mockUsers struct {
	player    *game.Player
	failMatch bool
}

func (m *mockUsers) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	if m.player == nil || m.player.UID != uid {
		return nil, game.ErrNotFound
	}
	p := *m.player
	return &p, nil
}
func (m *mockUsers) CreateUser(ctx context.Context, p *game.Player) error { return nil }
func (m *mockUsers) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	return nil
}
func (m *mockUsers) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	if m.failMatch {
		return false, nil
	}
	if m.player == nil || m.player.UID != uid || m.player.Version != version {
		return false, nil
	}
	if c, ok := patch["coins"].(int64); ok {
		m.player.Coins = c
	}
	if v, ok := patch["version"].(int); ok {
		m.player.Version = v
	}
	return true, nil
}
func (m *mockUsers) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	return nil, nil
}

type mockLedgers struct {
	rows []game.RewardLedger
}

func (m *mockLedgers) CreateLedger(ctx context.Context, ledger *game.RewardLedger) error {
	m.rows = append(m.rows, *ledger)
	return nil
}
func (m *mockLedgers) ListByUser(ctx context.Context, userID string) ([]game.RewardLedger, error) {
	out := []game.RewardLedger{}
	for _, r := range m.rows {
		if r.UserID == userID {
			out = append(out, r)
		}
	}
	return out, nil
}

type mockUnlocks struct {
	owned map[string]bool // key uid|cosmetic
}

func key(uid, id string) string { return uid + "|" + id }

func (m *mockUnlocks) ListByUser(ctx context.Context, uid string) ([]game.CosmeticUnlock, error) {
	var out []game.CosmeticUnlock
	for k, ok := range m.owned {
		if !ok {
			continue
		}
		const sep = "|"
		i := -1
		for j := 0; j < len(k); j++ {
			if k[j] == sep[0] {
				i = j
				break
			}
		}
		if i < 0 {
			continue
		}
		u, id := k[:i], k[i+1:]
		if u == uid {
			out = append(out, game.CosmeticUnlock{UID: u, CosmeticID: id})
		}
	}
	return out, nil
}
func (m *mockUnlocks) Has(ctx context.Context, uid, cosmeticID string) (bool, error) {
	return m.owned[key(uid, cosmeticID)], nil
}
func (m *mockUnlocks) CreateIfAbsent(ctx context.Context, uid, cosmeticID string, pricePaid int64) (bool, error) {
	k := key(uid, cosmeticID)
	if m.owned[k] {
		return false, nil
	}
	if m.owned == nil {
		m.owned = map[string]bool{}
	}
	m.owned[k] = true
	return true, nil
}
func (m *mockUnlocks) Delete(ctx context.Context, uid, cosmeticID string) error {
	delete(m.owned, key(uid, cosmeticID))
	return nil
}

type mockFrames struct {
	frame string
}

func (m *mockFrames) SetAvatarFrame(ctx context.Context, uid, frame string) error {
	m.frame = frame
	return nil
}

func newSvc(coins int64, unlocks *mockUnlocks) (*Service, *mockUsers, *mockLedgers, *mockFrames) {
	users := &mockUsers{player: &game.Player{UID: "u1", Coins: coins, Version: 1}}
	ledgers := &mockLedgers{}
	frames := &mockFrames{}
	if unlocks == nil {
		unlocks = &mockUnlocks{owned: map[string]bool{}}
	}
	return NewService(users, ledgers, unlocks, frames), users, ledgers, frames
}

func TestPurchase_Success(t *testing.T) {
	svc, users, ledgers, frames := newSvc(10, nil)
	res, err := svc.Purchase(context.Background(), "u1", CosmeticAvatarFrameGold)
	if err != nil {
		t.Fatalf("Purchase: %v", err)
	}
	if res.Status != "purchased" || res.AlreadyOwned {
		t.Fatalf("unexpected result: %+v", res)
	}
	if users.player.Coins != 7 {
		t.Fatalf("expected coins 7 (10-3), got %d", users.player.Coins)
	}
	if res.Coins != 7 {
		t.Fatalf("result coins %d", res.Coins)
	}
	if frames.frame != FrameGold {
		t.Fatalf("expected gold frame equipped, got %s", frames.frame)
	}
	if len(ledgers.rows) != 1 || ledgers.rows[0].Amount != -PriceAvatarFrameGold {
		t.Fatalf("expected spend ledger -3, got %+v", ledgers.rows)
	}
	if ledgers.rows[0].Source != SourceCosmeticPurchase {
		t.Fatalf("ledger source %s", ledgers.rows[0].Source)
	}
}

func TestPurchase_InsufficientCoins(t *testing.T) {
	svc, users, ledgers, _ := newSvc(2, nil)
	_, err := svc.Purchase(context.Background(), "u1", CosmeticAvatarFrameGold)
	if err != ErrInsufficientCoins {
		t.Fatalf("expected ErrInsufficientCoins, got %v", err)
	}
	if users.player.Coins != 2 {
		t.Fatalf("coins should be unchanged, got %d", users.player.Coins)
	}
	if len(ledgers.rows) != 0 {
		t.Fatalf("no ledger on failed purchase")
	}
}

func TestPurchase_AlreadyUnlocked(t *testing.T) {
	unlocks := &mockUnlocks{owned: map[string]bool{key("u1", CosmeticAvatarFrameGold): true}}
	svc, users, ledgers, _ := newSvc(10, unlocks)
	res, err := svc.Purchase(context.Background(), "u1", CosmeticAvatarFrameGold)
	if err != nil {
		t.Fatalf("Purchase: %v", err)
	}
	if !res.AlreadyOwned || res.Status != "already_owned" {
		t.Fatalf("expected already_owned, got %+v", res)
	}
	if users.player.Coins != 10 {
		t.Fatalf("no charge on already owned, got %d", users.player.Coins)
	}
	if len(ledgers.rows) != 0 {
		t.Fatalf("no ledger on already owned")
	}
}

func TestPurchase_RetryDoesNotDoubleCharge(t *testing.T) {
	svc, users, ledgers, _ := newSvc(10, nil)
	if _, err := svc.Purchase(context.Background(), "u1", CosmeticAvatarFrameGold); err != nil {
		t.Fatalf("first: %v", err)
	}
	res, err := svc.Purchase(context.Background(), "u1", CosmeticAvatarFrameGold)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if !res.AlreadyOwned {
		t.Fatal("second should be already_owned")
	}
	if users.player.Coins != 7 {
		t.Fatalf("expected single debit to 7, got %d", users.player.Coins)
	}
	if len(ledgers.rows) != 1 {
		t.Fatalf("expected one ledger entry, got %d", len(ledgers.rows))
	}
}

func TestPurchase_UnknownCosmetic(t *testing.T) {
	svc, _, _, _ := newSvc(10, nil)
	_, err := svc.Purchase(context.Background(), "u1", "nope")
	if err != ErrUnknownCosmetic {
		t.Fatalf("expected unknown, got %v", err)
	}
}

func TestListForUser_ShowsUnlockAndBalance(t *testing.T) {
	unlocks := &mockUnlocks{owned: map[string]bool{}}
	svc, _, _, _ := newSvc(6, unlocks)
	list, err := svc.ListForUser(context.Background(), "u1")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if list.Coins != 6 || len(list.Items) != 1 || list.Items[0].Unlocked {
		t.Fatalf("unexpected list: %+v", list)
	}
	// after purchase
	if _, err := svc.Purchase(context.Background(), "u1", CosmeticAvatarFrameGold); err != nil {
		t.Fatalf("purchase: %v", err)
	}
	list, _ = svc.ListForUser(context.Background(), "u1")
	if list.Coins != 3 || !list.Items[0].Unlocked || list.Frame != FrameGold {
		t.Fatalf("after purchase list: %+v", list)
	}
}

func TestPurchase_LedgerBalanceConsistency(t *testing.T) {
	svc, users, ledgers, _ := newSvc(5, nil)
	before := users.player.Coins
	_, err := svc.Purchase(context.Background(), "u1", CosmeticAvatarFrameGold)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var ledgerSum int64
	for _, r := range ledgers.rows {
		ledgerSum += r.Amount
	}
	if before+ledgerSum != users.player.Coins {
		t.Fatalf("balance inconsistency: before=%d ledgerSum=%d coins=%d", before, ledgerSum, users.player.Coins)
	}
	_ = time.Now()
}
