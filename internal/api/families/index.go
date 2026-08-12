package families

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	"odyssey/pkg/shared"
)

// CrewStore provides access to crew data.
type CrewStore interface {
	GetCrew(ctx context.Context, crewID string) (*game.Family, error)
	UpdateCrew(ctx context.Context, crewID string, patch map[string]any) error
}

// UserStore provides access to user data.
type UserStore interface {
	ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error)
}

// crewResponse is the JSON representation of a crew.
type crewResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"`
	BannerURL string    `json:"banner_url,omitempty"`
	Theme     string    `json:"theme,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapCrew(c *game.Family) crewResponse {
	return crewResponse{
		ID:        c.ID,
		Name:      c.Name,
		BannerURL: c.BannerURL,
		Theme:     c.Theme,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

type updateCrewRequest struct {
	BannerURL string `json:"banner_url,omitempty"`
	Theme     string `json:"theme,omitempty"`
}

var store CrewStore
var users UserStore

// Setup injects the crew and user stores. Must be called once at startup before
// the server serves requests.
func Setup(s CrewStore, u UserStore) {
	store = s
	users = u
}

// Handler serves the /api/families endpoint.
//
// GET /api/families — returns the authenticated user's crew information.
// PATCH /api/families — updates the authenticated user's crew banner and theme.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if store == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/api/families/members" {
			handleGetMembers(w, r, claims)
		} else {
			handleGetCrew(w, r, claims)
		}
	case http.MethodPatch:
		handlePatchCrew(w, r, claims)
	default:
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetCrew(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	crew, err := store.GetCrew(r.Context(), claims.FamilyID)
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

func handlePatchCrew(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	var req updateCrewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	patch := make(map[string]any)
	if req.BannerURL != "" {
		patch["banner_url"] = req.BannerURL
	}
	if req.Theme != "" {
		patch["theme"] = req.Theme
	}

	if len(patch) == 0 {
		shared.WriteJSONError(w, "no fields to update", http.StatusBadRequest)
		return
	}

	if err := store.UpdateCrew(r.Context(), claims.FamilyID, patch); err != nil {
		shared.WriteJSONError(w, "failed to update crew", http.StatusInternalServerError)
		return
	}

	crew, err := store.GetCrew(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to reload crew", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, mapCrew(crew))
}

func handleGetMembers(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	members, err := users.ListUsersByCrew(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to get crew members", http.StatusInternalServerError)
		return
	}

	if members == nil {
		members = []game.Player{}
	}

	shared.WriteJSON(w, http.StatusOK, members)
}
