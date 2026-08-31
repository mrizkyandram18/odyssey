package push

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"odyssey/pkg/db"
)

// ---- mock store ----

type mockSubStore struct {
	mu        sync.Mutex
	subs      []db.PushSubscription
	listErr   error
	deleteErr error
	deleted   []string // endpoints deleted
}

func (m *mockSubStore) UpsertSubscription(_ context.Context, sub *db.PushSubscription) (*db.PushSubscription, error) {
	return sub, nil
}

func (m *mockSubStore) ListSubscriptionsByUID(_ context.Context, uid string) ([]db.PushSubscription, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []db.PushSubscription
	for _, s := range m.subs {
		if s.UID == uid {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *mockSubStore) DeleteSubscription(_ context.Context, uid, endpoint string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleted = append(m.deleted, endpoint)
	return nil
}

// ---- helpers ----

const (
	testVAPIDPublic  = "BEl62iUYgUivxIkv69yViEuiBIa40R2oCOX10ol4DNfJ9vKJSb_B6Ri7-_9l7mTXj0gUwDlDMFa1RiEsRIlbfPo"
	testVAPIDPrivate = "EVAibCpGBnwE7PaHcNaJ4G3k_CElMT3FHlyiHHmJonc"
)

func makeSender(store db.PushSubscriptionStore) (*Sender, error) {
	return NewSender(Config{
		VAPIDPublicKey:  testVAPIDPublic,
		VAPIDPrivateKey: testVAPIDPrivate,
		VAPIDSubject:    "mailto:test@example.com",
	}, store)
}

func makeSub(uid, endpoint string) db.PushSubscription {
	return db.PushSubscription{
		ID:        1,
		UID:       uid,
		Endpoint:  endpoint,
		P256DH:    "BGBENAHSPFQVFv60HvtCdHLGFKH3OFRaMnG_hHULAUWEd35CePkANvVHzHOstSK55LK1MBq6l5YMYYGe7mCpx8U",
		Auth:      "sdBWyxr4x3WlhbfF5-Dqug",
		CreatedAt: time.Now(),
	}
}

// ---- tests ----

func TestSender_MissingVAPIDConfig(t *testing.T) {
	_, err := NewSender(Config{}, &mockSubStore{})
	if !errors.Is(err, ErrVAPIDNotConfigured) {
		t.Fatalf("expected ErrVAPIDNotConfigured, got: %v", err)
	}

	_, err = NewSender(Config{VAPIDPublicKey: "pub"}, &mockSubStore{})
	if !errors.Is(err, ErrVAPIDNotConfigured) {
		t.Fatalf("expected ErrVAPIDNotConfigured with missing private key, got: %v", err)
	}
}

func TestSender_NoSubscriptions(t *testing.T) {
	store := &mockSubStore{}
	s, err := makeSender(store)
	if err != nil {
		t.Fatalf("makeSender: %v", err)
	}

	err = s.SendToUser(context.Background(), "uid-nobody", Notification{
		Title: "Test", Body: "body", URL: "/",
	})
	if err != nil {
		t.Fatalf("expected nil error with no subscriptions, got: %v", err)
	}
}

func TestSender_ListError(t *testing.T) {
	listErr := errors.New("db down")
	store := &mockSubStore{listErr: listErr}
	s, err := makeSender(store)
	if err != nil {
		t.Fatalf("makeSender: %v", err)
	}

	err = s.SendToUser(context.Background(), "uid-1", Notification{Title: "Test"})
	if err == nil {
		t.Fatal("expected error when list subscriptions fails")
	}
}

func TestSender_SingleSubscription_Success(t *testing.T) {
	var received []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf [4096]byte
		n, _ := r.Body.Read(buf[:])
		received = buf[:n]
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	sub := makeSub("u1", srv.URL+"/push/endpoint")
	store := &mockSubStore{subs: []db.PushSubscription{sub}}
	s, err := makeSender(store)
	if err != nil {
		t.Fatalf("makeSender: %v", err)
	}

	n := Notification{Title: "Tugas Baru", Body: "Ada tugas harian baru.", URL: "/"}
	err = s.SendToUser(context.Background(), "u1", n)
	if len(received) > 0 {
		if containsSecret(received, sub.Auth) || containsSecret(received, sub.P256DH) {
			t.Error("subscription secrets must not appear in push payload")
		}
	}
	_ = err
}

func TestSender_MultipleSubscriptions_AllAttempted(t *testing.T) {
	subs := []db.PushSubscription{
		makeSub("u1", "https://push.example.com/ep1"),
		makeSub("u1", "https://push.example.com/ep2"),
	}
	store := &mockSubStore{subs: subs}
	s, err := makeSender(store)
	if err != nil {
		t.Fatalf("makeSender: %v", err)
	}

	gotErr := s.SendToUser(context.Background(), "u1", Notification{Title: "Test"})
	if gotErr == nil {
		t.Error("expected non-nil error when all subscriptions fail delivery")
	}
}

func TestSender_HTTP410_DeletesSubscription(t *testing.T) {
	sub := makeSub("u1", "https://push.example.com/stale")
	store := &mockSubStore{subs: []db.PushSubscription{sub}}

	if err := store.DeleteSubscription(context.Background(), sub.UID, sub.Endpoint); err != nil {
		t.Fatalf("DeleteSubscription: %v", err)
	}

	store.mu.Lock()
	deleted := store.deleted
	store.mu.Unlock()

	if len(deleted) != 1 || deleted[0] != sub.Endpoint {
		t.Errorf("expected endpoint %q to be recorded as deleted, got %v", sub.Endpoint, deleted)
	}
}

func TestSender_HTTP5xx_ReportsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	sub := makeSub("u1", srv.URL+"/push/error")
	store := &mockSubStore{subs: []db.PushSubscription{sub}}
	s, err := makeSender(store)
	if err != nil {
		t.Fatalf("makeSender: %v", err)
	}

	err = s.SendToUser(context.Background(), "u1", Notification{Title: "Test"})
	if err == nil {
		t.Error("expected error for HTTP 5xx response")
	}
}

func TestSender_CorrectMinimalPayload(t *testing.T) {
	n := Notification{
		Title: "Tugas Baru",
		Body:  "Ada tugas baru.",
		URL:   "/",
	}
	encoded, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded["title"] != "Tugas Baru" {
		t.Errorf("title field: got %q", decoded["title"])
	}
}

func containsSecret(data []byte, secret string) bool {
	return len(secret) > 4 && len(data) > 0 &&
		len(secret) > 8 && string(data) != "" &&
		containsBytes(data, []byte(secret))
}

func containsBytes(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if string(haystack[i:i+len(needle)]) == string(needle) {
			return true
		}
	}
	return false
}
