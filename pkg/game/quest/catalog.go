package quest

import (
	"odyssey/pkg/game/world"
)

// QuestTemplate is the code-embedded definition for a quest (see
// docs/domain-model.md: "Quest (template) referenced by slug in odyssey_quests").
// Templates are NOT persisted; they live in code for the MVP and provide the
// realm association and reward shape used by the quest completion flow.
type QuestTemplate struct {
	Slug          string
	Title         string
	Realm         string
	ChallengeDefs []ChallengeDef
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

// RealmProgressPerQuest is how much the shared realm progress bar advances
// when a quest in that realm is completed. Capped at the realm's MaxProgress.
// Deprecated: use RealmCatalog with configurable MaxProgress per realm.
const RealmProgressPerQuest = 25

// RealmCompletionThreshold is the progress value at which a realm is considered complete.
// Deprecated: use RealmCatalog.Get(realm).MaxProgress.
const RealmCompletionThreshold = 100

// realmsOrder defines the sequence in which realms are unlocked.
// Deprecated: use world.DefaultRealmCatalog.Order() instead.
var realmsOrder = world.DefaultRealmCatalog.Order()

// NextRealm returns the realm that follows the given realm in the unlock sequence.
// Returns an empty string if the given realm is the last one or is unknown.
// Deprecated: use RealmCatalog.NextRealm instead.
func NextRealm(current string) string {
	for i, r := range realmsOrder {
		if r == current && i+1 < len(realmsOrder) {
			return realmsOrder[i+1]
		}
	}
	return ""
}

// questCatalog holds the code-embedded quest templates for the MVP.
// Realms are embedded in code per docs/domain-model.md.
var questCatalog = map[string]QuestTemplate{
	"morning-light": {
		Slug:  "morning-light",
		Title: "Morning Light",
		Realm: "whispering-woods",
		ChallengeDefs: []ChallengeDef{
			{Slug: "find-the-dew", Description: "Find something glistening outside your door and describe it.", Type: ChallengeObservation},
			{Slug: "morning-fact", Description: "Look up one fact about morning sunlight and share it.", Type: ChallengeResearch},
		},
	},
	"gather-herbs": {
		Slug:  "gather-herbs",
		Title: "Gather Herbs",
		Realm: "whispering-woods",
		ChallengeDefs: []ChallengeDef{
			{Slug: "spot-the-green", Description: "Point out three shades of green you can see right now.", Type: ChallengeObservation},
			{Slug: "herb-lore", Description: "Name one use for a common houseplant.", Type: ChallengeResearch},
		},
	},
	"riddle-of-the-stones": {
		Slug:  "riddle-of-the-stones",
		Title: "Riddle of the Stones",
		Realm: "whispering-woods",
		ChallengeDefs: []ChallengeDef{
			{Slug: "stone-shape", Description: "Find a stone or brick and describe its shape.", Type: ChallengeObservation},
			{Slug: "solve-riddle", Description: "Solve: I have no voice, yet I answer every question. What am I?", Type: ChallengePuzzle},
		},
	},
}

// LookupTemplate returns the embedded template for a quest slug, if known.
func LookupTemplate(slug string) (QuestTemplate, bool) {
	t, ok := questCatalog[slug]
	return t, ok
}

// RealmForSlug returns the realm a quest template belongs to. Returns "" when
// the slug is unknown (callers treat an empty realm as "no realm progress").
func RealmForSlug(slug string) string {
	if t, ok := LookupTemplate(slug); ok {
		return t.Realm
	}
	return ""
}
