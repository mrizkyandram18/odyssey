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
		Families:               NewCrewStore(client),
		Missions:              NewQuestStore(client),
		JourneyProgress:       NewRealmProgressStore(client),
		Creatives:           NewCreativeStore(client),
		CreativeSubmissions: NewCreativeSubmissionStore(client),
		DailyTurns:          NewDailyTurnStore(client),
		Progression:         NewProgressionStore(client),
		Gifts:              NewChestStore(client),
		ChestDefinitions:    NewChestDefinitionStore(client),
		Collections:              NewRelicStore(client),
		RelicDefinitions:    NewRelicDefinitionStore(client),
		PlayerRelics:        NewPlayerRelicStore(client),
		CourseProgress:     NewChapterProgressStore(client),
		LoreUnlocks:         NewLoreUnlockStore(client),
		Achievements:        NewAchievementStore(client),
		Config:              NewConfigStore(client),
		Reactions:           NewReactionStore(client),
		Activity:            NewActivityStore(client),
		RewardLedgers:       NewRewardLedgerStore(client),
		CosmeticUnlocks:     NewCosmeticUnlockStore(client),
		PushSubscriptions:   NewPushSubscriptionStore(client),
	}, nil
}

// BuildContentRepository constructs a content.Repository from a SupabaseClient,
// wiring each content store to its Supabase-backed implementation.
func BuildContentRepository(client SupabaseClient) (*content.Repository, error) {
	if client == nil {
		return nil, fmt.Errorf("nil supabase client")
	}

	return &content.Repository{
		Journeys:       NewRealmDefinitionStore(client),
		Courses:     NewChapterDefinitionStore(client),
		Missions:       NewQuestDefinitionStore(client),
		Prompts:      NewCreativePromptStore(client),
		Achievements: NewAchievementDefinitionStore(client),
		Seasons:      NewSeasonDefinitionStore(client),
		Concept:         NewLoreDefinitionStore(client),
	}, nil
}
