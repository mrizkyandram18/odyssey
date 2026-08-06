package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"odyssey/pkg/game/content"
)

const publishedFilter = "published=eq.true"
const notDeletedFilter = "deleted_at=is.null"

func publishedParams() string {
	v := url.Values{}
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	return v.Encode()
}

func appendPublishedFilter(params string) string {
	v := url.Values{}
	if params != "" {
		parsed, err := url.ParseQuery(params)
		if err == nil {
			for k, vals := range parsed {
				for _, val := range vals {
					v.Set(k, val)
				}
			}
		}
	}
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	return v.Encode()
}

type supabaseRealmDefinitionStore struct {
	client SupabaseClient
}

func NewRealmDefinitionStore(client SupabaseClient) content.RealmDefinitionStore {
	return &supabaseRealmDefinitionStore{client: client}
}

func (s *supabaseRealmDefinitionStore) ListRealms(ctx context.Context) ([]content.RealmDefinition, error) {
	raw, err := s.client.Get(ctx, "odyssey_realm_definitions", publishedParams())
	if err != nil {
		return nil, fmt.Errorf("list realms: %w", err)
	}
	var dbDefs []RealmDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse realm definitions: %w", err)
	}
	result := make([]content.RealmDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		result = append(result, *mapRealmDefinition(dbDefs[i]))
	}
	return result, nil
}

func (s *supabaseRealmDefinitionStore) GetRealm(ctx context.Context, slug string) (*content.RealmDefinition, error) {
	v := url.Values{}
	v.Set("slug", "eq."+slug)
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_realm_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get realm: %w", err)
	}
	var defs []RealmDefinition
	if err := json.Unmarshal(raw, &defs); err != nil {
		return nil, fmt.Errorf("parse realm definition: %w", err)
	}
	if len(defs) == 0 {
		return nil, nil
	}
	return mapRealmDefinition(defs[0]), nil
}

func mapRealmDefinition(d RealmDefinition) *content.RealmDefinition {
	return &content.RealmDefinition{
		ID:          d.ID,
		Slug:        d.Slug,
		Name:        d.Name,
		Description: d.Description,
		Order:       d.Order,
		MaxProgress: d.MaxProgress,
		Icon:        d.Icon,
		Published:   d.Published,
		Version:     d.Version,
		UpdatedBy:   d.UpdatedBy,
		DeletedAt:   d.DeletedAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

type supabaseChapterDefinitionStore struct {
	client SupabaseClient
}

func NewChapterDefinitionStore(client SupabaseClient) content.ChapterDefinitionStore {
	return &supabaseChapterDefinitionStore{client: client}
}

func (s *supabaseChapterDefinitionStore) ListChapters(ctx context.Context, realm string) ([]content.ChapterDefinition, error) {
	v := url.Values{}
	if realm != "" {
		v.Set("realm", "eq."+realm)
	}
	v.Set("order", "order")
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_chapter_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}
	var dbDefs []ChapterDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse chapter definitions: %w", err)
	}
	result := make([]content.ChapterDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		result = append(result, *mapChapterDefinition(dbDefs[i]))
	}
	return result, nil
}

func (s *supabaseChapterDefinitionStore) GetChapter(ctx context.Context, slug string) (*content.ChapterDefinition, error) {
	v := url.Values{}
	v.Set("slug", "eq."+slug)
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_chapter_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get chapter: %w", err)
	}
	var defs []ChapterDefinition
	if err := json.Unmarshal(raw, &defs); err != nil {
		return nil, fmt.Errorf("parse chapter definition: %w", err)
	}
	if len(defs) == 0 {
		return nil, nil
	}
	return mapChapterDefinition(defs[0]), nil
}

