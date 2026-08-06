package validation

import (
	"testing"
	"time"

	"odyssey/pkg/game/content"
)

func TestValidator_ValidContent(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Realms: []content.RealmDefinition{
			{ID: 1, Slug: "forest", Name: "Enchanted Forest"},
		},
		Chapters: []content.ChapterDefinition{
			{ID: 1, Slug: "ch1", Realm: "forest", Title: "Chapter 1", Order: 1},
			{ID: 2, Slug: "ch2", Realm: "forest", Title: "Chapter 2", Order: 2},
		},
		Quests: []content.QuestDefinition{
			{ID: 1, Slug: "quest1", Realm: "forest", Chapter: "ch1", Title: "Quest 1"},
		},
		Seasons: []content.SeasonDefinition{
			{ID: 1, Slug: "s1", Realm: "forest", Name: "Season 1",
				StartAt: time.Now().Add(-24 * time.Hour),
				EndAt:   time.Now().Add(24 * time.Hour)},
		},
		Chests: []content.ChestDefinition{
			{ID: 1, Slug: "chest1", Name: "Chest 1", Rarity: "common"},
		},
		Relics: []content.RelicDefinition{
			{ID: 1, Slug: "relic1", Name: "Relic 1", Realm: "forest", Rarity: "rare"},
		},
		Achievements: []content.AchievementDefinition{
			{ID: 1, Code: "ACH001", Title: "Test Achievement", Threshold: 10},
		},
		Lore: []content.LoreDefinition{
			{ID: 1, Slug: "lore1", Title: "Lore 1", Realm: "forest", Chapter: "ch1"},
		},
		Prompts: []content.CreativePromptDefinition{
			{ID: 1, Slug: "prompt1", Prompt: "Write a story", Realm: "forest"},
		},
	}

	result := v.Validate(cs)
	if !result.Valid {
		t.Errorf("expected valid result, got %d errors: %v", len(result.Errors), result.Errors)
	}
}

func TestValidator_DuplicateSlug(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Realms: []content.RealmDefinition{
			{ID: 1, Slug: "forest", Name: "Forest"},
			{ID: 2, Slug: "forest", Name: "Duplicate"},
		},
	}
	result := v.Validate(cs)
	if result.Valid {
		t.Error("expected invalid result for duplicate slug")
	}
	found := false
	for _, e := range result.Errors {
		if e.Code == "DUP_SLUG" {
			found = true
		}
	}
	if !found {
		t.Error("expected DUP_SLUG error")
	}
}

func TestValidator_BrokenPrerequisite(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Realms: []content.RealmDefinition{
			{ID: 1, Slug: "forest", Name: "Forest"},
		},
		Chapters: []content.ChapterDefinition{
			{ID: 1, Slug: "ch1", Realm: "forest", Title: "Chapter 1", Order: 1},
		},
		Quests: []content.QuestDefinition{
			{ID: 1, Slug: "quest1", Realm: "forest", Chapter: "ch1",
				RequiredQuestSlug: "nonexistent", RequiredChapter: "ch1"},
		},
	}
	result := v.Validate(cs)
	if result.Valid {
		t.Error("expected invalid result for broken prerequisite")
	}
	found := false
	for _, e := range result.Errors {
		if e.Code == "BROKEN_PREREQ" {
			found = true
		}
	}
	if !found {
		t.Error("expected BROKEN_PREREQ error")
	}
}

func TestValidator_BrokenPrerequisite_MultiSlug(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Realms: []content.RealmDefinition{
			{ID: 1, Slug: "forest", Name: "Forest"},
		},
		Chapters: []content.ChapterDefinition{
			{ID: 1, Slug: "ch1", Realm: "forest", Title: "Chapter 1", Order: 1},
		},
		Quests: []content.QuestDefinition{
			{ID: 1, Slug: "quest1", Realm: "forest", Chapter: "ch1",
				RequiredQuestSlugs: []string{"qA", "nonexistent"}},
		},
	}
	result := v.Validate(cs)
	if result.Valid {
		t.Error("expected invalid result for broken multi prerequisite")
	}
	found := false
	for _, e := range result.Errors {
		if e.Code == "BROKEN_PREREQ" {
			found = true
		}
	}
	if !found {
		t.Error("expected BROKEN_PREREQ error for multi slug")
	}
}

