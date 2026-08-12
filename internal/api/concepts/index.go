package concept

import (
	"context"
	"net/http"
	"strings"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	gameLore "odyssey/pkg/game/concepts"
	"odyssey/pkg/shared"
)

type LoreHandler interface {
	GetSummary(ctx context.Context, crewID string) (*gameLore.LoreSummary, error)
	ListUnlocked(ctx context.Context, crewID string) ([]gameLore.LoreView, error)
	ListUnlocks(ctx context.Context, crewID string) ([]game.LoreUnlock, error)
}

var handler LoreHandler

func Setup(h LoreHandler) {
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

	path := strings.Trim(r.URL.Path, "/")
	rest := strings.TrimPrefix(path, "api/concepts")
	rest = strings.Trim(rest, "/")

	switch rest {
	case "", "summary":
		handleSummary(w, r, claims)
	case "unlocked":
		handleUnlocked(w, r, claims)
	case "unlocks":
		handleUnlocks(w, r, claims)
	default:
		shared.WriteJSONError(w, "not found", http.StatusNotFound)
	}
}

func handleSummary(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	summary, err := handler.GetSummary(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to get concept summary", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, summary)
}

func handleUnlocked(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	result, err := handler.ListUnlocked(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to list unlocked concept", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, result)
}

func handleUnlocks(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	result, err := handler.ListUnlocks(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to list concept unlocks", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, result)
}
