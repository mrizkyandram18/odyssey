package db

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type mockPushSupabaseClient struct {
	getErr      error
	mutateErr   error
	getResp     []byte
	mutateResp  []byte
	lastMethod  string
	lastTable   string
	lastParams  string
	lastPayload any
}

func (m *mockPushSupabaseClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	m.lastTable = table
	m.lastParams = params
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getResp, nil
}

func (m *mockPushSupabaseClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	m.lastMethod = method
	m.lastTable = table
	m.lastPayload = payload
	m.lastParams = params
	if m.mutateErr != nil {
		return nil, m.mutateErr
	}
	return m.mutateResp, nil
}

func (m *mockPushSupabaseClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	m.lastMethod = method
	m.lastTable = table
	m.lastPayload = payload
	m.lastParams = params
	if m.mutateErr != nil {
		return nil, m.mutateErr
	}
	return m.mutateResp, nil
}

func TestPushSubscriptionStore_Upsert(t *testing.T) {
	client := &mockPushSupabaseClient{}
	store := NewPushSubscriptionStore(client)

	subRow := pushSubscriptionRow{
		ID:        1,
		UID:       "u1",
		Endpoint:  "https://push.example.com/1",
		P256dh:    "p256",
		Auth:      "auth",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	client.mutateResp, _ = json.Marshal([]pushSubscriptionRow{subRow})

	sub := &game.PushSubscription{
		UID:      "u1",
		Endpoint: "https://push.example.com/1",
		P256dh:   "p256",
		Auth:     "auth",
	}

	result, err := store.UpsertSubscription(context.Background(), sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 || result.UID != "u1" || result.Endpoint != "https://push.example.com/1" {
		t.Errorf("unexpected result: %+v", result)
	}

	if client.lastTable != "odyssey_push_subscriptions" {
		t.Errorf("expected table odyssey_push_subscriptions, got %s", client.lastTable)
	}
}

func TestPushSubscriptionStore_ListByUID(t *testing.T) {
	client := &mockPushSupabaseClient{}
	store := NewPushSubscriptionStore(client)

	rows := []pushSubscriptionRow{
		{ID: 1, UID: "u1", Endpoint: "https://push.example.com/1", P256dh: "p1", Auth: "a1"},
		{ID: 2, UID: "u1", Endpoint: "https://push.example.com/2", P256dh: "p2", Auth: "a2"},
	}
	client.getResp, _ = json.Marshal(rows)

	list, err := store.ListSubscriptionsByUID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", len(list))
	}
	if !strings.Contains(client.lastParams, "uid=eq.u1") {
		t.Errorf("expected params to contain uid=eq.u1, got %s", client.lastParams)
	}
}

func TestPushSubscriptionStore_Delete(t *testing.T) {
	client := &mockPushSupabaseClient{}
	store := NewPushSubscriptionStore(client)

	err := store.DeleteSubscription(context.Background(), "u1", "https://push.example.com/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.lastMethod != "DELETE" || client.lastTable != "odyssey_push_subscriptions" {
		t.Errorf("unexpected method or table: %s %s", client.lastMethod, client.lastTable)
	}
	if !strings.Contains(client.lastParams, "endpoint=eq.https%3A%2F%2Fpush.example.com%2F1") && !strings.Contains(client.lastParams, "endpoint=") {
		t.Errorf("expected endpoint filter in params: %s", client.lastParams)
	}
}
