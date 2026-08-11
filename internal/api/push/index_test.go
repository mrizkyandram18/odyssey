package push

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
)

type mockStore struct {
	mu            sync.Mutex
	subscriptions map[string][]game.PushSubscription // key: uid
}

func newMockStore() *mockStore {
	return &mockStore{
		subscriptions: make(map[string][]game.PushSubscription),
	}
}

func (m *mockStore) UpsertSubscription(ctx context.Context, sub *game.PushSubscription) (*game.PushSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	subs := m.subscriptions[sub.UID]
	for i, existing := range subs {
		if existing.Endpoint == sub.Endpoint {
			subs[i] = *sub
			m.subscriptions[sub.UID] = subs
			return sub, nil
		}
	}
	sub.ID = int64(len(subs) + 1)
	m.subscriptions[sub.UID] = append(subs, *sub)
	return sub, nil
}

func (m *mockStore) ListSubscriptionsByUID(ctx context.Context, uid string) ([]game.PushSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.subscriptions[uid], nil
}

func (m *mockStore) DeleteSubscription(ctx context.Context, uid string, endpoint string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	subs := m.subscriptions[uid]
	if endpoint == "" {
		delete(m.subscriptions, uid)
		return nil
	}
	filtered := make([]game.PushSubscription, 0, len(subs))
	for _, s := range subs {
		if s.Endpoint != endpoint {
			filtered = append(filtered, s)
		}
	}
	m.subscriptions[uid] = filtered
	return nil
}

func withAuth(req *http.Request, uid string) *http.Request {
	ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID:    uid,
		CrewID: "crew-1",
		Role:   "SEEKER",
	})
	return req.WithContext(ctx)
}

func TestSubscribe_Success(t *testing.T) {
	store := newMockStore()
	Setup(store)

	body := map[string]any{
		"endpoint": "https://push.example.com/sub/123",
		"keys": map[string]string{
			"p256dh": "mock-p256dh-key",
			"auth":   "mock-auth-secret",
		},
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/push/subscribe", bytes.NewReader(data))
	req = withAuth(req, "user-1")
	w := httptest.NewRecorder()

	Handler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	subs, _ := store.ListSubscriptionsByUID(context.Background(), "user-1")
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
	if subs[0].Endpoint != "https://push.example.com/sub/123" {
		t.Errorf("unexpected endpoint: %s", subs[0].Endpoint)
	}
}

func TestSubscribe_IdempotentDuplicate(t *testing.T) {
	store := newMockStore()
	Setup(store)

	body := map[string]any{
		"endpoint": "https://push.example.com/sub/123",
		"keys": map[string]string{
			"p256dh": "mock-p256dh-key",
			"auth":   "mock-auth-secret",
		},
	}
	data, _ := json.Marshal(body)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/push/subscribe", bytes.NewReader(data))
		req = withAuth(req, "user-1")
		w := httptest.NewRecorder()
		Handler(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("iteration %d: expected 201, got %d", i, w.Code)
		}
	}

	subs, _ := store.ListSubscriptionsByUID(context.Background(), "user-1")
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription after duplicate post, got %d", len(subs))
	}
}

