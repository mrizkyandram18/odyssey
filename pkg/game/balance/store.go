package balance

import (
	"context"
	"errors"
)

var ErrConfigNotFound = errors.New("balance config not found")

// ConfigKey identifies a balance parameter that can be overridden at runtime.
type ConfigKey string

const (
	KeyXPPerLevel                ConfigKey = "xp_per_level"
	KeyMaxNewQuestsPerDay        ConfigKey = "max_new_quests_per_day"
	KeyChallengeXP               ConfigKey = "challenge_xp"
	KeyCompletionBonusXP         ConfigKey = "completion_bonus_xp"
	KeyDropRate                  ConfigKey = "drop_rate_multiplier"
	KeyDailyTurnXP               ConfigKey = "daily_turn_xp"
	KeyRealmProgressPerQuest     ConfigKey = "realm_progress_per_quest"
	KeyRealmCompletionThreshold  ConfigKey = "realm_completion_threshold"
	KeyAchievementThreshold      ConfigKey = "achievement_threshold_multiplier"
	KeyQuestRewardXP             ConfigKey = "quest_reward_xp_multiplier"
	KeyChestRewardCountCommon    ConfigKey = "chest_reward_count_common"
	KeyChestRewardCountUncommon  ConfigKey = "chest_reward_count_uncommon"
	KeyChestRewardCountRare      ConfigKey = "chest_reward_count_rare"
	KeyChestRewardCountEpic      ConfigKey = "chest_reward_count_epic"
	KeyChestRewardCountLegendary ConfigKey = "chest_reward_count_legendary"
)

// Override holds a single runtime balance override loaded from the DB.
type Override struct {
	Key       string `json:"key"`
	Value     int64  `json:"value"`
	UpdatedBy string `json:"updated_by,omitempty"`
}

// Store provides read access to balance overrides from the database.
type Store interface {
	GetOverride(ctx context.Context, key string) (*Override, error)
	ListOverrides(ctx context.Context) ([]Override, error)
}
