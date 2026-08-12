package board

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"odyssey/pkg/game"
)

// Slice 2.3 shared text board — append-only multi-entry posts per crew.
// Not a real-time collaborative editor (no OT/CRDT/live cursors).

const (
	MaxPayloadRunes = 2000
	MinPayloadRunes = 1
)

var (
	ErrEmptyContent   = errors.New("text is required")
	ErrContentTooLong = errors.New("text is too long")
	ErrNotFound       = errors.New("board post not found")
	ErrForbidden      = errors.New("cross-crew access denied")
)

// Service manages the shared crew text board.
type Service struct {
	items game.CreativeStore
}

func NewService(items game.CreativeStore) *Service {
	return &Service{items: items}
}

// PostText appends a SHARED_TEXT entry for the caller's crew.
func (s *Service) PostText(ctx context.Context, crewID, authorUID, content string) (*game.CreativeItem, error) {
	if crewID == "" || authorUID == "" {
		return nil, fmt.Errorf("crew and author required")
	}
	text := strings.TrimSpace(content)
	n := utf8.RuneCountInString(text)
	if n < MinPayloadRunes {
		return nil, ErrEmptyContent
	}
	if n > MaxPayloadRunes {
		return nil, ErrContentTooLong
	}

	item := &game.CreativeItem{
		FamilyID:    crewID,
		Journey:     game.RealmSharedBoard,
		AuthorUID: authorUID,
		Kind:      game.KindSharedText,
		Payload:   text,
	}
	created, err := s.items.CreateCreativeItem(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("create board post: %w", err)
	}
	return created, nil
}

// ListForCrew returns SHARED_TEXT posts for one crew only (newest first).
func (s *Service) ListForCrew(ctx context.Context, crewID string) ([]game.CreativeItem, error) {
	if crewID == "" {
		return nil, fmt.Errorf("crew required")
	}
	items, err := s.items.ListCreativeItemsByCrew(ctx, crewID, game.KindSharedText)
	if err != nil {
		return nil, fmt.Errorf("list board posts: %w", err)
	}
	return items, nil
}

// GetForCrew returns a post if it belongs to the crew (isolation check).
func (s *Service) GetForCrew(ctx context.Context, crewID string, id int64) (*game.CreativeItem, error) {
	item, err := s.items.GetCreativeItem(ctx, id)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if item.FamilyID != crewID || item.Kind != game.KindSharedText {
		return nil, ErrForbidden
	}
	return item, nil
}
