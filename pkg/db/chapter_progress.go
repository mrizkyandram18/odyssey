package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"odyssey/pkg/game"
)

type supabaseChapterProgressStore struct {
	client SupabaseClient
}

func NewChapterProgressStore(client SupabaseClient) game.ChapterProgressStore {
	return &supabaseChapterProgressStore{client: client}
}

func (s *supabaseChapterProgressStore) GetChapterProgress(ctx context.Context, crewID, chapter string) (*game.ChapterProgress, error) {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	v.Set("chapter", "eq."+chapter)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_chapter_progress", params)
	if err != nil {
		return nil, fmt.Errorf("get chapter progress: %w", err)
	}

	var rows []ChapterProgress
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse chapter progress: %w", err)
	}
	if len(rows) == 0 {
		return nil, game.ErrNotFound
	}
	return mapChapterProgress(rows[0]), nil
}

func (s *supabaseChapterProgressStore) CreateChapterProgress(ctx context.Context, cp *game.ChapterProgress) (*game.ChapterProgress, error) {
	payload := ChapterProgress{
		CrewID:      cp.CrewID,
		Chapter:     cp.Chapter,
		Realm:       cp.Realm,
		Status:      cp.Status,
		CompletedAt: cp.CompletedAt,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_chapter_progress", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create chapter progress: %w", err)
	}

	var rows []ChapterProgress
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse created chapter progress: %w", err)
	}
	if len(rows) == 0 {
		return cp, nil
	}
	return mapChapterProgress(rows[0]), nil
}

func (s *supabaseChapterProgressStore) UpdateChapterProgress(ctx context.Context, crewID, chapter string, patch map[string]any) error {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	v.Set("chapter", "eq."+chapter)
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_chapter_progress", patch, params)
	if err != nil {
		return fmt.Errorf("update chapter progress: %w", err)
	}
	return nil
}

func (s *supabaseChapterProgressStore) ListChapterProgressByCrew(ctx context.Context, crewID string) ([]game.ChapterProgress, error) {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_chapter_progress", params)
	if err != nil {
		return nil, fmt.Errorf("list chapter progress: %w", err)
	}

	var rows []ChapterProgress
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse chapter progress: %w", err)
	}

	result := make([]game.ChapterProgress, 0, len(rows))
	for i := range rows {
		result = append(result, *mapChapterProgress(rows[i]))
	}
	return result, nil
}

func mapChapterProgress(cp ChapterProgress) *game.ChapterProgress {
	return &game.ChapterProgress{
		CrewID:      cp.CrewID,
		Chapter:     cp.Chapter,
		Realm:       cp.Realm,
		Status:      cp.Status,
		CompletedAt: cp.CompletedAt,
		CreatedAt:   cp.CreatedAt,
		UpdatedAt:   cp.UpdatedAt,
	}
}

var _ game.ChapterProgressStore = (*supabaseChapterProgressStore)(nil)
