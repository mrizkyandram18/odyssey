package content

import (
	"context"
)

// Table names for content definition tables.
const (
	TableRealms       = "odyssey_journey_definitions"
	TableChapters     = "odyssey_course_definitions"
	TableQuests       = "odyssey_quest_definitions"
	TablePrompts      = "odyssey_creative_prompt_definitions"
	TableChests       = "odyssey_chest_definitions"
	TableDropTables   = "odyssey_drop_tables"
	TableRelics       = "odyssey_relic_definitions"
	TableAchievements = "odyssey_achievement_definitions"
	TableSeasons      = "odyssey_season_definitions"
	TableLore         = "odyssey_concept_definitions"
)

// ResourceInfo describes a content definition resource for admin operations.
type ResourceInfo struct {
	Table     string
	Resource  string
	IDField   string
	SlugField string
}

// AdminStore provides generic CRUD persistence for content definitions.
// It operates on arbitrary tables backing content definitions, using the
// PostgREST REST adapter. All writes are patch-based (map[string]any)
// with no hardcoded SQL.
type AdminStore interface {
	// GetByID retrieves a definition row by its primary key (published or draft).
	GetByID(ctx context.Context, table string, id int64) (map[string]any, error)

	// GetBySlug retrieves a definition row by slug (published or draft, for admin).
	GetBySlug(ctx context.Context, table, slug string) (map[string]any, error)

	// ListAll returns all rows including drafts and deleted (for admin listing).
	ListAll(ctx context.Context, table string, includeDeleted bool) ([]map[string]any, error)

	// ListPublished returns only published, non-deleted rows (for gameplay).
	ListPublished(ctx context.Context, table string) ([]map[string]any, error)

	// Create inserts a new definition row in draft state (published=false).
	Create(ctx context.Context, table string, data map[string]any) (map[string]any, error)

	// UpdateDraft patches a definition row by slug, incrementing version and
	// setting updated_by. The row's published flag is not changed.
	UpdateDraft(ctx context.Context, table, slug string, patch map[string]any, updatedBy string) error

	// Publish sets published=true, published_at=now, increments version.
	Publish(ctx context.Context, table, slug string, updatedBy string) error

	// SoftDelete sets deleted_at=now() and published=false.
	SoftDelete(ctx context.Context, table, slug string) error

	// Restore clears deleted_at. Preserves the published flag.
	Restore(ctx context.Context, table, slug string) error
}
