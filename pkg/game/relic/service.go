package relic

import (
	"context"
	"errors"

	"odyssey/pkg/content"
	"odyssey/pkg/game"
)

// ErrRelicNotFound is returned when a relic does not exist.
var ErrRelicNotFound = errors.New("relic not found")

// RelicService handles relic listing, detail, and inventory.
type RelicService struct {
	relics     game.RelicStore
	playerRel  game.PlayerRelicStore
	contentSvc *content.ContentService
	catalog    RelicCatalog
}

// NewRelicService constructs a RelicService.
func NewRelicService(
	relics game.RelicStore,
	playerRel game.PlayerRelicStore,
	defStore game.RelicDefinitionStore,
	contentSvc ...*content.ContentService,
) *RelicService {
	svc := &RelicService{
		relics:    relics,
		playerRel: playerRel,
		catalog:   NewContentRelicCatalog(defStore),
	}
	if len(contentSvc) > 0 && contentSvc[0] != nil {
		svc.contentSvc = contentSvc[0]
	}
	return svc
}

// GetRelicDefinition returns a relic definition, trying ContentService first,
// then the content-backed catalog, then hardcoded fallback.
func (s *RelicService) GetRelicDefinition(ctx context.Context, slug string) *RelicDefinition {
	if s.contentSvc != nil {
		def, err := s.contentSvc.GetRelic(ctx, slug)
		if err == nil && def != nil {
			return &RelicDefinition{
				ID:          def.ID,
				Slug:        def.Slug,
				Name:        def.Name,
				Description: def.Description,
				Realm:       def.Realm,
				Rarity:      game.Rarity(def.Rarity),
				Image:       def.Image,
				Lore:        def.Lore,
				CreatedAt:   def.CreatedAt,
			}
		}
	}
	return s.catalog.Get(ctx, slug)
}

// ListRelics returns all relic definitions.
func (s *RelicService) ListRelics(ctx context.Context) ([]RelicDefinition, error) {
	if s.contentSvc != nil {
		defs, err := s.contentSvc.ListRelics(ctx)
		if err == nil && len(defs) > 0 {
			result := make([]RelicDefinition, 0, len(defs))
			for _, d := range defs {
				result = append(result, RelicDefinition{
					ID:          d.ID,
					Slug:        d.Slug,
					Name:        d.Name,
					Description: d.Description,
					Realm:       d.Realm,
					Rarity:      game.Rarity(d.Rarity),
					Image:       d.Image,
					Lore:        d.Lore,
					CreatedAt:   d.CreatedAt,
				})
			}
			return result, nil
		}
	}
	result := s.catalog.ListAll(ctx)
	return result, nil
}

// GetRelic returns a single relic definition by slug.
func (s *RelicService) GetRelic(ctx context.Context, slug string) (*RelicDefinition, error) {
	def := s.GetRelicDefinition(ctx, slug)
	if def == nil {
		return nil, ErrRelicNotFound
	}
	return def, nil
}

// ListInventory returns the player's relic collection.
func (s *RelicService) ListInventory(ctx context.Context, uid string) ([]InventoryItem, error) {
	playerRelics, err := s.playerRel.ListPlayerRelics(ctx, uid)
	if err != nil {
		return nil, err
	}

	items := make([]InventoryItem, 0, len(playerRelics))
	for _, pr := range playerRelics {
		def := s.GetRelicDefinition(ctx, pr.RelicSlug)
		if def == nil {
			continue
		}
		items = append(items, InventoryItem{
			RelicID:      pr.RelicID,
			RelicSlug:    pr.RelicSlug,
			Name:         def.Name,
			Description:  def.Description,
			Realm:        def.Realm,
			Rarity:       def.Rarity,
			Image:        def.Image,
			Lore:         def.Lore,
			OwnedCount:   pr.OwnedCount,
			IsNew:        pr.IsNew,
			DiscoveredAt: pr.DiscoveredAt,
			CreatedAt:    pr.CreatedAt,
		})
	}
	return items, nil
}

// GetLatestRelic returns the most recently discovered relic for a player.
func (s *RelicService) GetLatestRelic(ctx context.Context, uid string) (*InventoryItem, error) {
	playerRelics, err := s.playerRel.ListPlayerRelics(ctx, uid)
	if err != nil {
		return nil, err
	}
	if len(playerRelics) == 0 {
		return nil, game.ErrNotFound
	}

	latest := &playerRelics[0]
	for i := range playerRelics {
		if playerRelics[i].DiscoveredAt.After(latest.DiscoveredAt) {
			latest = &playerRelics[i]
		}
	}

	def := s.GetRelicDefinition(ctx, latest.RelicSlug)
	if def == nil {
		return nil, ErrRelicNotFound
	}

	return &InventoryItem{
		RelicID:      latest.RelicID,
		RelicSlug:    latest.RelicSlug,
		Name:         def.Name,
		Description:  def.Description,
		Realm:        def.Realm,
		Rarity:       def.Rarity,
		Image:        def.Image,
		Lore:         def.Lore,
		OwnedCount:   latest.OwnedCount,
		IsNew:        latest.IsNew,
		DiscoveredAt: latest.DiscoveredAt,
		CreatedAt:    latest.CreatedAt,
	}, nil
}

// GetCollectionProgress returns collected/total counts.
func (s *RelicService) GetCollectionProgress(ctx context.Context, uid string) (collected int, total int, err error) {
	items, err := s.ListInventory(ctx, uid)
	if err != nil {
		return 0, 0, err
	}
	defs, err := s.ListRelics(ctx)
	if err != nil {
		return 0, 0, err
	}
	total = len(defs)
	collected = 0
	seen := make(map[string]bool)
	for _, item := range items {
		if !seen[item.RelicSlug] {
			collected++
			seen[item.RelicSlug] = true
		}
	}
	return collected, total, nil
}

// InventoryService is a facade combining relic listing and inventory.
type InventoryService struct {
	relicSvc *RelicService
}

// NewInventoryService constructs an InventoryService.
func NewInventoryService(relicSvc *RelicService) *InventoryService {
	return &InventoryService{relicSvc: relicSvc}
}

// ListInventory returns the player's full relic collection.
func (s *InventoryService) ListInventory(ctx context.Context, uid string) ([]InventoryItem, error) {
	return s.relicSvc.ListInventory(ctx, uid)
}

// GetCollectionProgress returns collected/total counts.
func (s *InventoryService) GetCollectionProgress(ctx context.Context, uid string) (collected int, total int, err error) {
	items, err := s.ListInventory(ctx, uid)
	if err != nil {
		return 0, 0, err
	}
	defs, err := s.relicSvc.ListRelics(ctx)
	if err != nil {
		return 0, 0, err
	}
	total = len(defs)
	collected = 0
	seen := make(map[string]bool)
	for _, item := range items {
		if !seen[item.RelicSlug] {
			collected++
			seen[item.RelicSlug] = true
		}
	}
	return collected, total, nil
}
