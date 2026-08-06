package db

import (
	"testing"
)

func TestBuildRepository_Success(t *testing.T) {
	client := NewClient("https://test.supabase.co", "test-key")
	repo, err := BuildRepository(client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
	if repo.Users == nil {
		t.Error("expected Users store to be set")
	}
	if repo.Quests == nil {
		t.Error("expected Quests store to be set")
	}
	if repo.DailyTurns == nil {
		t.Error("expected DailyTurns store to be set")
	}
	if repo.Progression == nil {
		t.Error("expected Progression store to be set")
	}
	if repo.RealmProgress == nil {
		t.Error("expected RealmProgress store to be set")
	}
	if repo.Creatives == nil {
		t.Error("expected Creatives store to be set")
	}
	if repo.Crews == nil {
		t.Error("expected Crews store to be set")
	}
	if repo.Config == nil {
		t.Error("expected Config store to be set")
	}
	if repo.Chests == nil {
		t.Error("expected Chests store to be set")
	}
	if repo.ChestDefinitions == nil {
		t.Error("expected ChestDefinitions store to be set")
	}
	if repo.Relics == nil {
		t.Error("expected Relics store to be set")
	}
	if repo.RelicDefinitions == nil {
		t.Error("expected RelicDefinitions store to be set")
	}
	if repo.PlayerRelics == nil {
		t.Error("expected PlayerRelics store to be set")
	}
}