func TestValidator_MissingRealm(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Quests: []content.QuestDefinition{
			{ID: 1, Slug: "quest1", Realm: "unknown", Title: "Quest"},
		},
	}
	result := v.Validate(cs)
	if result.Valid {
		t.Error("expected invalid result for missing realm")
	}
	found := false
	for _, e := range result.Errors {
		if e.Code == "MISSING_REALM" {
			found = true
		}
	}
	if !found {
		t.Error("expected MISSING_REALM error")
	}
}

func TestValidator_QuestPrereqCycle(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Realms: []content.RealmDefinition{
			{ID: 1, Slug: "forest", Name: "Forest"},
		},
		Quests: []content.QuestDefinition{
			{ID: 1, Slug: "qA", Realm: "forest", RequiredQuestSlug: "qB"},
			{ID: 2, Slug: "qB", Realm: "forest", RequiredQuestSlug: "qA"},
		},
	}
	result := v.Validate(cs)
	if result.Valid {
		t.Error("expected invalid result for quest prerequisite cycle")
	}
	found := false
	for _, e := range result.Errors {
		if e.Code == "CHAPTER_LOOP" {
			found = true
		}
	}
	if !found {
		t.Error("expected CHAPTER_LOOP error")
	}
}

func TestValidator_QuestPrereqCycle_MultiSlug(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Realms: []content.RealmDefinition{
			{ID: 1, Slug: "forest", Name: "Forest"},
		},
		Quests: []content.QuestDefinition{
			{ID: 1, Slug: "qA", Realm: "forest", RequiredQuestSlugs: []string{"qB", "qC"}},
			{ID: 2, Slug: "qB", Realm: "forest", RequiredQuestSlugs: []string{"qC"}},
			{ID: 3, Slug: "qC", Realm: "forest", RequiredQuestSlugs: []string{"qA"}},
		},
	}
	result := v.Validate(cs)
	if result.Valid {
		t.Error("expected invalid result for multi-slug quest prerequisite cycle")
	}
	found := false
	for _, e := range result.Errors {
		if e.Code == "CHAPTER_LOOP" {
			found = true
		}
	}
	if !found {
		t.Error("expected CHAPTER_LOOP error for multi-slug cycle")
	}
}

func TestValidator_SeasonOverlap(t *testing.T) {
	v := NewValidator()
	now := time.Now().UTC()
	cs := ContentSet{
		Realms: []content.RealmDefinition{
			{ID: 1, Slug: "forest", Name: "Forest"},
		},
		Seasons: []content.SeasonDefinition{
			{ID: 1, Slug: "s1", Realm: "forest", Name: "S1",
				StartAt: now.Add(-24 * time.Hour), EndAt: now.Add(24 * time.Hour)},
			{ID: 2, Slug: "s2", Realm: "forest", Name: "S2",
				StartAt: now.Add(-12 * time.Hour), EndAt: now.Add(12 * time.Hour)},
		},
	}
	result := v.Validate(cs)
	found := false
	for _, e := range result.Warnings {
		if e.Code == "SEASON_OVERLAP" {
			found = true
		}
	}
	if !found {
		t.Error("expected SEASON_OVERLAP warning")
	}
}

func TestValidator_InvalidDropWeight(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		DropTables: []content.DropTableEntry{
			{ChestSlug: "chest1", Rarity: "common", Weight: 0},
			{ChestSlug: "chest1", Rarity: "rare", Weight: -1},
		},
	}
	result := v.Validate(cs)
	if result.Valid {
		t.Error("expected invalid result for invalid drop weights")
	}
	count := 0
	for _, e := range result.Errors {
		if e.Code == "INVALID_DROP_WEIGHT" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 INVALID_DROP_WEIGHT errors, got %d", count)
	}
}

