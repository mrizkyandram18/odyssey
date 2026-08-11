package relic

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"odyssey/pkg/content"
	"odyssey/pkg/game"
)

// ErrRelicNotFound is returned when a relic does not exist.
var ErrRelicNotFound = errors.New("relic not found")

// Gifting errors
var (
	ErrRelicNotOwned     = errors.New("relic not owned or insufficient count")
	ErrSelfGift          = errors.New("cannot gift relic to yourself")
	ErrCrossCrewGift     = errors.New("cross-crew gifting not allowed")
	ErrRecipientNotFound = errors.New("recipient not found")
)

// Gifting ledger sources
const (
	SourceRelicGiftSent     = "RELIC_GIFT_SENT"
	SourceRelicGiftReceived = "RELIC_GIFT_RECEIVED"
	RewardTypeRelicItem     = "RELIC_ITEM"
)

// RelicService handles relic listing, detail, and inventory.
type RelicService struct {
	relics     game.RelicStore
	playerRel  game.PlayerRelicStore
	contentSvc *content.ContentService
	catalog    RelicCatalog

	users   game.UserStore
	ledgers game.RewardLedgerStore
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

// SetUserStore injects the user store for crew validation.
func (s *RelicService) SetUserStore(u game.UserStore) {
	s.users = u
}

// SetLedgerStore injects the ledger store for audit logs.
func (s *RelicService) SetLedgerStore(l game.RewardLedgerStore) {
	s.ledgers = l
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
		if pr.OwnedCount <= 0 {
			continue
		}
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
	var validRelics []game.PlayerRelic
	for _, pr := range playerRelics {
		if pr.OwnedCount > 0 {
			validRelics = append(validRelics, pr)
		}
	}
	if len(validRelics) == 0 {
		return nil, game.ErrNotFound
	}

	latest := &validRelics[0]
	for i := range validRelics {
		if validRelics[i].DiscoveredAt.After(latest.DiscoveredAt) {
			latest = &validRelics[i]
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

// GiftRelic transfers a relic from one explorer to another within the same crew.
func (s *RelicService) GiftRelic(ctx context.Context, senderUID, recipientUID, relicSlug, crewID string) (*GiftResult, error) {
	if senderUID == "" || recipientUID == "" || relicSlug == "" {
		return nil, errors.New("invalid gift payload")
	}
	if senderUID == recipientUID {
		return nil, ErrSelfGift
	}

	if s.users == nil || s.ledgers == nil {
		return nil, errors.New("relic service missing required dependencies for gifting")
	}

	// 1. Verify Sender (auth layer verified UID, we verify crew)
	sender, err := s.users.GetUser(ctx, senderUID)
	if err != nil {
		if err == game.ErrNotFound {
			return nil, errors.New("sender not found")
		}
		return nil, err
	}
	if sender.CrewID != crewID {
		return nil, errors.New("sender not in specified crew context")
	}

	// 2. Verify Recipient
	recipient, err := s.users.GetUser(ctx, recipientUID)
	if err != nil {
		if err == game.ErrNotFound {
			return nil, ErrRecipientNotFound
		}
		return nil, err
	}
	if recipient.CrewID != crewID {
		return nil, ErrCrossCrewGift
	}

	// 3. Verify Relic Definition
	def := s.GetRelicDefinition(ctx, relicSlug)
	if def == nil {
		return nil, ErrRelicNotFound
	}

	// 4. Check Sender Inventory
	senderRelic, err := s.playerRel.GetPlayerRelic(ctx, senderUID, relicSlug)
	if err != nil {
		if err == game.ErrNotFound {
			return nil, ErrRelicNotOwned
		}
		return nil, err
	}
	if senderRelic.OwnedCount < 1 {
		return nil, ErrRelicNotOwned
	}

	// 5. Decrement Sender
	newSenderCount := senderRelic.OwnedCount - 1
	patchSender := map[string]any{
		"owned_count": newSenderCount,
	}
	// Note: In a fully atomic system we would use a compare-and-swap here.
	// Since PostgREST layer lacks optimistic locking on this table, we proceed
	// best-effort. If two gifts happen concurrently on the last item, the count
	// goes negative or hits 0, which we can detect and report in logs.
	if err := s.playerRel.UpdatePlayerRelic(ctx, senderUID, relicSlug, patchSender); err != nil {
		return nil, err
	}

	// 6. Increment/Create Recipient
	recipientRelic, err := s.playerRel.GetPlayerRelic(ctx, recipientUID, relicSlug)
	if err != nil {
		if err == game.ErrNotFound {
			_, err = s.playerRel.CreatePlayerRelic(ctx, &game.PlayerRelic{
				UID:          recipientUID,
				RelicSlug:    relicSlug,
				RelicID:      senderRelic.RelicID,
				OwnedCount:   1,
				IsNew:        true,
				DiscoveredAt: senderRelic.DiscoveredAt,
			})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		patchRecipient := map[string]any{
			"owned_count": recipientRelic.OwnedCount + 1,
			"is_new":      true,
		}
		if err := s.playerRel.UpdatePlayerRelic(ctx, recipientUID, relicSlug, patchRecipient); err != nil {
			return nil, err
		}
	}

	// 7. Write Ledger Audit
	metaSent := fmt.Sprintf(`{"relic_slug":"%s","recipient_uid":"%s"}`, relicSlug, recipientUID)
	metaRecv := fmt.Sprintf(`{"relic_slug":"%s","sender_uid":"%s"}`, relicSlug, senderUID)

	_ = s.ledgers.CreateLedger(ctx, &game.RewardLedger{
		ID:         uuid.New().String(),
		UserID:     senderUID,
		Source:     SourceRelicGiftSent,
		Amount:     0,
		RewardType: RewardTypeRelicItem,
		Metadata:   &metaSent,
	})

	_ = s.ledgers.CreateLedger(ctx, &game.RewardLedger{
		ID:         uuid.New().String(),
		UserID:     recipientUID,
		Source:     SourceRelicGiftReceived,
		Amount:     0,
		RewardType: RewardTypeRelicItem,
		Metadata:   &metaRecv,
	})

	return &GiftResult{
		RelicSlug:     relicSlug,
		RelicName:     def.Name,
		RecipientUID:  recipientUID,
		RecipientName: recipient.ExplorerName,
		SenderCount:   newSenderCount,
	}, nil
}
