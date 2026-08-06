package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestCreativeSubmissionStore_Create(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]CreativeSubmission{
		{
			ID:          1,
			QuestID:     1,
			ChallengeID: 1,
			CrewID:      "crew-1",
			AuthorUID:   "user-1",
			Kind:        "STORY",
			Content:     "A short story",
			Status:      "PENDING",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	})
	client := &mockSupabaseClient{data: data}
	store := NewCreativeSubmissionStore(client)

	sub := &game.Submission{
		QuestID:     1,
		ChallengeID: 1,
		CrewID:      "crew-1",
		AuthorUID:   "user-1",
		Kind:        game.SubmissionStory,
		Content:     "A short story",
		Status:      game.SubmissionStatusPending,
	}
	created, err := store.CreateSubmission(context.Background(), sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected ID to be set")
	}
	if created.Content != "A short story" {
		t.Errorf("expected content 'A short story', got %s", created.Content)
	}
}

func TestCreativeSubmissionStore_ListByQuest(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]CreativeSubmission{
		{ID: 1, QuestID: 1, ChallengeID: 1, CrewID: "c1", AuthorUID: "u1", Kind: "STORY", Content: "s1", Status: "PENDING", CreatedAt: now, UpdatedAt: now},
		{ID: 2, QuestID: 1, ChallengeID: 1, CrewID: "c1", AuthorUID: "u1", Kind: "COMIC", Content: "c1", Status: "PENDING", CreatedAt: now, UpdatedAt: now},
	})
	store := NewCreativeSubmissionStore(&mockSupabaseClient{data: data})

	subs, err := store.ListByQuest(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 submissions, got %d", len(subs))
	}
}

func TestCreativeSubmissionStore_ListByCrew(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]CreativeSubmission{
		{ID: 1, QuestID: 1, ChallengeID: 1, CrewID: "c1", AuthorUID: "u1", Kind: "STORY", Content: "s1", Status: "PENDING", CreatedAt: now, UpdatedAt: now},
		{ID: 2, QuestID: 1, ChallengeID: 1, CrewID: "c1", AuthorUID: "u1", Kind: "COMIC", Content: "c1", Status: "PENDING", CreatedAt: now, UpdatedAt: now},
	})
	store := NewCreativeSubmissionStore(&mockSupabaseClient{data: data})

	subs, err := store.ListByCrew(context.Background(), "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 submissions, got %d", len(subs))
	}
}

func TestCreativeSubmissionStore_GetSubmission(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]CreativeSubmission{
		{ID: 1, QuestID: 1, ChallengeID: 1, CrewID: "c1", AuthorUID: "u1", Kind: "STORY", Content: "s1", Status: "PENDING", CreatedAt: now, UpdatedAt: now},
	})
	store := NewCreativeSubmissionStore(&mockSupabaseClient{data: data})

	sub, err := store.GetSubmission(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Content != "s1" {
		t.Errorf("expected content 's1', got %s", sub.Content)
	}
}

func TestCreativeSubmissionStore_GetSubmission_NotFound(t *testing.T) {
	data, _ := json.Marshal([]CreativeSubmission{})
	store := NewCreativeSubmissionStore(&mockSupabaseClient{data: data})

	_, err := store.GetSubmission(context.Background(), 999)
	if err != game.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreativeSubmissionStore_UpdateSubmission(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]CreativeSubmission{
		{ID: 1, QuestID: 1, ChallengeID: 1, CrewID: "c1", AuthorUID: "u1", Kind: "STORY", Content: "s1", Status: "PENDING", CreatedAt: now, UpdatedAt: now},
	})
	store := NewCreativeSubmissionStore(&mockSupabaseClient{data: data})

	reviewedAt := now.Add(time.Hour)
	err := store.UpdateSubmission(context.Background(), 1, map[string]any{
		"status":      string(game.SubmissionStatusApproved),
		"reviewed_by": "reviewer-1",
		"reviewed_at": reviewedAt,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
