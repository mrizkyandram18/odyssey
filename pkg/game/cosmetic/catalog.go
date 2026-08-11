package cosmetic

// Fixed Slice 2.2 catalog — intentionally tiny (one paid cosmetic).
// Business rules live here; handlers must not hardcode prices.

const (
	// CosmeticAvatarFrameGold is the only paid cosmetic in Slice 2.2.
	CosmeticAvatarFrameGold   = "avatar_frame_gold"
	CosmeticAvatarFrameSilver = "avatar_frame_silver"

	// PriceAvatarFrameGold is the fixed coin price for the gold frame.
	PriceAvatarFrameGold   int64 = 3
	PriceAvatarFrameSilver int64 = 2

	// Frame values stored on odyssey_user_profiles.avatar_frame.
	FrameNone   = "none"
	FrameGold   = "gold"
	FrameSilver = "silver"

	// Explorer effect items.
	EffectSparkle = "explorer_effect_sparkle"
	EffectFloat   = "explorer_effect_float"
	EffectTrail   = "explorer_effect_trail"

	// Effect values stored on odyssey_user_profiles.equipped_explorer_effect.
	EffectNone       = "none"
	EffectSparkleVal = "sparkle"
	EffectFloatVal   = "float"
	EffectTrailVal   = "trail"

	SourceCosmeticPurchase = "COSMETIC_PURCHASE"
	RewardTypeCoins        = "COINS"
)

// Item is a purchasable cosmetic definition.
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Kind        string `json:"kind"`  // avatar_frame | explorer_effect
	Value       string `json:"value"` // equip value (e.g. gold, sparkle)
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
	{
		ID:          CosmeticAvatarFrameSilver,
		Name:        "Silver Avatar Frame",
		Description: "A sleek silver ring around your explorer portrait.",
		Price:       PriceAvatarFrameSilver,
		Kind:        "avatar_frame",
		Value:       FrameSilver,
	},
	{
		ID:          EffectSparkle,
		Name:        "Sparkle Effect",
		Description: "A subtle sparkle around your explorer portrait.",
		Price:       0,
		Kind:        "explorer_effect",
		Value:       EffectSparkleVal,
	},
	{
		ID:          EffectFloat,
		Name:        "Float Effect",
		Description: "Your explorer portrait gently floats up and down.",
		Price:       0,
		Kind:        "explorer_effect",
		Value:       EffectFloatVal,
	},
	{
		ID:          EffectTrail,
		Name:        "Trail Effect",
		Description: "A playful swaying motion trails your explorer.",
		Price:       0,
		Kind:        "explorer_effect",
		Value:       EffectTrailVal,
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
