package chest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/balance"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
	"odyssey/pkg/game/relic"
	"odyssey/pkg/observability"
)

// ErrChestNotFound is returned when a chest does not exist.
var ErrChestNotFound = errors.New("chest not found")

// ChestService handles chest listing, retrieval, creation, and opening.
type ChestService struct {
	chests    game.ChestStore
	playerRel game.PlayerRelicStore
	relics    game.RelicStore
	users     game.UserStore
	engine    *RewardEngine
	publisher events.Publisher
	content   ContentGateway
	balance   *balance.Service
	metrics   *observability.Metrics
}

// SetMetrics attaches an optional metrics sink for chest/reward telemetry.
// Safe to call with nil (metrics recording is disabled).
func (s *ChestService) SetMetrics(m *observability.Metrics) {
	s.metrics = m
}

// ContentGateway provides quest definition lookups for chest creation handlers.
type ContentGateway interface {
	GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error)
}

// NewChestService constructs a ChestService.
func NewChestService(
	chests game.ChestStore,
	playerRel game.PlayerRelicStore,
	relics game.RelicStore,
	users game.UserStore,
	engine *RewardEngine,
) *ChestService {
	return &ChestService{
		chests:    chests,
		playerRel: playerRel,
		relics:    relics,
		users:     users,
		engine:    engine,
		publisher: events.NopPublisher{},
	}
}

func NewChestServiceWithPublisher(
	chests game.ChestStore,
	playerRel game.PlayerRelicStore,
	relics game.RelicStore,
	users game.UserStore,
	engine *RewardEngine,
	publisher events.Publisher,
) *ChestService {
	if publisher == nil {
		publisher = events.NopPublisher{}
	}
	return &ChestService{
		chests:    chests,
		playerRel: playerRel,
		relics:    relics,
		users:     users,
		engine:    engine,
		publisher: publisher,
	}
}

// SetContentGateway attaches a content gateway for quest definition lookups.
func (s *ChestService) SetContentGateway(cg ContentGateway) {
	s.content = cg
}

// SetBalance attaches a balance service for drop rate and reward overrides.
func (s *ChestService) SetBalance(b *balance.Service) {
	s.balance = b
	if s.engine != nil {
		s.engine.SetBalance(b)
	}
}

// CreateChest creates a chest instance for a player from a chest definition slug.
// The source indicates how the chest was earned (e.g. "QUEST", "LEVEL_UP").
func (s *ChestService) CreateChest(ctx context.Context, uid, chestSlug, source string) (*game.Chest, error) {
	ct := s.engine.GetChestType(ctx, chestSlug)
	if ct == nil {
		return nil, fmt.Errorf("chest definition not found: %s", chestSlug)
	}

	now := time.Now().UTC()
	ch := &game.Chest{
		UID:         uid,
		Source:      source,
		ChestSlug:   chestSlug,
		Rarity:      string(ct.Rarity),
		Icon:        ct.Icon,
		Description: ct.Description,
		DropTable:   "",
		Opened:      false,
		CreatedAt:   now,
	}

	created, err := s.chests.CreateChest(ctx, ch)
	if err != nil {
		return nil, fmt.Errorf("create chest: %w", err)
	}

	if s.metrics != nil {
		s.metrics.RecordChestCreated()
	}

	return created, nil
}

// ListChests returns all chests owned by a user.
func (s *ChestService) ListChests(ctx context.Context, uid string) ([]game.Chest, error) {
	return s.chests.ListChestsByUser(ctx, uid)
}

// GetChest returns a single chest by ID, ensuring it belongs to the user.
func (s *ChestService) GetChest(ctx context.Context, chestID int64, uid string) (*game.Chest, error) {
	ch, err := s.chests.GetChest(ctx, chestID)
	if err != nil {
		return nil, fmt.Errorf("get chest: %w", err)
	}
	if ch == nil {
		return nil, ErrChestNotFound
	}
	if ch.UID != uid {
		return nil, ErrChestNotFound
	}
	return ch, nil
}

// rewardCountByRarity maps chest rarity to reward count using defaults.
// Callers with a balance service should use ChestService.rewardCountForRarity
// to apply runtime overrides.
func rewardCountByRarity(r game.Rarity) int {
	switch r {
	case game.RarityCommon:
		return 1
	case game.RarityUncommon:
		return 2
	case game.RarityRare:
		return 2
	case game.RarityEpic:
		return 3
	case game.RarityLegendary:
		return 4
	}
	return 1
}

// rewardCountForRarity returns the reward count for a chest of the given rarity,
// applying balance overrides when a balance service is available.
func (s *ChestService) rewardCountForRarity(r game.Rarity) int {
	def := rewardCountByRarity(r)
	if s.balance != nil {
		return s.balance.OverrideChestRewardCount(r, def)
	}
	return def
}

