package relic

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type mockRelicDefinitionStore struct {
	defs []game.RelicDefinition
	err  error
}

func (m *mockRelicDefinitionStore) ListRelicDefinitions(ctx context.Context) ([]game.RelicDefinition, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.defs, nil
}

func (m *mockRelicDefinitionStore) GetRelicDefinition(ctx context.Context, slug string) (*game.RelicDefinition, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.defs {
		if m.defs[i].Slug == slug {
			return &m.defs[i], nil
		}
	}
	return nil, game.ErrNotFound
}

func TestContentRelicCatalog_Get_FromStore(t *testing.T) {
	store := &mockRelicDefinitionStore{
		defs: []game.RelicDefinition{
			{ID: 1, Slug: "ancient-compass", Name: "Ancient Compass", Rarity: "COMMON", Image: "🧭", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
	catalog := NewContentRelicCatalog(store)

	rd := catalog.Get(context.Background(), "ancient-compass")
	if rd == nil {
		t.Fatal("expected ancient-compass to exist")
	}
	if rd.Name != "Ancient Compass" {
		t.Errorf("expected Ancient Compass, got %s", rd.Name)
	}
}

func TestContentRelicCatalog_Get_Fallback(t *testing.T) {
	store := &mockRelicDefinitionStore{
		defs: []game.RelicDefinition{},
	}
	catalog := NewContentRelicCatalog(store)

	rd := catalog.Get(context.Background(), "ancient-compass")
	if rd == nil {
		t.Fatal("expected ancient-compass to exist via fallback")
	}
	if rd.Name != "Ancient Compass" {
		t.Errorf("expected Ancient Compass, got %s", rd.Name)
	}
}

func TestContentRelicCatalog_Get_NotFound(t *testing.T) {
	store := &mockRelicDefinitionStore{
		defs: []game.RelicDefinition{},
	}
	catalog := NewContentRelicCatalog(store)

	rd := catalog.Get(context.Background(), "nonexistent")
	if rd != nil {
		t.Error("expected nil for nonexistent relic")
	}
}

func TestContentRelicCatalog_Get_StoreError(t *testing.T) {
	store := &mockRelicDefinitionStore{
		err: errors.New("db error"),
	}
	catalog := NewContentRelicCatalog(store)

	rd := catalog.Get(context.Background(), "ancient-compass")
	if rd == nil {
		t.Fatal("expected fallback when store errors")
	}
	if rd.Slug != "ancient-compass" {
		t.Errorf("expected ancient-compass from fallback, got %s", rd.Slug)
	}
}

func TestContentRelicCatalog_ListAll_FromStore(t *testing.T) {
	store := &mockRelicDefinitionStore{
		defs: []game.RelicDefinition{
			{ID: 1, Slug: "ancient-compass", Name: "Ancient Compass", Rarity: "COMMON", Image: "🧭", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
			{ID: 2, Slug: "crystal-shard", Name: "Crystal Shard", Rarity: "UNCOMMON", Image: "💎", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
	catalog := NewContentRelicCatalog(store)

	relics := catalog.ListAll(context.Background())
	if len(relics) != 2 {
		t.Errorf("expected 2 relics, got %d", len(relics))
	}
}

func TestContentRelicCatalog_ListAll_Fallback(t *testing.T) {
	store := &mockRelicDefinitionStore{
		defs: []game.RelicDefinition{},
	}
	catalog := NewContentRelicCatalog(store)

	relics := catalog.ListAll(context.Background())
	if len(relics) == 0 {
		t.Error("expected fallback relics when store is empty")
	}
}

func TestContentRelicCatalog_ListAll_StoreError(t *testing.T) {
	store := &mockRelicDefinitionStore{
		err: errors.New("db error"),
	}
	catalog := NewContentRelicCatalog(store)

	relics := catalog.ListAll(context.Background())
	if len(relics) == 0 {
		t.Error("expected fallback relics when store errors")
	}
}

func TestContentRelicCatalog_NilStore(t *testing.T) {
	catalog := NewContentRelicCatalog(nil)

	rd := catalog.Get(context.Background(), "ancient-compass")
	if rd == nil {
		t.Fatal("expected ancient-compass from fallback when store is nil")
	}

	relics := catalog.ListAll(context.Background())
	if len(relics) == 0 {
		t.Error("expected fallback relics when store is nil")
	}
}
