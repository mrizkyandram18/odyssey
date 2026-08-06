package relic

import (
	"time"

	"odyssey/pkg/game"
)

// RelicDefinition is the catalog definition for a relic.
type RelicDefinition struct {
	ID          int64
	Slug        string
	Name        string
	Description string
	Realm       string
	Rarity      game.Rarity
	Image       string
	Lore        string
	CreatedAt   time.Time
}

// RelicInstance represents a single time a relic was awarded to a player.
type RelicInstance struct {
	ID        int64
	UID       string
	Code      string
	Name      string
	Rarity    game.Rarity
	Realm     string
	AwardedAt time.Time
	CreatedAt time.Time
}

// InventoryItem represents a relic in the player's collection.
type InventoryItem struct {
	RelicID      int64       `json:"relic_id"`
	RelicSlug    string      `json:"relic_slug"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Realm        string      `json:"realm"`
	Rarity       game.Rarity `json:"rarity"`
	Image        string      `json:"image"`
	Lore         string      `json:"lore"`
	OwnedCount   int         `json:"owned_count"`
	IsNew        bool        `json:"is_new"`
	DiscoveredAt time.Time   `json:"discovered_at"`
	CreatedAt    time.Time   `json:"created_at"`
}
