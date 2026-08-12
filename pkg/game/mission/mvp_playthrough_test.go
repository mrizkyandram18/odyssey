package quest

import (
	"context"
	"testing"
	"time"

	"odyssey/pkg/game"
)

// TestMVPPlaythrough_AllSixQuests exercises start + complete for every MVP
// Whispering Woods template using an in-memory store. This is the unit-level
// smoke for "all 6 missions are playable", not a live DB E2E.
func TestMVPPlaythrough_AllSixQuests(t *testing.T) {
	store := newMockQuestStore()
	svc := NewQuestService(store)
	ctx := context.Background()

	var nextQuestID int64 = 1
	var nextChallengeID int64 = 100

	for _, slug := range MVPQuestSlugs() {
		tpl, ok := LookupTemplate(slug)
		if !ok {
			t.Fatalf("missing template %s", slug)
		}

		qid := nextQuestID
		nextQuestID++
		now := time.Now().UTC()
		q := &game.Mission{
			ID:           qid,
			FamilyID:       "crew-mvp",
			TemplateSlug: tpl.Slug,
			Title:        tpl.Title,
			Status:       string(QuestStatusPending),
			CreatedAt:    now,
		}
		store.missions[qid] = q
		store.questsByCrew["crew-mvp"] = append(store.questsByCrew["crew-mvp"], *q)

		chs := make([]game.Exercise, 0, len(tpl.ChallengeDefs))
		for _, def := range tpl.ChallengeDefs {
			cid := nextChallengeID
			nextChallengeID++
			chs = append(chs, game.Exercise{
				ID:          cid,
				MissionID:     qid,
				Slug:        def.Slug,
				Description: def.Description,
				Status:      string(ChallengeStatusPending),
				CreatedAt:   now,
			})
		}
		store.exercises[qid] = chs

		if err := svc.StartQuest(ctx, qid, "user-mvp"); err != nil {
			t.Fatalf("%s: start: %v", slug, err)
		}
		got, err := svc.GetByCrewAndID(ctx, qid, "crew-mvp")
		if err != nil {
			t.Fatalf("%s: get after start: %v", slug, err)
		}
		if got.Status != string(QuestStatusActive) {
			t.Fatalf("%s: expected ACTIVE after start, got %s", slug, got.Status)
		}
		if got.QuestType != string(tpl.Type) {
			t.Fatalf("%s: quest_type want %s got %s", slug, tpl.Type, got.QuestType)
		}

		for i, ch := range store.exercises[qid] {
			status, progressed, completed, err := svc.CompleteChallengeForQuest(ctx, qid, ch.ID, "user-mvp", "")
			if err != nil {
				t.Fatalf("%s challenge %d: complete: %v", slug, ch.ID, err)
			}
			if !progressed {
				t.Fatalf("%s challenge %d: expected progressed", slug, ch.ID)
			}
			last := i == len(store.exercises[qid])-1
			if last {
				if status != QuestStatusDone || !completed {
					t.Fatalf("%s: expected DONE completed=true, status=%s completed=%v", slug, status, completed)
				}
			} else {
				if status != QuestStatusActive {
					t.Fatalf("%s mid-quest: expected ACTIVE, got %s", slug, status)
				}
				if completed {
					t.Fatalf("%s mid-quest: unexpected completed=true", slug)
				}
			}
		}
	}

	list, err := svc.List(ctx, "crew-mvp")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 6 {
		t.Fatalf("list want 6, got %d", len(list))
	}
	for _, v := range list {
		if v.Status != string(QuestStatusDone) {
			t.Errorf("quest %s status = %s, want DONE", v.TemplateSlug, v.Status)
		}
		if v.QuestType == "" {
			t.Errorf("quest %s missing quest_type", v.TemplateSlug)
		}
	}
}
