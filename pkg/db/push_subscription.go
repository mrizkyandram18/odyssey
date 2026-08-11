package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"odyssey/pkg/game"
)

type pushSubscriptionStore struct {
	client SupabaseClient
}

func NewPushSubscriptionStore(client SupabaseClient) game.PushSubscriptionStore {
	return &pushSubscriptionStore{client: client}
}

type pushSubscriptionRow struct {
	ID        int64     `json:"id"`
	UID       string    `json:"uid"`
	Endpoint  string    `json:"endpoint"`
	P256dh    string    `json:"p256dh"`
	Auth      string    `json:"auth"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *pushSubscriptionRow) toDomain() *game.PushSubscription {
	if r == nil {
		return nil
	}
	return &game.PushSubscription{
		ID:        r.ID,
		UID:       r.UID,
		Endpoint:  r.Endpoint,
		P256dh:    r.P256dh,
		Auth:      r.Auth,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func (s *pushSubscriptionStore) UpsertSubscription(ctx context.Context, sub *game.PushSubscription) (*game.PushSubscription, error) {
	if sub == nil || sub.UID == "" || sub.Endpoint == "" {
		return nil, fmt.Errorf("invalid subscription payload")
	}
	payload := map[string]any{
		"uid":        sub.UID,
		"endpoint":   sub.Endpoint,
		"p256dh":     sub.P256dh,
		"auth":       sub.Auth,
		"updated_at": time.Now().UTC(),
	}
	raw, err := s.client.MutateAtomic(ctx, "POST", "odyssey_push_subscriptions", payload,
		"on_conflict=uid,endpoint",
		"return=representation,resolution=merge-duplicates")
	if err != nil {
		return nil, fmt.Errorf("upsert push subscription: %w", err)
	}
	var rows []pushSubscriptionRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse push subscription: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("empty response on push subscription upsert")
	}
	return rows[0].toDomain(), nil
}

func (s *pushSubscriptionStore) ListSubscriptionsByUID(ctx context.Context, uid string) ([]game.PushSubscription, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("order", "created_at.desc")
	raw, err := s.client.Get(ctx, "odyssey_push_subscriptions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("list push subscriptions: %w", err)
	}
	var rows []pushSubscriptionRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse push subscriptions: %w", err)
	}
	out := make([]game.PushSubscription, len(rows))
	for i, r := range rows {
		out[i] = *r.toDomain()
	}
	return out, nil
}

func (s *pushSubscriptionStore) DeleteSubscription(ctx context.Context, uid string, endpoint string) error {
	if uid == "" {
		return fmt.Errorf("empty uid")
	}
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	if endpoint != "" {
		v.Set("endpoint", "eq."+endpoint)
	}
	_, err := s.client.Mutate(ctx, "DELETE", "odyssey_push_subscriptions", nil, v.Encode())
	if err != nil {
		return fmt.Errorf("delete push subscription: %w", err)
	}
	return nil
}
