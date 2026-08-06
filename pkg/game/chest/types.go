package chest

import (
	"time"

	"odyssey/pkg/game"
)

// ChestType defines a type of chest.
type ChestType struct {
	ID           int64
	Slug         string
	Name         string
	Rarity       game.Rarity
	Icon         string
	Description  string
	DropTable    map[game.Rarity]float64
	RelicWeights []RelicWeight
	CreatedAt    time.Time
}

// RewardItem represents a single reward from opening a chest.
type RewardItem struct {
	RelicSlug string
	Name      string
	Rarity    game.Rarity
	Realm     string
	IsNew     bool
}

// OpenResult is the response from opening a chest.
type OpenResult struct {
	Chest          *ChestView
	Rewards        []RewardItem
	NewCount       int
	DuplicateCount int
}

// ChestView is the API representation of a chest.
type ChestView struct {
	ID          int64       `json:"id"`
	UID         string      `json:"uid"`
	Source      string      `json:"source"`
	ChestSlug   string      `json:"chest_slug"`
	Name        string      `json:"name"`
	Rarity      game.Rarity `json:"rarity"`
	Icon        string      `json:"icon"`
	Description string      `json:"description"`
	Opened      bool        `json:"opened"`
	OpenedAt    *time.Time  `json:"opened_at"`
	CreatedAt   time.Time   `json:"created_at"`
}
