package journey_progress

import (
	"context"
	"net/http"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	"odyssey/pkg/shared"
)

// RealmProgressStore provides access to journey progress data.
type RealmProgressStore interface {
	ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.JourneyProgress, error)
}

var store RealmProgressStore

// Setup injects the journey progress store. Must be called once at startup before
// the server serves requests.
func Setup(s RealmProgressStore) {
	store = s
}

// Handler serves the /api/journey_progress endpoint.
//
// GET /api/journey_progress — returns the family's shared progress across all
// realms. Each entry represents the crew's status, story branch, and progress
// percentage within a journey.
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

	progress, err := store.ListRealmProgressByCrew(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to list journey progress", http.StatusInternalServerError)
		return
	}

	if progress == nil {
		progress = []game.JourneyProgress{}
	}

	shared.WriteJSON(w, http.StatusOK, progress)
}
