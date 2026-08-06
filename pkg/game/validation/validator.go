package validation

import (
	"fmt"

	"odyssey/pkg/game/content"
)

// ErrorLevel distinguishes validation severities.
type ErrorLevel string

const (
	LevelError   ErrorLevel = "ERROR"
	LevelWarning ErrorLevel = "WARNING"
)

// ValidationIssue is a single validation finding.
type ValidationIssue struct {
	Level    ErrorLevel `json:"level"`
	Code     string     `json:"code"`
	Message  string     `json:"message"`
	Resource string     `json:"resource,omitempty"`
	Field    string     `json:"field,omitempty"`
}

func (vi ValidationIssue) String() string {
	if vi.Resource != "" {
		return fmt.Sprintf("[%s] %s: %s (resource=%s, field=%s)",
			vi.Level, vi.Code, vi.Message, vi.Resource, vi.Field)
	}
	return fmt.Sprintf("[%s] %s: %s", vi.Level, vi.Code, vi.Message)
}

// ValidationResult is the outcome of validating all content.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationIssue `json:"errors"`
	Warnings []ValidationIssue `json:"warnings"`
}

// ContentSet holds all loaded content definitions for validation.
type ContentSet struct {
	Realms       []content.RealmDefinition
	Chapters     []content.ChapterDefinition
	Quests       []content.QuestDefinition
	Prompts      []content.CreativePromptDefinition
	Achievements []content.AchievementDefinition
	Seasons      []content.SeasonDefinition
	Lore         []content.LoreDefinition
	Chests       []content.ChestDefinition
	Relics       []content.RelicDefinition
	DropTables   []content.DropTableEntry
}

func NewResult() *ValidationResult {
	return &ValidationResult{}
}

func (r *ValidationResult) AddError(code, message, resource, field string) {
	r.Errors = append(r.Errors, ValidationIssue{
		Level:    LevelError,
		Code:     code,
		Message:  message,
		Resource: resource,
		Field:    field,
	})
}

func (r *ValidationResult) AddWarning(code, message, resource, field string) {
	r.Warnings = append(r.Warnings, ValidationIssue{
		Level:    LevelWarning,
		Code:     code,
		Message:  message,
		Resource: resource,
		Field:    field,
	})
}

func (r *ValidationResult) Finalize() {
	r.Valid = len(r.Errors) == 0
}

// Validator checks content definitions for integrity violations.
type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// Validate runs all checks against the provided content set.
func (v *Validator) Validate(cs ContentSet) *ValidationResult {
	result := NewResult()

	v.checkDuplicateSlug(cs, result)
	v.checkBrokenPrerequisite(cs, result)
	v.checkMissingRealm(cs, result)
	v.checkQuestPrereqCycle(cs, result)
	v.checkSeasonOverlap(cs, result)
	v.checkInvalidDropWeight(cs, result)
	v.checkAchievementThreshold(cs, result)
	v.checkUnusedContent(cs, result)
	v.checkChestDefinitionValidity(cs, result)
	v.checkRelicReferenceIntegrity(cs, result)
	v.checkWeightTotals(cs, result)
	v.checkDuplicateDropTableEntries(cs, result)
	v.checkInvalidRarity(cs, result)
	v.checkOrphanContent(cs, result)

	result.Finalize()
	return result
}

