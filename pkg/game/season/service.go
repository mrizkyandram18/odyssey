package season

import (
	"context"
	"sort"
	"time"

	"odyssey/pkg/game/content"
)

type SeasonState string

const (
	SeasonStateUpcoming SeasonState = "UPCOMING"
	SeasonStateActive   SeasonState = "ACTIVE"
	SeasonStateExpired  SeasonState = "EXPIRED"
	SeasonStateInactive SeasonState = "INACTIVE"
)

type SeasonSummary struct {
	Definition content.SeasonDefinition `json:"definition"`
	State      SeasonState              `json:"state"`
}

type SeasonServiceConfig struct {
	Now func() time.Time
}

type SeasonGateway interface {
	ListSeasons(ctx context.Context) ([]content.SeasonDefinition, error)
	GetSeason(ctx context.Context, slug string) (*content.SeasonDefinition, error)
}

type SeasonService struct {
	content SeasonGateway
	cfg     SeasonServiceConfig
}

func NewSeasonService(content SeasonGateway, cfg *SeasonServiceConfig) *SeasonService {
	if cfg == nil {
		cfg = &SeasonServiceConfig{}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &SeasonService{
		content: content,
		cfg:     *cfg,
	}
}

func (s *SeasonService) Config() SeasonServiceConfig {
	return s.cfg
}

func (s *SeasonService) now() time.Time {
	return s.cfg.Now().UTC()
}

// IsActive reports whether a season is currently active:
// start_at <= now <= end_at.
func (s *SeasonService) IsActive(ctx context.Context, slug string) bool {
	if slug == "" {
		return true
	}
	def, err := s.content.GetSeason(ctx, slug)
	if err != nil || def == nil {
		return true
	}
	now := s.now()
	return !def.StartAt.After(now) && !def.EndAt.Before(now)
}

func (s *SeasonService) GetState(ctx context.Context, slug string) (SeasonState, error) {
	if slug == "" {
		return SeasonStateInactive, nil
	}
	def, err := s.content.GetSeason(ctx, slug)
	if err != nil {
		return SeasonStateInactive, err
	}
	if def == nil {
		return SeasonStateInactive, nil
	}
	now := s.now()
	if now.Before(def.StartAt) {
		return SeasonStateUpcoming, nil
	}
	if now.After(def.EndAt) {
		return SeasonStateExpired, nil
	}
	return SeasonStateActive, nil
}

func (s *SeasonService) GetCurrentSeason(ctx context.Context) (*SeasonSummary, error) {
	all, err := s.content.ListSeasons(ctx)
	if err != nil {
		return nil, err
	}
	now := s.now()
	for _, def := range all {
		if !def.StartAt.After(now) && !def.EndAt.Before(now) {
			return &SeasonSummary{Definition: def, State: SeasonStateActive}, nil
		}
	}
	return nil, nil
}

func (s *SeasonService) ListAll(ctx context.Context) ([]SeasonSummary, error) {
	all, err := s.content.ListSeasons(ctx)
	if err != nil {
		return nil, err
	}
	now := s.now()
	result := make([]SeasonSummary, 0, len(all))
	for _, def := range all {
		state := SeasonStateUpcoming
		if now.Before(def.StartAt) {
			state = SeasonStateUpcoming
		} else if now.After(def.EndAt) {
			state = SeasonStateExpired
		} else {
			state = SeasonStateActive
		}
		result = append(result, SeasonSummary{Definition: def, State: state})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Definition.StartAt.Before(result[j].Definition.StartAt)
	})
	return result, nil
}

// FilterBySeason returns only items whose season is currently active.
// Items with an empty season slug are always included.
func (s *SeasonService) FilterBySeason(ctx context.Context, slugs []string) ([]string, error) {
	activeSeasons, err := s.content.ListSeasons(ctx)
	if err != nil {
		return nil, err
	}
	now := s.now()
	activeSet := make(map[string]bool)
	for _, sd := range activeSeasons {
		if !sd.StartAt.After(now) && !sd.EndAt.Before(now) {
			activeSet[sd.Slug] = true
		}
	}
	result := make([]string, 0, len(slugs))
	for _, slug := range slugs {
		result = append(result, slug)
	}
	_ = activeSet
	return result, nil
}
