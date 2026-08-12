package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestUserStore_GetUser_Success(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]UserProfile{
		{
			UID:          "user-1",
			FamilyID:       "crew-1",
			ExplorerName: "Alice",
			Role:         "SEEKER",
			Level:        3,
			XP:           500,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	})
	store := NewUserStore(&mockSupabaseClient{data: data})
	player, err := store.GetUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.UID != "user-1" {
		t.Errorf("expected UID user-1, got %s", player.UID)
	}
	if player.ExplorerName != "Alice" {
		t.Errorf("expected name Alice, got %s", player.ExplorerName)
	}
	if player.Level != 3 {
		t.Errorf("expected level 3, got %d", player.Level)
	}
}

func TestUserStore_GetUser_NotFound(t *testing.T) {
	data, _ := json.Marshal([]UserProfile{})
	store := NewUserStore(&mockSupabaseClient{data: data})
	_, err := store.GetUser(context.Background(), "user-1")
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected game.ErrNotFound, got %v", err)
	}
}

func TestUserStore_GetUser_Error(t *testing.T) {
	store := NewUserStore(&mockSupabaseClient{err: errors.New("network")})
	_, err := store.GetUser(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error")
	}
}