// OpenChest opens a chest and generates rewards based on its drop table.
func (s *ChestService) OpenChest(ctx context.Context, chestID int64, uid string) (*OpenResult, error) {
	ch, err := s.GetChest(ctx, chestID, uid)
	if err != nil {
		return nil, err
	}

	if ch.Opened {
		if s.metrics != nil {
			s.metrics.RecordReplayIgnored()
		}
		return nil, errors.New("chest already opened")
	}

	rewardCount := s.rewardCountForRarity(game.Rarity(ch.Rarity))
	rewards := s.engine.GenerateRewardsForChest(ctx, ch.ChestSlug, rewardCount)
	if s.metrics != nil {
		s.metrics.RecordRewardsGenerated(len(rewards))
	}

	now := time.Now().UTC()
	newCount := 0
	duplicateCount := 0

	for i := range rewards {
		rewards[i].IsNew = false
		existing, err := s.playerRel.GetPlayerRelic(ctx, uid, rewards[i].RelicSlug)
		if err != nil && !errors.Is(err, game.ErrNotFound) {
			return nil, fmt.Errorf("check player relic: %w", err)
		}

		if existing == nil {
			rewards[i].IsNew = true
			newCount++
			def := relic.GetRelicDefinition(rewards[i].RelicSlug)
			relicID := int64(0)
			if def != nil {
				relicID = def.ID
			}
			pr := &game.PlayerRelic{
				UID:          uid,
				RelicSlug:    rewards[i].RelicSlug,
				RelicID:      relicID,
				OwnedCount:   1,
				IsNew:        true,
				DiscoveredAt: now,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if _, err := s.playerRel.CreatePlayerRelic(ctx, pr); err != nil {
				return nil, fmt.Errorf("create player relic: %w", err)
			}

			s.publisher.Publish(ctx, events.RelicCollectedEvent{
				UID:       uid,
				CrewID:    "",
				RelicSlug: rewards[i].RelicSlug,
				Realm:     rewards[i].Realm,
			})
		} else {
			duplicateCount++
			patch := map[string]any{
				"owned_count": existing.OwnedCount + 1,
				"is_new":      false,
				"updated_at":  now,
			}
			if err := s.playerRel.UpdatePlayerRelic(ctx, uid, rewards[i].RelicSlug, patch); err != nil {
				return nil, fmt.Errorf("update player relic: %w", err)
			}
		}

		instance := &game.Relic{
			UID:         uid,
			Code:        rewards[i].RelicSlug,
			Name:        rewards[i].Name,
			Description: "",
			Realm:       "",
			Rarity:      string(rewards[i].Rarity),
			AwardedAt:   now,
		}
		if def := relic.GetRelicDefinition(rewards[i].RelicSlug); def != nil {
			instance.Description = def.Description
			instance.Realm = def.Realm
			instance.Image = def.Image
			instance.Lore = def.Lore
		}
		if _, err := s.relics.CreateRelic(ctx, instance); err != nil {
			return nil, fmt.Errorf("create relic instance: %w", err)
		}
	}

	openedAt := now
	chestPatch := map[string]any{
		"opened":    true,
		"opened_at": openedAt,
	}
	matched, err := s.chests.UpdateChestIfMatch(ctx, chestID, false, chestPatch)
	if err != nil {
		return nil, fmt.Errorf("update chest: %w", err)
	}
	if !matched {
		if s.metrics != nil {
			s.metrics.RecordLockConflict()
		}
		return nil, errors.New("chest already opened")
	}

	if s.metrics != nil {
		s.metrics.RecordDuplicatePrevented(duplicateCount)
	}

	openedChest, _ := s.chests.GetChest(ctx, chestID)
	view := ToChestView(*openedChest)

	return &OpenResult{
		Chest:          &view,
		Rewards:        rewards,
		NewCount:       newCount,
		DuplicateCount: duplicateCount,
	}, nil
}

func ToChestView(ch game.Chest) ChestView {
	return ChestView{
		ID:          ch.ID,
		UID:         ch.UID,
		Source:      ch.Source,
		ChestSlug:   ch.ChestSlug,
		Name:        ch.Description,
		Rarity:      game.Rarity(ch.Rarity),
		Icon:        ch.Icon,
		Description: ch.Description,
		Opened:      ch.Opened,
		OpenedAt:    ch.OpenedAt,
		CreatedAt:   ch.CreatedAt,
	}
}

// QuestCompletedHandler creates a chest instance when a quest is completed
// and the quest definition specifies a reward_chest.
type QuestCompletedHandler struct {
	svc     *ChestService
	content ContentGateway
}

func NewQuestCompletedHandler(svc *ChestService, content ContentGateway) *QuestCompletedHandler {
	return &QuestCompletedHandler{svc: svc, content: content}
}

func (h *QuestCompletedHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.QuestCompletedEvent)
	if !ok {
		return nil
	}
	if h.content == nil {
		return nil
	}
	qd, err := h.content.GetQuest(ctx, e.TemplateSlug)
	if err != nil {
		return err
	}
	if qd == nil {
		return nil
	}
	if qd.RewardChest == "" {
		return nil
	}

	existing, err := h.svc.chests.ListChestsByUser(ctx, e.PlayerUID)
	if err == nil {
		for _, ch := range existing {
			if ch.Source == "QUEST" && ch.ChestSlug == qd.RewardChest && ch.UID == e.PlayerUID {
				if h.svc.metrics != nil {
					h.svc.metrics.RecordReplayIgnored()
				}
				return nil
			}
		}
	}

	_, err = h.svc.CreateChest(ctx, e.PlayerUID, qd.RewardChest, "QUEST")
	return err
}
