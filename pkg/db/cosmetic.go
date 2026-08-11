package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"odyssey/pkg/game"
)

type cosmeticUnlockStore struct {
	client SupabaseClient
}

func NewCosmeticUnlockStore(client SupabaseClient) *cosmeticUnlockStore {
	return &cosmeticUnlockStore{client: client}
}

type cosmeticUnlockRow struct {
	ID          int64     `json:"id"`
	UID         string    `json:"uid"`
	CosmeticID  string    `json:"cosmetic_id"`
	PricePaid   int64     `json:"price_paid"`
	PurchasedAt time.Time `json:"purchased_at"`
}

func (s *cosmeticUnlockStore) ListByUser(ctx context.Context, uid string) ([]game.CosmeticUnlock, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("order", "purchased_at.desc")
	raw, err := s.client.Get(ctx, "odyssey_cosmetic_unlocks", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("list cosmetic unlocks: %w", err)
	}
	var rows []cosmeticUnlockRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse cosmetic unlocks: %w", err)
	}
	out := make([]game.CosmeticUnlock, len(rows))
	for i, r := range rows {
		out[i] = game.CosmeticUnlock{
			ID:          r.ID,
			UID:         r.UID,
			CosmeticID:  r.CosmeticID,
			PricePaid:   r.PricePaid,
			PurchasedAt: r.PurchasedAt,
		}
	}
	return out, nil
}

func (s *cosmeticUnlockStore) Has(ctx context.Context, uid, cosmeticID string) (bool, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("cosmetic_id", "eq."+cosmeticID)
	v.Set("select", "id")
	v.Set("limit", "1")
	raw, err := s.client.Get(ctx, "odyssey_cosmetic_unlocks", v.Encode())
	if err != nil {
		return false, fmt.Errorf("has cosmetic unlock: %w", err)
	}
	var rows []struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return false, fmt.Errorf("parse has cosmetic unlock: %w", err)
	}
	return len(rows) > 0, nil
}

func (s *cosmeticUnlockStore) CreateIfAbsent(ctx context.Context, uid, cosmeticID string, pricePaid int64) (bool, error) {
	payload := map[string]any{
		"uid":         uid,
		"cosmetic_id": cosmeticID,
		"price_paid":  pricePaid,
	}
	// ignore-duplicates + representation: empty body means already owned.
	raw, err := s.client.MutateAtomic(ctx, "POST", "odyssey_cosmetic_unlocks", payload,
		"on_conflict=uid,cosmetic_id",
		"return=representation,resolution=ignore-duplicates")
	if err != nil {
		// Unique violation fallback if prefer resolution not applied.
		if strings.Contains(err.Error(), "409") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return false, nil
		}
		return false, fmt.Errorf("create cosmetic unlock: %w", err)
	}
	var rows []cosmeticUnlockRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return false, fmt.Errorf("parse create cosmetic unlock: %w", err)
	}
	return len(rows) > 0, nil
}

func (s *cosmeticUnlockStore) Delete(ctx context.Context, uid, cosmeticID string) error {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("cosmetic_id", "eq."+cosmeticID)
	_, err := s.client.Mutate(ctx, "DELETE", "odyssey_cosmetic_unlocks", nil, v.Encode())
	if err != nil {
		return fmt.Errorf("delete cosmetic unlock: %w", err)
	}
	return nil
}

// SetAvatarFrame equips a frame value on the user profile.
func (s *supabaseProfileStore) SetAvatarFrame(ctx context.Context, uid, frame string) error {
	payload := map[string]string{"avatar_frame": frame}
	params := s.buildFilter(uid)
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_user_profiles", payload, params)
	if err != nil {
		return fmt.Errorf("set avatar frame: %w", err)
	}
	return nil
}

// SetExplorerEffect equips an explorer effect value on the user profile.
func (s *supabaseProfileStore) SetExplorerEffect(ctx context.Context, uid, effect string) error {
	payload := map[string]string{"equipped_explorer_effect": effect}
	params := s.buildFilter(uid)
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_user_profiles", payload, params)
	if err != nil {
		return fmt.Errorf("set explorer effect: %w", err)
	}
	return nil
}
