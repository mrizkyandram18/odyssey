package families

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/shared"
)

type FamilyStore interface {
	GetFamily(ctx context.Context, familyID string) (*db.Family, error)
	UpdateFamily(ctx context.Context, familyID string, patch map[string]any) error
}

type UserStore interface {
	ListUsersByFamily(ctx context.Context, familyID string) ([]db.UserProfile, error)
}

type familyResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"`
	BannerURL string    `json:"banner_url,omitempty"`
	Theme     string    `json:"theme,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapFamily(c *db.Family) familyResponse {
	return familyResponse{
		ID:        c.ID,
		Name:      c.Name,
		BannerURL: c.BannerURL,
		Theme:     c.Theme,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

type updateFamilyRequest struct {
	BannerURL string `json:"banner_url,omitempty"`
	Theme     string `json:"theme,omitempty"`
}

var store FamilyStore
var users UserStore

func Setup(s FamilyStore, u UserStore) {
	store = s
	users = u
}

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
			handleGetFamily(w, r, claims)
		}
	case http.MethodPatch:
		handlePatchFamily(w, r, claims)
	default:
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetFamily(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	fam, err := store.GetFamily(r.Context(), claims.FamilyID)
	if err != nil {
		if errors.Is(err, db.ErrFamilyNotFound) {
			shared.WriteJSONError(w, "family not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to get family", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, mapFamily(fam))
}

func handlePatchFamily(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	var req updateFamilyRequest
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

	if err := store.UpdateFamily(r.Context(), claims.FamilyID, patch); err != nil {
		shared.WriteJSONError(w, "failed to update family", http.StatusInternalServerError)
		return
	}

	fam, err := store.GetFamily(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to reload family", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, mapFamily(fam))
}

func handleGetMembers(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	if users == nil {
		shared.WriteJSON(w, http.StatusOK, []db.UserProfile{})
		return
	}
	members, err := users.ListUsersByFamily(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to get family members", http.StatusInternalServerError)
		return
	}

	if members == nil {
		members = []db.UserProfile{}
	}

	shared.WriteJSON(w, http.StatusOK, members)
}
