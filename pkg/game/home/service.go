package home

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"odyssey/pkg/game"
	"odyssey/pkg/game/achievement"
	"odyssey/pkg/game/chapter"
	"odyssey/pkg/game/chest"
	"odyssey/pkg/game/creative"
	"odyssey/pkg/game/crewstreak"
	"odyssey/pkg/game/dailyturn"
	"odyssey/pkg/game/lore"
	"odyssey/pkg/game/progression"
	"odyssey/pkg/game/quest"
	"odyssey/pkg/game/relic"
)

// HomeResponse is the aggregated data returned by the Home endpoint.
type HomeResponse struct {
	Player                game.Player                   `json:"player"`
	Quests                []quest.QuestView             `json:"quests"`
	DailyTurn             DailyTurnView                 `json:"daily_turn"`
	RealmProgress         []game.RealmProgress          `json:"realm_progress"`
	RelicCount            int                           `json:"relic_count"`
	ActiveQuests          []quest.QuestView             `json:"active_quests"`
	CompletedQuestsToday  []quest.QuestView             `json:"completed_quests_today"`
	PendingCreativeReview int                           `json:"pending_creative_review"`
	LastSubmission        *creative.SubmissionView      `json:"last_submission,omitempty"`
	AvailableChests       []chest.ChestView             `json:"available_chests"`
	LatestRelic           *relic.InventoryItem          `json:"latest_relic,omitempty"`
	CollectionProgress    CollectionProgress            `json:"collection_progress"`
	ChapterProgress       *chapter.ChapterProgressView  `json:"chapter_progress,omitempty"`
	LoreSummary           *lore.LoreSummary             `json:"lore_summary,omitempty"`
	Achievements          []achievement.AchievementView `json:"achievements,omitempty"`
	Sections              HomeSections                  `json:"sections"`
}

// CollectionProgress tracks relic collection completion.
type CollectionProgress struct {
	Collected int `json:"collected"`
	Total     int `json:"total"`
}

// HomeSections groups HomeResponse data into logical views.
type HomeSections struct {
	Player       PlayerSection       `json:"player"`
	Quests       QuestsSection       `json:"quests"`
	DailyTurn    DailyTurnSection    `json:"daily_turn"`
	Realm        RealmSection        `json:"realm"`
	World        WorldSection        `json:"world"`
	Creative     CreativeSection     `json:"creative"`
	Chests       ChestsSection       `json:"chests"`
	Relics       RelicsSection       `json:"relics"`
	Lore         LoreSection         `json:"lore"`
	Achievements AchievementsSection `json:"achievements"`
}

// PlayerSection holds the explorer's personal progression summary.
type PlayerSection struct {
	game.Player `json:"-"`
	XPToNext    int64 `json:"xp_to_next"`
}

// QuestsSection holds all quest data with derived summary counts.
type QuestsSection struct {
	All       []quest.QuestView `json:"all"`
	Active    []quest.QuestView `json:"active"`
	Done      []quest.QuestView `json:"done"`
	DoneToday []quest.QuestView `json:"done_today"`
}

// DailyTurnSection holds the daily turn state for the home screen.
type DailyTurnSection struct {
	Today          string `json:"today"`
	Completed      bool   `json:"completed"`
	Available      bool   `json:"available"`
	StreakDays     int    `json:"streak_days"`
	CrewStreak     int    `json:"crew_streak"`
	RemainingTurns int    `json:"remaining_turns"`
	QuestSlug      string `json:"quest_slug,omitempty"`
}

// RealmSection holds the crew's realm progression summary.
type RealmSection struct {
	Progress []game.RealmProgress `json:"progress"`
}

// CreativeSection holds the crew's creative summary.
type CreativeSection struct {
	PendingReviewCount int                      `json:"pending_review_count"`
	LastSubmission     *creative.SubmissionView `json:"last_submission,omitempty"`
}

// ChestsSection holds the user's available chests.
type ChestsSection struct {
	Available []chest.ChestView `json:"available"`
}

// RelicsSection holds the user's relic summary.
type RelicsSection struct {
	Latest             *relic.InventoryItem `json:"latest,omitempty"`
	CollectionProgress CollectionProgress   `json:"collection_progress"`
}

