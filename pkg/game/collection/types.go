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
	Journey       string
	Rarity      game.Rarity
	Image       string
	Concept        string
	CreatedAt   time.Time
}

// RelicInstance represents a single time a relic was awarded to a player.
type RelicInstance struct {
	ID        int64
	UID       string
	Code      string
	Name      string
	Rarity    game.Rarity
	Journey     string
	AwardedAt time.Time
	CreatedAt time.Time
}

// InventoryItem represents a relic in the player's collection.
type InventoryItem struct {
	CollectionID      int64       `json:"collection_id"`
	CollectionSlug    string      `json:"collection_slug"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Journey        string      `json:"journey"`
	Rarity       game.Rarity `json:"rarity"`
	Image        string      `json:"image"`
	Concept         string      `json:"concept"`
	OwnedCount   int         `json:"owned_count"`
	IsNew        bool        `json:"is_new"`
	DiscoveredAt time.Time   `json:"discovered_at"`
	CreatedAt    time.Time   `json:"created_at"`
}

// GiftResult represents the outcome of a successful relic gift.
type GiftResult struct {
	CollectionSlug     string `json:"collection_slug"`
	RelicName     string `json:"relic_name"`
	RecipientUID  string `json:"recipient_uid"`
	RecipientName string `json:"recipient_name"`
	SenderCount   int    `json:"sender_remaining_count"`
}
