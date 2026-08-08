package db

import (
	"fmt"

	"odyssey/pkg/game"
	"odyssey/pkg/game/content"
)

// BuildRepository constructs a game.Repository from a SupabaseClient,
// wiring each persistence interface to its Supabase-backed implementation.
func BuildRepository(client SupabaseClient) (*game.Repository, error) {
	if client == nil {
		return nil, fmt.Errorf("nil supabase client")
	}

	return &game.Repository{
		Users:               NewUserStore(client),
		Crews:               NewCrewStore(client),
		Quests:              NewQuestStore(client),
		RealmProgress:       NewRealmProgressStore(client),
		Creatives:           NewCreativeStore(client),
		CreativeSubmissions: NewCreativeSubmissionStore(client),
		DailyTurns:          NewDailyTurnStore(client),
		Progression:         NewProgressionStore(client),
		Chests:              NewChestStore(client),
		ChestDefinitions:    NewChestDefinitionStore(client),
		Relics:              NewRelicStore(client),
		RelicDefinitions:    NewRelicDefinitionStore(client),
		PlayerRelics:        NewPlayerRelicStore(client),
		ChapterProgress:     NewChapterProgressStore(client),
		LoreUnlocks:         NewLoreUnlockStore(client),
		Achievements:        NewAchievementStore(client),
		Config:              NewConfigStore(client),
		Reactions:           NewReactionStore(client),
		Activity:            NewActivityStore(client),
		RewardLedgers:       NewRewardLedgerStore(client),
	}, nil
}

// BuildContentRepository constructs a content.Repository from a SupabaseClient,
// wiring each content store to its Supabase-backed implementation.
func BuildContentRepository(client SupabaseClient) (*content.Repository, error) {
	if client == nil {
		return nil, fmt.Errorf("nil supabase client")
	}

	return &content.Repository{
		Realms:       NewRealmDefinitionStore(client),
		Chapters:     NewChapterDefinitionStore(client),
		Quests:       NewQuestDefinitionStore(client),
		Prompts:      NewCreativePromptStore(client),
		Achievements: NewAchievementDefinitionStore(client),
		Seasons:      NewSeasonDefinitionStore(client),
		Lore:         NewLoreDefinitionStore(client),
	}, nil
}
