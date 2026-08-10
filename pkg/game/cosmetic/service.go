package cosmetic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"odyssey/pkg/game"
)

var (
	ErrUnknownCosmetic   = errors.New("unknown cosmetic")
	ErrInsufficientCoins = errors.New("insufficient coins")
	ErrConcurrent        = errors.New("concurrent modification")
)

// UnlockStore persists cosmetic ownership.
type UnlockStore interface {
	ListByUser(ctx context.Context, uid string) ([]game.CosmeticUnlock, error)
	Has(ctx context.Context, uid, cosmeticID string) (bool, error)
	// CreateIfAbsent inserts ownership; returns created=false if already owned.
	CreateIfAbsent(ctx context.Context, uid, cosmeticID string, pricePaid int64) (created bool, err error)
	Delete(ctx context.Context, uid, cosmeticID string) error
}

// FrameStore updates the equipped avatar frame on a profile.
type FrameStore interface {
	SetAvatarFrame(ctx context.Context, uid, frame string) error
}

// Service handles listing and purchasing cosmetics.
type Service struct {
	users   game.UserStore
	ledgers game.RewardLedgerStore
	unlocks UnlockStore
	frames  FrameStore
}

func NewService(users game.UserStore, ledgers game.RewardLedgerStore, unlocks UnlockStore, frames FrameStore) *Service {
	return &Service{users: users, ledgers: ledgers, unlocks: unlocks, frames: frames}
}

// CatalogItemView is a catalog entry with ownership for the caller.
type CatalogItemView struct {
	Item
	Unlocked bool `json:"unlocked"`
}

// ListResult is returned by ListForUser.
type ListResult struct {
	Coins int64             `json:"coins"`
	Items []CatalogItemView `json:"items"`
	Frame string            `json:"avatar_frame"`
}

// PurchaseResult is returned by Purchase.
type PurchaseResult struct {
	Status       string `json:"status"` // purchased | already_owned
	CosmeticID   string `json:"cosmetic_id"`
	Price        int64  `json:"price"`
	Coins        int64  `json:"coins"`
	AvatarFrame  string `json:"avatar_frame"`
	AlreadyOwned bool   `json:"already_owned"`
}

// ListForUser returns the fixed catalog with unlock flags and current balance.
func (s *Service) ListForUser(ctx context.Context, uid string) (*ListResult, error) {
	player, err := s.users.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	owned, err := s.unlocks.ListByUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("list unlocks: %w", err)
	}
	ownedSet := map[string]bool{}
	for _, u := range owned {
		ownedSet[u.CosmeticID] = true
	}
	items := make([]CatalogItemView, 0, len(Catalog))
	for _, it := range Catalog {
		items = append(items, CatalogItemView{Item: it, Unlocked: ownedSet[it.ID]})
	}
	frame := FrameNone
	// Prefer equipped frame from unlock of gold if we only equip via purchase.
	if ownedSet[CosmeticAvatarFrameGold] {
		frame = FrameGold
	}
	return &ListResult{Coins: player.Coins, Items: items, Frame: frame}, nil
}

// Purchase buys a cosmetic with coins. Safe against double-click:
// unique (uid, cosmetic_id) + balance checks + compensate unlock if debit fails.
func (s *Service) Purchase(ctx context.Context, uid, cosmeticID string) (*PurchaseResult, error) {
	item, ok := Lookup(cosmeticID)
	if !ok {
		return nil, ErrUnknownCosmetic
	}

	// Fast path: already owned.
	has, err := s.unlocks.Has(ctx, uid, cosmeticID)
	if err != nil {
		return nil, fmt.Errorf("check ownership: %w", err)
	}
	if has {
		player, err := s.users.GetUser(ctx, uid)
		if err != nil {
			return nil, err
		}
		return &PurchaseResult{
			Status:       "already_owned",
			CosmeticID:   item.ID,
			Price:        item.Price,
			Coins:        player.Coins,
			AvatarFrame:  item.Value,
			AlreadyOwned: true,
		}, nil
	}

	player, err := s.users.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if player.Coins < item.Price {
		return nil, ErrInsufficientCoins
	}

	// Claim ownership first (unique constraint = retry-safe).
	created, err := s.unlocks.CreateIfAbsent(ctx, uid, item.ID, item.Price)
	if err != nil {
		return nil, fmt.Errorf("create unlock: %w", err)
	}
	if !created {
		// Lost race to another request — no charge.
		player, err = s.users.GetUser(ctx, uid)
		if err != nil {
			return nil, err
		}
		return &PurchaseResult{
			Status:       "already_owned",
			CosmeticID:   item.ID,
			Price:        item.Price,
			Coins:        player.Coins,
			AvatarFrame:  item.Value,
			AlreadyOwned: true,
		}, nil
	}

	// Re-read balance under ownership claim; compensate unlock if insufficient/race.
	player, err = s.users.GetUser(ctx, uid)
	if err != nil {
		_ = s.unlocks.Delete(ctx, uid, item.ID)
		return nil, fmt.Errorf("get user for debit: %w", err)
	}
	if player.Coins < item.Price {
		_ = s.unlocks.Delete(ctx, uid, item.ID)
		return nil, ErrInsufficientCoins
	}

	newCoins := player.Coins - item.Price
	okMatch, err := s.users.UpdateUserIfMatch(ctx, uid, player.Version, map[string]any{
		"coins":   newCoins,
		"version": player.Version + 1,
	})
	if err != nil {
		_ = s.unlocks.Delete(ctx, uid, item.ID)
		return nil, fmt.Errorf("debit coins: %w", err)
	}
	if !okMatch {
		_ = s.unlocks.Delete(ctx, uid, item.ID)
		return nil, ErrConcurrent
	}

	// Equip frame (best-effort after payment; ownership is source of truth).
	if s.frames != nil && item.Kind == "avatar_frame" {
		_ = s.frames.SetAvatarFrame(ctx, uid, item.Value)
	}

	// Spend ledger (negative amount).
	meta, _ := json.Marshal(map[string]any{"cosmetic_id": item.ID, "kind": item.Kind})
	metaStr := string(meta)
	_ = s.ledgers.CreateLedger(ctx, &game.RewardLedger{
		ID:         uuid.New().String(),
		UserID:     uid,
		Source:     SourceCosmeticPurchase,
		Amount:     -item.Price,
		RewardType: RewardTypeCoins,
		Metadata:   &metaStr,
		CreatedAt:  time.Now().UTC(),
	})

	return &PurchaseResult{
		Status:       "purchased",
		CosmeticID:   item.ID,
		Price:        item.Price,
		Coins:        newCoins,
		AvatarFrame:  item.Value,
		AlreadyOwned: false,
	}, nil
}
