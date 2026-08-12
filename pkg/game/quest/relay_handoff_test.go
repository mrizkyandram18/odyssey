package quest

// relay_handoff_test.go exercises RelayHandoffEvent emission from QuestAPIHandler.
// This file is in the same package (quest) as handler_test.go, so it can reuse
// the capturePublisher type and mockQuestStore declared there.

import (
	"context"
	"testing"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/events"
	"odyssey/pkg/game/progression"
	"odyssey/pkg/game/world"
)

// relayHandoffs filters RelayHandoffEvents from a capturePublisher.
func relayHandoffs(pub *capturePublisher) []events.RelayHandoffEvent {
	var out []events.RelayHandoffEvent
	for _, e := range pub.published {
		if rhe, ok := e.(events.RelayHandoffEvent); ok {
			out = append(out, rhe)
		}
	}
	return out
}

// buildRelayHandoffHandler creates a minimal QuestAPIHandler wired with a
// publisher, a two-member crew, and a relay quest with two pending challenges.
func buildRelayHandoffHandler(store *mockQuestStore) (*QuestAPIHandler, *capturePublisher) {
	pub := &capturePublisher{}

	users := &mockListUserStore{
		players: []game.Player{
			{UID: "alice", ExplorerName: "Alice", XP: 0, Version: 1},
			{UID: "bob", ExplorerName: "Bob", XP: 0, Version: 1},
		},
	}
	progSvc := progression.NewProgressionService(users, nil)
	realmCatalog := world.DefaultRealmCatalog
	qs := NewQuestService(store)
	qs.SetUserStore(users)

	h := NewQuestAPIHandler(qs, progSvc, &mockRealmProgressStoreForHandler{
		progress: map[string]*game.RealmProgress{},
		updates:  []string{},
	}, realmCatalog, nil)
	h.SetPublisher(pub)
	return h, pub
}

// mockListUserStore satisfies game.UserStore including ListUsersByCrew.
type mockListUserStore struct {
	players []game.Player
}

func (m *mockListUserStore) GetUser(_ context.Context, uid string) (*game.Player, error) {
	for i := range m.players {
		if m.players[i].UID == uid {
			return &m.players[i], nil
		}
	}
	return nil, game.ErrNotFound
}
func (m *mockListUserStore) CreateUser(_ context.Context, p *game.Player) error { return nil }
func (m *mockListUserStore) UpdateUser(_ context.Context, uid string, patch map[string]any) error {
	for i := range m.players {
		if m.players[i].UID == uid {
			if v, ok := patch["xp"].(int64); ok {
				m.players[i].XP = v
			}
			if v, ok := patch["level"].(int); ok {
				m.players[i].Level = v
			}
			if v, ok := patch["version"].(int); ok {
				m.players[i].Version = v
			}
		}
	}
	return nil
}
func (m *mockListUserStore) UpdateUserIfMatch(_ context.Context, uid string, version int, patch map[string]any) (bool, error) {
	for i := range m.players {
		if m.players[i].UID == uid {
			if m.players[i].Version != version {
				return false, nil
			}
			_ = m.UpdateUser(context.Background(), uid, patch)
			return true, nil
		}
	}
	return false, game.ErrNotFound
}
func (m *mockListUserStore) ListUsersByCrew(_ context.Context, crewID string) ([]game.Player, error) {
	return m.players, nil
}

func makeRelayQuestStore() *mockQuestStore {
	bobUID := "bob"
	return &mockQuestStore{
		quests: map[int64]*game.Quest{
			1: {
				ID:           1,
				CrewID:       "crew-1",
				TemplateSlug: "relay-family-adventure",
				Title:        "Family Adventure",
				Status:       "ACTIVE",
				StartedAt:    func() *time.Time { t := time.Now(); return &t }(),
				CreatedAt:    time.Now(),
			},
		},
		challenges: map[int64][]game.Challenge{
			1: {
				{ID: 10, QuestID: 1, Status: "PENDING", AssignedTo: func() *string { s := "alice"; return &s }()},
				{ID: 11, QuestID: 1, Status: "PENDING", AssignedTo: &bobUID},
			},
		},
	}
}

