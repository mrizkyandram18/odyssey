package status

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"odyssey/pkg/shared"
)

type staticProvider struct {
	info map[string]any
}

func (p *staticProvider) StatusInfo(ctx context.Context) map[string]any {
	out := map[string]any{
		"app":       "odyssey",
		"timestamp": "t",
	}
	for k, v := range p.info {
		out[k] = v
	}
	return out
}

func TestStatus_Handler(t *testing.T) {
	Setup(&staticProvider{info: map[string]any{
		"version":        "v9.9.9",
		"schema_version": "9",
		"uptime_seconds": int64(42),
	}})
	defer func() { provider = nil }()

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected content-type application/json, got %s", ct)
	}
	body := w.Body.String()
	if !contains(body, "odyssey") || !contains(body, "v9.9.9") || !contains(body, "42") {
		t.Fatalf("expected status payload with app/version/uptime, got: %s", body)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestStatus_MethodNotAllowed(t *testing.T) {
	Setup(&staticProvider{})
	defer func() { provider = nil }()

	req := httptest.NewRequest(http.MethodPost, "/api/status", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestStatus_NoProvider_StillWorks(t *testing.T) {
	// No Setup call -> provider is nil.
	provider = nil
	defer func() { provider = nil }()

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	// shared.WriteJSON sets content type
	_ = shared.WriteJSON
}
