package chest

import (
	"context"
	"log"
	"sync"
	"time"

	"odyssey/pkg/game"
)

// ChestCatalog provides access to chest definitions.
type ChestCatalog interface {
	Get(ctx context.Context, slug string) *ChestType
	ListAll(ctx context.Context) []ChestType
}

// ContentChestCatalog is a content-backed chest catalog that loads
// definitions from a ChestDefinitionStore, caches them, and falls
// back to hardcoded defaults when the store is unavailable or returns no data.
type ContentChestCatalog struct {
	store    game.ChestDefinitionStore
	cache    map[string]*ChestType
	fallback []ChestType
	mu       sync.RWMutex
}

// NewContentChestCatalog creates a ContentChestCatalog backed by the given store.
// If store is nil, the catalog falls back to hardcoded defaults only.
func NewContentChestCatalog(store game.ChestDefinitionStore) *ContentChestCatalog {
	return &ContentChestCatalog{
		store:    store,
		cache:    make(map[string]*ChestType),
		fallback: DefaultChestCatalog,
	}
}

// lookupFallback searches the fallback catalog for a specific chest slug.
func (c *ContentChestCatalog) lookupFallback(slug string) (*ChestType, bool) {
	for i := range c.fallback {
		if c.fallback[i].Slug == slug {
			return &c.fallback[i], true
		}
	}
	return nil, false
}

// Get returns a chest definition by slug, trying the store first then falling back.
func (c *ContentChestCatalog) Get(ctx context.Context, slug string) *ChestType {
	c.mu.RLock()
	if ct, ok := c.cache[slug]; ok {
		c.mu.RUnlock()
		return ct
	}
	c.mu.RUnlock()

	if c.store != nil {
		def, err := c.store.GetChestDefinition(ctx, slug)
		if err == nil && def != nil {
			entries, err := c.store.ListDropTableEntries(ctx, slug)
			if err == nil {
				ct := c.buildChestType(def, entries)
				c.mu.Lock()
				c.cache[slug] = ct
				c.mu.Unlock()
				return ct
			}
			log.Printf("WARN: loading drop table %s: %v", def.Slug, err)
		}
	}

	if fb, ok := c.lookupFallback(slug); ok {
		return fb
	}
	log.Printf("WARN: no fallback exists for %s", slug)
	return nil
}

// ListAll returns all chest definitions, trying the store first then falling back.
func (c *ContentChestCatalog) ListAll(ctx context.Context) []ChestType {
	c.mu.RLock()
	if len(c.cache) > 0 {
		result := make([]ChestType, 0, len(c.cache))
		for _, ct := range c.cache {
			result = append(result, *ct)
		}
		c.mu.RUnlock()
		return result
	}
	c.mu.RUnlock()

	if c.store != nil {
		defs, err := c.store.ListChestDefinitions(ctx)
		if err == nil && len(defs) > 0 {
			result := make([]ChestType, 0, len(defs))
			for _, def := range defs {
				entries, err := c.store.ListDropTableEntries(ctx, def.Slug)
				if err != nil {
					log.Printf("WARN: loading drop table %s: %v", def.Slug, err)
					if fb, ok := c.lookupFallback(def.Slug); ok {
						result = append(result, *fb)
						continue
					}
					log.Printf("WARN: no fallback exists for %s", def.Slug)
					continue
				}
				ct := c.buildChestType(&def, entries)
				result = append(result, *ct)
				c.mu.Lock()
				c.cache[def.Slug] = ct
				c.mu.Unlock()
			}
			return result
		}
	}

	result := make([]ChestType, len(c.fallback))
	copy(result, c.fallback)
	return result
}

func (c *ContentChestCatalog) buildChestType(def *game.ChestDefinition, entries []game.DropTableEntry) *ChestType {
	weights := make(map[game.Rarity]float64, len(entries))
	for _, e := range entries {
		weights[game.Rarity(e.Rarity)] = e.Weight
	}
	return &ChestType{
		ID:          def.ID,
		Slug:        def.Slug,
		Name:        def.Name,
		Rarity:      game.Rarity(def.Rarity),
		Icon:        def.Icon,
		Description: def.Description,
		DropTable:   weights,
		CreatedAt:   def.CreatedAt,
	}
}

// DefaultChestCatalog is the hardcoded fallback chest catalog.
// Preserved for backward compatibility and as a fallback when DB content is unavailable.
var DefaultChestCatalog = []ChestType{
	{
		ID:          1,
		Slug:        "wooden-chest",
		Name:        "Wooden Chest",
		Rarity:      game.RarityCommon,
		Icon:        "📦",
		Description: "A simple wooden chest, worn by time.",
		DropTable: map[game.Rarity]float64{
			game.RarityCommon:   0.70,
			game.RarityUncommon: 0.25,
			game.RarityRare:     0.05,
		},
		CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		ID:          2,
		Slug:        "bronze-chest",
		Name:        "Bronze Chest",
		Rarity:      game.RarityUncommon,
		Icon:        "🟤",
		Description: "A sturdy bronze chest with ornate fittings.",
		DropTable: map[game.Rarity]float64{
			game.RarityCommon:   0.50,
			game.RarityUncommon: 0.35,
			game.RarityRare:     0.12,
			game.RarityEpic:     0.03,
		},
		CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		ID:          3,
		Slug:        "silver-chest",
		Name:        "Silver Chest",
		Rarity:      game.RarityRare,
		Icon:        "⚪",
		Description: "A polished silver chest that gleams in the dark.",
		DropTable: map[game.Rarity]float64{
			game.RarityCommon:    0.30,
			game.RarityUncommon:  0.35,
			game.RarityRare:      0.25,
			game.RarityEpic:      0.08,
			game.RarityLegendary: 0.02,
		},
		CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		ID:          4,
		Slug:        "golden-chest",
		Name:        "Golden Chest",
		Rarity:      game.RarityEpic,
		Icon:        "🟡",
		Description: "A magnificent golden chest, warm to the touch.",
		DropTable: map[game.Rarity]float64{
			game.RarityCommon:    0.15,
			game.RarityUncommon:  0.25,
			game.RarityRare:      0.30,
			game.RarityEpic:      0.22,
			game.RarityLegendary: 0.08,
		},
		CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		ID:          5,
		Slug:        "mystic-chest",
		Name:        "Mystic Chest",
		Rarity:      game.RarityLegendary,
		Icon:        "🔮",
		Description: "An otherworldly chest that hums with arcane energy.",
		DropTable: map[game.Rarity]float64{
			game.RarityCommon:    0.05,
			game.RarityUncommon:  0.15,
			game.RarityRare:      0.25,
			game.RarityEpic:      0.30,
			game.RarityLegendary: 0.25,
		},
		CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	},
}

var defaultCatalog = NewContentChestCatalog(nil)

// GetChestType returns a chest definition by slug using the default catalog.
func GetChestType(slug string) *ChestType {
	return defaultCatalog.Get(context.Background(), slug)
}
