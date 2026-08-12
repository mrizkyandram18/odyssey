package achievement

import (
	"context"

	"odyssey/pkg/game"
)

type progressReader struct {
	missions   game.QuestStore
	realms   game.RealmProgressStore
	chapters game.ChapterProgressStore
	users    game.UserStore
	relocs   game.PlayerRelicStore
	daily    game.DailyTurnStore
	creative game.CreativeSubmissionStore
}

func NewProgressReader(
	missions game.QuestStore,
	realms game.RealmProgressStore,
	users game.UserStore,
	relocs game.PlayerRelicStore,
	daily game.DailyTurnStore,
	creative game.CreativeSubmissionStore,
	chapters game.ChapterProgressStore,
) ProgressReader {
	return &progressReader{
		missions:   missions,
		realms:   realms,
		chapters: chapters,
		users:    users,
		relocs:   relocs,
		daily:    daily,
		creative: creative,
	}
}

func (r *progressReader) CountCompletedQuests(ctx context.Context, crewID string) (int, error) {
	missions, err := r.missions.ListQuestByCrew(ctx, crewID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, q := range missions {
		if q.Status == "DONE" {
			count++
		}
	}
	return count, nil
}

func (r *progressReader) CountCompletedRealms(ctx context.Context, crewID string) (int, error) {
	rps, err := r.realms.ListRealmProgressByCrew(ctx, crewID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, rp := range rps {
		if rp.Status == "COMPLETE" {
			count++
		}
	}
	return count, nil
}

func (r *progressReader) CountCompletedChapters(ctx context.Context, crewID string) (int, error) {
	cps, err := r.chapters.ListChapterProgressByCrew(ctx, crewID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, cp := range cps {
		if cp.Status == "COMPLETE" {
			count++
		}
	}
	return count, nil
}

func (r *progressReader) CountCollectedRelics(ctx context.Context, uid string) (int, error) {
	return r.relocs.CountUniqueRelics(ctx, uid)
}

func (r *progressReader) CountDailyStreak(ctx context.Context, uid string) (int, error) {
	turns, err := r.daily.ListDailyTurns(ctx, uid)
	if err != nil {
		return 0, err
	}
	streak := 0
	for _, t := range turns {
		if t.Completed {
			streak++
		}
	}
	return streak, nil
}

func (r *progressReader) CountCreativeSubmissions(ctx context.Context, crewID string) (int, error) {
	subs, err := r.creative.ListByCrew(ctx, crewID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, s := range subs {
		if s.Status == game.SubmissionStatusApproved {
			count++
		}
	}
	return count, nil
}

func (r *progressReader) GetPlayerLevel(ctx context.Context, uid string) (int, error) {
	p, err := r.users.GetUser(ctx, uid)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, nil
	}
	return p.Level, nil
}
