package relics

import (
	"context"
	"net/http"
	"strings"

	"odyssey/pkg/auth"
	gamerelic "odyssey/pkg/game/relic"
	"odyssey/pkg/shared"
)

// Service is the interface the handler depends on.
type Service interface {
	ListRelics(ctx context.Context) ([]gamerelic.RelicDefinition, error)
	GetRelic(ctx context.Context, slug string) (*gamerelic.RelicDefinition, error)
	ListInventory(ctx context.Context, uid string) ([]gamerelic.InventoryItem, error)
}

var svc Service

func Setup(s Service) {
	svc = s
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	default:
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if svc == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/relics")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	if path == "" || path == "inventory" {
		items, err := svc.ListInventory(r.Context(), claims.UID)
		if err != nil {
			shared.WriteJSONError(w, "failed to load inventory", http.StatusInternalServerError)
			return
		}
		shared.WriteJSON(w, http.StatusOK, items)
		return
	}

	if strings.Contains(path, "/") {
		shared.WriteJSONError(w, "not found", http.StatusNotFound)
		return
	}

	if !shared.ValidateSlug(path) || len(path) > 256 {
		shared.WriteJSONError(w, "invalid slug", http.StatusBadRequest)
		return
	}

	def, err := svc.GetRelic(r.Context(), path)
	if err != nil {
		shared.WriteJSONError(w, "relic not found", http.StatusNotFound)
		return
	}
	shared.WriteJSON(w, http.StatusOK, def)
}