func mapChapterDefinition(d ChapterDefinition) *content.ChapterDefinition {
	var publishedAt time.Time
	if d.PublishedAt != nil {
		publishedAt = *d.PublishedAt
	}
	return &content.ChapterDefinition{
		ID:          d.ID,
		Slug:        d.Slug,
		Realm:       d.Realm,
		Title:       d.Title,
		Description: d.Description,
		Order:       d.Order,
		Published:   d.Published,
		Version:     d.Version,
		UpdatedBy:   d.UpdatedBy,
		PublishedAt: publishedAt,
		DeletedAt:   d.DeletedAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

type supabaseQuestDefinitionStore struct {
	client SupabaseClient
}

func NewQuestDefinitionStore(client SupabaseClient) content.QuestDefinitionStore {
	return &supabaseQuestDefinitionStore{client: client}
}

func (s *supabaseQuestDefinitionStore) ListQuests(ctx context.Context) ([]content.QuestDefinition, error) {
	raw, err := s.client.Get(ctx, "odyssey_quest_definitions", publishedParams())
	if err != nil {
		return nil, fmt.Errorf("list quests: %w", err)
	}
	var dbDefs []QuestDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse quest definitions: %w", err)
	}
	result := make([]content.QuestDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		qd, err := mapQuestDefinition(dbDefs[i])
		if err != nil {
			return nil, fmt.Errorf("map quest %s: %w", dbDefs[i].Slug, err)
		}
		result = append(result, *qd)
	}
	return result, nil
}

func (s *supabaseQuestDefinitionStore) GetQuest(ctx context.Context, slug string) (*content.QuestDefinition, error) {
	v := url.Values{}
	v.Set("slug", "eq."+slug)
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_quest_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get quest: %w", err)
	}
	var defs []QuestDefinition
	if err := json.Unmarshal(raw, &defs); err != nil {
		return nil, fmt.Errorf("parse quest definition: %w", err)
	}
	if len(defs) == 0 {
		return nil, nil
	}
	return mapQuestDefinition(defs[0])
}

func (s *supabaseQuestDefinitionStore) ListQuestsByRealm(ctx context.Context, realm string) ([]content.QuestDefinition, error) {
	v := url.Values{}
	v.Set("realm", "eq."+realm)
	v.Set("order", "order")
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_quest_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("list quests by realm: %w", err)
	}
	var dbDefs []QuestDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse quest definitions: %w", err)
	}
	result := make([]content.QuestDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		qd, err := mapQuestDefinition(dbDefs[i])
		if err != nil {
			return nil, fmt.Errorf("map quest %s: %w", dbDefs[i].Slug, err)
		}
		result = append(result, *qd)
	}
	return result, nil
}

func mapQuestDefinition(d QuestDefinition) (*content.QuestDefinition, error) {
	var challengeDefs []content.ChallengeDef
	if len(d.ChallengeDefs) > 0 {
		if err := json.Unmarshal(d.ChallengeDefs, &challengeDefs); err != nil {
			return nil, fmt.Errorf("parse challenge defs: %w", err)
		}
	}
	var publishedAt time.Time
	if d.PublishedAt != nil {
		publishedAt = *d.PublishedAt
	}
	var requiredQuestSlugs []string
	if d.RequiredQuestSlugs != nil {
		requiredQuestSlugs = d.RequiredQuestSlugs
	} else if d.RequiredQuestSlug != "" {
		requiredQuestSlugs = []string{d.RequiredQuestSlug}
	}
	return &content.QuestDefinition{
		ID:                 d.ID,
		Slug:               d.Slug,
		Realm:              d.Realm,
		Chapter:            d.Chapter,
		Title:              d.Title,
		Description:        d.Description,
		QuestType:          d.QuestType,
		ChallengeDefs:      challengeDefs,
		RewardXP:           d.RewardXP,
		RewardChest:        d.RewardChest,
		IsMandatory:        d.IsMandatory,
		RequiredQuestSlug:  d.RequiredQuestSlug,
		RequiredQuestSlugs: requiredQuestSlugs,
		RequiredChapter:    d.RequiredChapter,
		RequiredRealm:      d.RequiredRealm,
		RequiredLevel:      d.RequiredLevel,
		SeasonSlug:         d.SeasonSlug,
		Published:          d.Published,
		Version:            d.Version,
		UpdatedBy:          d.UpdatedBy,
		PublishedAt:        publishedAt,
		DeletedAt:          d.DeletedAt,
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
	}, nil
}

type supabaseCreativePromptStore struct {
	client SupabaseClient
}

func NewCreativePromptStore(client SupabaseClient) content.CreativePromptStore {
	return &supabaseCreativePromptStore{client: client}
}

func (s *supabaseCreativePromptStore) ListPrompts(ctx context.Context) ([]content.CreativePromptDefinition, error) {
	raw, err := s.client.Get(ctx, "odyssey_creative_prompt_definitions", publishedParams())
	if err != nil {
		return nil, fmt.Errorf("list prompts: %w", err)
	}
	var dbDefs []CreativePromptDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse creative prompt definitions: %w", err)
	}
	result := make([]content.CreativePromptDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		result = append(result, *mapCreativePromptDefinition(dbDefs[i]))
	}
	return result, nil
}

