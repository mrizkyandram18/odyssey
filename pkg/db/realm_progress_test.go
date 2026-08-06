package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestRealmProgressStore_GetRealmProgress_Success(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]RealmProgress{
		{
			CrewID:         "crew-1",
			Realm:          "forest",
			Status:         "ACTIVE",
			StoryBranch:    "main-path",
			Progress:       3,
			LastUnlockedAt: now,
			UpdatedAt:      now,
		},
	})
	store := NewRealmProgressStore(&mockSupabaseClient{data: data})
	rp, err := store.GetRealmProgress(context.Background(), "crew-1", "forest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rp.Realm != "forest" {
		t.Errorf("expected realm forest, got %s", rp.Realm)
	}
	if rp.Progress != 3 {
		t.Errorf("expected progress 3, got %d", rp.Progress)
	}
}

func TestRealmProgressStore_CreateRealmProgress(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]RealmProgress{
		{
			CrewID:         "crew-1",
			Realm:          "cave",
			Status:         "ACTIVE",
			StoryBranch:    "",
			Progress:       0,
			LastUnlockedAt: now,
			UpdatedAt:      now,
		},
	})
	store := NewRealmProgressStore(&mockSupabaseClient{data: data})
	rp, err := store.CreateRealmProgress(context.Background(), &game.RealmProgress{
		CrewID:         "crew-1",
		Realm:          "cave",
		Status:         "ACTIVE",
		Progress:       0,
		LastUnlockedAt: now,
		UpdatedAt:      now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rp.Realm != "cave" {
		t.Errorf("expected realm cave, got %s", rp.Realm)
	}
	if rp.Status != "ACTIVE" {
		t.Errorf("expected status ACTIVE, got %s", rp.Status)
	}
}

func TestRealmProgressStore_CreateRealmProgress_Error(t *testing.T) {
	store := NewRealmProgressStore(&mockSupabaseClient{err: errors.New("network")})
	_, err := store.CreateRealmProgress(context.Background(), &game.RealmProgress{
		CrewID: "crew-1",
		Realm:  "cave",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRealmProgressStore_GetRealmProgress_NotFound(t *testing.T) {
	data, _ := json.Marshal([]RealmProgress{})
	store := NewRealmProgressStore(&mockSupabaseClient{data: data})
	_, err := store.GetRealmProgress(context.Background(), "crew-1", "forest")
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRealmProgressStore_ListByCrew(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]RealmProgress{
		{CrewID: "crew-1", Realm: "forest", Status: "ACTIVE", Progress: 1, UpdatedAt: now},
		{CrewID: "crew-1", Realm: "cave", Status: "LOCKED", Progress: 0, UpdatedAt: now},
	})
	store := NewRealmProgressStore(&mockSupabaseClient{data: data})
	progress, err := store.ListRealmProgressByCrew(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(progress) != 2 {
		t.Fatalf("expected 2 progress entries, got %d", len(progress))
	}
	if progress[0].Realm != "forest" {
		t.Errorf("expected first realm forest, got %s", progress[0].Realm)
	}
}

func TestRealmProgressStore_UpdateRealmProgress(t *testing.T) {
	store := NewRealmProgressStore(&mockSupabaseClient{data: []byte("[]")})
	err := store.UpdateRealmProgress(context.Background(), "crew-1", "forest", map[string]any{"status": "ACTIVE"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRealmProgressStore_ImplementsInterface(t *testing.T) {
	var _ game.RealmProgressStore = (*supabaseRealmProgressStore)(nil)
}
