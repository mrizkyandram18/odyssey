package quest

import (
	"odyssey/pkg/game/world"
)

// BranchOption defines a narrative choice option for a quest.
type BranchOption struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// QuestTemplate is the code-embedded definition for a quest (see
// docs/domain-model.md: "Mission (template) referenced by slug in odyssey_missions").
// Templates are NOT persisted; they live in code for the MVP and provide the
// journey association and reward shape used by the quest completion flow.
type QuestTemplate struct {
	Slug          string
	Title         string
	Journey         string
	Type          QuestType
	ChallengeDefs []ChallengeDef
	BranchOptions []BranchOption
}

// ChallengeDef describes a single challenge within a template.
type ChallengeDef struct {
	Slug        string
	Description string
	Type        ChallengeType
}

// DefaultXPValues is the set of XP rewards used by the quest completion flow
// when no ProgressionConfig overrides are supplied. These are the MVP defaults
// and may be changed via environment variables or system config.
var DefaultXPValues = struct {
	ChallengeXP       int64
	CompletionBonusXP int64
}{
	ChallengeXP:       20,
	CompletionBonusXP: 60,
}

// ChallengeXP is granted to the Explorer who completes an individual challenge.
// Deprecated: use DefaultXPValues.ChallengeXP or ProgressionConfig.ChallengeXP.
const ChallengeXP int64 = 20

// CompletionBonusXP is granted to the Explorer who completes the final challenge
// (i.e. who triggers quest completion).
// Deprecated: use DefaultXPValues.CompletionBonusXP or ProgressionConfig.CompletionBonusXP.
const CompletionBonusXP int64 = 60

// RealmProgressPerQuest is how much the shared journey progress bar advances
// when a quest in that journey is completed. Capped at the journey's MaxProgress.
// Deprecated: use RealmCatalog with configurable MaxProgress per journey.
const RealmProgressPerQuest = 25

// RealmCompletionThreshold is the progress value at which a journey is considered complete.
// Deprecated: use RealmCatalog.Get(journey).MaxProgress.
const RealmCompletionThreshold = 100

// MVPRealm is the single playable journey for Phase 1.
const MVPRealm = "whispering-woods"

// realmsOrder defines the sequence in which realms are unlocked.
// Deprecated: use world.DefaultRealmCatalog.Order() instead.
var realmsOrder = world.DefaultRealmCatalog.Order()

// NextRealm returns the journey that follows the given journey in the unlock sequence.
// Returns an empty string if the given journey is the last one or is unknown.
// Deprecated: use RealmCatalog.NextRealm instead.
func NextRealm(current string) string {
	for i, r := range realmsOrder {
		if r == current && i+1 < len(realmsOrder) {
			return realmsOrder[i+1]
		}
	}
	return ""
}

