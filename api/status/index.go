package status

import (
	"context"
	"net/http"
	"time"

	"odyssey/pkg/shared"
)

// StatusProvider supplies the payload for the public /api/status endpoint.
// Implementations should be safe for concurrent use.
type StatusProvider interface {
	StatusInfo(ctx context.Context) map[string]any
}

// FuncStatusProvider adapts a function into a StatusProvider.
func FuncStatusProvider(fn func(ctx context.Context) map[string]any) StatusProvider {
	return &funcProvider{fn: fn}
}

type funcProvider struct {
	fn func(ctx context.Context) map[string]any
}

func (p *funcProvider) StatusInfo(ctx context.Context) map[string]any {
	if p == nil || p.fn == nil {
		return map[string]any{
			"app":       "odyssey",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
	}
	return p.fn(ctx)
}

var provider StatusProvider

// Setup injects the status provider. Must be called once at startup before
// the server serves requests.
func Setup(p StatusProvider) {
	provider = p
}

// Handler serves the public service status: app name, version, uptime, and
// (when available) the live content-generation pointer and cache health.
// It is lightweight and does not require admin authentication.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := map[string]any{
		"app":       "odyssey",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if provider != nil {
		for k, v := range provider.StatusInfo(r.Context()) {
			if _, exists := info[k]; !exists {
				info[k] = v
			}
		}
	}

	shared.WriteJSON(w, http.StatusOK, info)
}