func (s *supabaseCreativePromptStore) GetPrompt(ctx context.Context, slug string) (*content.CreativePromptDefinition, error) {
	v := url.Values{}
	v.Set("slug", "eq."+slug)
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_creative_prompt_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get prompt: %w", err)
	}
	var defs []CreativePromptDefinition
	if err := json.Unmarshal(raw, &defs); err != nil {
		return nil, fmt.Errorf("parse creative prompt definition: %w", err)
	}
	if len(defs) == 0 {
		return nil, nil
	}
	return mapCreativePromptDefinition(defs[0]), nil
}

func (s *supabaseCreativePromptStore) ListPromptsByRealm(ctx context.Context, realm string) ([]content.CreativePromptDefinition, error) {
	v := url.Values{}
	v.Set("realm", "eq."+realm)
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_creative_prompt_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("list prompts by realm: %w", err)
	}
	var dbDefs []CreativePromptDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse creative prompt definitions: %w", err)
	}
	result := make([]content.CreativePromptDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		result = append(result, *mapCreativePromptDefinition(dbDefs[i]))
	}
	return result, nil
}

func mapCreativePromptDefinition(d CreativePromptDefinition) *content.CreativePromptDefinition {
	return &content.CreativePromptDefinition{
		ID:          d.ID,
		Slug:        d.Slug,
		Realm:       d.Realm,
		Title:       d.Title,
		Description: d.Description,
		Prompt:      d.Prompt,
		Kind:        d.Kind,
		SeasonSlug:  d.SeasonSlug,
		Published:   d.Published,
		Version:     d.Version,
		UpdatedBy:   d.UpdatedBy,
		DeletedAt:   d.DeletedAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

type supabaseAchievementDefinitionStore struct {
	client SupabaseClient
}

func NewAchievementDefinitionStore(client SupabaseClient) content.AchievementDefinitionStore {
	return &supabaseAchievementDefinitionStore{client: client}
}

func (s *supabaseAchievementDefinitionStore) ListAchievements(ctx context.Context) ([]content.AchievementDefinition, error) {
	raw, err := s.client.Get(ctx, "odyssey_achievement_definitions", publishedParams())
	if err != nil {
		return nil, fmt.Errorf("list achievements: %w", err)
	}
	var dbDefs []AchievementDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse achievement definitions: %w", err)
	}
	result := make([]content.AchievementDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		result = append(result, *mapAchievementDefinition(dbDefs[i]))
	}
	return result, nil
}

func (s *supabaseAchievementDefinitionStore) GetAchievement(ctx context.Context, code string) (*content.AchievementDefinition, error) {
	v := url.Values{}
	v.Set("code", "eq."+code)
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_achievement_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get achievement: %w", err)
	}
	var defs []AchievementDefinition
	if err := json.Unmarshal(raw, &defs); err != nil {
		return nil, fmt.Errorf("parse achievement definition: %w", err)
	}
	if len(defs) == 0 {
		return nil, nil
	}
	return mapAchievementDefinition(defs[0]), nil
}

func mapAchievementDefinition(d AchievementDefinition) *content.AchievementDefinition {
	return &content.AchievementDefinition{
		ID:          d.ID,
		Code:        d.Code,
		Title:       d.Title,
		Description: d.Description,
		Kind:        d.Kind,
		Trigger:     d.Trigger,
		Threshold:   d.Threshold,
		RewardXP:    d.RewardXP,
		RewardRelic: d.RewardRelic,
		SeasonSlug:  d.SeasonSlug,
		Published:   d.Published,
		Version:     d.Version,
		UpdatedBy:   d.UpdatedBy,
		DeletedAt:   d.DeletedAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

type supabaseSeasonDefinitionStore struct {
	client SupabaseClient
}

func NewSeasonDefinitionStore(client SupabaseClient) content.SeasonDefinitionStore {
	return &supabaseSeasonDefinitionStore{client: client}
}

func (s *supabaseSeasonDefinitionStore) ListSeasons(ctx context.Context) ([]content.SeasonDefinition, error) {
	raw, err := s.client.Get(ctx, "odyssey_season_definitions", publishedParams())
	if err != nil {
		return nil, fmt.Errorf("list seasons: %w", err)
	}
	var dbDefs []SeasonDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse season definitions: %w", err)
	}
	result := make([]content.SeasonDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		result = append(result, *mapSeasonDefinition(dbDefs[i]))
	}
	return result, nil
}

func (s *supabaseSeasonDefinitionStore) GetSeason(ctx context.Context, slug string) (*content.SeasonDefinition, error) {
	v := url.Values{}
	v.Set("slug", "eq."+slug)
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_season_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get season: %w", err)
	}
	var defs []SeasonDefinition
	if err := json.Unmarshal(raw, &defs); err != nil {
		return nil, fmt.Errorf("parse season definition: %w", err)
	}
	if len(defs) == 0 {
		return nil, nil
	}
	return mapSeasonDefinition(defs[0]), nil
}

