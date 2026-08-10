package quest

import "testing"

func TestMVPCatalog_ExactlySixWhisperingWoods(t *testing.T) {
	slugs := MVPQuestSlugs()
	if len(slugs) != 6 {
		t.Fatalf("expected 6 MVP quests, got %d", len(slugs))
	}
	if len(questCatalog) != 6 {
		t.Fatalf("expected catalog size 6, got %d", len(questCatalog))
	}

	types := map[QuestType]int{}
	for _, slug := range slugs {
		tpl, ok := LookupTemplate(slug)
		if !ok {
			t.Fatalf("missing template for %s", slug)
		}
		if tpl.Realm != MVPRealm {
			t.Errorf("%s: realm want %s got %s", slug, MVPRealm, tpl.Realm)
		}
		if len(tpl.ChallengeDefs) < 2 {
			t.Errorf("%s: want at least 2 challenges, got %d", slug, len(tpl.ChallengeDefs))
		}
		types[tpl.Type]++
	}

	if types[QuestTypeSolo] < 1 {
		t.Error("MVP catalog missing SOLO quest")
	}
	if types[QuestTypeRelay] < 1 {
		t.Error("MVP catalog missing RELAY quest")
	}
	if types[QuestTypeCreative] < 1 {
		t.Error("MVP catalog missing CREATIVE quest")
	}

	if TypeForSlug("shadow-trail") != QuestTypeRelay {
		t.Errorf("shadow-trail type = %s, want RELAY", TypeForSlug("shadow-trail"))
	}
	if TypeForSlug("the-old-growth") != QuestTypeCreative {
		t.Errorf("the-old-growth type = %s, want CREATIVE", TypeForSlug("the-old-growth"))
	}
	if TypeForSlug("morning-light") != QuestTypeSolo {
		t.Errorf("morning-light type = %s, want SOLO", TypeForSlug("morning-light"))
	}
}
