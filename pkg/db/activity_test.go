package db

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestActivityStore_ListActivityDatesByUsers(t *testing.T) {
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]game.DailyActivity{
		{UserID: "u1", ActivityDate: "2026-08-11", ActivityType: "daily_turn", CreatedAt: now},
		{UserID: "u2", ActivityDate: "2026-08-10", ActivityType: "daily_turn", CreatedAt: now},
	})
	mock := &mockSupabaseClient{data: data}
	store := NewActivityStore(mock)

	acts, err := store.ListActivityDatesByUsers(context.Background(), []string{"u1", "u2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acts) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(acts))
	}
	if acts[0].ActivityDate != "2026-08-11" || acts[1].UserID != "u2" {
		t.Errorf("unexpected activity rows: %+v", acts)
	}

	if len(mock.getCalls) != 1 {
		t.Fatalf("expected exactly 1 GET call, got %d", len(mock.getCalls))
	}
	call := mock.getCalls[0]
	if !strings.HasPrefix(call, "odyssey_daily_activity?") {
		t.Errorf("expected odyssey_daily_activity query, got %q", call)
	}
	// The OR filter must restrict the query to the given crew member uids.
	if !strings.Contains(call, "user_id.eq.u1") || !strings.Contains(call, "user_id.eq.u2") {
		t.Errorf("expected OR filter to contain both uids, got %q", call)
	}
	if !strings.Contains(call, "or=") {
		t.Errorf("expected OR filter, got %q", call)
	}
	decoded, err := url.QueryUnescape(call)
	if err != nil {
		t.Fatalf("unescape query: %v", err)
	}
	if !strings.Contains(decoded, "select=user_id,activity_date") {
		t.Errorf("expected user_id,activity_date select, got %q", decoded)
	}
}

func TestActivityStore_ListActivityDatesByUsers_Empty(t *testing.T) {
	mock := &mockSupabaseClient{}
	store := NewActivityStore(mock)

	acts, err := store.ListActivityDatesByUsers(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(acts) != 0 {
		t.Errorf("expected no activities, got %d", len(acts))
	}
	if len(mock.getCalls) != 0 {
		t.Errorf("expected no GET call for empty uids, got %d", len(mock.getCalls))
	}
}

func TestActivityStore_ListActivityDatesByUsers_Error(t *testing.T) {
	mock := &mockSupabaseClient{err: errors.New("network")}
	store := NewActivityStore(mock)

	_, err := store.ListActivityDatesByUsers(context.Background(), []string{"u1"})
	if err == nil {
		t.Fatal("expected error")
	}
}
