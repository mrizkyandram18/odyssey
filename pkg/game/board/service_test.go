package board

import (
	"context"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type mockItems struct {
	byID  map[int64]*game.CreativeItem
	seq   int64
	calls int
}

func (m *mockItems) CreateCreativeItem(ctx context.Context, item *game.CreativeItem) (*game.CreativeItem, error) {
	m.calls++
	m.seq++
	c := *item
	c.ID = m.seq
	c.CreatedAt = time.Now().UTC()
	if m.byID == nil {
		m.byID = map[int64]*game.CreativeItem{}
	}
	m.byID[c.ID] = &c
	return &c, nil
}
func (m *mockItems) GetCreativeItem(ctx context.Context, id int64) (*game.CreativeItem, error) {
	it, ok := m.byID[id]
	if !ok {
		return nil, game.ErrNotFound
	}
	c := *it
	return &c, nil
}
func (m *mockItems) ListCreativeItemsByCrew(ctx context.Context, crewID, kind string) ([]game.CreativeItem, error) {
	var out []game.CreativeItem
	for _, it := range m.byID {
		if it.FamilyID == crewID && (kind == "" || it.Kind == kind) {
			out = append(out, *it)
		}
	}
	return out, nil
}

func TestPostText_Success(t *testing.T) {
	svc := NewService(&mockItems{})
	item, err := svc.PostText(context.Background(), "crew-A", "u1", "  Hello family  ")
	if err != nil {
		t.Fatalf("PostText: %v", err)
	}
	if item.Kind != game.KindSharedText || item.Journey != game.RealmSharedBoard {
		t.Fatalf("unexpected kind/journey: %+v", item)
	}
	if item.Payload != "Hello family" {
		t.Fatalf("payload trimmed expected, got %q", item.Payload)
	}
	if item.FamilyID != "crew-A" || item.AuthorUID != "u1" {
		t.Fatalf("crew/author wrong: %+v", item)
	}
}

func TestPostText_EmptyAndTooLong(t *testing.T) {
	svc := NewService(&mockItems{})
	if _, err := svc.PostText(context.Background(), "c", "u", "   "); err != ErrEmptyContent {
		t.Fatalf("empty: %v", err)
	}
	long := strings.Repeat("あ", MaxPayloadRunes+1)
	if _, err := svc.PostText(context.Background(), "c", "u", long); err != ErrContentTooLong {
		t.Fatalf("too long: %v", err)
	}
}

func TestListForCrew_Isolation(t *testing.T) {
	store := &mockItems{}
	svc := NewService(store)
	_, _ = svc.PostText(context.Background(), "crew-A", "u1", "from A")
	_, _ = svc.PostText(context.Background(), "crew-B", "u2", "from B")

	a, err := svc.ListForCrew(context.Background(), "crew-A")
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 1 || a[0].Payload != "from A" {
		t.Fatalf("crew-A list: %+v", a)
	}
	b, _ := svc.ListForCrew(context.Background(), "crew-B")
	if len(b) != 1 || b[0].Payload != "from B" {
		t.Fatalf("crew-B list: %+v", b)
	}
}

func TestGetForCrew_CrossCrewForbidden(t *testing.T) {
	store := &mockItems{}
	svc := NewService(store)
	item, _ := svc.PostText(context.Background(), "crew-A", "u1", "secret")
	_, err := svc.GetForCrew(context.Background(), "crew-B", item.ID)
	if err != ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
	got, err := svc.GetForCrew(context.Background(), "crew-A", item.ID)
	if err != nil || got.Payload != "secret" {
		t.Fatalf("same crew: %v %+v", err, got)
	}
}
