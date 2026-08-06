package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	gamecontent "odyssey/pkg/game/content"
)

func makeQuestRow(challengeDefsJSON json.RawMessage) QuestDefinition {
	return QuestDefinition{
		ID:            1,
		Slug:          "test-quest",
		Realm:         "whispering-woods",
		Chapter:       "the-awakening",
		Title:         "Test Quest",
		QuestType:     "SOLO",
		ChallengeDefs: challengeDefsJSON,
		RewardXP:      100,
		RewardChest:   "wooden-chest",
		Published:     true,
		Version:       1,
		CreatedAt:     time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func TestMapQuestDefinition_OneChallenge(t *testing.T) {
	defs := `[{"slug":"find-the-dew","description":"Find something glistening.","type":"OBSERVATION"}]`
	d := makeQuestRow(json.RawMessage(defs))

	result, err := mapQuestDefinition(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.ChallengeDefs) != 1 {
		t.Fatalf("expected 1 challenge def, got %d", len(result.ChallengeDefs))
	}
	if result.ChallengeDefs[0].Slug != "find-the-dew" {
		t.Errorf("expected slug find-the-dew, got %s", result.ChallengeDefs[0].Slug)
	}
	if result.ChallengeDefs[0].Type != "OBSERVATION" {
		t.Errorf("expected type OBSERVATION, got %s", result.ChallengeDefs[0].Type)
	}
}

func TestMapQuestDefinition_MultipleChallenges(t *testing.T) {
	defs := `[
		{"slug":"find-the-dew","description":"Find something glistening.","type":"OBSERVATION"},
		{"slug":"morning-fact","description":"Look up a fact.","type":"RESEARCH"},
		{"slug":"solve-riddle","description":"Solve a riddle.","type":"PUZZLE"}
	]`
	d := makeQuestRow(json.RawMessage(defs))

	result, err := mapQuestDefinition(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.ChallengeDefs) != 3 {
		t.Fatalf("expected 3 challenge defs, got %d", len(result.ChallengeDefs))
	}
	if result.ChallengeDefs[0].Slug != "find-the-dew" {
		t.Errorf("expected first slug find-the-dew, got %s", result.ChallengeDefs[0].Slug)
	}
	if result.ChallengeDefs[1].Slug != "morning-fact" {
		t.Errorf("expected second slug morning-fact, got %s", result.ChallengeDefs[1].Slug)
	}
	if result.ChallengeDefs[2].Slug != "solve-riddle" {
		t.Errorf("expected third slug solve-riddle, got %s", result.ChallengeDefs[2].Slug)
	}
}

func TestMapQuestDefinition_EmptyArray(t *testing.T) {
	d := makeQuestRow(json.RawMessage(`[]`))

	result, err := mapQuestDefinition(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.ChallengeDefs) != 0 {
		t.Errorf("expected 0 challenge defs, got %d", len(result.ChallengeDefs))
	}
}

func TestMapQuestDefinition_NilRawMessage(t *testing.T) {
	d := makeQuestRow(json.RawMessage(nil))

	result, err := mapQuestDefinition(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.ChallengeDefs) != 0 {
		t.Errorf("expected 0 challenge defs for nil, got %d", len(result.ChallengeDefs))
	}
}

func TestMapQuestDefinition_NullJSON(t *testing.T) {
	d := makeQuestRow(json.RawMessage(`null`))

	result, err := mapQuestDefinition(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.ChallengeDefs) != 0 {
		t.Errorf("expected 0 challenge defs for null, got %d", len(result.ChallengeDefs))
	}
}

func TestMapQuestDefinition_MalformedJSON(t *testing.T) {
	d := makeQuestRow(json.RawMessage(`[{invalid json}]`))

	_, err := mapQuestDefinition(d)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestMapQuestDefinition_UnknownFields(t *testing.T) {
	defs := `[{"slug":"find-the-dew","description":"Find something.","type":"OBSERVATION","unknown_field":"ignored"}]`
	d := makeQuestRow(json.RawMessage(defs))

	result, err := mapQuestDefinition(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.ChallengeDefs) != 1 {
		t.Fatalf("expected 1 challenge def, got %d", len(result.ChallengeDefs))
	}
	if result.ChallengeDefs[0].Slug != "find-the-dew" {
		t.Errorf("expected slug find-the-dew, got %s", result.ChallengeDefs[0].Slug)
	}
}

func TestMapQuestDefinition_RoundTripSerialization(t *testing.T) {
	original := []gamecontent.ChallengeDef{
		{Slug: "find-the-dew", Description: "Find something glistening.", Type: "OBSERVATION"},
		{Slug: "morning-fact", Description: "Look up a fact.", Type: "RESEARCH"},
	}
	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal challenge defs: %v", err)
	}
	d := makeQuestRow(raw)

	result, err := mapQuestDefinition(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ChallengeDefs) != len(original) {
		t.Fatalf("expected %d challenge defs, got %d", len(original), len(result.ChallengeDefs))
	}
	for i, cd := range result.ChallengeDefs {
		if cd.Slug != original[i].Slug {
			t.Errorf("challenge %d: expected slug %s, got %s", i, original[i].Slug, cd.Slug)
		}
		if cd.Description != original[i].Description {
			t.Errorf("challenge %d: expected description %s, got %s", i, original[i].Description, cd.Description)
		}
		if cd.Type != original[i].Type {
			t.Errorf("challenge %d: expected type %s, got %s", i, original[i].Type, cd.Type)
		}
	}

	reSerialized, err := json.Marshal(result.ChallengeDefs)
	if err != nil {
		t.Fatalf("failed to re-marshal: %v", err)
	}
	var roundTripped []gamecontent.ChallengeDef
	if err := json.Unmarshal(reSerialized, &roundTripped); err != nil {
		t.Fatalf("failed to unmarshal round-tripped: %v", err)
	}
	if len(roundTripped) != len(original) {
		t.Fatalf("round-trip length mismatch: expected %d, got %d", len(original), len(roundTripped))
	}
}

func TestQuestStore_GetQuest_WithChallengeDefs(t *testing.T) {
	questJSON := `[{"id":1,"slug":"morning-light","realm":"whispering-woods","chapter":"the-awakening","title":"Morning Light","quest_type":"SOLO","challenge_defs":[{"slug":"find-the-dew","description":"Find something glistening.","type":"OBSERVATION"},{"slug":"morning-fact","description":"Look up a fact.","type":"RESEARCH"}],"reward_xp":80,"reward_chest":"wooden-chest","is_mandatory":true,"published":true,"version":1,"created_at":"2026-08-03T12:00:00Z","updated_at":"2026-08-03T12:00:00Z"}]`
	store := NewQuestDefinitionStore(&mockSupabaseClient{data: []byte(questJSON)})

	def, err := store.GetQuest(context.Background(), "morning-light")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def == nil {
		t.Fatal("expected non-nil quest definition")
	}
	if len(def.ChallengeDefs) != 2 {
		t.Fatalf("expected 2 challenge defs, got %d", len(def.ChallengeDefs))
	}
	if def.ChallengeDefs[0].Slug != "find-the-dew" {
		t.Errorf("expected first challenge slug find-the-dew, got %s", def.ChallengeDefs[0].Slug)
	}
}

func TestQuestStore_ListQuests_WithChallengeDefs(t *testing.T) {
	questListJSON := `[
		{"id":1,"slug":"morning-light","realm":"whispering-woods","chapter":"the-awakening","title":"Morning Light","quest_type":"SOLO","challenge_defs":[{"slug":"find-the-dew","description":"Find something.","type":"OBSERVATION"}],"reward_xp":80,"published":true,"version":1,"created_at":"2026-08-03T12:00:00Z","updated_at":"2026-08-03T12:00:00Z"},
		{"id":2,"slug":"gather-herbs","realm":"whispering-woods","chapter":"the-awakening","title":"Gather Herbs","quest_type":"SOLO","challenge_defs":[{"slug":"spot-the-green","description":"Point out greens.","type":"OBSERVATION"},{"slug":"herb-lore","description":"Name a use.","type":"RESEARCH"}],"reward_xp":80,"published":true,"version":1,"created_at":"2026-08-03T12:00:00Z","updated_at":"2026-08-03T12:00:00Z"}
	]`
	store := NewQuestDefinitionStore(&mockSupabaseClient{data: []byte(questListJSON)})

	quests, err := store.ListQuests(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(quests) != 2 {
		t.Fatalf("expected 2 quests, got %d", len(quests))
	}
	if len(quests[0].ChallengeDefs) != 1 {
		t.Errorf("expected 1 challenge def for quest 0, got %d", len(quests[0].ChallengeDefs))
	}
	if len(quests[1].ChallengeDefs) != 2 {
		t.Errorf("expected 2 challenge defs for quest 1, got %d", len(quests[1].ChallengeDefs))
	}
}

func TestQuestStore_GetQuest_MalformedJSONReturnsError(t *testing.T) {
	questJSON := `[{"id":1,"slug":"morning-light","challenge_defs":"[invalid json]","published":true,"version":1,"created_at":"2026-08-03T12:00:00Z","updated_at":"2026-08-03T12:00:00Z"}]`
	store := NewQuestDefinitionStore(&mockSupabaseClient{data: []byte(questJSON)})

	_, err := store.GetQuest(context.Background(), "morning-light")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestQuestStore_ListQuests_MalformedJSONReturnsError(t *testing.T) {
	questListJSON := `[
		{"id":1,"slug":"morning-light","challenge_defs":[{"slug":"ok","type":"OBSERVATION"}],"published":true,"version":1,"created_at":"2026-08-03T12:00:00Z","updated_at":"2026-08-03T12:00:00Z"},
		{"id":2,"slug":"gather-herbs","challenge_defs":"not-json","published":true,"version":1,"created_at":"2026-08-03T12:00:00Z","updated_at":"2026-08-03T12:00:00Z"}
	]`
	store := NewQuestDefinitionStore(&mockSupabaseClient{data: []byte(questListJSON)})

	_, err := store.ListQuests(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed JSON in list")
	}
}

func TestQuestStore_GetQuest_PostgRESTNativeJSONB(t *testing.T) {
	questJSON := `[{"id":1,"slug":"morning-light","realm":"whispering-woods","chapter":"the-awakening","title":"Morning Light","quest_type":"SOLO","challenge_defs":[{"slug":"find-the-dew","description":"Find something glistening outside your door.","type":"OBSERVATION"},{"slug":"morning-fact","description":"Look up one fact about morning sunlight.","type":"RESEARCH"}],"reward_xp":80,"reward_chest":"wooden-chest","is_mandatory":true,"required_quest_slug":"","required_quest_slugs":[],"required_chapter":"","required_realm":"","required_level":0,"season_slug":"","published":true,"version":1,"updated_by":"system","created_at":"2026-08-03T12:00:00Z","updated_at":"2026-08-03T12:00:00Z"}]`
	store := NewQuestDefinitionStore(&mockSupabaseClient{data: []byte(questJSON)})

	def, err := store.GetQuest(context.Background(), "morning-light")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def == nil {
		t.Fatal("expected non-nil quest definition")
	}
	if len(def.ChallengeDefs) != 2 {
		t.Fatalf("expected 2 challenge defs, got %d", len(def.ChallengeDefs))
	}
	if def.ChallengeDefs[0].Slug != "find-the-dew" {
		t.Errorf("expected first slug find-the-dew, got %s", def.ChallengeDefs[0].Slug)
	}
	if def.ChallengeDefs[1].Slug != "morning-fact" {
		t.Errorf("expected second slug morning-fact, got %s", def.ChallengeDefs[1].Slug)
	}
}

func TestQuestStore_GetQuest_ClientError(t *testing.T) {
	clientErr := errors.New("network error")
	store := NewQuestDefinitionStore(&mockSupabaseClient{err: clientErr})

	_, err := store.GetQuest(context.Background(), "morning-light")
	if err == nil {
		t.Fatal("expected error from client")
	}
	if !errors.Is(err, clientErr) {
		t.Errorf("expected wrapped client error, got: %v", err)
	}
}
