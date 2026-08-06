package balance

import (
	"context"
	"sync"
	"time"

	"odyssey/pkg/game"
)

// Service provides runtime balance overrides for gameplay parameters.
// All overrides are loaded from the database; no rebuild is required.
// Defaults are used when no override exists.
type Service struct {
	mu        sync.RWMutex
	store     Store
	overrides map[string]int64
	loadedAt  time.Time
}

func NewService(store Store) *Service {
	return &Service{
		store:     store,
		overrides: make(map[string]int64),
	}
}

// Load reads all overrides from the DB and caches them in memory.
func (s *Service) Load(ctx context.Context) error {
	overrides, err := s.store.ListOverrides(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.overrides = make(map[string]int64, len(overrides))
	for _, ov := range overrides {
		s.overrides[ov.Key] = ov.Value
	}
	s.loadedAt = time.Now().UTC()
	return nil
}

// GetOverride returns the override value for a key, falling back to the
// provided default if no override exists.
func (s *Service) GetOverride(key ConfigKey, def int64) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.overrides[string(key)]; ok {
		return v
	}
	return def
}

// Overrides returns a copy of the currently applied overrides keyed by name.
func (s *Service) Overrides() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]int64, len(s.overrides))
	for k, v := range s.overrides {
		out[k] = v
	}
	return out
}

// OverrideXPForLevel returns the XP needed per level.
func (s *Service) OverrideXPForLevel(def int64) int64 {
	return s.GetOverride(KeyXPPerLevel, def)
}

// OverrideDropRateMultiplier returns the drop rate multiplier.
func (s *Service) OverrideDropRateMultiplier(def float64) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.overrides[string(KeyDropRate)]; ok {
		return float64(v) / 100.0
	}
	return def
}

// OverrideDailyTurnXP returns the XP awarded per daily turn.
func (s *Service) OverrideDailyTurnXP(def int64) int64 {
	return s.GetOverride(KeyDailyTurnXP, def)
}

// OverrideRealmProgressPerQuest returns the realm progress per completed quest.
func (s *Service) OverrideRealmProgressPerQuest(def int) int {
	return int(s.GetOverride(KeyRealmProgressPerQuest, int64(def)))
}

// OverrideRealmCompletionThreshold returns the threshold for realm completion.
func (s *Service) OverrideRealmCompletionThreshold(def int) int {
	return int(s.GetOverride(KeyRealmCompletionThreshold, int64(def)))
}

// OverrideChestRewardCount returns the number of rewards for a chest of the
// given rarity, falling back to def if no override exists.
func (s *Service) OverrideChestRewardCount(rarity game.Rarity, def int) int {
	var key ConfigKey
	switch rarity {
	case game.RarityCommon:
		key = KeyChestRewardCountCommon
	case game.RarityUncommon:
		key = KeyChestRewardCountUncommon
	case game.RarityRare:
		key = KeyChestRewardCountRare
	case game.RarityEpic:
		key = KeyChestRewardCountEpic
	case game.RarityLegendary:
		key = KeyChestRewardCountLegendary
	default:
		return def
	}
	return int(s.GetOverride(key, int64(def)))
}

// OverrideAchievementThresholdMultiplier returns the multiplier for achievement thresholds.
func (s *Service) OverrideAchievementThresholdMultiplier(def float64) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.overrides[string(KeyAchievementThreshold)]; ok {
		return float64(v) / 100.0
	}
	return def
}

// OverrideQuestRewardXP returns the XP multiplier for quest rewards.
func (s *Service) OverrideQuestRewardXP(def float64) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.overrides[string(KeyQuestRewardXP)]; ok {
		return float64(v) / 100.0
	}
	return def
}

// OverrideCompletionBonusXP returns the completion bonus XP.
func (s *Service) OverrideCompletionBonusXP(def int64) int64 {
	return s.GetOverride(KeyCompletionBonusXP, def)
}

// OverrideChallengeXP returns the per-challenge XP.
func (s *Service) OverrideChallengeXP(def int64) int64 {
	return s.GetOverride(KeyChallengeXP, def)
}

// Reload re-reads overrides from the DB.
func (s *Service) Reload(ctx context.Context) error {
	return s.Load(ctx)
}

// LoadedAt returns when overrides were last loaded.
func (s *Service) LoadedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadedAt
}
