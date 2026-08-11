package fragment

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/progression"
)

var ErrFragmentNotFound = errors.New("story fragment not found")
var ErrUnauthorized = errors.New("unauthorized fragment access")

// FragmentStore abstracts persistence for story fragments.
type FragmentStore interface {
	ListCatalog(ctx context.Context) ([]game.StoryFragment, error)
	ListByPlayer(ctx context.Context, uid string) ([]game.PlayerStoryFragment, error)
	GetFragment(ctx context.Context, slug string) (*game.StoryFragment, error)
	DiscoverFragment(ctx context.Context, uid, crewID, slug string) (*game.PlayerStoryFragment, bool, error)
}

// RealmProgressReader provides realm status checks for replay detection.
type RealmProgressReader interface {
	GetRealmProgress(ctx context.Context, crewID, realm string) (*game.RealmProgress, error)
}

// InMemoryFragmentStore provides a thread-safe in-memory store for testing and fallbacks.
type InMemoryFragmentStore struct {
	mu          sync.RWMutex
	catalog     map[string]game.StoryFragment
	playerStash map[string]map[string]game.PlayerStoryFragment // uid -> slug -> row
	nextID      int64
}

func NewInMemoryFragmentStore() *InMemoryFragmentStore {
	cat := make(map[string]game.StoryFragment)
	for _, f := range DefaultFragments {
		cat[f.Slug] = f
	}
	return &InMemoryFragmentStore{
		catalog:     cat,
		playerStash: make(map[string]map[string]game.PlayerStoryFragment),
		nextID:      1,
	}
}

func (s *InMemoryFragmentStore) ListCatalog(ctx context.Context) ([]game.StoryFragment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]game.StoryFragment, 0, len(s.catalog))
	for _, f := range s.catalog {
		res = append(res, f)
	}
	return res, nil
}

func (s *InMemoryFragmentStore) ListByPlayer(ctx context.Context, uid string) ([]game.PlayerStoryFragment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := s.playerStash[uid]
	res := make([]game.PlayerStoryFragment, 0, len(rows))
	for _, r := range rows {
		res = append(res, r)
	}
	return res, nil
}

func (s *InMemoryFragmentStore) GetFragment(ctx context.Context, slug string) (*game.StoryFragment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.catalog[slug]
	if !ok {
		return nil, ErrFragmentNotFound
	}
	return &f, nil
}

func (s *InMemoryFragmentStore) DiscoverFragment(ctx context.Context, uid, crewID, slug string) (*game.PlayerStoryFragment, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.catalog[slug]; !ok {
		return nil, false, ErrFragmentNotFound
	}

	if s.playerStash[uid] == nil {
		s.playerStash[uid] = make(map[string]game.PlayerStoryFragment)
	}

	if existing, ok := s.playerStash[uid][slug]; ok {
		return &existing, false, nil
	}

	now := time.Now().UTC()
	row := game.PlayerStoryFragment{
		ID:           s.nextID,
		UID:          uid,
		CrewID:       crewID,
		FragmentSlug: slug,
		DiscoveredAt: now,
	}
	s.nextID++
	s.playerStash[uid][slug] = row
	return &row, true, nil
}

// FragmentService manages story fragment discovery, replay detection, and journal views.
type FragmentService struct {
	store       FragmentStore
	realmReader RealmProgressReader
	progression *progression.ProgressionService
}

func NewFragmentService(store FragmentStore, realmReader RealmProgressReader, prog *progression.ProgressionService) *FragmentService {
	if store == nil {
		store = NewInMemoryFragmentStore()
	}
	return &FragmentService{
		store:       store,
		realmReader: realmReader,
		progression: prog,
	}
}