func TestSubscribe_Unauthenticated(t *testing.T) {
	store := newMockStore()
	Setup(store)

	body := map[string]any{
		"endpoint": "https://push.example.com/sub/123",
		"keys": map[string]string{
			"p256dh": "mock-p256dh-key",
			"auth":   "mock-auth-secret",
		},
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/push/subscribe", bytes.NewReader(data))
	w := httptest.NewRecorder()

	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestSubscribe_InvalidPayload(t *testing.T) {
	store := newMockStore()
	Setup(store)

	tests := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing endpoint",
			body: map[string]any{
				"keys": map[string]string{"p256dh": "k", "auth": "a"},
			},
		},
		{
			name: "non-http endpoint",
			body: map[string]any{
				"endpoint": "http://insecure.example.com/push",
				"keys":     map[string]string{"p256dh": "k", "auth": "a"},
			},
		},
		{
			name: "missing p256dh",
			body: map[string]any{
				"endpoint": "https://push.example.com/sub/1",
				"keys":     map[string]string{"auth": "a"},
			},
		},
		{
			name: "missing auth",
			body: map[string]any{
				"endpoint": "https://push.example.com/sub/1",
				"keys":     map[string]string{"p256dh": "k"},
			},
		},
		{
			name: "oversized endpoint",
			body: map[string]any{
				"endpoint": "https://push.example.com/" + strings.Repeat("a", 2050),
				"keys":     map[string]string{"p256dh": "k", "auth": "a"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/push/subscribe", bytes.NewReader(data))
			req = withAuth(req, "user-1")
			w := httptest.NewRecorder()

			Handler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 Bad Request, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestSubscribe_OwnershipEnforced(t *testing.T) {
	store := newMockStore()
	Setup(store)

	// User body trying to specify a different uid is ignored since claims.UID is used.
	body := map[string]any{
		"uid":      "victim-user",
		"endpoint": "https://push.example.com/sub/attacker",
		"keys": map[string]string{
			"p256dh": "k",
			"auth":   "a",
		},
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/push/subscribe", bytes.NewReader(data))
	req = withAuth(req, "attacker-user")
	w := httptest.NewRecorder()

	Handler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// Verify it was stored under attacker-user, NOT victim-user
	attackerSubs, _ := store.ListSubscriptionsByUID(context.Background(), "attacker-user")
	if len(attackerSubs) != 1 {
		t.Fatalf("expected subscription under attacker-user")
	}
	victimSubs, _ := store.ListSubscriptionsByUID(context.Background(), "victim-user")
	if len(victimSubs) != 0 {
		t.Fatalf("cross-user injection occurred! Found subscription under victim-user")
	}
}

func TestUnsubscribe_Success(t *testing.T) {
	store := newMockStore()
	Setup(store)

	store.UpsertSubscription(context.Background(), &game.PushSubscription{
		UID:      "user-1",
		Endpoint: "https://push.example.com/sub/1",
		P256dh:   "k1",
		Auth:     "a1",
	})
	store.UpsertSubscription(context.Background(), &game.PushSubscription{
		UID:      "user-1",
		Endpoint: "https://push.example.com/sub/2",
		P256dh:   "k2",
		Auth:     "a2",
	})

	// Delete specific endpoint
	body := map[string]string{"endpoint": "https://push.example.com/sub/1"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodDelete, "/api/push/subscribe", bytes.NewReader(data))
	req = withAuth(req, "user-1")
	w := httptest.NewRecorder()

	Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	remaining, _ := store.ListSubscriptionsByUID(context.Background(), "user-1")
	if len(remaining) != 1 || remaining[0].Endpoint != "https://push.example.com/sub/2" {
		t.Fatalf("expected only sub/2 remaining, got %v", remaining)
	}
}

func TestUnsubscribe_CannotDeleteOtherUserSubscription(t *testing.T) {
	store := newMockStore()
	Setup(store)

	store.UpsertSubscription(context.Background(), &game.PushSubscription{
		UID:      "victim-user",
		Endpoint: "https://push.example.com/sub/victim",
		P256dh:   "k",
		Auth:     "a",
	})

	// Attacker tries to delete victim's endpoint
	body := map[string]string{"endpoint": "https://push.example.com/sub/victim"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodDelete, "/api/push/subscribe", bytes.NewReader(data))
	req = withAuth(req, "attacker-user")
	w := httptest.NewRecorder()

	Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Victim's subscription must remain untouched
	victimSubs, _ := store.ListSubscriptionsByUID(context.Background(), "victim-user")
	if len(victimSubs) != 1 {
		t.Fatalf("attacker was able to delete victim's subscription!")
	}
}