// checkDuplicateSlug scans each definition set for duplicate slugs/codes.
func (v *Validator) checkDuplicateSlug(cs ContentSet, result *ValidationResult) {
	seen := make(map[string]string)

	for _, r := range cs.Realms {
		key := "realm:" + r.Slug
		if _, ok := seen[key]; ok {
			result.AddError("DUP_SLUG",
				fmt.Sprintf("duplicate realm slug %q", r.Slug), "realm", "slug")
		}
		seen[key] = r.Slug
	}

	for _, c := range cs.Chapters {
		key := "chapter:" + c.Slug
		if _, ok := seen[key]; ok {
			result.AddError("DUP_SLUG",
				fmt.Sprintf("duplicate chapter slug %q", c.Slug), "chapter", "slug")
		}
		seen[key] = c.Slug
	}

	for _, q := range cs.Quests {
		key := "quest:" + q.Slug
		if _, ok := seen[key]; ok {
			result.AddError("DUP_SLUG",
				fmt.Sprintf("duplicate quest slug %q", q.Slug), "quest", "slug")
		}
		seen[key] = q.Slug
	}

	for _, p := range cs.Prompts {
		key := "prompt:" + p.Slug
		if _, ok := seen[key]; ok {
			result.AddError("DUP_SLUG",
				fmt.Sprintf("duplicate prompt slug %q", p.Slug), "prompt", "slug")
		}
		seen[key] = p.Slug
	}

	for _, l := range cs.Lore {
		key := "lore:" + l.Slug
		if _, ok := seen[key]; ok {
			result.AddError("DUP_SLUG",
				fmt.Sprintf("duplicate lore slug %q", l.Slug), "lore", "slug")
		}
		seen[key] = l.Slug
	}

	for _, a := range cs.Achievements {
		key := "achievement:" + a.Code
		if _, ok := seen[key]; ok {
			result.AddError("DUP_SLUG",
				fmt.Sprintf("duplicate achievement code %q", a.Code), "achievement", "code")
		}
		seen[key] = a.Code
	}

	for _, s := range cs.Seasons {
		key := "season:" + s.Slug
		if _, ok := seen[key]; ok {
			result.AddError("DUP_SLUG",
				fmt.Sprintf("duplicate season slug %q", s.Slug), "season", "slug")
		}
		seen[key] = s.Slug
	}

	for _, c := range cs.Chests {
		key := "chest:" + c.Slug
		if _, ok := seen[key]; ok {
			result.AddError("DUP_SLUG",
				fmt.Sprintf("duplicate chest slug %q", c.Slug), "chest", "slug")
		}
		seen[key] = c.Slug
	}

	for _, r := range cs.Relics {
		key := "relic:" + r.Slug
		if _, ok := seen[key]; ok {
			result.AddError("DUP_SLUG",
				fmt.Sprintf("duplicate relic slug %q", r.Slug), "relic", "slug")
		}
		seen[key] = r.Slug
	}
}

// checkBrokenPrerequisite checks quest prerequisites for dangling references.
func (v *Validator) checkBrokenPrerequisite(cs ContentSet, result *ValidationResult) {
	questSlugs := make(map[string]bool)
	for _, q := range cs.Quests {
		questSlugs[q.Slug] = true
	}

	chapterSlugs := make(map[string]bool)
	for _, c := range cs.Chapters {
		chapterSlugs[c.Slug] = true
	}

	realmSlugs := make(map[string]bool)
	for _, r := range cs.Realms {
		realmSlugs[r.Slug] = true
	}

	for _, q := range cs.Quests {
		prereqs := q.RequiredQuestSlugs
		if len(prereqs) == 0 && q.RequiredQuestSlug != "" {
			prereqs = []string{q.RequiredQuestSlug}
		}
		for _, prereq := range prereqs {
			if prereq != "" && !questSlugs[prereq] {
				result.AddError("BROKEN_PREREQ",
					fmt.Sprintf("quest %q references unknown required quest %q",
						q.Slug, prereq), "quest", "required_quest_slug")
			}
		}
		if q.RequiredChapter != "" && !chapterSlugs[q.RequiredChapter] {
			result.AddError("BROKEN_PREREQ",
				fmt.Sprintf("quest %q references unknown chapter %q",
					q.Slug, q.RequiredChapter), "quest", "required_chapter")
		}
		if q.RequiredRealm != "" && !realmSlugs[q.RequiredRealm] {
			result.AddError("BROKEN_PREREQ",
				fmt.Sprintf("quest %q references unknown realm %q",
					q.Slug, q.RequiredRealm), "quest", "required_realm")
		}
	}
}

// checkMissingRealm verifies that all content references point to existing realms.
func (v *Validator) checkMissingRealm(cs ContentSet, result *ValidationResult) {
	realmSlugs := make(map[string]bool)
	for _, r := range cs.Realms {
		realmSlugs[r.Slug] = true
	}

	for _, c := range cs.Chapters {
		if !realmSlugs[c.Realm] {
			result.AddError("MISSING_REALM",
				fmt.Sprintf("chapter %q references unknown realm %q",
					c.Slug, c.Realm), "chapter", "realm")
		}
	}
	for _, q := range cs.Quests {
		if !realmSlugs[q.Realm] {
			result.AddError("MISSING_REALM",
				fmt.Sprintf("quest %q references unknown realm %q",
					q.Slug, q.Realm), "quest", "realm")
		}
	}
	for _, p := range cs.Prompts {
		if !realmSlugs[p.Realm] {
			result.AddError("MISSING_REALM",
				fmt.Sprintf("prompt %q references unknown realm %q",
					p.Slug, p.Realm), "prompt", "realm")
		}
	}
	for _, l := range cs.Lore {
		if !realmSlugs[l.Realm] {
			result.AddError("MISSING_REALM",
				fmt.Sprintf("lore %q references unknown realm %q",
					l.Slug, l.Realm), "lore", "realm")
		}
	}
	for _, r := range cs.Relics {
		if !realmSlugs[r.Realm] {
			result.AddError("MISSING_REALM",
				fmt.Sprintf("relic %q references unknown realm %q",
					r.Slug, r.Realm), "relic", "realm")
		}
	}
	for _, s := range cs.Seasons {
		if !realmSlugs[s.Realm] {
			result.AddError("MISSING_REALM",
				fmt.Sprintf("season %q references unknown realm %q",
					s.Slug, s.Realm), "season", "realm")
		}
	}
}