func TestPublishRelayHandoff_EmittedAfterSuccessfulAssignment(t *testing.T) {
	store := makeRelayQuestStore()
	h, pub := buildRelayHandoffHandler(store)

	_, err := h.CompleteChallenge(context.Background(), 1, 10, "crew-1", "alice", "")
	if err != nil {
		t.Fatalf("CompleteChallenge: %v", err)
	}

	handoffs := relayHandoffs(pub)
	if len(handoffs) != 1 {
		t.Fatalf("expected 1 RelayHandoffEvent, got %d", len(handoffs))
	}
	e := handoffs[0]
	if e.FromUID != "alice" {
		t.Errorf("FromUID: got %q, want alice", e.FromUID)
	}
	if e.ToUID != "bob" {
		t.Errorf("ToUID: got %q, want bob", e.ToUID)
	}
	if e.QuestTitle != "Family Adventure" {
		t.Errorf("QuestTitle: got %q, want Family Adventure", e.QuestTitle)
	}
	if e.QuestID != 1 {
		t.Errorf("QuestID: got %d, want 1", e.QuestID)
	}
}

func TestPublishRelayHandoff_RecipientDiffersFromSender(t *testing.T) {
	store := makeRelayQuestStore()
	h, pub := buildRelayHandoffHandler(store)

	_, err := h.CompleteChallenge(context.Background(), 1, 10, "crew-1", "alice", "")
	if err != nil {
		t.Fatalf("CompleteChallenge: %v", err)
	}

	for _, rhe := range relayHandoffs(pub) {
		if rhe.ToUID == rhe.FromUID {
			t.Error("ToUID must differ from FromUID (sender should not receive their own push)")
		}
	}
}

func TestPublishRelayHandoff_NoEventOnQuestCompletion(t *testing.T) {
	// When the last challenge is completed, quest becomes DONE — no handoff event.
	bobUID := "bob"
	store := &mockQuestStore{
		quests: map[int64]*game.Quest{
			2: {
				ID: 2, CrewID: "crew-1", TemplateSlug: "relay-quest", Title: "Relay Quest",
				Status: "ACTIVE", StartedAt: func() *time.Time { t := time.Now(); return &t }(),
			},
		},
		challenges: map[int64][]game.Challenge{
			2: {
				{ID: 20, QuestID: 2, Status: "DONE", CompletedBy: "alice"},
				{ID: 21, QuestID: 2, Status: "PENDING", AssignedTo: &bobUID},
			},
		},
	}

	h, pub := buildRelayHandoffHandler(store)

	_, err := h.CompleteChallenge(context.Background(), 2, 21, "crew-1", "bob", "")
	if err != nil {
		t.Fatalf("CompleteChallenge: %v", err)
	}

	// Quest becomes DONE — questCompleted=true → no relay handoff emitted.
	if len(relayHandoffs(pub)) != 0 {
		t.Errorf("expected no RelayHandoffEvent when quest completes, got %d", len(relayHandoffs(pub)))
	}
}

func TestPublishRelayHandoff_NoEventOnReplay(t *testing.T) {
	// When challenge is already DONE, no progression — no handoff event.
	store := &mockQuestStore{
		quests: map[int64]*game.Quest{
			3: {
				ID: 3, CrewID: "crew-1", TemplateSlug: "relay-quest", Title: "Relay",
				Status: "ACTIVE", StartedAt: func() *time.Time { t := time.Now(); return &t }(),
			},
		},
		challenges: map[int64][]game.Challenge{
			3: {
				{ID: 30, QuestID: 3, Status: "DONE", CompletedBy: "alice"},
				{ID: 31, QuestID: 3, Status: "PENDING"},
			},
		},
	}

	h, pub := buildRelayHandoffHandler(store)

	// Replay the already-completed challenge.
	_, err := h.CompleteChallenge(context.Background(), 3, 30, "crew-1", "alice", "")
	if err != nil {
		t.Fatalf("CompleteChallenge (replay): %v", err)
	}

	if len(relayHandoffs(pub)) != 0 {
		t.Errorf("expected no RelayHandoffEvent on replay, got %d", len(relayHandoffs(pub)))
	}
}

func TestPublishRelayHandoff_QuestTitlePropagated(t *testing.T) {
	store := makeRelayQuestStore()
	h, pub := buildRelayHandoffHandler(store)

	_, err := h.CompleteChallenge(context.Background(), 1, 10, "crew-1", "alice", "")
	if err != nil {
		t.Fatalf("CompleteChallenge: %v", err)
	}

	handoffs := relayHandoffs(pub)
	if len(handoffs) == 0 {
		t.Fatal("expected at least 1 RelayHandoffEvent")
	}
	if handoffs[0].QuestTitle == "" {
		t.Error("QuestTitle must be propagated in RelayHandoffEvent")
	}
}
