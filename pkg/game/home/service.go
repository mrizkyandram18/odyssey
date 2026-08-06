package home

import (
	"context"
	"fmt"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/achievement"
	"odyssey/pkg/game/chapter"
	"odyssey/pkg/game/chest"
	"odyssey/pkg/game/creative"
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
	player, err := s.users.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	quests, err := s.qs.List(ctx, crewID)
	if err != nil {
		return nil, fmt.Errorf("failed to list quests: %w", err)
	}

	today := s.dts.TodayDate()
	hasCompleted, err := s.dts.HasCompletedToday(ctx, uid, today)
	if err != nil {
		return nil, fmt.Errorf("failed to check daily turn completion: %w", err)
	}
	available, err := s.dts.IsAvailableToday(ctx, uid, today)
	if err != nil {
		return nil, fmt.Errorf("failed to check daily turn availability: %w", err)
	}
	remainingTurns := 0
	if available && !hasCompleted {
		remainingTurns = 1
	}
	streak, err := s.dts.ComputeStreak(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to compute streak: %w", err)
	}

	realmProgress, err := s.realm.ListRealmProgressByCrew(ctx, crewID)
	if err != nil {
		return nil, fmt.Errorf("failed to list realm progress: %w", err)
	}

	if s.chapterSvc != nil {
		activeRealm := firstActiveRealm(realmProgress)
		if activeRealm != "" {
			_ = s.chapterSvc.EnsureFirstChapterUnlocked(ctx, crewID, activeRealm)
		}
	}

	relicCount, err := s.prog.CountRelics(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to count relics: %w", err)
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

	var availableChests []chest.ChestView
	if s.chests != nil {
		allChests, err := s.chests.ListChestsByUser(ctx, uid)
		if err != nil {
			return nil, fmt.Errorf("failed to list chests: %w", err)
		}
		for _, ch := range allChests {
			if !ch.Opened {
				availableChests = append(availableChests, chest.ToChestView(ch))
			}
		}
	}

	var latestRelic *relic.InventoryItem
	var collectionProgress CollectionProgress
	if s.relSvc != nil {
		latest, err := s.relSvc.GetLatestRelic(ctx, uid)
		if err == nil {
			latestRelic = latest
		}
		collected, total, err := s.relSvc.GetCollectionProgress(ctx, uid)
		if err == nil {
			collectionProgress = CollectionProgress{Collected: collected, Total: total}
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
		allSubs, err := s.crea.ListByCrew(ctx, crewID)
		if err != nil {
			return nil, fmt.Errorf("failed to list creative submissions: %w", err)
		}
		pendingCount := 0
		var lastSub *game.Submission
		for i := range allSubs {
			sub := &allSubs[i]
			if sub.Status == game.SubmissionStatusPending {
				pendingCount++
			}
			if sub.AuthorUID == uid && (lastSub == nil || sub.CreatedAt.After(lastSub.CreatedAt)) {
				lastSub = sub
			}
		}
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

	if s.chapterSvc != nil {
		progressView, err := s.chapterSvc.GetProgressView(ctx, crewID)
		if err == nil {
			resp.ChapterProgress = progressView
			resp.Sections.World = WorldSection{
				CurrentChapter:    progressView.CurrentChapter,
				NextChapter:       progressView.NextChapter,
				CompletedChapters: progressView.CompletedChapters,
				UnlockedChapters:  progressView.UnlockedChapters,
				AllChapters:       progressView.UnlockedChapters,
			}
			allChapters, err := s.chapterSvc.ListProgress(ctx, crewID)
			if err == nil {
				resp.Sections.World.AllChapters = allChapters
			}
		}
	}

	if s.loreSvc != nil {
		loreSummary, err := s.loreSvc.GetSummary(ctx, crewID)
		if err == nil {
			resp.LoreSummary = loreSummary
			unlocked, err := s.loreSvc.ListUnlocked(ctx, crewID)
			if err == nil {
				resp.Sections.Lore = LoreSection{
					Summary:  loreSummary,
					Unlocked: unlocked,
				}
			}
		}
	}

	if s.achieveSvc != nil {
		achieves, err := s.achieveSvc.ListByPlayer(ctx, uid)
		if err == nil {
			resp.Achievements = achieves
			count := 0
			for _, a := range achieves {
				if a.Unlocked {
					count++
				}
			}
			resp.Sections.Achievements = AchievementsSection{
				All:   achieves,
				Count: count,
			}
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