// checkQuestPrereqCycle detects circular quest prerequisite chains using DFS.
func (v *Validator) checkQuestPrereqCycle(cs ContentSet, result *ValidationResult) {
	slugMap := make(map[string]content.QuestDefinition)
	for _, q := range cs.Quests {
		slugMap[q.Slug] = q
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(slug string) bool
	dfs = func(slug string) bool {
		if recStack[slug] {
			return true
		}
		if visited[slug] {
			return false
		}
		visited[slug] = true
		recStack[slug] = true

		def, ok := slugMap[slug]
		if !ok {
			recStack[slug] = false
			return false
		}

		prereqs := def.RequiredQuestSlugs
		if len(prereqs) == 0 && def.RequiredQuestSlug != "" {
			prereqs = []string{def.RequiredQuestSlug}
		}
		for _, req := range prereqs {
			if _, exists := slugMap[req]; exists {
				if dfs(req) {
					return true
				}
			}
		}
		recStack[slug] = false
		return false
	}

	for slug := range slugMap {
		if dfs(slug) {
			result.AddError("CHAPTER_LOOP",
				fmt.Sprintf("circular quest prerequisite chain involving %q",
					slug), "quest", "required_quest_slug")
		}
	}
}

// checkSeasonOverlap detects seasons that overlap in time for the same realm.
func (v *Validator) checkSeasonOverlap(cs ContentSet, result *ValidationResult) {
	for i, s1 := range cs.Seasons {
		for j := i + 1; j < len(cs.Seasons); j++ {
			s2 := cs.Seasons[j]
			if s1.Realm != s2.Realm || s1.Slug == s2.Slug {
				continue
			}
			if s1.StartAt.Before(s2.EndAt) && s2.StartAt.Before(s1.EndAt) {
				result.AddWarning("SEASON_OVERLAP",
					fmt.Sprintf("season %q overlaps with season %q in realm %q",
						s1.Slug, s2.Slug, s1.Realm), "season", "start_at")
			}
		}
	}
}

// checkInvalidDropWeight checks for non-positive drop table weights.
func (v *Validator) checkInvalidDropWeight(cs ContentSet, result *ValidationResult) {
	for _, dt := range cs.DropTables {
		if dt.Weight <= 0 {
			result.AddError("INVALID_DROP_WEIGHT",
				fmt.Sprintf("drop table entry for chest %q rarity %q has weight %.4f (must be > 0)",
					dt.ChestSlug, dt.Rarity, dt.Weight), "drop_table", "weight")
		}
	}
}

// checkAchievementThreshold verifies achievement thresholds are positive.
func (v *Validator) checkAchievementThreshold(cs ContentSet, result *ValidationResult) {
	for _, a := range cs.Achievements {
		if a.Threshold < 1 {
			result.AddError("INVALID_THRESHOLD",
				fmt.Sprintf("achievement %q has threshold %d (must be >= 1)",
					a.Code, a.Threshold), "achievement", "threshold")
		}
	}
}

// checkUnusedContent flags content definitions that are not referenced
// by any quest, achievement, or other content.
func (v *Validator) checkUnusedContent(cs ContentSet, result *ValidationResult) {
	// Check for quests that are not part of any chapter
	chapterTitles := make(map[string]bool)
	for _, c := range cs.Chapters {
		chapterTitles[c.Slug] = true
	}
	for _, q := range cs.Quests {
		if q.Chapter != "" && !chapterTitles[q.Chapter] {
			result.AddWarning("UNUSED_CONTENT",
				fmt.Sprintf("quest %q references chapter %q that is not defined",
					q.Slug, q.Chapter), "quest", "chapter")
		}
	}

	// Check for lore not assigned to a chapter
	for _, l := range cs.Lore {
		if l.Chapter == "" {
			result.AddWarning("UNUSED_CONTENT",
				fmt.Sprintf("lore %q has no chapter assignment", l.Slug),
				"lore", "chapter")
		}
	}

	// Check for prompts not linked to a realm
	for _, p := range cs.Prompts {
		if p.Realm == "" {
			result.AddWarning("UNUSED_CONTENT",
				fmt.Sprintf("prompt %q has no realm assignment", p.Slug),
				"prompt", "realm")
		}
	}
}

// checkChestDefinitionValidity verifies that all chest definitions have valid fields.
func (v *Validator) checkChestDefinitionValidity(cs ContentSet, result *ValidationResult) {
	validRarities := map[string]bool{
		"common":    true,
		"uncommon":  true,
		"rare":      true,
		"epic":      true,
		"legendary": true,
	}
	for _, c := range cs.Chests {
		if c.Slug == "" {
			result.AddError("INVALID_CHEST",
				fmt.Sprintf("chest definition has empty slug"), "chest", "slug")
		}
		if c.Name == "" {
			result.AddError("INVALID_CHEST",
				fmt.Sprintf("chest %q has empty name", c.Slug), "chest", "name")
		}
		if !validRarities[c.Rarity] {
			result.AddError("INVALID_RARITY",
				fmt.Sprintf("chest %q has invalid rarity %q", c.Slug, c.Rarity), "chest", "rarity")
		}
	}
}

// checkRelicReferenceIntegrity verifies that all referenced relics in drop tables exist.
func (v *Validator) checkRelicReferenceIntegrity(cs ContentSet, result *ValidationResult) {
	relicSlugs := make(map[string]bool)
	for _, r := range cs.Relics {
		relicSlugs[r.Slug] = true
	}
	relicIDs := make(map[int64]bool)
	for _, r := range cs.Relics {
		relicIDs[r.ID] = true
	}

	for _, dt := range cs.DropTables {
		if dt.RelicID != 0 && !relicIDs[dt.RelicID] {
			result.AddError("MISSING_RELIC",
				fmt.Sprintf("drop table entry for chest %q references unknown relic ID %d",
					dt.ChestSlug, dt.RelicID), "drop_table", "relic_id")
		}
	}
}

// checkWeightTotals verifies that each chest's drop table has a positive total weight.
func (v *Validator) checkWeightTotals(cs ContentSet, result *ValidationResult) {
	weightSums := make(map[string]float64)
	for _, dt := range cs.DropTables {
		weightSums[dt.ChestSlug] += dt.Weight
	}
	for slug, total := range weightSums {
		if total <= 0 {
			result.AddError("ZERO_TOTAL_WEIGHT",
				fmt.Sprintf("chest %q has total drop weight %.4f (must be > 0)",
					slug, total), "drop_table", "weight")
		}
	}
}

// checkDuplicateDropTableEntries checks for duplicate relic IDs in the same drop table.
func (v *Validator) checkDuplicateDropTableEntries(cs ContentSet, result *ValidationResult) {
	type entryKey struct {
		ChestSlug string
		RelicID   int64
	}
	seen := make(map[entryKey]bool)
	for _, dt := range cs.DropTables {
		if dt.RelicID == 0 {
			continue
		}
		key := entryKey{ChestSlug: dt.ChestSlug, RelicID: dt.RelicID}
		if seen[key] {
			result.AddError("DUPLICATE_DROP_ENTRY",
				fmt.Sprintf("drop table for chest %q has duplicate relic_id %d",
					dt.ChestSlug, dt.RelicID), "drop_table", "relic_id")
		}
		seen[key] = true
	}
}

// checkInvalidRarity verifies that all rarity values in drop tables are valid.
func (v *Validator) checkInvalidRarity(cs ContentSet, result *ValidationResult) {
	validRarities := map[string]bool{
		"common":    true,
		"uncommon":  true,
		"rare":      true,
		"epic":      true,
		"legendary": true,
	}
	for _, dt := range cs.DropTables {
		if dt.RelicID != 0 {
			continue
		}
		if dt.Rarity == "" {
			result.AddError("INVALID_RARITY",
				fmt.Sprintf("drop table entry for chest %q has empty rarity",
					dt.ChestSlug), "drop_table", "rarity")
		} else if !validRarities[dt.Rarity] {
			result.AddError("INVALID_RARITY",
				fmt.Sprintf("drop table entry for chest %q has invalid rarity %q",
					dt.ChestSlug, dt.Rarity), "drop_table", "rarity")
		}
	}
}

// checkOrphanContent flags content definitions that are not referenced
// by any chest, quest, achievement, or other content.
func (v *Validator) checkOrphanContent(cs ContentSet, result *ValidationResult) {
	chestSlugs := make(map[string]bool)
	for _, c := range cs.Chests {
		chestSlugs[c.Slug] = true
	}

	for _, dt := range cs.DropTables {
		if !chestSlugs[dt.ChestSlug] {
			result.AddWarning("ORPHAN_CONTENT",
				fmt.Sprintf("drop table entry references unknown chest %q",
					dt.ChestSlug), "drop_table", "chest_slug")
		}
	}
}
