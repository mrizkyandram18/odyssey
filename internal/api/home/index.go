package home

import (
	"context"
	"net/http"

	"odyssey/pkg/auth"
	gamesvc "odyssey/pkg/game/home"
	"odyssey/pkg/shared"
)

// Service is the interface the handler depends on.
// *gamesvc.HomeService satisfies it.
type Service interface {
	GetHome(ctx context.Context, uid string, crewID string) (*gamesvc.HomeResponse, error)
}

var svc Service

func Setup(s Service) {
	svc = s
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

	if svc == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	resp, err := svc.GetHome(r.Context(), claims.UID, claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to load home data", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, resp)
}