// WorldSection holds the crew's chapter progression summary.
type WorldSection struct {
	CurrentChapter    *chapter.ChapterSummary  `json:"current_chapter,omitempty"`
	NextChapter       *chapter.ChapterSummary  `json:"next_chapter,omitempty"`
	CompletedChapters []chapter.ChapterSummary `json:"completed_chapters"`
	UnlockedChapters  []chapter.ChapterSummary `json:"unlocked_chapters"`
	AllChapters       []chapter.ChapterSummary `json:"all_chapters"`
}

// LoreSection holds the crew's lore discovery summary.
type LoreSection struct {
	Summary  *lore.LoreSummary `json:"summary,omitempty"`
	Unlocked []lore.LoreView   `json:"unlocked"`
}

// AchievementsSection holds the player's achievement progress.
type AchievementsSection struct {
	All    []achievement.AchievementView `json:"all"`
	Recent []achievement.AchievementView `json:"recent"`
	Count  int                           `json:"count"`
}

// DailyTurnView is the daily-turn summary for the home screen.
type DailyTurnView struct {
	Today          string `json:"today"`
	Completed      bool   `json:"completed"`
	Available      bool   `json:"available"`
	StreakDays     int    `json:"streak_days"`
	CrewStreak     int    `json:"crew_streak"`
	RemainingTurns int    `json:"remaining_turns"`
	QuestSlug      string `json:"quest_slug,omitempty"`
}

// HomeService aggregates data from multiple sub-services into a
// single home-screen response.
type HomeService struct {
	qs         *quest.QuestService
	dts        *dailyturn.DailyTurnService
	prog       game.ProgressionStore
	realm      game.RealmProgressStore
	users      game.UserStore
	crea       game.CreativeSubmissionStore
	chests     game.ChestStore
	relSvc     *relic.RelicService
	chapterSvc *chapter.ChapterService
	loreSvc    *lore.LoreService
	achieveSvc *achievement.AchievementService
	crewStreak *crewstreak.Service
}

func (s *HomeService) SetChapterService(cs *chapter.ChapterService) {
	s.chapterSvc = cs
}

func (s *HomeService) SetLoreService(ls *lore.LoreService) {
	s.loreSvc = ls
}

func (s *HomeService) SetAchievementService(as *achievement.AchievementService) {
	s.achieveSvc = as
}

// SetCrewStreakService attaches the crew-level streak computation. Safe to
// call with nil (crew streak simply stays 0).
func (s *HomeService) SetCrewStreakService(cs *crewstreak.Service) {
	s.crewStreak = cs
}

// NewHomeService constructs a HomeService from its collaborators.
func NewHomeService(
	qs *quest.QuestService,
	dts *dailyturn.DailyTurnService,
	prog game.ProgressionStore,
	realm game.RealmProgressStore,
	users game.UserStore,
	crea game.CreativeSubmissionStore,
	chests game.ChestStore,
	relSvc *relic.RelicService,
) *HomeService {
	return &HomeService{
		qs:     qs,
		dts:    dts,
		prog:   prog,
		realm:  realm,
		users:  users,
		crea:   crea,
		chests: chests,
		relSvc: relSvc,
	}
}

