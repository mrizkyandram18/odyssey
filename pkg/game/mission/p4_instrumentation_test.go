package quest

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/progression"
	"odyssey/pkg/observability"
)

func setupOrchestratorWithMetrics(t *testing.T, qStore *mockQuestStore, player *game.Player) (*QuestAPIHandler, *mockRealmProgressStoreForHandler, *observability.Metrics) {
	t.Helper()
	qs := NewQuestService(qStore)
	userStore := &mockUserStoreForHandler{player: player}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	journey := newMockRealmStore()
	journey.progress["crew-1|whispering-woods"].Progress = 75
	m := observability.NewMetrics()
	prog.SetMetrics(m)
	h := NewQuestAPIHandler(qs, prog, journey, defaultRealmCfg, &defaultProgCfg)
	h.SetMetrics(m)
	return h, journey, m
}

func TestCompleteChallenge_RealmCompletedMetric(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.missions[1] = q
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	h, journey, m := setupOrchestratorWithMetrics(t, qStore, makePlayerForHandler(1, 90))
	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.QuestCompleted {
		t.Fatal("expected quest completed")
	}
	if len(journey.updates) != 1 {
		t.Errorf("expected 1 journey update, got %d", len(journey.updates))
	}
	snap := m.Snapshot()
	if snap.XPAwarded <= 0 {
		t.Errorf("expected XP awarded > 0, got %d", snap.XPAwarded)
	}
	if snap.RealmCompleted != 1 {
		t.Errorf("expected 1 journey completed, got %d", snap.RealmCompleted)
	}
}

func TestCompleteChallenge_ServiceCallLogging(t *testing.T) {
	qStore := newMockQuestStore()
	qStore.missions[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusPending))
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusPending), string(ChallengeStatusPending))

	var buf bytes.Buffer
	logger := observability.NewLogger(&buf)
	qs := NewQuestService(qStore)
	prog := progression.NewProgressionService(&mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}, &defaultProgCfg)
	journey := newMockRealmStore()
	h := NewQuestAPIHandler(qs, prog, journey, defaultRealmCfg, &defaultProgCfg)
	h.SetLogger(logger)

	// Non-completing challenge: still exercises the service-call log path.
	if _, err := h.CompleteChallenge(context.Background(), 1, 11, "crew-1", "user-1", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	logger.Flush()
	out := buf.String()
	if !strings.Contains(out, "service_call") || !strings.Contains(out, "quest.complete_challenge") {
		t.Fatalf("expected service_call log for quest.complete_challenge, got: %s", out)
	}
}

func TestQuestAPIHandler_SetMetricsNilSafe(t *testing.T) {
	h, _ := setupOrchestrator(t, newMockQuestStore(), makePlayerForHandler(1, 0))
	h.SetMetrics(nil)
	h.SetLogger(nil)
	// Completing an unknown quest exercises the early error-return log path.
	_, _ = h.CompleteChallenge(context.Background(), 999, 1, "crew-1", "user-1", "")
}
