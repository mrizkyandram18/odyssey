package realm_progress

import (
	"context"
	"net/http"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	"odyssey/pkg/shared"
)

// RealmProgressStore provides access to realm progress data.
type RealmProgressStore interface {
	ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.RealmProgress, error)
}

var store RealmProgressStore

// Setup injects the realm progress store. Must be called once at startup before
// the server serves requests.
func Setup(s RealmProgressStore) {
	store = s
}

// Handler serves the /api/realm_progress endpoint.
//
// GET /api/realm_progress — returns the family's shared progress across all
// realms. Each entry represents the crew's status, story branch, and progress
// percentage within a realm.
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

	if store == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	progress, err := store.ListRealmProgressByCrew(r.Context(), claims.CrewID)
	if err != nil {
		shared.WriteJSONError(w, "failed to list realm progress", http.StatusInternalServerError)
		return
	}

	if progress == nil {
		progress = []game.RealmProgress{}
	}

	shared.WriteJSON(w, http.StatusOK, progress)
}