// questCatalog holds the code-embedded quest templates for Phase 1 MVP.
// Exactly 6 missions in Whispering Woods with SOLO + RELAY + CREATIVE variety.
// Future-journey templates live in odyssey_quest_definitions only until later phases.
var questCatalog = map[string]QuestTemplate{
	// SOLO — observation / research
	"morning-light": {
		Slug:  "morning-light",
		Title: "Morning Light",
		Journey: MVPRealm,
		Type:  QuestTypeSolo,
		ChallengeDefs: []ChallengeDef{
			{Slug: "find-the-dew", Description: "Find something glistening outside your door and describe it.", Type: ChallengeObservation},
			{Slug: "morning-fact", Description: "Look up one fact about morning sunlight and share it.", Type: ChallengeResearch},
		},
	},
	"gather-herbs": {
		Slug:  "gather-herbs",
		Title: "Gather Herbs",
		Journey: MVPRealm,
		Type:  QuestTypeSolo,
		ChallengeDefs: []ChallengeDef{
			{Slug: "spot-the-green", Description: "Point out three shades of green you can see right now.", Type: ChallengeObservation},
			{Slug: "herb-concept", Description: "Name one use for a common houseplant.", Type: ChallengeResearch},
		},
	},
	// SOLO — puzzle / observation with narrative branching choices
	"riddle-of-the-stones": {
		Slug:  "riddle-of-the-stones",
		Title: "Riddle of the Stones",
		Journey: MVPRealm,
		Type:  QuestTypeSolo,
		ChallengeDefs: []ChallengeDef{
			{Slug: "stone-shape", Description: "Find a stone or brick and describe its shape.", Type: ChallengeObservation},
			{Slug: "solve-riddle", Description: "Solve: I have no voice, yet I answer every question. What am I?", Type: ChallengePuzzle},
		},
		BranchOptions: []BranchOption{
			{Slug: "path-of-echoes", Title: "Langkah Gema Berbisik", Description: "Kru memilih menyusuri tebing gaung bertekstur lumut."},
			{Slug: "path-of-moss", Title: "Langkah Karpet Hijau", Description: "Kru memilih menelusuri karpet hijau di dasar lembah."},
		},
	},
	// RELAY — sequential legs; next leg is assigned round-robin after each completion
	"shadow-trail": {
		Slug:  "shadow-trail",
		Title: "Shadow Trail",
		Journey: MVPRealm,
		Type:  QuestTypeRelay,
		ChallengeDefs: []ChallengeDef{
			{Slug: "trace-shadow", Description: "Trace the shape of a shadow on the ground and describe it.", Type: ChallengeObservation},
			{Slug: "shadow-story", Description: "Invent a short story about where the shadow leads.", Type: ChallengeWrite},
		},
	},
	// CREATIVE — group Story submission (CREATE_MEMORY on quest completion)
	"the-old-growth": {
		Slug:  "the-old-growth",
		Title: "The Old Growth",
		Journey: MVPRealm,
		Type:  QuestTypeCreative,
		ChallengeDefs: []ChallengeDef{
			{Slug: "draw-tree", Description: "Draw or describe the oldest tree you can see.", Type: ChallengeDraw},
			{Slug: "tree-history", Description: "Write a short paragraph about what this tree has witnessed.", Type: ChallengeWrite},
		},
	},
	// SOLO — puzzle / observation
	"forest-riddle": {
		Slug:  "forest-riddle",
		Title: "Forest Riddle",
		Journey: MVPRealm,
		Type:  QuestTypeSolo,
		ChallengeDefs: []ChallengeDef{
			{Slug: "riddle-solve", Description: "Solve the riddle: I am always hungry, I must always be fed. The finger I touch will soon turn red. What am I?", Type: ChallengePuzzle},
			{Slug: "find-marker", Description: "Find a natural marker (stone, stick, leaf) that matches the riddle answer.", Type: ChallengeObservation},
		},
	},
	// SOLO — movement / observation challenge in Clockwork City
	"clockwork-expedition": {
		Slug:  "clockwork-expedition",
		Title: "Clockwork Expedition",
		Journey: "clockwork-city",
		Type:  QuestTypeSolo,
		ChallengeDefs: []ChallengeDef{
			{Slug: "step-count", Description: "Berjalan 100 langkah atau periksa sudut bayangan menara jam.", Type: ChallengeMovement},
			{Slug: "gear-observation", Description: "Catat tiga objek berputar yang kamu temukan.", Type: ChallengeObservation},
		},
		BranchOptions: []BranchOption{
			{Slug: "path-of-copper", Title: "Jalur Tembaga Kuning", Description: "Kru menelusuri lorong tembaga kuno yang berkilau."},
			{Slug: "path-of-steam", Title: "Jalur Uap Bertekanan", Description: "Kru menelusuri lorong uap panas bertekanan tinggi."},
		},
	},
}

// LookupTemplate returns the embedded template for a quest slug, if known.
func LookupTemplate(slug string) (QuestTemplate, bool) {
	t, ok := questCatalog[slug]
	return t, ok
}

// RealmForSlug returns the journey a quest template belongs to. Returns "" when
// the slug is unknown (callers treat an empty journey as "no journey progress").
func RealmForSlug(slug string) string {
	if t, ok := LookupTemplate(slug); ok {
		return t.Journey
	}
	return ""
}

// TypeForSlug returns the quest type for a template slug, or empty if unknown.
func TypeForSlug(slug string) QuestType {
	if t, ok := LookupTemplate(slug); ok {
		return t.Type
	}
	return ""
}

// MVPQuestSlugs returns the ordered list of the 6 Phase 1 Whispering Woods missions.
func MVPQuestSlugs() []string {
	return []string{
		"morning-light",
		"gather-herbs",
		"riddle-of-the-stones",
		"shadow-trail",
		"the-old-growth",
		"forest-riddle",
	}
}
