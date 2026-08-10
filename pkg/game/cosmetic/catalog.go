package cosmetic

// Fixed Slice 2.2 catalog — intentionally tiny (one paid cosmetic).
// Business rules live here; handlers must not hardcode prices.

const (
	// CosmeticAvatarFrameGold is the only paid cosmetic in Slice 2.2.
	CosmeticAvatarFrameGold = "avatar_frame_gold"

	// PriceAvatarFrameGold is the fixed coin price for the gold frame.
	PriceAvatarFrameGold int64 = 3

	// Frame values stored on odyssey_user_profiles.avatar_frame.
	FrameNone = "none"
	FrameGold = "gold"

	SourceCosmeticPurchase = "COSMETIC_PURCHASE"
	RewardTypeCoins        = "COINS"
)

// Item is a purchasable cosmetic definition.
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Kind        string `json:"kind"`  // avatar_frame
	Value       string `json:"value"` // equip value (e.g. gold)
}

// Catalog is the closed list of cosmetics available in Slice 2.2.
var Catalog = []Item{
	{
		ID:          CosmeticAvatarFrameGold,
		Name:        "Gold Avatar Frame",
		Description: "A golden ring around your explorer portrait.",
		Price:       PriceAvatarFrameGold,
		Kind:        "avatar_frame",
		Value:       FrameGold,
	},
}

// Lookup returns a catalog item by id.
func Lookup(id string) (Item, bool) {
	for _, it := range Catalog {
		if it.ID == id {
			return it, true
		}
	}
	return Item{}, false
}
