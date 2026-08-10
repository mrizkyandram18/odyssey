package game

import (
	"context"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

// Repository interfaces — focused, cohesive, single-purpose.
// Each interface follows the Interface Segregation Principle: callers
// depend only on the methods they need.
//
// The db layer (pkg/db) will implement these via a Supabase adapter.
// At runtime, the api/handlers wire a concrete implementation into each
// interactor. For now, these interfaces stand alone — no import of pkg/db.

// UserStore manages player identities.
type UserStore interface {
	GetUser(ctx context.Context, uid string) (*Player, error)
	CreateUser(ctx context.Context, p *Player) error
	UpdateUser(ctx context.Context, uid string, patch map[string]any) error
	UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error)
	ListUsersByCrew(ctx context.Context, crewID string) ([]Player, error)
}

// CrewStore provides persistence operations for crew (family group) data.
type CrewStore interface {
	GetCrew(ctx context.Context, crewID string) (*Crew, error)
	CreateCrew(ctx context.Context, c *Crew) error
}

// QuestStore provides persistence operations for quests and their challenges.
type QuestStore interface {
	GetQuest(ctx context.Context, questID int64) (*Quest, error)
	CreateQuest(ctx context.Context, q *Quest) (*Quest, error)
	ListQuestByCrew(ctx context.Context, crewID string) ([]Quest, error)
	UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error
	UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error)

	GetChallenges(ctx context.Context, questID int64) ([]Challenge, error)
	CreateChallenge(ctx context.Context, c *Challenge) (*Challenge, error)
	UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error
	UpdateChallengeIfMatch(ctx context.Context, challengeID int64, oldStatus string, patch map[string]any) (bool, error)
}

// RealmProgressStore provides persistence for shared realm progress.
// Multiple rows per crew are supported (one per realm).
type RealmProgressStore interface {
	GetRealmProgress(ctx context.Context, crewID, realm string) (*RealmProgress, error)
	CreateRealmProgress(ctx context.Context, rp *RealmProgress) (*RealmProgress, error)
	UpdateRealmProgress(ctx context.Context, crewID, realm string, patch map[string]any) error
	UpdateRealmProgressIfMatch(ctx context.Context, crewID, realm string, oldProgress int, patch map[string]any) (bool, error)
	ListRealmProgressByCrew(ctx context.Context, crewID string) ([]RealmProgress, error)
}

// CreativeStore provides persistence for creative-space contributions.
// Items are append-only.
type CreativeStore interface {
	CreateCreativeItem(ctx context.Context, item *CreativeItem) (*CreativeItem, error)
	GetCreativeItem(ctx context.Context, id int64) (*CreativeItem, error)
	ListCreativeItemsByCrew(ctx context.Context, crewID, kind string) ([]CreativeItem, error)
}

// CreativeSubmissionStore provides persistence for creative quest submissions.
type CreativeSubmissionStore interface {
	CreateSubmission(ctx context.Context, s *Submission) (*Submission, error)
	ListByQuest(ctx context.Context, questID int64) ([]Submission, error)
	ListByCrew(ctx context.Context, crewID string) ([]Submission, error)
	GetSubmission(ctx context.Context, submissionID int64) (*Submission, error)
	UpdateSubmission(ctx context.Context, submissionID int64, patch map[string]any) error
}

// DailyTurnStore provides persistence for daily turn tracking.
type DailyTurnStore interface {
	CreateDailyTurn(ctx context.Context, dt *DailyTurn) (*DailyTurn, error)
	UpdateDailyTurn(ctx context.Context, turnID int64, patch map[string]any) error
	ListDailyTurns(ctx context.Context, uid string) ([]DailyTurn, error)
}

// ProgressionStore provides persistence for rewards and achievements.
type ProgressionStore interface {
	CreateRelic(ctx context.Context, r *Relic) (*Relic, error)
	CreateChest(ctx context.Context, ch *Chest) (*Chest, error)
	UpdateChest(ctx context.Context, chestID int64, patch map[string]any) error
	CreateAchievement(ctx context.Context, a *Achievement) (*Achievement, error)
	CountRelics(ctx context.Context, uid string) (int, error)
}

// ChestStore provides persistence operations for chest instances.
type ChestStore interface {
	CreateChest(ctx context.Context, ch *Chest) (*Chest, error)
	GetChest(ctx context.Context, chestID int64) (*Chest, error)
	UpdateChest(ctx context.Context, chestID int64, patch map[string]any) error
	UpdateChestIfMatch(ctx context.Context, chestID int64, oldOpened bool, patch map[string]any) (bool, error)
	ListChestsByUser(ctx context.Context, uid string) ([]Chest, error)
}

// ChestDefinitionStore provides persistence for chest template definitions.
type ChestDefinitionStore interface {
	ListChestDefinitions(ctx context.Context) ([]ChestDefinition, error)
	GetChestDefinition(ctx context.Context, slug string) (*ChestDefinition, error)
	ListDropTableEntries(ctx context.Context, chestSlug string) ([]DropTableEntry, error)
}