func mapSeasonDefinition(d SeasonDefinition) *content.SeasonDefinition {
	return &content.SeasonDefinition{
		ID:          d.ID,
		Slug:        d.Slug,
		Name:        d.Name,
		Description: d.Description,
		StartAt:     d.StartAt,
		EndAt:       d.EndAt,
		Realm:       d.Realm,
		Published:   d.Published,
		Version:     d.Version,
		UpdatedBy:   d.UpdatedBy,
		DeletedAt:   d.DeletedAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

type supabaseLoreDefinitionStore struct {
	client SupabaseClient
}

func NewLoreDefinitionStore(client SupabaseClient) content.LoreDefinitionStore {
	return &supabaseLoreDefinitionStore{client: client}
}

func (s *supabaseLoreDefinitionStore) ListLore(ctx context.Context) ([]content.LoreDefinition, error) {
	raw, err := s.client.Get(ctx, "odyssey_lore_definitions", publishedParams())
	if err != nil {
		return nil, fmt.Errorf("list lore: %w", err)
	}
	var dbDefs []LoreDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse lore definitions: %w", err)
	}
	result := make([]content.LoreDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		result = append(result, *mapLoreDefinition(dbDefs[i]))
	}
	return result, nil
}

func (s *supabaseLoreDefinitionStore) GetLore(ctx context.Context, slug string) (*content.LoreDefinition, error) {
	v := url.Values{}
	v.Set("slug", "eq."+slug)
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_lore_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get lore: %w", err)
	}
	var defs []LoreDefinition
	if err := json.Unmarshal(raw, &defs); err != nil {
		return nil, fmt.Errorf("parse lore definition: %w", err)
	}
	if len(defs) == 0 {
		return nil, nil
	}
	return mapLoreDefinition(defs[0]), nil
}

func (s *supabaseLoreDefinitionStore) ListLoreByRealm(ctx context.Context, realm string) ([]content.LoreDefinition, error) {
	v := url.Values{}
	v.Set("realm", "eq."+realm)
	v.Set("order", "order")
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
	raw, err := s.client.Get(ctx, "odyssey_lore_definitions", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("list lore by realm: %w", err)
	}
	var dbDefs []LoreDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse lore definitions: %w", err)
	}
	result := make([]content.LoreDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		result = append(result, *mapLoreDefinition(dbDefs[i]))
	}
	return result, nil
}

func mapLoreDefinition(d LoreDefinition) *content.LoreDefinition {
	return &content.LoreDefinition{
		ID:         d.ID,
		Slug:       d.Slug,
		Realm:      d.Realm,
		Chapter:    d.Chapter,
		Title:      d.Title,
		Content:    d.Content,
		Order:      d.Order,
		SeasonSlug: d.SeasonSlug,
		Published:  d.Published,
		Version:    d.Version,
		UpdatedBy:  d.UpdatedBy,
		DeletedAt:  d.DeletedAt,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
}

var _ content.RealmDefinitionStore = (*supabaseRealmDefinitionStore)(nil)
var _ content.ChapterDefinitionStore = (*supabaseChapterDefinitionStore)(nil)
var _ content.QuestDefinitionStore = (*supabaseQuestDefinitionStore)(nil)
var _ content.CreativePromptStore = (*supabaseCreativePromptStore)(nil)
var _ content.AchievementDefinitionStore = (*supabaseAchievementDefinitionStore)(nil)
var _ content.SeasonDefinitionStore = (*supabaseSeasonDefinitionStore)(nil)
var _ content.LoreDefinitionStore = (*supabaseLoreDefinitionStore)(nil)
