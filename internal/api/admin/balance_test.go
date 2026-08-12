package admin

import (
	"context"
	"testing"
	"time"

	"odyssey/pkg/observability"
)

type mockBalanceService struct {
	reloaded  bool
	err       error
	overrides map[string]int64
	loadedAt  time.Time
}

func (m *mockBalanceService) Reload(ctx context.Context) error {
	m.reloaded = true
	return m.err
}

func (m *mockBalanceService) Overrides() map[string]int64 {
	if m.overrides == nil {
		return map[string]int64{}
	}
	out := make(map[string]int64, len(m.overrides))
	for k, v := range m.overrides {
		out[k] = v
	}
	return out
}

func (m *mockBalanceService) LoadedAt() time.Time { return m.loadedAt }

func TestAdmin_BalanceOverrides(t *testing.T) {
	mockContent := &mockContentServiceForAdmin{}
	svc := NewAdminService(mockContent, newMockAdminStore(), nil)
	svc.SetBalance(&mockBalanceService{overrides: map[string]int64{"daily_mission_xp": 20}, loadedAt: time.Now()})

	got, err := svc.BalanceOverrides(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["daily_mission_xp"] != int64(20) {
		t.Errorf("expected daily_mission_xp=20, got %v", got["daily_mission_xp"])
	}
	if got["count"] != 1 {
		t.Errorf("expected count=1, got %v", got["count"])
	}
}

func TestAdmin_BalanceReload(t *testing.T) {
	b := &mockBalanceService{overrides: map[string]int64{}}
	svc := NewAdminService(&mockContentServiceForAdmin{}, newMockAdminStore(), nil)
	svc.SetBalance(b)
	m := observability.NewMetrics()
	svc.SetMetrics(m)

	if err := svc.ReloadBalance(context.Background()); err != nil {
		t.Fatalf("ReloadBalance error: %v", err)
	}
	if !b.reloaded {
		t.Error("expected reload to be called")
	}
}

func TestAdmin_BalanceNotConfigured(t *testing.T) {
	svc := NewAdminService(&mockContentServiceForAdmin{}, newMockAdminStore(), nil)
	if _, err := svc.BalanceOverrides(context.Background()); err == nil {
		t.Error("expected error when balance not configured")
	}
	if err := svc.ReloadBalance(context.Background()); err == nil {
		t.Error("expected error when balance not configured")
	}
}

func TestAdmin_Validate_ValidContent(t *testing.T) {
	svc := NewAdminService(&mockContentServiceForAdmin{}, newMockAdminStore(), nil)
	m := observability.NewMetrics()
	svc.SetMetrics(m)

	result, err := svc.Validate(context.Background())
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid content, errors: %v", result.Errors)
	}
	if m.Snapshot().ValidationFailures != 0 {
		t.Errorf("expected 0 validation failures, got %d", m.Snapshot().ValidationFailures)
	}
}
