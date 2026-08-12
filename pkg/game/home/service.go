package home

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"odyssey/pkg/game"
	"odyssey/pkg/game/achievement"
	"odyssey/pkg/game/course"
	"odyssey/pkg/game/gift"
	"odyssey/pkg/game/creative"
	"odyssey/pkg/game/familystreak"
	"odyssey/pkg/game/dailymission"
	"odyssey/pkg/game/concepts"
	"odyssey/pkg/game/progression"
	"odyssey/pkg/game/mission"
	"odyssey/pkg/game/collection"
	"odyssey/pkg/game/season"
)

// HomeResponse is the aggregated data returned by the Home endpoint.
type HomeResponse struct {
	Player                game.Player                   `json:"player"`
	Missions                []quest.QuestView             `json:"missions"`
	DailyMission             DailyTurnView                 `json:"daily_mission"`
	JourneyProgress         []game.JourneyProgress          `json:"journey_progress"`
	RelicCount            int                           `json:"relic_count"`
	ActiveQuests          []quest.QuestView             `json:"active_missions"`
	CompletedQuestsToday  []quest.QuestView             `json:"completed_missions_today"`
	PendingCreativeReview int                           `json:"pending_creative_review"`
	LastSubmission        *creative.SubmissionView      `json:"last_submission,omitempty"`
	AvailableChests       []chest.ChestView             `json:"available_gifts"`
	LatestRelic           *relic.InventoryItem          `json:"latest_relic,omitempty"`
	CollectionProgress    CollectionProgress            `json:"collection_progress"`
	CourseProgress       *course.ChapterProgressView  `json:"course_progress,omitempty"`
	LoreSummary           *concept.LoreSummary             `json:"concept_summary,omitempty"`
	Achievements          []achievement.AchievementView `json:"achievements,omitempty"`
	CurrentSeason         *season.SeasonSummary         `json:"current_season,omitempty"`
	SeasonProgress        SeasonProgress                `json:"season_progress"`
	Sections              HomeSections                  `json:"sections"`
}

// SeasonProgress tracks crew progress within the current season.
type SeasonProgress struct {
	SeasonSlug      string `json:"season_slug"`
	SeasonName      string `json:"season_name"`
	QuestsCompleted int    `json:"missions_completed"`
	JourneyProgress   int    `json:"journey_progress"`
	RealmStatus     string `json:"journey_status"`
}

// CollectionProgress tracks relic collection completion.
type CollectionProgress struct {
	Collected int `json:"collected"`
	Total     int `json:"total"`
}

// HomeSections groups HomeResponse data into logical views.
type HomeSections struct {
	Player       PlayerSection       `json:"player"`
	Missions       QuestsSection       `json:"missions"`
	DailyMission    DailyTurnSection    `json:"daily_mission"`
	Journey        RealmSection        `json:"journey"`
	World        WorldSection        `json:"world"`
	Creative     CreativeSection     `json:"creative"`
	Gifts       ChestsSection       `json:"gifts"`
	Collections       RelicsSection       `json:"collections"`
	Concept         LoreSection         `json:"concept"`
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
	MissionSlug      string `json:"mission_slug,omitempty"`
}

// RealmSection holds the crew's journey progression summary.
type RealmSection struct {
	Progress []game.JourneyProgress `json:"progress"`
}

// CreativeSection holds the crew's creative summary.
type CreativeSection struct {
	PendingReviewCount int                      `json:"pending_review_count"`
	LastSubmission     *creative.SubmissionView `json:"last_submission,omitempty"`
}

// ChestsSection holds the user's available gifts.
type ChestsSection struct {
	Available []chest.ChestView `json:"available"`
}

// RelicsSection holds the user's relic summary.
type RelicsSection struct {
	Latest             *relic.InventoryItem `json:"latest,omitempty"`
	CollectionProgress CollectionProgress   `json:"collection_progress"`
}

// WorldSection holds the crew's course progression summary.
type WorldSection struct {
	CurrentChapter    *course.ChapterSummary  `json:"current_course,omitempty"`
	NextChapter       *course.ChapterSummary  `json:"next_course,omitempty"`
	CompletedChapters []course.ChapterSummary `json:"completed_chapters"`
	UnlockedChapters  []course.ChapterSummary `json:"unlocked_chapters"`
	AllChapters       []course.ChapterSummary `json:"all_chapters"`
}

// LoreSection holds the crew's concept discovery summary.
type LoreSection struct {
	Summary  *concept.LoreSummary `json:"summary,omitempty"`
	Unlocked []concept.LoreView   `json:"unlocked"`
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
	MissionSlug      string `json:"mission_slug,omitempty"`
}

// HomeService aggregates data from multiple sub-services into a
// single home-screen response.
type HomeService struct {
	qs         *quest.QuestService
	dts        *dailymission.DailyTurnService
	prog       game.ProgressionStore
	journey      game.RealmProgressStore
	users      game.UserStore
	crea       game.CreativeSubmissionStore
	gifts     game.ChestStore
	relSvc     *relic.RelicService
	chapterSvc *course.ChapterService
	loreSvc    *concept.LoreService
	achieveSvc *achievement.AchievementService
	crewStreak *familystreak.Service
	seasonSvc  *season.SeasonService
}

func (s *HomeService) SetChapterService(cs *course.ChapterService) {
	s.chapterSvc = cs
}

func (s *HomeService) SetLoreService(ls *concept.LoreService) {
	s.loreSvc = ls
}

func (s *HomeService) SetAchievementService(as *achievement.AchievementService) {
	s.achieveSvc = as
}

