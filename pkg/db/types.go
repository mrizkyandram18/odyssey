package db

import (
	"encoding/json"
	"time"
)

type UserProfile struct {
	UID          string    `json:"uid"`
	CrewID       string    `json:"crew_id"`
	ExplorerName string    `json:"explorer_name"`
	Role         string    `json:"role"`
	Level        int       `json:"level"`
	XP           int64     `json:"xp"`
	Version      int       `json:"version"`
	PasswordHash string    `json:"-"`
	DeviceID     string    `json:"device_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Crew struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type QuestInstance struct {
	ID           int64      `json:"id"`
	CrewID       string     `json:"crew_id"`
	TemplateSlug string     `json:"template_slug"`
	Title        string     `json:"title"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Challenge struct {
	ID          int64      `json:"id"`
	QuestID     int64      `json:"quest_id"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CompletedBy string     `json:"completed_by,omitempty"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type RealmProgress struct {
	CrewID         string    `json:"crew_id"`
	Realm          string    `json:"realm"`
	Status         string    `json:"status"`
	StoryBranch    string    `json:"story_branch"`
	Progress       int       `json:"progress"`
	LastUnlockedAt time.Time `json:"last_unlocked_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreativeItem struct {
	ID        int64     `json:"id"`
	CrewID    string    `json:"crew_id"`
	Realm     string    `json:"realm"`
	AuthorUID string    `json:"author_uid"`
	Kind      string    `json:"kind"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type CreativeSubmission struct {
	ID              int64      `json:"id"`
	QuestID         int64      `json:"quest_id"`
	ChallengeID     int64      `json:"challenge_id"`
	CrewID          string     `json:"crew_id"`
	AuthorUID       string     `json:"author_uid"`
	Kind            string     `json:"kind"`
	Content         string     `json:"content"`
	Status          string     `json:"status"`
	ReviewedBy      string     `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type DailyTurn struct {
	ID        int64     `json:"id"`
	UID       string    `json:"uid"`
	Date      string    `json:"date"`
	QuestSlug string    `json:"quest_slug"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

type Achievement struct {
	ID              int64     `json:"id"`
	UID             string    `json:"uid"`
	CrewID          string    `json:"crew_id"`
	Code            string    `json:"code"`
	Kind            string    `json:"kind"`
	Trigger         string    `json:"trigger"`
	CompletionCount int       `json:"completion_count"`
	AwardedAt       time.Time `json:"awarded_at"`
	CreatedAt       time.Time `json:"created_at"`
}

type Relic struct {
	ID          int64     `json:"id"`
	UID         string    `json:"uid"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Realm       string    `json:"realm"`
	Rarity      string    `json:"rarity"`
	Image       string    `json:"image"`
	Lore        string    `json:"lore"`
	AwardedAt   time.Time `json:"awarded_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Chest struct {
	ID          int64      `json:"id"`
	UID         string     `json:"uid"`
	Source      string     `json:"source"`
	ChestSlug   string     `json:"chest_slug"`
	Rarity      string     `json:"rarity"`
	Icon        string     `json:"icon"`
	Description string     `json:"description"`
	DropTable   string     `json:"drop_table"`
	Opened      bool       `json:"opened"`
	OpenedAt    *time.Time `json:"opened_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type PlayerRelic struct {
	ID           int64     `json:"id"`
	UID          string    `json:"uid"`
	RelicSlug    string    `json:"relic_slug"`
	RelicID      int64     `json:"relic_id"`
	OwnedCount   int       `json:"owned_count"`
	IsNew        bool      `json:"is_new"`
	DiscoveredAt time.Time `json:"discovered_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ChestDefinition struct {
	ID          int64           `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Rarity      string          `json:"rarity"`
	Icon        string          `json:"icon"`
	Description string          `json:"description"`
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

type DropTableEntry struct {
	ID        int64     `json:"id"`
	ChestSlug string    `json:"chest_slug"`
	RelicID   int64     `json:"relic_id,omitempty"`
	Rarity    string    `json:"rarity,omitempty"`
	Weight    float64   `json:"weight"`
	CreatedAt time.Time `json:"created_at"`
}

type RelicDefinition struct {
	ID          int64           `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Realm       string          `json:"realm"`
	Rarity      string          `json:"rarity"`
	Image       string          `json:"image"`
	Lore        string          `json:"lore"`
	Published   bool            `json:"published"`
	Draft       json.RawMessage `json:"draft"`
	Version     int             `json:"version"`
	UpdatedBy   string          `json:"updated_by"`
	PublishedAt *time.Time      `json:"published_at"`
	DeletedAt   *time.Time      `json:"deleted_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ChapterProgress struct {
	CrewID      string     `json:"crew_id"`
	Chapter     string     `json:"chapter"`
	Realm       string     `json:"realm"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type LoreUnlock struct {
	CrewID     string    `json:"crew_id"`
	LoreSlug   string    `json:"lore_slug"`
	Realm      string    `json:"realm"`
	Chapter    string    `json:"chapter"`
	UnlockedAt time.Time `json:"unlocked_at"`
	CreatedAt  time.Time `json:"created_at"`
}