func TestValidator_InvalidAchievementThreshold(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Achievements: []content.AchievementDefinition{
			{ID: 1, Code: "ACH0", Title: "Bad", Threshold: 0},
			{ID: 2, Code: "ACH1", Title: "Good", Threshold: 5},
		},
	}
	result := v.Validate(cs)
	found := false
	for _, e := range result.Errors {
		if e.Code == "INVALID_THRESHOLD" {
			found = true
		}
	}
	if !found {
		t.Error("expected INVALID_THRESHOLD error")
	}
}

func TestValidator_InvalidChestDefinition(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Chests: []content.ChestDefinition{
			{ID: 1, Slug: "", Name: "", Rarity: "invalid"},
		},
	}
	result := v.Validate(cs)
	found := false
	for _, e := range result.Errors {
		if e.Code == "INVALID_CHEST" || e.Code == "INVALID_RARITY" {
			found = true
		}
	}
	if !found {
		t.Error("expected INVALID_CHEST or INVALID_RARITY error")
	}
}

func TestValidator_MissingRelicReference(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Relics: []content.RelicDefinition{
			{ID: 1, Slug: "relic1", Name: "Relic 1", Realm: "forest", Rarity: "rare"},
		},
		DropTables: []content.DropTableEntry{
			{ChestSlug: "chest1", RelicID: 999, Weight: 1.0},
		},
	}
	result := v.Validate(cs)
	found := false
	for _, e := range result.Errors {
		if e.Code == "MISSING_RELIC" {
			found = true
		}
	}
	if !found {
		t.Error("expected MISSING_RELIC error")
	}
}

func TestValidator_ZeroTotalWeight(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		DropTables: []content.DropTableEntry{
			{ChestSlug: "chest1", Rarity: "common", Weight: 0},
		},
	}
	result := v.Validate(cs)
	found := false
	for _, e := range result.Errors {
		if e.Code == "ZERO_TOTAL_WEIGHT" {
			found = true
		}
	}
	if !found {
		t.Error("expected ZERO_TOTAL_WEIGHT error")
	}
}

func TestValidator_DuplicateDropTableEntry(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		DropTables: []content.DropTableEntry{
			{ChestSlug: "chest1", RelicID: 1, Weight: 0.5},
			{ChestSlug: "chest1", RelicID: 1, Weight: 0.5},
		},
	}
	result := v.Validate(cs)
	found := false
	for _, e := range result.Errors {
		if e.Code == "DUPLICATE_DROP_ENTRY" {
			found = true
		}
	}
	if !found {
		t.Error("expected DUPLICATE_DROP_ENTRY error")
	}
}

func TestValidator_InvalidRarity(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		DropTables: []content.DropTableEntry{
			{ChestSlug: "chest1", Rarity: "invalid-rarity", Weight: 1.0},
		},
	}
	result := v.Validate(cs)
	found := false
	for _, e := range result.Errors {
		if e.Code == "INVALID_RARITY" {
			found = true
		}
	}
	if !found {
		t.Error("expected INVALID_RARITY error")
	}
}

func TestValidator_OrphanDropTable(t *testing.T) {
	v := NewValidator()
	cs := ContentSet{
		Chests: []content.ChestDefinition{
			{ID: 1, Slug: "chest1", Name: "Chest 1", Rarity: "common"},
		},
		DropTables: []content.DropTableEntry{
			{ChestSlug: "unknown-chest", Rarity: "common", Weight: 1.0},
		},
	}
	result := v.Validate(cs)
	found := false
	for _, e := range result.Warnings {
		if e.Code == "ORPHAN_CONTENT" {
			found = true
		}
	}
	if !found {
		t.Error("expected ORPHAN_CONTENT warning")
	}
}
