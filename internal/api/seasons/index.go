package seasons

import (
	"context"
	"net/http"

	"odyssey/pkg/auth"
	"odyssey/pkg/game/season"
	"odyssey/pkg/shared"
)

// SeasonHandler is the interface the handler depends on.
// *season.SeasonService satisfies it.
type SeasonHandler interface {
	GetCurrentSeason(ctx context.Context) (*season.SeasonSummary, error)
	ListAll(ctx context.Context) ([]season.SeasonSummary, error)
}

var handler SeasonHandler

func Setup(h SeasonHandler) {
	handler = h
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if handler == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	path := r.URL.Path
	if path == "/api/seasons" || path == "/api/seasons/" {
		handleList(w, r, claims)
		return
	}

	shared.WriteJSONError(w, "not found", http.StatusNotFound)
}

func handleList(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	_ = claims
	list, err := handler.ListAll(r.Context())
	if err != nil {
		shared.WriteJSONError(w, "failed to list seasons: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, list)
}