// RelicStore provides persistence for relic definitions and instances.
type RelicStore interface {
	CreateRelic(ctx context.Context, r *Relic) (*Relic, error)
	GetRelic(ctx context.Context, relicID int64) (*Relic, error)
	ListRelics(ctx context.Context) ([]Relic, error)
	CountRelics(ctx context.Context, uid string) (int, error)
}

// RelicDefinitionStore provides persistence for relic template definitions.
type RelicDefinitionStore interface {
	ListRelicDefinitions(ctx context.Context) ([]RelicDefinition, error)
	GetRelicDefinition(ctx context.Context, slug string) (*RelicDefinition, error)
}

// PlayerRelicStore provides persistence for player-owned relic tracking.
type PlayerRelicStore interface {
	GetPlayerRelic(ctx context.Context, uid, relicSlug string) (*PlayerRelic, error)
	CreatePlayerRelic(ctx context.Context, pr *PlayerRelic) (*PlayerRelic, error)
	UpdatePlayerRelic(ctx context.Context, uid, relicSlug string, patch map[string]any) error
	ListPlayerRelics(ctx context.Context, uid string) ([]PlayerRelic, error)
	CountUniqueRelics(ctx context.Context, uid string) (int, error)
}

// ChapterProgressStore provides persistence for crew chapter progress.
type ChapterProgressStore interface {
	GetChapterProgress(ctx context.Context, crewID, chapter string) (*ChapterProgress, error)
	CreateChapterProgress(ctx context.Context, cp *ChapterProgress) (*ChapterProgress, error)
	UpdateChapterProgress(ctx context.Context, crewID, chapter string, patch map[string]any) error
	ListChapterProgressByCrew(ctx context.Context, crewID string) ([]ChapterProgress, error)
}

// LoreUnlockStore provides persistence for crew lore unlocks.
type LoreUnlockStore interface {
	GetLoreUnlock(ctx context.Context, crewID, loreSlug string) (*LoreUnlock, error)
	CreateLoreUnlock(ctx context.Context, lu *LoreUnlock) (*LoreUnlock, error)
	UpdateLoreUnlock(ctx context.Context, crewID, loreSlug string, patch map[string]any) error
	ListLoreUnlocksByCrew(ctx context.Context, crewID string) ([]LoreUnlock, error)
}

// AchievementStore provides persistence for earned achievements.
type AchievementStore interface {
	CreateAchievement(ctx context.Context, a *Achievement) (*Achievement, error)
	GetAchievementByCode(ctx context.Context, uid, code string) (*Achievement, error)
	ListAchievementsByPlayer(ctx context.Context, uid string) ([]Achievement, error)
	CountAchievementsByKind(ctx context.Context, uid string, kind string) (int, error)
}

// ConfigStore provides read access to system configuration.
type ConfigStore interface {
	GetSystemConfig(ctx context.Context, key string) (string, error)
}

// ReactionStore provides persistence for peer reactions.
type ReactionStore interface {
	UpsertReaction(ctx context.Context, r *Reaction) (*Reaction, error)
	ListReactionsForTarget(ctx context.Context, crewID, targetType string, targetID int64) ([]Reaction, error)
}

// ActivityStore provides persistence for daily user activity and streak calculation.
type ActivityStore interface {
	RecordActivity(ctx context.Context, act *DailyActivity) (*DailyActivity, error)
	GetStreak(ctx context.Context, uid string) (int, error)
}

// RewardLedgerStore manages reward history.
type RewardLedgerStore interface {
	CreateLedger(ctx context.Context, ledger *RewardLedger) error
	ListByUser(ctx context.Context, userID string) ([]RewardLedger, error)
}

// CosmeticUnlockStore manages paid cosmetic ownership.
type CosmeticUnlockStore interface {
	ListByUser(ctx context.Context, uid string) ([]CosmeticUnlock, error)
	Has(ctx context.Context, uid, cosmeticID string) (bool, error)
	CreateIfAbsent(ctx context.Context, uid, cosmeticID string, pricePaid int64) (created bool, err error)
	Delete(ctx context.Context, uid, cosmeticID string) error
}

// Repository groups all persistence interfaces for convenient dependency injection.
type Repository struct {
	Users               UserStore
	Crews               CrewStore
	Quests              QuestStore
	RealmProgress       RealmProgressStore
	Creatives           CreativeStore
	CreativeSubmissions CreativeSubmissionStore
	DailyTurns          DailyTurnStore
	Progression         ProgressionStore
	Chests              ChestStore
	ChestDefinitions    ChestDefinitionStore
	Relics              RelicStore
	RelicDefinitions    RelicDefinitionStore
	PlayerRelics        PlayerRelicStore
	ChapterProgress     ChapterProgressStore
	LoreUnlocks         LoreUnlockStore
	Achievements        AchievementStore
	Config              ConfigStore
	Reactions           ReactionStore
	Activity            ActivityStore
	RewardLedgers       RewardLedgerStore
	CosmeticUnlocks     CosmeticUnlockStore
}

// BuildRepository is a package placeholder. Concrete repository construction
// is implemented in pkg/db/wire.go (BuildRepository) to avoid an import cycle.
func BuildRepository(_ any) (*Repository, error) {
	return nil, fmt.Errorf("use pkg/db.BuildRepository for repository construction")
}
