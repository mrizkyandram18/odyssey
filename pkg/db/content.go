package db

import (
	"encoding/json"
	"time"
)

type RealmDefinition struct {
	ID          int64           `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Order       int             `json:"order"`
	MaxProgress int             `json:"max_progress"`
	Icon        string          `json:"icon"`
	Published   bool            `json:"published"`
	Draft       json.RawMessage `json:"draft"`
	Version     int             `json:"version"`
	UpdatedBy   string          `json:"updated_by"`
	PublishedAt *time.Time      `json:"published_at"`
	DeletedAt   *time.Time      `json:"deleted_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ChapterDefinition struct {
	ID          int64           `json:"id"`
	Slug        string          `json:"slug"`
	Realm       string          `json:"realm"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Order       int             `json:"order"`
	Published   bool            `json:"published"`
	Draft       json.RawMessage `json:"draft"`
	Version     int             `json:"version"`
	UpdatedBy   string          `json:"updated_by"`
	PublishedAt *time.Time      `json:"published_at"`
	DeletedAt   *time.Time      `json:"deleted_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type QuestDefinition struct {
	ID                 int64           `json:"id"`
	Slug               string          `json:"slug"`
	Realm              string          `json:"realm"`
	Chapter            string          `json:"chapter"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	QuestType          string          `json:"quest_type"`
	ChallengeDefs      json.RawMessage `json:"challenge_defs"`
	RewardXP           int64           `json:"reward_xp"`
	RewardChest        string          `json:"reward_chest"`
	IsMandatory        bool            `json:"is_mandatory"`
	RequiredQuestSlug  string          `json:"required_quest_slug"`
	RequiredQuestSlugs []string        `json:"required_quest_slugs"`
	RequiredChapter    string          `json:"required_chapter"`
	RequiredRealm      string          `json:"required_realm"`
	RequiredLevel      int             `json:"required_level"`
	SeasonSlug         string          `json:"season_slug"`
	Published          bool            `json:"published"`
	Draft              json.RawMessage `json:"draft"`
	Version            int             `json:"version"`
	UpdatedBy          string          `json:"updated_by"`
	PublishedAt        *time.Time      `json:"published_at"`
	DeletedAt          *time.Time      `json:"deleted_at"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type CreativePromptDefinition struct {
	ID          int64           `json:"id"`
	Slug        string          `json:"slug"`
	Realm       string          `json:"realm"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Prompt      string          `json:"prompt"`
	Kind        string          `json:"kind"`
	SeasonSlug  string          `json:"season_slug"`
	Published   bool            `json:"published"`
	Draft       json.RawMessage `json:"draft"`
	Version     int             `json:"version"`
	UpdatedBy   string          `json:"updated_by"`
	PublishedAt *time.Time      `json:"published_at"`
	DeletedAt   *time.Time      `json:"deleted_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type AchievementDefinition struct {
	ID               int64           `json:"id"`
	Code             string          `json:"code"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	Kind             string          `json:"kind"`
	Trigger          string          `json:"trigger"`
	Threshold        int             `json:"threshold"`
	RewardXP         int64           `json:"reward_xp"`
	RewardRelic      string          `json:"reward_relic"`
	RewardCosmeticID string          `json:"reward_cosmetic_id"`
	SeasonSlug       string          `json:"season_slug"`
	Published        bool            `json:"published"`
	Draft            json.RawMessage `json:"draft"`
	Version          int             `json:"version"`
	UpdatedBy        string          `json:"updated_by"`
	PublishedAt      *time.Time      `json:"published_at"`
	DeletedAt        *time.Time      `json:"deleted_at"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type SeasonDefinition struct {
	ID          int64           `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	StartAt     time.Time       `json:"start_at"`
	EndAt       time.Time       `json:"end_at"`
	Realm       string          `json:"realm"`
	Published   bool            `json:"published"`
	Draft       json.RawMessage `json:"draft"`
	Version     int             `json:"version"`
	UpdatedBy   string          `json:"updated_by"`
	PublishedAt *time.Time      `json:"published_at"`
	DeletedAt   *time.Time      `json:"deleted_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type LoreDefinition struct {
	ID          int64           `json:"id"`
	Slug        string          `json:"slug"`
	Realm       string          `json:"realm"`
	Chapter     string          `json:"chapter"`
	Title       string          `json:"title"`
	Content     string          `json:"content"`
	Order       int             `json:"order"`
	SeasonSlug  string          `json:"season_slug"`
	Published   bool            `json:"published"`
	Draft       json.RawMessage `json:"draft"`
	Version     int             `json:"version"`
	UpdatedBy   string          `json:"updated_by"`
	PublishedAt *time.Time      `json:"published_at"`
	DeletedAt   *time.Time      `json:"deleted_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
