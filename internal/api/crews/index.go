package crews

import (
	"context"
	"net/http"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	"odyssey/pkg/shared"
)

// CrewStore provides access to crew data.
type CrewStore interface {
	GetCrew(ctx context.Context, crewID string) (*game.Crew, error)
}

// crewResponse is the JSON representation of a crew.
type crewResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapCrew(c *game.Crew) crewResponse {
	return crewResponse{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

var store CrewStore

// Setup injects the crew store. Must be called once at startup before
// the server serves requests.
func Setup(s CrewStore) {
	store = s
}

// Handler serves the /api/crews endpoint.
//
// GET /api/crews — returns the authenticated user's crew information.
// This is the family group the user belongs to.
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

	crew, err := store.GetCrew(r.Context(), claims.CrewID)
	if err != nil {
		if err == game.ErrNotFound {
			shared.WriteJSONError(w, "crew not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to get crew", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, mapCrew(crew))
}
