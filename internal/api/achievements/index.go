package achievements

import (
	"context"
	"net/http"

	"odyssey/pkg/auth"
	"odyssey/pkg/game/achievement"
	"odyssey/pkg/shared"
)

type AchievementHandler interface {
	ListByPlayer(ctx context.Context, uid string) ([]achievement.AchievementView, error)
}

var handler AchievementHandler

func Setup(h AchievementHandler) {
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

	result, err := handler.ListByPlayer(r.Context(), claims.UID)
	if err != nil {
		shared.WriteJSONError(w, "failed to list achievements", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, result)
}
