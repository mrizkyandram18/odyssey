package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestDailyTurnStore_ListDailyTurns_Success(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]DailyTurn{
		{
			ID:        1,
			UID:       "user-1",
			Date:      "2026-08-01",
			QuestSlug: "morning-light",
			Completed: true,
			CreatedAt: now,
		},
		{
			ID:        2,
			UID:       "user-1",
			Date:      "2026-08-02",
			QuestSlug: "forest-walk",
			Completed: false,
			CreatedAt: now,
		},
	})
	store := NewDailyTurnStore(&mockSupabaseClient{data: data})
	turns, err := store.ListDailyTurns(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}
	if turns[0].Date != "2026-08-01" {
		t.Errorf("expected date 2026-08-01, got %s", turns[0].Date)
	}
	if turns[1].Completed != false {
		t.Errorf("expected second turn not completed")
	}
}

func TestDailyTurnStore_ListDailyTurns_Error(t *testing.T) {
	store := NewDailyTurnStore(&mockSupabaseClient{err: errors.New("network")})
	_, err := store.ListDailyTurns(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDailyTurnStore_CreateDailyTurn(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]DailyTurn{
		{ID: 1, UID: "user-1", Date: "2026-08-03", QuestSlug: "morning-light", Completed: false, CreatedAt: now},
	})
	store := NewDailyTurnStore(&mockSupabaseClient{data: data})
	dt := &game.DailyTurn{
		UID:       "user-1",
		Date:      "2026-08-03",
		QuestSlug: "morning-light",
	}
	result, err := store.CreateDailyTurn(context.Background(), dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.QuestSlug != "morning-light" {
		t.Errorf("expected slug morning-light, got %s", result.QuestSlug)
	}
}

func TestDailyTurnStore_UpdateDailyTurn(t *testing.T) {
	store := NewDailyTurnStore(&mockSupabaseClient{data: []byte("[]")})
	err := store.UpdateDailyTurn(context.Background(), 1, map[string]any{"completed": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDailyTurnStore_ImplementsInterface(t *testing.T) {
	var _ game.DailyTurnStore = (*supabaseDailyTurnStore)(nil)
}
