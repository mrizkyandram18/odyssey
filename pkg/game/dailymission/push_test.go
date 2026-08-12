package dailymission

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/events"
	"odyssey/pkg/game/progression"
)

// capturePub records published events.
type capturePub struct {
	events []events.Event
}

func (p *capturePub) Publish(_ context.Context, e events.Event) {
	p.events = append(p.events, e)
}

func (p *capturePub) dailyTurnEvents() []events.DailyTurnCompletedEvent {
	var out []events.DailyTurnCompletedEvent
	for _, e := range p.events {
		if ev, ok := e.(events.DailyTurnCompletedEvent); ok {
			out = append(out, ev)
		}
	}
	return out
}

// mockUserForDailyPush satisfies game.UserStore for ProgressionService.
type mockUserForDailyPush struct {
	player *game.Player
}

func (m *mockUserForDailyPush) GetUser(_ context.Context, uid string) (*game.Player, error) {
	return m.player, nil
}
func (m *mockUserForDailyPush) CreateUser(_ context.Context, p *game.Player) error { return nil }
func (m *mockUserForDailyPush) UpdateUser(_ context.Context, uid string, patch map[string]any) error {
	if v, ok := patch["xp"].(int64); ok {
		m.player.XP = v
	}
	if v, ok := patch["version"].(int); ok {
		m.player.Version = v
	}
	return nil
}
func (m *mockUserForDailyPush) UpdateUserIfMatch(_ context.Context, uid string, ver int, patch map[string]any) (bool, error) {
	if m.player.Version != ver {
		return false, nil
	}
	_ = m.UpdateUser(context.Background(), uid, patch)
	return true, nil
}
func (m *mockUserForDailyPush) ListUsersByCrew(_ context.Context, crewID string) ([]game.Player, error) {
	return []game.Player{*m.player}, nil
}

func buildHandlerWithPublisher(turns []game.DailyMission, pub events.Publisher) *DailyTurnAPIHandler {
	store := &mockDailyTurnStore{turns: turns}
	cfg := DailyTurnConfig{XP: 30, MaxTurnsPerDay: 1, Timezone: "UTC", Now: time.Now}
	dts := NewDailyTurnService(store, &cfg)
	userStore := &mockUserForDailyPush{
		player: &game.Player{UID: "u1", XP: 0, Level: 1, Version: 1},
	}
	prog := progression.NewProgressionService(userStore, nil)
	return NewDailyTurnAPIHandlerWithPublisher(dts, prog, pub)
}

// TestDailyTurnPush_AssignedExplorerReceivesEvent verifies that the DailyTurnCompletedEvent
// carries the UID of the explorer who consumed the turn (the one who should receive push).
func TestDailyTurnPush_AssignedExplorerReceivesEvent(t *testing.T) {
	pub := &capturePub{}
	h := buildHandlerWithPublisher(nil, pub)

	_, err := h.Consume(context.Background(), "u1", "morning-light")
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}

	evts := pub.dailyTurnEvents()
	if len(evts) != 1 {
		t.Fatalf("expected 1 DailyTurnCompletedEvent, got %d", len(evts))
	}
	if evts[0].UID != "u1" {
		t.Errorf("event UID: got %q, want u1", evts[0].UID)
	}
}

// TestDailyTurnPush_OtherCrewMembersNotInEvent verifies only the consuming UID
// is in the event (not all crew members — push must be scoped to this UID).
func TestDailyTurnPush_OtherCrewMembersNotInEvent(t *testing.T) {
	pub := &capturePub{}
	h := buildHandlerWithPublisher(nil, pub)

	_, err := h.Consume(context.Background(), "u1", "morning-light")
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}

	for _, e := range pub.dailyTurnEvents() {
		if e.UID == "u2" {
			t.Error("UID u2 (non-consuming user) must not appear in DailyTurnCompletedEvent")
		}
	}
}

// TestDailyTurnPush_NoAssignment_NoEvent verifies that when Consume fails
// (e.g. already completed), no event is published.
func TestDailyTurnPush_NoAssignment_NoEvent(t *testing.T) {
	// Seed an already-completed turn for today.
	today := TodayDate()
	alreadyDone := game.DailyMission{
		ID: 1, UID: "u1", Date: today, MissionSlug: "morning-light", Completed: true,
	}
	pub := &capturePub{}
	h := buildHandlerWithPublisher([]game.DailyMission{alreadyDone}, pub)

	_, err := h.Consume(context.Background(), "u1", "morning-light")
	if !errors.Is(err, ErrNoTurnsRemaining) {
		t.Fatalf("expected ErrNoTurnsRemaining, got: %v", err)
	}

	if len(pub.events) != 0 {
		t.Errorf("expected no events on failed Consume, got %d", len(pub.events))
	}
}

// TestDailyTurnPush_EconomyUnchanged verifies that adding push delivery does not
// alter XP/economy behavior: the XP field in the result must remain correct.
func TestDailyTurnPush_EconomyUnchanged(t *testing.T) {
	pub := &capturePub{}
	h := buildHandlerWithPublisher(nil, pub)

	result, err := h.Consume(context.Background(), "u1", "morning-light")
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}

	if result.XP <= 0 {
		t.Errorf("XP must be > 0 (economy unchanged), got %d", result.XP)
	}
	if result.Turn.UID != "u1" {
		t.Errorf("Turn.UID: got %q, want u1", result.Turn.UID)
	}
}
