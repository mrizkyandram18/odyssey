package daily_missions

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	"odyssey/pkg/game/dailymission"
	"odyssey/pkg/shared"
)

// DailyTurnHandler is the interface the handler depends on.
type DailyTurnHandler interface {
	List(ctx context.Context, uid string) ([]game.DailyMission, error)
	Consume(ctx context.Context, uid string, questSlug string) (*dailymission.ConsumeDailyTurnResult, error)
}

var handler DailyTurnHandler

func Setup(h DailyTurnHandler) {
	handler = h
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch {
	case r.Method == http.MethodGet && !strings.Contains(r.URL.Path, "/consume"):
		handleList(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/consume"):
		handleConsume(w, r)
	default:
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func handleList(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if handler == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	turns, err := handler.List(r.Context(), claims.UID)
	if err != nil {
		if isNotFound(err) {
			shared.WriteJSONError(w, "daily turns not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to load daily turns", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, turns)
}

func handleConsume(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if handler == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		MissionSlug string `json:"mission_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.MissionSlug = shared.SanitizeString(req.MissionSlug, 256)
	if req.MissionSlug == "" {
		req.MissionSlug = "daily-turn"
	} else if !shared.ValidateSlug(req.MissionSlug) {
		shared.WriteJSONError(w, "invalid mission_slug", http.StatusBadRequest)
		return
	}

	result, err := handler.Consume(r.Context(), claims.UID, req.MissionSlug)
	if err != nil {
		if isNoTurnsRemaining(err) {
			shared.WriteJSONError(w, "no turns remaining", http.StatusConflict)
			return
		}
		shared.WriteJSONError(w, "failed to consume daily turn", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, result)
}

func isNotFound(err error) bool {
	return err != nil && (err == dailymission.ErrDailyTurnNotFound || strings.Contains(err.Error(), "not found"))
}

func isNoTurnsRemaining(err error) bool {
	return err != nil && (err == dailymission.ErrNoTurnsRemaining || strings.Contains(err.Error(), "no turns remaining"))
}
