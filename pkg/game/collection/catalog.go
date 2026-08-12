package relic

import (
	"context"
	"sync"
	"time"

	"odyssey/pkg/game"
)

// RelicCatalog provides access to relic definitions.
type RelicCatalog interface {
	Get(ctx context.Context, slug string) *RelicDefinition
	ListAll(ctx context.Context) []RelicDefinition
}

// ContentRelicCatalog is a content-backed relic catalog that loads
// definitions from a RelicDefinitionStore, caches them, and falls
// back to hardcoded defaults when the store is unavailable or returns no data.
type ContentRelicCatalog struct {
	store    game.RelicDefinitionStore
	cache    map[string]*RelicDefinition
	fallback []RelicDefinition
	mu       sync.RWMutex
}

// NewContentRelicCatalog creates a ContentRelicCatalog backed by the given store.
// If store is nil, the catalog falls back to hardcoded defaults only.
func NewContentRelicCatalog(store game.RelicDefinitionStore) *ContentRelicCatalog {
	return &ContentRelicCatalog{
		store:    store,
		cache:    make(map[string]*RelicDefinition),
		fallback: DefaultRelicCatalog,
	}
}

// Get returns a relic definition by slug, trying the store first then falling back.
func (c *ContentRelicCatalog) Get(ctx context.Context, slug string) *RelicDefinition {
	c.mu.RLock()
	if rd, ok := c.cache[slug]; ok {
		c.mu.RUnlock()
		return rd
	}
	c.mu.RUnlock()

	if c.store != nil {
		def, err := c.store.GetRelicDefinition(ctx, slug)
		if err == nil && def != nil {
			rd := &RelicDefinition{
				ID:          def.ID,
				Slug:        def.Slug,
				Name:        def.Name,
				Description: def.Description,
				Journey:       def.Journey,
				Rarity:      game.Rarity(def.Rarity),
				Image:       def.Image,
				Concept:        def.Concept,
				CreatedAt:   def.CreatedAt,
			}
			c.mu.Lock()
			c.cache[slug] = rd
			c.mu.Unlock()
			return rd
		}
	}

	for i := range c.fallback {
		if c.fallback[i].Slug == slug {
			return &c.fallback[i]
		}
	}
	return nil
}

// ListAll returns all relic definitions, trying the store first then falling back.
func (c *ContentRelicCatalog) ListAll(ctx context.Context) []RelicDefinition {
	c.mu.RLock()
	if len(c.cache) > 0 {
		result := make([]RelicDefinition, 0, len(c.cache))
		for _, rd := range c.cache {
			result = append(result, *rd)
		}
		c.mu.RUnlock()
		return result
	}
	c.mu.RUnlock()

	if c.store != nil {
		defs, err := c.store.ListRelicDefinitions(ctx)
		if err == nil && len(defs) > 0 {
			result := make([]RelicDefinition, 0, len(defs))
			for _, def := range defs {
				rd := RelicDefinition{
					ID:          def.ID,
					Slug:        def.Slug,
					Name:        def.Name,
					Description: def.Description,
					Journey:       def.Journey,
					Rarity:      game.Rarity(def.Rarity),
					Image:       def.Image,
					Concept:        def.Concept,
					CreatedAt:   def.CreatedAt,
				}
				result = append(result, rd)
				c.mu.Lock()
				c.cache[def.Slug] = &result[len(result)-1]
				c.mu.Unlock()
			}
			return result
		}
	}

	result := make([]RelicDefinition, len(c.fallback))
	copy(result, c.fallback)
	return result
}

