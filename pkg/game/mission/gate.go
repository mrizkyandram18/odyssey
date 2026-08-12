package quest

import (
	"context"
	"errors"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
)

var ErrCircularPrerequisite = errors.New("circular quest prerequisite detected")

type SeasonChecker func(ctx context.Context, slug string) bool

type questGate struct {
	chapters    game.ChapterProgressStore
	realms      game.RealmProgressStore
	users       game.UserStore
	missions      game.QuestStore
	checkSeason SeasonChecker
}

func NewQuestGate(
	chapters game.ChapterProgressStore,
	realms game.RealmProgressStore,
	users game.UserStore,
	missions game.QuestStore,
	checkSeason SeasonChecker,
) QuestGate {
	return &questGate{
		chapters:    chapters,
		realms:      realms,
		users:       users,
		missions:      missions,
		checkSeason: checkSeason,
	}
}

func (g *questGate) IsChapterUnlocked(ctx context.Context, crewID, course string) bool {
	cp, err := g.chapters.GetChapterProgress(ctx, crewID, course)
	if err != nil || cp == nil {
		return false
	}
	return cp.Status == "ACTIVE" || cp.Status == "COMPLETE"
}

func (g *questGate) IsRealmActive(ctx context.Context, crewID, journey string) bool {
	rp, err := g.realms.GetRealmProgress(ctx, crewID, journey)
	if err != nil || rp == nil {
		return false
	}
	return rp.Status == "ACTIVE" || rp.Status == "COMPLETE"
}

func (g *questGate) IsSeasonActive(ctx context.Context, seasonSlug string) bool {
	if g.checkSeason == nil {
		return true
	}
	return g.checkSeason(ctx, seasonSlug)
}

func (g *questGate) GetPlayerLevel(ctx context.Context, uid string) (int, error) {
	p, err := g.users.GetUser(ctx, uid)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, nil
	}
	return p.Level, nil
}

func (g *questGate) IsQuestCompleted(ctx context.Context, crewID, templateSlug string) bool {
	missions, err := g.missions.ListQuestByCrew(ctx, crewID)
	if err != nil {
		return false
	}
	for _, q := range missions {
		if q.TemplateSlug == templateSlug && q.Status == "DONE" {
			return true
		}
	}
	return false
}

func (g *questGate) ValidatePrerequisites(ctx context.Context, defs []gamecontent.QuestDefinition) error {
	slugMap := make(map[string]gamecontent.QuestDefinition, len(defs))
	for i := range defs {
		slugMap[defs[i].Slug] = defs[i]
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(slug string) bool
	dfs = func(slug string) bool {
		if recStack[slug] {
			return true
		}
		if visited[slug] {
			return false
		}
		visited[slug] = true
		recStack[slug] = true

		def, ok := slugMap[slug]
		if !ok {
			recStack[slug] = false
			return false
		}

		prereqs := def.RequiredQuestSlugs
		if len(prereqs) == 0 && def.RequiredQuestSlug != "" {
			prereqs = []string{def.RequiredQuestSlug}
		}
		for _, prereq := range prereqs {
			if dfs(prereq) {
				return true
			}
		}
		recStack[slug] = false
		return false
	}

	for slug := range slugMap {
		if dfs(slug) {
			return ErrCircularPrerequisite
		}
	}
	return nil
}
