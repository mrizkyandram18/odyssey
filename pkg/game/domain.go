package game

import "time"

// Domain entities — pure business types with no infrastructure dependencies.
// These exist independently of any persistence or transport concern.
// The db layer (pkg/db) holds DTOs that a future adapter will map to/from these.

// Player is a single family member's game identity.
// Identity is the shared UID from Gatekeeper; game state lives here.
type Player struct {
	UID          string    `json:"uid"`
	CrewID       string    `json:"crew_id"`
	ExplorerName string    `json:"explorer_name"`
	Role         string    `json:"role"`
	Level        int       `json:"level"`
	XP           int64     `json:"xp"`
	Coins        int64     `json:"coins"`
	Version      int       `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Crew is the family group — the shared party for all progression.
type Crew struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Quest is an instance of a quest template, activated by a specific crew.
// The template_slug references code-embedded definitions (see docs/domain-model.md).
type Quest struct {
	ID           int64      `json:"id"`
	CrewID       string     `json:"crew_id"`
	TemplateSlug string     `json:"template_slug"`
	Title        string     `json:"title"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Challenge is a single task within a Quest.
type Challenge struct {
	ID          int64      `json:"id"`
	QuestID     int64      `json:"quest_id"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	AssignedTo  *string    `json:"assigned_to,omitempty"`
	CompletedBy string     `json:"completed_by,omitempty"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// RealmProgress is the family's shared progress within a specific realm.
// Multiple rows per crew are expected — one per realm.
type RealmProgress struct {
	CrewID         string    `json:"crew_id"`
	Realm          string    `json:"realm"`
	Status         string    `json:"status"`
	StoryBranch    string    `json:"story_branch,omitempty"`
	Progress       int       `json:"progress"`
	LastUnlockedAt time.Time `json:"last_unlocked_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreativeItem is an append-only contribution to a creative space.
type CreativeItem struct {
	ID        int64
	CrewID    string
	Realm     string
	AuthorUID string
	Kind      string
	Payload   string
	CreatedAt time.Time
}

// SubmissionKind is the type of creative submission.
type SubmissionKind string

const (
	SubmissionStory SubmissionKind = "STORY"
	SubmissionComic SubmissionKind = "COMIC"
	SubmissionPhoto SubmissionKind = "PHOTO"
	SubmissionVideo SubmissionKind = "VIDEO"
	SubmissionDrawing SubmissionKind = "DRAWING"
)

// SubmissionStatus is the review state of a creative submission.
type SubmissionStatus string

const (
	SubmissionStatusPending  SubmissionStatus = "PENDING"
	SubmissionStatusApproved SubmissionStatus = "APPROVED"
	SubmissionStatusRejected SubmissionStatus = "REJECTED"
)

// Submission is a creative quest submission awaiting review.
type Submission struct {
	ID              int64
	QuestID         int64
	ChallengeID     int64
	CrewID          string
	AuthorUID       string
	Kind            SubmissionKind
	Content         string
	Status          SubmissionStatus
	ReviewedBy      string
	ReviewedAt      *time.Time
	RejectionReason string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Rarity represents item rarity.
type Rarity string

const (
	RarityCommon    Rarity = "COMMON"
	RarityUncommon  Rarity = "UNCOMMON"
	RarityRare      Rarity = "RARE"
	RarityEpic      Rarity = "EPIC"
	RarityLegendary Rarity = "LEGENDARY"
)

// DailyTurn is one-per-user-per-calendar-day engagement.
type DailyTurn struct {
	ID        int64     `json:"id"`
	UID       string    `json:"uid"`
	Date      string    `json:"date"`
	QuestSlug string    `json:"quest_slug"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// Achievement is a personal or group milestone.
type Achievement struct {
	ID              int64
	UID             string
	CrewID          string
	Code            string
	Kind            string
	Trigger         string
	CompletionCount int
	AwardedAt       time.Time
	CreatedAt       time.Time
}

// Relic is a personal collectible displayed in the crew gallery.
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

// Chest is a reward container with known, fixed contents.
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

// PlayerRelic tracks a player's ownership state for a specific relic.
type PlayerRelic struct {
	UID          string    `json:"uid"`
	RelicSlug    string    `json:"relic_slug"`
	RelicID      int64     `json:"relic_id"`
	OwnedCount   int       `json:"owned_count"`
	IsNew        bool      `json:"is_new"`
	DiscoveredAt time.Time `json:"discovered_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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

// ChapterProgress tracks a crew's progression through a single chapter.
// One row per (crew_id, chapter) — created when the chapter is unlocked.
type ChapterProgress struct {
	CrewID      string     `json:"crew_id"`
	Chapter     string     `json:"chapter"`
	Realm       string     `json:"realm"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// LoreUnlock tracks a crew's discovered lore entries.
// One row per (crew_id, lore_slug) — created when the lore is unlocked.
type LoreUnlock struct {
	CrewID     string    `json:"crew_id"`
	LoreSlug   string    `json:"lore_slug"`
	Realm      string    `json:"realm"`
	Chapter    string    `json:"chapter"`
	UnlockedAt time.Time `json:"unlocked_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// PlayerAchievement tracks an achievement earned by a player.
// Mirrors the existing odyssey_achievements table but adds completion_count
// for achievements with thresholds.
type PlayerAchievement struct {
	ID        int64     `json:"id"`
	UID       string    `json:"uid"`
	CrewID    string    `json:"crew_id"`
	Code      string    `json:"code"`
	Kind      string    `json:"kind"`
	Trigger   string    `json:"trigger"`
	Count     int       `json:"count"`
	AwardedAt time.Time `json:"awarded_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Reaction represents a peer reaction to a crew's action (e.g. quest, submission).
type Reaction struct {
	ID           int64     `json:"id"`
	CrewID       string    `json:"crew_id"`
	TargetType   string    `json:"target_type"`
	TargetID     int64     `json:"target_id"`
	ActorUID     string    `json:"actor_uid"`
	ReactionType string    `json:"reaction_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// DailyActivity tracks a user's activity for streaks and history.
type DailyActivity struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ActivityDate string    `json:"activity_date"`
	ActivityType string    `json:"activity_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// RewardLedger tracks a user's reward history.
type RewardLedger struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Source     string    `json:"source"`
	Amount     int64     `json:"amount"`
	RewardType string    `json:"reward_type"`
	Metadata   *string   `json:"metadata"`
	CreatedAt  time.Time `json:"created_at"`
}