// ListPlayerFragments returns all catalog fragments with their player discovery status.
func (s *FragmentService) ListPlayerFragments(ctx context.Context, uid string) ([]StoryFragmentView, error) {
	catalog, err := s.store.ListCatalog(ctx)
	if err != nil {
		return nil, fmt.Errorf("list catalog: %w", err)
	}

	discoveredRows, err := s.store.ListByPlayer(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("list player fragments: %w", err)
	}

	discoveredMap := make(map[string]time.Time, len(discoveredRows))
	for _, r := range discoveredRows {
		discoveredMap[r.FragmentSlug] = r.DiscoveredAt
	}

	views := make([]StoryFragmentView, 0, len(catalog))
	for _, f := range catalog {
		discTime, found := discoveredMap[f.Slug]
		v := StoryFragmentView{
			Slug:       f.Slug,
			Realm:      f.Realm,
			Title:      f.Title,
			Content:    f.Content,
			SetName:    f.SetName,
			IsHidden:   f.IsHidden,
			Discovered: found,
		}
		if found {
			v.DiscoveredAt = &discTime
		}
		views = append(views, v)
	}
	return views, nil
}

// DiscoverFragment collects a fragment by slug, enforcing catalog membership and granting XP idempotently.
func (s *FragmentService) DiscoverFragment(ctx context.Context, uid, crewID, slug string) (*DiscoverResult, error) {
	if uid == "" || crewID == "" {
		return nil, ErrUnauthorized
	}

	fragDef, err := s.store.GetFragment(ctx, slug)
	if err != nil {
		return nil, ErrFragmentNotFound
	}

	pf, newlyDiscovered, err := s.store.DiscoverFragment(ctx, uid, crewID, slug)
	if err != nil {
		return nil, err
	}

	var xpGranted int64 = 0
	if newlyDiscovered {
		xpGranted = 20
		if s.progression != nil {
			_, _, _ = s.progression.AwardXP(ctx, uid, xpGranted)
		}
	}

	discTime := pf.DiscoveredAt
	return &DiscoverResult{
		Fragment: StoryFragmentView{
			Slug:         fragDef.Slug,
			Realm:        fragDef.Realm,
			Title:        fragDef.Title,
			Content:      fragDef.Content,
			SetName:      fragDef.SetName,
			IsHidden:     fragDef.IsHidden,
			Discovered:   true,
			DiscoveredAt: &discTime,
		},
		Discovered: newlyDiscovered,
		XPGranted:  xpGranted,
	}, nil
}

// ReplayRealm detects if a realm is completed, unlocking bonus dialogue and hidden story fragments.
func (s *FragmentService) ReplayRealm(ctx context.Context, uid, crewID, realm string) (*ReplayResult, error) {
	if uid == "" || crewID == "" {
		return nil, ErrUnauthorized
	}

	var rp *game.RealmProgress
	if s.realmReader != nil {
		var err error
		rp, err = s.realmReader.GetRealmProgress(ctx, crewID, realm)
		if err != nil {
			return nil, fmt.Errorf("get realm progress: %w", err)
		}
	}

	isComplete := rp != nil && (rp.Status == "COMPLETE" || rp.Progress >= 100)
	if !isComplete {
		return &ReplayResult{
			Realm:             realm,
			IsReplay:          false,
			BonusDialogue:     "Ranah ini masih aktif! Selesaikan petualangan utama kru untuk membuka rahasia tersembunyi.",
			UnlockedFragments: []StoryFragmentView{},
		}, nil
	}

	// Realm is completed — trigger realm replay logic
	catalog, err := s.store.ListCatalog(ctx)
	if err != nil {
		return nil, fmt.Errorf("list catalog: %w", err)
	}

	var unlocked []StoryFragmentView
	for _, f := range catalog {
		if f.Realm == realm && f.IsHidden {
			res, err := s.DiscoverFragment(ctx, uid, crewID, f.Slug)
			if err == nil && res != nil {
				unlocked = append(unlocked, res.Fragment)
			}
		}
	}

	dialogue := fmt.Sprintf("Selamat datang kembali di %s! Sebagai penjelajah veteran, kamu menemukan rahasia tersembunyi yang tak terlihat saat kunjungan pertama.", realm)

	return &ReplayResult{
		Realm:             realm,
		IsReplay:          true,
		BonusDialogue:     dialogue,
		UnlockedFragments: unlocked,
	}, nil
}
