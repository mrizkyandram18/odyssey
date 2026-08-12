package chest

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type mockChestDefinitionStore struct {
	defs         []game.ChestDefinition
	entries      map[string][]game.DropTableEntry
	err          error
	dropTableErr error
}

func (m *mockChestDefinitionStore) ListChestDefinitions(ctx context.Context) ([]game.ChestDefinition, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.defs, nil
}

func (m *mockChestDefinitionStore) GetChestDefinition(ctx context.Context, slug string) (*game.ChestDefinition, error) {
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

func (m *mockChestDefinitionStore) ListDropTableEntries(ctx context.Context, chestSlug string) ([]game.DropTableEntry, error) {
	if m.dropTableErr != nil {
		return nil, m.dropTableErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.entries[chestSlug], nil
}

func TestContentChestCatalog_Get_FromStore(t *testing.T) {
	store := &mockChestDefinitionStore{
		defs: []game.ChestDefinition{
			{ID: 1, Slug: "wooden-chest", Name: "Wooden Gift", Rarity: "COMMON", Icon: "📦", Description: "A chest", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		},
		entries: map[string][]game.DropTableEntry{
			"wooden-chest": {
				{ID: 1, GiftSlug: "wooden-chest", Rarity: "COMMON", Weight: 0.7, CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
				{ID: 2, GiftSlug: "wooden-chest", Rarity: "UNCOMMON", Weight: 0.3, CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
			},
		},
	}
	catalog := NewContentChestCatalog(store)

	ct := catalog.Get(context.Background(), "wooden-chest")
	if ct == nil {
		t.Fatal("expected wooden-chest to exist")
	}
	if ct.Name != "Wooden Gift" {
		t.Errorf("expected Wooden Gift, got %s", ct.Name)
	}
	if ct.Slug != "wooden-chest" {
		t.Errorf("expected slug wooden-chest, got %s", ct.Slug)
	}
}

func TestContentChestCatalog_Get_Fallback(t *testing.T) {
	store := &mockChestDefinitionStore{
		defs:    []game.ChestDefinition{},
		entries: map[string][]game.DropTableEntry{},
	}
	catalog := NewContentChestCatalog(store)

	ct := catalog.Get(context.Background(), "wooden-chest")
	if ct == nil {
		t.Fatal("expected wooden-chest to exist via fallback")
	}
	if ct.Name != "Wooden Gift" {
		t.Errorf("expected Wooden Gift, got %s", ct.Name)
	}
}

func TestContentChestCatalog_Get_NotFound(t *testing.T) {
	store := &mockChestDefinitionStore{
		defs:    []game.ChestDefinition{},
		entries: map[string][]game.DropTableEntry{},
	}
	catalog := NewContentChestCatalog(store)

	ct := catalog.Get(context.Background(), "nonexistent")
	if ct != nil {
		t.Error("expected nil for nonexistent chest type")
	}
}

func TestContentChestCatalog_Get_StoreError(t *testing.T) {
	store := &mockChestDefinitionStore{
		err: errors.New("db error"),
	}
	catalog := NewContentChestCatalog(store)

	ct := catalog.Get(context.Background(), "wooden-chest")
	if ct == nil {
		t.Fatal("expected fallback when store errors")
	}
	if ct.Slug != "wooden-chest" {
		t.Errorf("expected wooden-chest from fallback, got %s", ct.Slug)
	}
}

func TestContentChestCatalog_ListAll_FromStore(t *testing.T) {
	store := &mockChestDefinitionStore{
		defs: []game.ChestDefinition{
			{ID: 1, Slug: "wooden-chest", Name: "Wooden Gift", Rarity: "COMMON", Icon: "📦", Description: "A chest", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
			{ID: 2, Slug: "bronze-chest", Name: "Bronze Gift", Rarity: "UNCOMMON", Icon: "🟤", Description: "A sturdy chest", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		},
		entries: map[string][]game.DropTableEntry{
			"wooden-chest": {
				{ID: 1, GiftSlug: "wooden-chest", Rarity: "COMMON", Weight: 0.7, CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
			},
			"bronze-chest": {
				{ID: 2, GiftSlug: "bronze-chest", Rarity: "UNCOMMON", Weight: 0.5, CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
			},
		},
	}
	catalog := NewContentChestCatalog(store)

	gifts := catalog.ListAll(context.Background())
	if len(gifts) != 2 {
		t.Errorf("expected 2 gifts, got %d", len(gifts))
	}
}

func TestContentChestCatalog_ListAll_Fallback(t *testing.T) {
	store := &mockChestDefinitionStore{
		defs:    []game.ChestDefinition{},
		entries: map[string][]game.DropTableEntry{},
	}
	catalog := NewContentChestCatalog(store)

	gifts := catalog.ListAll(context.Background())
	if len(gifts) == 0 {
		t.Error("expected fallback gifts")
	}
}

func TestContentChestCatalog_ListAll_StoreError(t *testing.T) {
	store := &mockChestDefinitionStore{
		err: errors.New("db error"),
	}
	catalog := NewContentChestCatalog(store)

	gifts := catalog.ListAll(context.Background())
	if len(gifts) == 0 {
		t.Error("expected fallback gifts when store errors")
	}
}

func TestContentChestCatalog_Caching(t *testing.T) {
	store := &mockChestDefinitionStore{
		defs: []game.ChestDefinition{
			{ID: 1, Slug: "wooden-chest", Name: "Wooden Gift", Rarity: "COMMON", Icon: "📦", Description: "A chest", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		},
		entries: map[string][]game.DropTableEntry{
			"wooden-chest": {
				{ID: 1, GiftSlug: "wooden-chest", Rarity: "COMMON", Weight: 0.7, CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
			},
		},
	}
	catalog := NewContentChestCatalog(store)

	ct1 := catalog.Get(context.Background(), "wooden-chest")
	ct2 := catalog.Get(context.Background(), "wooden-chest")
	if ct1 != ct2 {
		t.Error("expected cached result to be the same pointer")
	}
}

func TestContentChestCatalog_NilStore(t *testing.T) {
	catalog := NewContentChestCatalog(nil)

	ct := catalog.Get(context.Background(), "wooden-chest")
	if ct == nil {
		t.Fatal("expected wooden-chest from fallback when store is nil")
	}

	gifts := catalog.ListAll(context.Background())
	if len(gifts) == 0 {
		t.Error("expected fallback gifts when store is nil")
	}
}

func TestContentChestCatalog_ListAll_UsesFallbackWhenDropTableLoadFails(t *testing.T) {
	store := &mockChestDefinitionStore{
		defs: []game.ChestDefinition{
			{ID: 1, Slug: "wooden-chest", Name: "Wooden Gift", Rarity: "COMMON", Icon: "📦", Description: "A chest", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		},
		entries: map[string][]game.DropTableEntry{},
	}
	// Make ListDropTableEntries return an error specifically
	store.dropTableErr = errors.New("db error on drop table")

	// Intercept log output to check for warning
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	catalog := NewContentChestCatalog(store)
	gifts := catalog.ListAll(context.Background())

	if len(gifts) != 1 {
		t.Fatalf("expected 1 chest, got %d", len(gifts))
	}
	if gifts[0].Slug != "wooden-chest" {
		t.Errorf("expected wooden-chest fallback, got %s", gifts[0].Slug)
	}

	logStr := buf.String()
	if !strings.Contains(logStr, "WARN: loading drop table wooden-chest") {
		t.Errorf("expected warning log about drop table load failure, got %s", logStr)
	}
}

func TestContentChestCatalog_ListAll_LogsWarningWhenNoFallbackExists(t *testing.T) {
	store := &mockChestDefinitionStore{
		defs: []game.ChestDefinition{
			{ID: 99, Slug: "unknown-chest", Name: "Unknown", Rarity: "COMMON", Icon: "?", Description: "?", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		},
		entries: map[string][]game.DropTableEntry{},
	}
	store.dropTableErr = errors.New("db error")

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	catalog := NewContentChestCatalog(store)
	gifts := catalog.ListAll(context.Background())

	if len(gifts) != 0 {
		t.Fatalf("expected 0 gifts (skipped because no fallback), got %d", len(gifts))
	}

	logStr := buf.String()
	if !strings.Contains(logStr, "WARN: loading drop table unknown-chest") {
		t.Errorf("expected warning log about drop table load failure, got %s", logStr)
	}
	if !strings.Contains(logStr, "WARN: no fallback exists for unknown-chest") {
		t.Errorf("expected warning log about no fallback, got %s", logStr)
	}
}

func TestContentChestCatalog_Get_UsesFallbackWhenDropTableLoadFails(t *testing.T) {
	store := &mockChestDefinitionStore{
		defs: []game.ChestDefinition{
			{ID: 1, Slug: "wooden-chest", Name: "Wooden Gift", Rarity: "COMMON", Icon: "📦", Description: "A chest", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		},
		entries: map[string][]game.DropTableEntry{},
	}
	store.dropTableErr = errors.New("db error")

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	catalog := NewContentChestCatalog(store)
	ct := catalog.Get(context.Background(), "wooden-chest")

	if ct == nil {
		t.Fatal("expected fallback chest, got nil")
	}
	if ct.Slug != "wooden-chest" {
		t.Errorf("expected wooden-chest, got %s", ct.Slug)
	}

	logStr := buf.String()
	if !strings.Contains(logStr, "WARN: loading drop table wooden-chest") {
		t.Errorf("expected warning log about drop table load failure, got %s", logStr)
	}
}
