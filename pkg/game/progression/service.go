package progression

import (
	"context"
	"fmt"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/balance"
	"odyssey/pkg/game/events"
	"odyssey/pkg/observability"
)

// XPPerLevel is the XP stride between consecutive Explorer Levels.
// Level 1 begins at 0 XP; reaching level N requires (N-1)*XPPerLevel XP.
const XPPerLevel int64 = 500

// XPForLevel returns the total XP required to *reach* a given level.
// Level 1 requires 0 XP. This is a pure function with no side effects.
func XPForLevel(level int) int64 {
	if level < 1 {
		level = 1
	}
	return int64(level-1) * XPPerLevel
}

// LevelFromXP derives the Explorer Level from a total XP value.
// It is the inverse of XPForLevel (rounded down). Minimum level is 1.
// This is a pure function with no side effects.
func LevelFromXP(xp int64) int {
	if xp < 0 {
		xp = 0
	}
	return int(xp/XPPerLevel) + 1
}

// XPToNext returns the additional XP needed to reach the next level.
func XPToNext(xp int64) int64 {
	level := LevelFromXP(xp)
	return XPForLevel(level+1) - xp
}

// ProgressionConfig carries tunable progression parameters loaded from
// environment or system config. All fields have safe defaults.
type ProgressionConfig struct {
	// XPPerLevel is the XP stride between levels (default: 100).
	XPPerLevel int64
	// ChallengeXP is granted per individual challenge completion (default: 20).
	ChallengeXP int64
	// CompletionBonusXP is granted when a quest is completed (default: 60).
	CompletionBonusXP int64
}

// DefaultProgressionConfig returns the standard MVP progression values.
func DefaultProgressionConfig() ProgressionConfig {
	return ProgressionConfig{
		XPPerLevel:        500,
		ChallengeXP:       20,
		CompletionBonusXP: 60,
	}
}

// ProgressionService owns individual Explorer progression: XP accumulation,
// level computation, and level-up detection. It depends only on the
// UserStore interface — never on a concrete db adapter, following the same
// dependency-injection pattern as DailyTurnService.
type ProgressionService struct {
	users     game.UserStore
	cfg       ProgressionConfig
	publisher events.Publisher
	balance   *balance.Service
	metrics   *observability.Metrics
}

// SetMetrics attaches an optional metrics sink for XP and lock-conflict
// telemetry. Safe to call with nil (metrics recording is disabled).
func (s *ProgressionService) SetMetrics(m *observability.Metrics) {
	s.metrics = m
}

func NewProgressionService(users game.UserStore, cfg *ProgressionConfig) *ProgressionService {
	if cfg == nil {
		c := DefaultProgressionConfig()
		cfg = &c
	}
	return &ProgressionService{users: users, cfg: *cfg, publisher: events.NopPublisher{}}
}

func NewProgressionServiceWithPublisher(users game.UserStore, cfg *ProgressionConfig, publisher events.Publisher, balanceSvc ...*balance.Service) *ProgressionService {
	if cfg == nil {
		c := DefaultProgressionConfig()
		cfg = &c
	}
	if publisher == nil {
		publisher = events.NopPublisher{}
	}
	s := &ProgressionService{users: users, cfg: *cfg, publisher: publisher}
	if len(balanceSvc) > 0 && balanceSvc[0] != nil {
		s.balance = balanceSvc[0]
		s.applyBalanceOverrides()
	}
	return s
}

// Config returns the effective progression configuration with balance overrides applied.
func (s *ProgressionService) Config() ProgressionConfig {
	if s.balance != nil {
		def := s.cfg
		if def.XPPerLevel == 0 {
			def.XPPerLevel = 500
		}
		if def.ChallengeXP == 0 {
			def.ChallengeXP = 20
		}
		if def.CompletionBonusXP == 0 {
			def.CompletionBonusXP = 60
		}
		return ProgressionConfig{
			XPPerLevel:        s.balance.OverrideXPForLevel(def.XPPerLevel),
			ChallengeXP:       s.balance.OverrideChallengeXP(def.ChallengeXP),
			CompletionBonusXP: s.balance.OverrideCompletionBonusXP(def.CompletionBonusXP),
		}
	}
	return s.cfg
}

func (s *ProgressionService) applyBalanceOverrides() {
	def := s.cfg
	if def.XPPerLevel == 0 {
		def.XPPerLevel = 500
	}
	if def.ChallengeXP == 0 {
		def.ChallengeXP = 20
	}
	if def.CompletionBonusXP == 0 {
		def.CompletionBonusXP = 60
	}
	s.cfg = ProgressionConfig{
		XPPerLevel:        s.balance.OverrideXPForLevel(def.XPPerLevel),
		ChallengeXP:       s.balance.OverrideChallengeXP(def.ChallengeXP),
		CompletionBonusXP: s.balance.OverrideCompletionBonusXP(def.CompletionBonusXP),
	}
}

// AwardXP grants XP to the player identified by uid via an atomic compare-and-set,
// persisting the updated XP and level. Returns the updated player and whether
// a level-up occurred. The compare-and-set makes this call idempotent and
// protects against lost-update races: re-executing with the same uid and amount
// yields the same result without double-counting.
func (s *ProgressionService) AwardXP(ctx context.Context, uid string, amount int64) (*game.Player, bool, error) {
	player, err := s.users.GetUser(ctx, uid)
	if err != nil {
		return nil, false, fmt.Errorf("award xp: get user: %w", err)
	}
	if player == nil {
		return nil, false, fmt.Errorf("award xp: %w", game.ErrNotFound)
	}

	oldLevel := player.Level
	oldVersion := player.Version
	newXP := player.XP + amount
	newLevel := LevelFromXP(newXP)
	levelUp := newLevel > oldLevel
	if levelUp {
		player.Level = newLevel
	}

	patch := map[string]any{
		"xp":      newXP,
		"level":   player.Level,
		"version": oldVersion + 1,
	}

	matched, err := s.users.UpdateUserIfMatch(ctx, uid, oldVersion, patch)
	if err != nil {
		return nil, false, fmt.Errorf("award xp: update user: %w", err)
	}
	if !matched {
		// Version mismatch: another request updated the user concurrently.
		// Re-read and return current state without double-counting.
		if s.metrics != nil {
			s.metrics.RecordLockConflict()
		}
		updated, err := s.users.GetUser(ctx, uid)
		if err != nil {
			return nil, false, fmt.Errorf("award xp: re-read user after conflict: %w", err)
		}
		return updated, updated.Level > oldLevel, nil
	}

	if s.metrics != nil {
		s.metrics.RecordXP(amount)
	}

	player.XP = newXP
	player.UpdatedAt = time.Now().UTC()
	player.Version = oldVersion + 1

	if levelUp {
		s.publisher.Publish(ctx, events.LevelReachedEvent{
			UID:      uid,
			FamilyID:   player.FamilyID,
			OldLevel: oldLevel,
			NewLevel: newLevel,
		})
	}
	return player, levelUp, nil
}
