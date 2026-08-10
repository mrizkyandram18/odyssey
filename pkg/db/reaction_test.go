package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestReactionStore_UpsertOmitsIdentityAndTimestamp(t *testing.T) {
	// Evidence: upsert must not send id=0 / zero created_at (prod bug class).
	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	resp, _ := json.Marshal([]game.Reaction{
		{
			ID:           42,
			CrewID:       "crew-A",
			TargetType:   "JOURNAL",
			TargetID:     1001,
			ActorUID:     "user-1",
			ReactionType: "HEART",
			CreatedAt:    now,
		},
	})
	client := &mockSupabaseClient{data: resp}
	store := NewReactionStore(client)

	in := &game.Reaction{
		// Explicit zero id / zero time — must not appear in outbound payload.
		ID:           0,
		CrewID:       "crew-A",
		TargetType:   "JOURNAL",
		TargetID:     1001,
		ActorUID:     "user-1",
		ReactionType: "HEART",
	}
	out, err := store.UpsertReaction(context.Background(), in)
	if err != nil {
		t.Fatalf("UpsertReaction: %v", err)
	}
	if out.ID != 42 {
		t.Fatalf("expected server-assigned id 42, got %d", out.ID)
	}
	if out.ReactionType != "HEART" {
		t.Fatalf("expected HEART, got %s", out.ReactionType)
	}

	if client.lastMutateTable != "odyssey_reactions" {
		t.Fatalf("expected table odyssey_reactions, got %s", client.lastMutateTable)
	}
	if client.lastMutateMethod != "POST" {
		t.Fatalf("expected POST, got %s", client.lastMutateMethod)
	}
	if client.lastMutateParams != "on_conflict=crew_id,target_type,target_id,actor_uid" {
		t.Fatalf("unexpected on_conflict params: %s", client.lastMutateParams)
	}
	if client.lastMutatePrefer != "return=representation,resolution=merge-duplicates" {
		t.Fatalf("unexpected prefer: %s", client.lastMutatePrefer)
	}

	payload, ok := client.lastMutatePayload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", client.lastMutatePayload)
	}
	if _, hasID := payload["id"]; hasID {
		t.Error("payload must omit id so PostgREST assigns identity")
	}
	if _, hasCreated := payload["created_at"]; hasCreated {
		t.Error("payload must omit created_at so DB default applies")
	}
	if payload["crew_id"] != "crew-A" || payload["target_type"] != "JOURNAL" {
		t.Errorf("unexpected payload keys: %+v", payload)
	}
	if payload["target_id"] != int64(1001) {
		// JSON numbers from map literals stay as int64 when set as int64
		if n, ok := payload["target_id"].(int64); !ok || n != 1001 {
			// also accept plain int
			if n2, ok2 := payload["target_id"].(int); !ok2 || n2 != 1001 {
				t.Errorf("expected target_id 1001, got %#v", payload["target_id"])
			}
		}
	}
	if payload["actor_uid"] != "user-1" || payload["reaction_type"] != "HEART" {
		t.Errorf("unexpected actor/type: %+v", payload)
	}
}

func TestReactionStore_ListForTarget(t *testing.T) {
	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	resp, _ := json.Marshal([]game.Reaction{
		{ID: 1, CrewID: "crew-A", TargetType: "JOURNAL", TargetID: 1001, ActorUID: "u2", ReactionType: "CLAP", CreatedAt: now},
		{ID: 2, CrewID: "crew-A", TargetType: "JOURNAL", TargetID: 1001, ActorUID: "u3", ReactionType: "STAR", CreatedAt: now},
	})
	client := &mockSupabaseClient{data: resp}
	store := NewReactionStore(client)

	list, err := store.ListReactionsForTarget(context.Background(), "crew-A", "JOURNAL", 1001)
	if err != nil {
		t.Fatalf("ListReactionsForTarget: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 reactions, got %d", len(list))
	}
	if len(client.getCalls) != 1 {
		t.Fatalf("expected 1 get call, got %d", len(client.getCalls))
	}
	if client.getCalls[0] == "" || client.getCalls[0][:len("odyssey_reactions")] != "odyssey_reactions" {
		t.Errorf("unexpected get call: %s", client.getCalls[0])
	}
}
