package fragment

import (
	"context"
	"testing"

	"odyssey/pkg/game"
)

type mockRealmReader struct {
	rp *game.JourneyProgress
}

func (m *mockRealmReader) GetRealmProgress(ctx context.Context, crewID, journey string) (*game.JourneyProgress, error) {
	return m.rp, nil
}

func TestFragmentService_DiscoverValid(t *testing.T) {
	svc := NewFragmentService(nil, nil, nil)
	ctx := context.Background()

	res, err := svc.DiscoverFragment(ctx, "user-1", "crew-1", "ancient-bark-whisper")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Discovered || res.XPGranted != 20 || res.Fragment.Slug != "ancient-bark-whisper" {
		t.Errorf("unexpected discover result: %+v", res)
	}
}

func TestFragmentService_DiscoverIdempotent(t *testing.T) {
	svc := NewFragmentService(nil, nil, nil)
	ctx := context.Background()

	res1, err := svc.DiscoverFragment(ctx, "user-1", "crew-1", "ancient-bark-whisper")
	if err != nil {
		t.Fatalf("first discover error: %v", err)
	}
	if !res1.Discovered || res1.XPGranted != 20 {
		t.Errorf("expected new discovery with 20 XP, got %+v", res1)
	}

	res2, err := svc.DiscoverFragment(ctx, "user-1", "crew-1", "ancient-bark-whisper")
	if err != nil {
		t.Fatalf("second discover error: %v", err)
	}
	if res2.Discovered {
		t.Errorf("expected Discovered=false on repeat, got %+v", res2)
	}
	if res2.XPGranted != 0 {
		t.Errorf("expected 0 XP on repeat, got %d", res2.XPGranted)
	}
}

func TestFragmentService_RejectArbitrarySlug(t *testing.T) {
	svc := NewFragmentService(nil, nil, nil)
	ctx := context.Background()

	_, err := svc.DiscoverFragment(ctx, "user-1", "crew-1", "hacked-arbitrary-fragment")
	if err == nil || err != ErrFragmentNotFound {
		t.Fatalf("expected ErrFragmentNotFound for unseeded slug, got %v", err)
	}
}

func TestFragmentService_RejectUnauthorized(t *testing.T) {
	svc := NewFragmentService(nil, nil, nil)
	ctx := context.Background()

	_, err := svc.DiscoverFragment(ctx, "", "crew-1", "ancient-bark-whisper")
	if err == nil || err != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestFragmentService_ListPlayerFragments(t *testing.T) {
	svc := NewFragmentService(nil, nil, nil)
	ctx := context.Background()

	_, _ = svc.DiscoverFragment(ctx, "user-1", "crew-1", "ancient-bark-whisper")

	frags, err := svc.ListPlayerFragments(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(frags) < 4 {
		t.Fatalf("expected at least 4 catalog fragments, got %d", len(frags))
	}

	foundDiscovered := false
	foundLocked := false
	for _, f := range frags {
		if f.Slug == "ancient-bark-whisper" && f.Discovered {
			foundDiscovered = true
		}
		if f.Slug == "copper-cog-diagram" && !f.Discovered {
			foundLocked = true
		}
	}

	if !foundDiscovered || !foundLocked {
		t.Errorf("expected discovered and locked status mapping, foundDiscovered=%v foundLocked=%v", foundDiscovered, foundLocked)
	}
}

func TestFragmentService_ReplayIncompleteRealm(t *testing.T) {
	rr := &mockRealmReader{
		rp: &game.JourneyProgress{FamilyID: "crew-1", Journey: "whispering-woods", Status: "ACTIVE", Progress: 50},
	}
	svc := NewFragmentService(nil, rr, nil)
	ctx := context.Background()

	res, err := svc.ReplayRealm(ctx, "user-1", "crew-1", "whispering-woods")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.IsReplay {
		t.Errorf("expected IsReplay=false for incomplete journey, got %+v", res)
	}
	if len(res.UnlockedFragments) != 0 {
		t.Errorf("expected 0 unlocked fragments, got %d", len(res.UnlockedFragments))
	}
}

func TestFragmentService_ReplayCompletedRealm_UnlocksHiddenFragment(t *testing.T) {
	rr := &mockRealmReader{
		rp: &game.JourneyProgress{FamilyID: "crew-1", Journey: "whispering-woods", Status: "COMPLETE", Progress: 100},
	}
	svc := NewFragmentService(nil, rr, nil)
	ctx := context.Background()

	res, err := svc.ReplayRealm(ctx, "user-1", "crew-1", "whispering-woods")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.IsReplay {
		t.Fatalf("expected IsReplay=true for completed journey, got %+v", res)
	}

	if len(res.UnlockedFragments) == 0 {
		t.Fatalf("expected hidden fragment unlocked during replay")
	}

	if res.UnlockedFragments[0].Slug != "echo-of-the-first-explorer" {
		t.Errorf("expected echo-of-the-first-explorer, got %s", res.UnlockedFragments[0].Slug)
	}
}
