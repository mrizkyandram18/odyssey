package progression

type ResourceType string

const (
	ResourceXP            ResourceType = "XP"
	ResourceRelic         ResourceType = "RELIC"
	ResourceInspiration   ResourceType = "INSPIRATION"
	ResourceStoryFragment ResourceType = "STORY_FRAGMENT"
)

type ChestSource string

const (
	ChestSourceQuest          ChestSource = "QUEST"
	ChestSourceLevelUp        ChestSource = "LEVEL_UP"
	ChestSourceRealmMilestone ChestSource = "REALM_MILESTONE"
)

type AchievementKind string

const (
	AchievementKindPersonal AchievementKind = "PERSONAL"
	AchievementKindGroup    AchievementKind = "GROUP"
)