var DefaultRelicCatalog = []RelicDefinition{
	{ID: 1, Slug: "ancient-compass", Name: "Ancient Compass", Description: "A weathered compass that points to wonder.", Journey: "whispering-woods", Rarity: game.RarityCommon, Image: "🧭", Concept: "Found at the base of the oldest tree.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 2, Slug: "crystal-shard", Name: "Crystal Shard", Description: "A fragment of crystalline light.", Journey: "whispering-woods", Rarity: game.RarityUncommon, Image: "💎", Concept: "It hums softly in your palm.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 3, Slug: "dragon-scale", Name: "Dragon Scale", Description: "A shimmering scale from a ancient wyrm.", Journey: "whispering-woods", Rarity: game.RarityRare, Image: "🐉", Concept: "Warm as a summer stone.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 4, Slug: "starlight-essence", Name: "Starlight Essence", Description: "Captured light from a distant star.", Journey: "whispering-woods", Rarity: game.RarityEpic, Image: "✨", Concept: "It never dims.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 5, Slug: "world-seed", Name: "World Seed", Description: "A tiny seed that holds an entire ecosystem.", Journey: "whispering-woods", Rarity: game.RarityLegendary, Image: "🌱", Concept: "Planted, it grows a forest.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 6, Slug: "mossy-pebble", Name: "Mossy Pebble", Description: "A smooth stone covered in soft moss.", Journey: "whispering-woods", Rarity: game.RarityCommon, Image: "🪨", Concept: "It smells like rain.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 7, Slug: "feather-quill", Name: "Feather Quill", Description: "A quill that writes on its own.", Journey: "whispering-woods", Rarity: game.RarityUncommon, Image: "🪶", Concept: "It tells stories.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 8, Slug: "shadow-leaf", Name: "Shadow Leaf", Description: "A leaf that absorbs all light.", Journey: "whispering-woods", Rarity: game.RarityRare, Image: "🍂", Concept: "Dark as midnight.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 9, Slug: "echo-stone", Name: "Echo Stone", Description: "A stone that repeats the last sound it heard.", Journey: "clockwork-city", Rarity: game.RarityCommon, Image: "🪨", Concept: "It echoes whispers.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 10, Slug: "gear-heart", Name: "Gear Heart", Description: "A mechanical heart that ticks steadily.", Journey: "clockwork-city", Rarity: game.RarityUncommon, Image: "⚙️", Concept: "It never stops.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 11, Slug: "steam-vial", Name: "Steam Vial", Description: "A vial of pressurized steam.", Journey: "clockwork-city", Rarity: game.RarityRare, Image: "🧪", Concept: "It hisses when opened.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 12, Slug: "cogwheel-crown", Name: "Cogwheel Crown", Description: "A crown made of interlocking gears.", Journey: "clockwork-city", Rarity: game.RarityEpic, Image: "👑", Concept: "It fits only the worthy.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 13, Slug: "time-thread", Name: "Time Thread", Description: "A single thread from the fabric of time.", Journey: "clockwork-city", Rarity: game.RarityLegendary, Image: "🧵", Concept: "It weaves moments.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 14, Slug: "copper-nut", Name: "Copper Nut", Description: "A small but sturdy copper nut.", Journey: "clockwork-city", Rarity: game.RarityCommon, Image: "🥜", Concept: "Useful in a pinch.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 15, Slug: "inkwell-spirit", Name: "Inkwell Spirit", Description: "An inkwell filled with living ink.", Journey: "starlit-library", Rarity: game.RarityUncommon, Image: "🖋️", Concept: "It writes your dreams.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 16, Slug: "page-turner", Name: "Page Turner", Description: "A brass tool that turns pages by itself.", Journey: "starlit-library", Rarity: game.RarityRare, Image: "📖", Concept: "It knows what to read next.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 17, Slug: "concept-flame", Name: "Concept Flame", Description: "A small flame that burns with stored knowledge.", Journey: "starlit-library", Rarity: game.RarityEpic, Image: "🔥", Concept: "It lights the way to truth.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 18, Slug: "story-orb", Name: "Story Orb", Description: "An orb containing an entire unwritten story.", Journey: "starlit-library", Rarity: game.RarityLegendary, Image: "🔮", Concept: "It waits for the right author.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 19, Slug: "bookmark", Name: "Silver Bookmark", Description: "A simple silver bookmark.", Journey: "starlit-library", Rarity: game.RarityCommon, Image: "🔖", Concept: "It saves your place.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	{ID: 20, Slug: "quill-of-whispers", Name: "Quill of Whispers", Description: "A quill that records secrets.", Journey: "starlit-library", Rarity: game.RarityRare, Image: "🪶", Concept: "It knows all stories.", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
}

var defaultRelicCatalog = NewContentRelicCatalog(nil)

// GetRelicDefinition returns a relic definition by slug using the default catalog.
func GetRelicDefinition(slug string) *RelicDefinition {
	return defaultRelicCatalog.Get(context.Background(), slug)
}
