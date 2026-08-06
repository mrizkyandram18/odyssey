package content

import (
	"context"
	"time"
)

// RealmDefinitionStore provides persistence for realm definitions.
type RealmDefinitionStore interface {
	ListRealms(ctx context.Context) ([]RealmDefinition, error)
	GetRealm(ctx context.Context, slug string) (*RealmDefinition, error)
}

// ChapterDefinitionStore provides persistence for chapter definitions.
type ChapterDefinitionStore interface {
	ListChapters(ctx context.Context, realm string) ([]ChapterDefinition, error)
	GetChapter(ctx context.Context, slug string) (*ChapterDefinition, error)
}

// QuestDefinitionStore provides persistence for quest definitions.
type QuestDefinitionStore interface {
	ListQuests(ctx context.Context) ([]QuestDefinition, error)
	GetQuest(ctx context.Context, slug string) (*QuestDefinition, error)
	ListQuestsByRealm(ctx context.Context, realm string) ([]QuestDefinition, error)
}

// CreativePromptStore provides persistence for creative prompt definitions.
type CreativePromptStore interface {
	ListPrompts(ctx context.Context) ([]CreativePromptDefinition, error)
	GetPrompt(ctx context.Context, slug string) (*CreativePromptDefinition, error)
	ListPromptsByRealm(ctx context.Context, realm string) ([]CreativePromptDefinition, error)
}

// AchievementDefinitionStore provides persistence for achievement definitions.
type AchievementDefinitionStore interface {
	ListAchievements(ctx context.Context) ([]AchievementDefinition, error)
	GetAchievement(ctx context.Context, code string) (*AchievementDefinition, error)
}

// SeasonDefinitionStore provides persistence for season definitions.
type SeasonDefinitionStore interface {
	ListSeasons(ctx context.Context) ([]SeasonDefinition, error)
	GetSeason(ctx context.Context, slug string) (*SeasonDefinition, error)
}

// LoreDefinitionStore provides persistence for lore definitions.
type LoreDefinitionStore interface {
	ListLore(ctx context.Context) ([]LoreDefinition, error)
	GetLore(ctx context.Context, slug string) (*LoreDefinition, error)
	ListLoreByRealm(ctx context.Context, realm string) ([]LoreDefinition, error)
}

// Repository groups all content store interfaces for convenient
// dependency injection, mirroring the pattern in pkg/game/store.go.
type Repository struct {
	Realms       RealmDefinitionStore
	Chapters     ChapterDefinitionStore
	Quests       QuestDefinitionStore
	Prompts      CreativePromptStore
	Achievements AchievementDefinitionStore
	Seasons      SeasonDefinitionStore
	Lore         LoreDefinitionStore
}

// RealmDefinition describes a themed area of the shared world.
type RealmDefinition struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Order       int        `json:"order"`
	MaxProgress int        `json:"max_progress"`
	Icon        string     `json:"icon"`
	Published   bool       `json:"published"`
	Version     int        `json:"version"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	PublishedAt time.Time  `json:"published_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ChapterDefinition describes a story chapter within a Realm.
type ChapterDefinition struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Realm       string     `json:"realm"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Order       int        `json:"order"`
	Published   bool       `json:"published"`
	Version     int        `json:"version"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	PublishedAt time.Time  `json:"published_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// QuestDefinition is a reusable quest template definition.
type QuestDefinition struct {
	ID                 int64          `json:"id"`
	Slug               string         `json:"slug"`
	Realm              string         `json:"realm"`
	Chapter            string         `json:"chapter"`
	Title              string         `json:"title"`
	Description        string         `json:"description"`
	QuestType          string         `json:"quest_type"`
	ChallengeDefs      []ChallengeDef `json:"challenge_defs"`
	RewardXP           int64          `json:"reward_xp"`
	RewardChest        string         `json:"reward_chest"`
	IsMandatory        bool           `json:"is_mandatory"`
	RequiredQuestSlug  string         `json:"required_quest_slug,omitempty"`
	RequiredQuestSlugs []string       `json:"required_quest_slugs,omitempty"`
	RequiredChapter    string         `json:"required_chapter,omitempty"`
	RequiredRealm      string         `json:"required_realm,omitempty"`
	RequiredLevel      int            `json:"required_level,omitempty"`
	SeasonSlug         string         `json:"season_slug,omitempty"`
	Published          bool           `json:"published"`
	Version            int            `json:"version"`
	UpdatedBy          string         `json:"updated_by,omitempty"`
	PublishedAt        time.Time      `json:"published_at"`
	DeletedAt          *time.Time     `json:"deleted_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

// ChallengeDef describes a single challenge within a quest definition.
type ChallengeDef struct {
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// CreativePromptDefinition describes a creative prompt for a creative quest.
type CreativePromptDefinition struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Realm       string     `json:"realm"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Prompt      string     `json:"prompt"`
	Kind        string     `json:"kind"`
	SeasonSlug  string     `json:"season_slug,omitempty"`
	Published   bool       `json:"published"`
	Version     int        `json:"version"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	PublishedAt time.Time  `json:"published_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ChestDefinition is the admin-managed template for a chest type.
type ChestDefinition struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	Rarity      string     `json:"rarity"`
	Icon        string     `json:"icon"`
	Description string     `json:"description"`
	SeasonSlug  string     `json:"season_slug,omitempty"`
	Published   bool       `json:"published"`
	Version     int        `json:"version"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	PublishedAt time.Time  `json:"published_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// DropTableEntry is a single rarity-weight pair for a chest definition.
type DropTableEntry struct {
	ID        int64     `json:"id"`
	ChestSlug string    `json:"chest_slug"`
	RelicID   int64     `json:"relic_id,omitempty"`
	Rarity    string    `json:"rarity,omitempty"`
	Weight    float64   `json:"weight"`
	Published bool      `json:"published"`
	Version   int       `json:"version"`
	UpdatedBy string    `json:"updated_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// RelicDefinition is the admin-managed template for a relic.
type RelicDefinition struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Realm       string     `json:"realm"`
	Rarity      string     `json:"rarity"`
	Image       string     `json:"image"`
	Lore        string     `json:"lore"`
	Published   bool       `json:"published"`
	Version     int        `json:"version"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	PublishedAt time.Time  `json:"published_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// AchievementDefinition describes a milestone definition.
type AchievementDefinition struct {
	ID          int64      `json:"id"`
	Code        string     `json:"code"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Kind        string     `json:"kind"`
	Trigger     string     `json:"trigger"`
	Threshold   int        `json:"threshold"`
	RewardXP    int64      `json:"reward_xp"`
	RewardRelic string     `json:"reward_relic"`
	SeasonSlug  string     `json:"season_slug,omitempty"`
	Published   bool       `json:"published"`
	Version     int        `json:"version"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	PublishedAt time.Time  `json:"published_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// SeasonDefinition describes a time-bounded progression arc.
type SeasonDefinition struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	StartAt     time.Time  `json:"start_at"`
	EndAt       time.Time  `json:"end_at"`
	Realm       string     `json:"realm"`
	Published   bool       `json:"published"`
	Version     int        `json:"version"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	PublishedAt time.Time  `json:"published_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// LoreDefinition describes a piece of narrative lore.
type LoreDefinition struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Realm       string     `json:"realm"`
	Chapter     string     `json:"chapter"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Order       int        `json:"order"`
	SeasonSlug  string     `json:"season_slug,omitempty"`
	Published   bool       `json:"published"`
	Version     int        `json:"version"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	PublishedAt time.Time  `json:"published_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