// SetCrewStreakService attaches the crew-level streak computation. Safe to
// call with nil (crew streak simply stays 0).
func (s *HomeService) SetCrewStreakService(cs *familystreak.Service) {
	s.crewStreak = cs
}

func (s *HomeService) SetSeasonService(ss *season.SeasonService) {
	s.seasonSvc = ss
}

// NewHomeService constructs a HomeService from its collaborators.
func NewHomeService(
	qs *quest.QuestService,
	dts *dailymission.DailyTurnService,
	prog game.ProgressionStore,
	journey game.RealmProgressStore,
	users game.UserStore,
	crea game.CreativeSubmissionStore,
	gifts game.ChestStore,
	relSvc *relic.RelicService,
) *HomeService {
	return &HomeService{
		qs:     qs,
		dts:    dts,
		prog:   prog,
		journey:  journey,
		users:  users,
		crea:   crea,
		gifts: gifts,
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
		missions             []quest.QuestView
		today              = s.dts.TodayDate()
		hasCompleted       bool
		available          bool
		streak             int
		crewStreak         int
		realmProgress      []game.JourneyProgress
		relicCount         int
		availableChests    []chest.ChestView
		latestRelic        *relic.InventoryItem
		collectionProgress CollectionProgress
		pendingCount       int
		lastSub            *game.Submission
		chapterProgView    *course.ChapterProgressView
		allChapters        []course.ChapterSummary
		loreSummary        *concept.LoreSummary
		unlockedLore       []concept.LoreView
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

	// 2. Missions
	g.Go(func() error {
		var err error
		missions, err = s.qs.List(gCtx, crewID)
		if err != nil {
			return fmt.Errorf("failed to list missions: %w", err)
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

	// 3b. Family streak (display-only: failures never break Home)
	g.Go(func() error {
		if s.crewStreak == nil {
			return nil
		}
		crewStreak, _ = s.crewStreak.ComputeCrewStreak(gCtx, crewID)
		return nil
	})

	// 4. Journey Progress
	g.Go(func() error {
		var err error
		realmProgress, err = s.journey.ListRealmProgressByCrew(gCtx, crewID)
		if err != nil {
			return fmt.Errorf("failed to list journey progress: %w", err)
		}
		if s.chapterSvc != nil {
			activeRealm := firstActiveRealm(realmProgress)
			if activeRealm != "" {
				_ = s.chapterSvc.EnsureFirstChapterUnlocked(gCtx, crewID, activeRealm)
			}
		}
		return nil
	})

	// 5. Collections count
	g.Go(func() error {
		var err error
		relicCount, err = s.prog.CountRelics(gCtx, uid)
		if err != nil {
			return fmt.Errorf("failed to count collections: %w", err)
		}
		return nil
	})

	// 6. Gifts
	if s.gifts != nil {
		g.Go(func() error {
			allChests, err := s.gifts.ListChestsByUser(gCtx, uid)
			if err != nil {
				return fmt.Errorf("failed to list gifts: %w", err)
			}
			for _, ch := range allChests {
				if !ch.Opened {
					availableChests = append(availableChests, chest.ToChestView(ch))
				}
			}
			return nil
		})
	}

	// 7. Collections Service
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

	// 9. Course Service
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

	// 10. Concept Service
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

	var currentSeason *season.SeasonSummary
	seasonProgress := SeasonProgress{}
	if s.seasonSvc != nil {
		summary, err := s.seasonSvc.GetCurrentSeason(gCtx)
		if err == nil && summary != nil {
			currentSeason = summary
			seasonProgress.SeasonSlug = summary.Definition.Slug
			seasonProgress.SeasonName = summary.Definition.Name
			for _, rp := range realmProgress {
				if rp.Journey == summary.Definition.Journey {
					seasonProgress.JourneyProgress = rp.Progress
					seasonProgress.RealmStatus = rp.Status
					break
				}
			}
			for _, q := range missions {
				if q.Status == string(quest.QuestStatusDone) && q.SeasonSlug == summary.Definition.Slug {
					seasonProgress.QuestsCompleted++
				}
			}
		}
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
	for _, q := range missions {
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
		Missions: missions,
		DailyMission: DailyTurnView{
			Today:          today,
			Completed:      hasCompleted,
			Available:      available,
			StreakDays:     streak,
			CrewStreak:     crewStreak,
			RemainingTurns: remainingTurns,
		},
		JourneyProgress:        realmProgress,
		RelicCount:           relicCount,
		ActiveQuests:         activeQuests,
		CompletedQuestsToday: completedToday,
		AvailableChests:      availableChests,
		LatestRelic:          latestRelic,
		CollectionProgress:   collectionProgress,
		CurrentSeason:        currentSeason,
		SeasonProgress:       seasonProgress,
		Sections: HomeSections{
			Player: PlayerSection{
				Player:   *player,
				XPToNext: progression.XPForLevel(player.Level+1) - player.XP,
			},
			Missions: QuestsSection{
				All:       missions,
				Active:    activeQuests,
				Done:      doneQuests,
				DoneToday: completedToday,
			},
			DailyMission: DailyTurnSection{
				Today:          today,
				Completed:      hasCompleted,
				Available:      available,
				StreakDays:     streak,
				CrewStreak:     crewStreak,
				RemainingTurns: remainingTurns,
			},
			Journey: RealmSection{
				Progress: realmProgress,
			},
			Gifts: ChestsSection{
				Available: availableChests,
			},
			Collections: RelicsSection{
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
		resp.CourseProgress = chapterProgView
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
			resp.Sections.Concept = LoreSection{
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

func firstActiveRealm(progress []game.JourneyProgress) string {
	for _, rp := range progress {
		if rp.Status == "ACTIVE" {
			return rp.Journey
		}
	}
	return ""
}