// isToday reports whether t falls on the current date in the given location.
func isToday(t time.Time, loc *time.Location) bool {
	now := time.Now().In(loc)
	y1, m1, d1 := now.Date()
	y2, m2, d2 := t.In(loc).Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// GetHome builds the aggregated home-screen response for a given user.
func (s *HomeService) GetHome(ctx context.Context, uid string, crewID string) (*HomeResponse, error) {
	g, gCtx := errgroup.WithContext(ctx)

	var (
		player             *game.Player
		quests             []quest.QuestView
		today              = s.dts.TodayDate()
		hasCompleted       bool
		available          bool
		streak             int
		crewStreak         int
		realmProgress      []game.RealmProgress
		relicCount         int
		availableChests    []chest.ChestView
		latestRelic        *relic.InventoryItem
		collectionProgress CollectionProgress
		pendingCount       int
		lastSub            *game.Submission
		chapterProgView    *chapter.ChapterProgressView
		allChapters        []chapter.ChapterSummary
		loreSummary        *lore.LoreSummary
		unlockedLore       []lore.LoreView
		achievements       []achievement.AchievementView
	)

	// 1. Player
	g.Go(func() error {
		var err error
		player, err = s.users.GetUser(gCtx, uid)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}
		return nil
	})

	// 2. Quests
	g.Go(func() error {
		var err error
		quests, err = s.qs.List(gCtx, crewID)
		if err != nil {
			return fmt.Errorf("failed to list quests: %w", err)
		}
		return nil
	})

	// 3. Daily Turn
	g.Go(func() error {
		var err error
		hasCompleted, err = s.dts.HasCompletedToday(gCtx, uid, today)
		if err != nil {
			return fmt.Errorf("failed to check daily turn completion: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		available, err = s.dts.IsAvailableToday(gCtx, uid, today)
		if err != nil {
			return fmt.Errorf("failed to check daily turn availability: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		streak, err = s.dts.ComputeStreak(gCtx, uid)
		if err != nil {
			return fmt.Errorf("failed to compute streak: %w", err)
		}
		return nil
	})

	// 3b. Crew streak (display-only: failures never break Home)
	g.Go(func() error {
		if s.crewStreak == nil {
			return nil
		}
		crewStreak, _ = s.crewStreak.ComputeCrewStreak(gCtx, crewID)
		return nil
	})

	// 4. Realm Progress
	g.Go(func() error {
		var err error
		realmProgress, err = s.realm.ListRealmProgressByCrew(gCtx, crewID)
		if err != nil {
			return fmt.Errorf("failed to list realm progress: %w", err)
		}
		if s.chapterSvc != nil {
			activeRealm := firstActiveRealm(realmProgress)
			if activeRealm != "" {
				_ = s.chapterSvc.EnsureFirstChapterUnlocked(gCtx, crewID, activeRealm)
			}
		}
		return nil
	})

	// 5. Relics count
	g.Go(func() error {
		var err error
		relicCount, err = s.prog.CountRelics(gCtx, uid)
		if err != nil {
			return fmt.Errorf("failed to count relics: %w", err)
		}
		return nil
	})

	// 6. Chests
	if s.chests != nil {
		g.Go(func() error {
			allChests, err := s.chests.ListChestsByUser(gCtx, uid)
			if err != nil {
				return fmt.Errorf("failed to list chests: %w", err)
			}
			for _, ch := range allChests {
				if !ch.Opened {
					availableChests = append(availableChests, chest.ToChestView(ch))
				}
			}
			return nil
		})
	}

	// 7. Relics Service
	if s.relSvc != nil {
		g.Go(func() error {
			latestRelic, _ = s.relSvc.GetLatestRelic(gCtx, uid)
			return nil
		})
		g.Go(func() error {
			col, tot, err := s.relSvc.GetCollectionProgress(gCtx, uid)
			if err == nil {
				collectionProgress = CollectionProgress{Collected: col, Total: tot}
			}
			return nil
		})
	}

	// 8. Creative
	if s.crea != nil {
		g.Go(func() error {
			allSubs, err := s.crea.ListByCrew(gCtx, crewID)
			if err != nil {
				return fmt.Errorf("failed to list creative submissions: %w", err)
			}
			for i := range allSubs {
				sub := &allSubs[i]
				if sub.Status == game.SubmissionStatusPending {
					pendingCount++
				}
				if sub.AuthorUID == uid && (lastSub == nil || sub.CreatedAt.After(lastSub.CreatedAt)) {
					lastSub = sub
				}
			}
			return nil
		})
	}

	// 9. Chapter Service
	if s.chapterSvc != nil {
		g.Go(func() error {
			chapterProgView, _ = s.chapterSvc.GetProgressView(gCtx, crewID)
			return nil
		})
		g.Go(func() error {
			allChapters, _ = s.chapterSvc.ListProgress(gCtx, crewID)
			return nil
		})
	}

	// 10. Lore Service
	if s.loreSvc != nil {
		g.Go(func() error {
			loreSummary, _ = s.loreSvc.GetSummary(gCtx, crewID)
			return nil
		})
		g.Go(func() error {
			unlockedLore, _ = s.loreSvc.ListUnlocked(gCtx, crewID)
			return nil
		})
	}

	// 11. Achievements Service
	if s.achieveSvc != nil {
		g.Go(func() error {
			achievements, _ = s.achieveSvc.ListByPlayer(gCtx, uid)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Assembly phase
	remainingTurns := 0
	if available && !hasCompleted {
		remainingTurns = 1
	}

	tz := s.dts.Config().Timezone
	tzLoc, tzErr := time.LoadLocation(tz)
	if tzErr != nil {
		tzLoc = time.UTC
	}
	activeQuests := make([]quest.QuestView, 0)
	completedToday := make([]quest.QuestView, 0)
	doneQuests := make([]quest.QuestView, 0)
	for _, q := range quests {
		if q.Status != string(quest.QuestStatusDone) {
			activeQuests = append(activeQuests, q)
		} else {
			doneQuests = append(doneQuests, q)
		}
		if q.CompletedAt != nil && isToday(*q.CompletedAt, tzLoc) {
			completedToday = append(completedToday, q)
		}
	}

	resp := &HomeResponse{
		Player: *player,
		Quests: quests,
		DailyTurn: DailyTurnView{
			Today:          today,
			Completed:      hasCompleted,
			Available:      available,
			StreakDays:     streak,
			CrewStreak:     crewStreak,
			RemainingTurns: remainingTurns,
		},
		RealmProgress:        realmProgress,
		RelicCount:           relicCount,
		ActiveQuests:         activeQuests,
		CompletedQuestsToday: completedToday,
		AvailableChests:      availableChests,
		LatestRelic:          latestRelic,
		CollectionProgress:   collectionProgress,
		Sections: HomeSections{
			Player: PlayerSection{
				Player:   *player,
				XPToNext: progression.XPForLevel(player.Level+1) - player.XP,
			},
			Quests: QuestsSection{
				All:       quests,
				Active:    activeQuests,
				Done:      doneQuests,
				DoneToday: completedToday,
			},
			DailyTurn: DailyTurnSection{
				Today:          today,
				Completed:      hasCompleted,
				Available:      available,
				StreakDays:     streak,
				CrewStreak:     crewStreak,
				RemainingTurns: remainingTurns,
			},
			Realm: RealmSection{
				Progress: realmProgress,
			},
			Chests: ChestsSection{
				Available: availableChests,
			},
			Relics: RelicsSection{
				Latest:             latestRelic,
				CollectionProgress: collectionProgress,
			},
		},
	}

	if s.crea != nil {
		creativeSection := CreativeSection{
			PendingReviewCount: pendingCount,
		}
		if lastSub != nil {
			creativeSection.LastSubmission = creative.ToView(lastSub)
		}
		resp.PendingCreativeReview = pendingCount
		resp.LastSubmission = creativeSection.LastSubmission
		resp.Sections.Creative = creativeSection
	}

	if s.chapterSvc != nil && chapterProgView != nil {
		resp.ChapterProgress = chapterProgView
		resp.Sections.World = WorldSection{
			CurrentChapter:    chapterProgView.CurrentChapter,
			NextChapter:       chapterProgView.NextChapter,
			CompletedChapters: chapterProgView.CompletedChapters,
			UnlockedChapters:  chapterProgView.UnlockedChapters,
			AllChapters:       chapterProgView.UnlockedChapters,
		}
		if allChapters != nil {
			resp.Sections.World.AllChapters = allChapters
		}
	}

	if s.loreSvc != nil && loreSummary != nil {
		resp.LoreSummary = loreSummary
		if unlockedLore != nil {
			resp.Sections.Lore = LoreSection{
				Summary:  loreSummary,
				Unlocked: unlockedLore,
			}
		}
	}

	if s.achieveSvc != nil && achievements != nil {
		resp.Achievements = achievements
		count := 0
		for _, a := range achievements {
			if a.Unlocked {
				count++
			}
		}
		resp.Sections.Achievements = AchievementsSection{
			All:   achievements,
			Count: count,
		}
	}

	return resp, nil
}

func firstActiveRealm(progress []game.RealmProgress) string {
	for _, rp := range progress {
		if rp.Status == "ACTIVE" {
			return rp.Realm
		}
	}
	return ""
}
