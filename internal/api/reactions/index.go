package reactions

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	"odyssey/pkg/shared"
)

// Service is the interface the handler depends on.
type Service interface {
	AddReaction(ctx context.Context, crewID, actorUID string, targetType string, targetID int64, reactionType string) (*game.Reaction, error)
	ListReactionsForTarget(ctx context.Context, crewID, targetType string, targetID int64) ([]game.Reaction, error)
}

var svc Service

func Setup(s Service) {
	svc = s
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if svc == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	if r.Method == http.MethodGet {
		targetType := r.URL.Query().Get("target_type")
		targetIDStr := r.URL.Query().Get("target_id")

		if targetType == "" || targetIDStr == "" {
			shared.WriteJSONError(w, "missing target_type or target_id", http.StatusBadRequest)
			return
		}

		targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
		if err != nil {
			shared.WriteJSONError(w, "invalid target_id", http.StatusBadRequest)
			return
		}

		reactions, err := svc.ListReactionsForTarget(r.Context(), claims.FamilyID, targetType, targetID)
		if err != nil {
			shared.WriteJSONError(w, "failed to list reactions", http.StatusInternalServerError)
			return
		}
		shared.WriteJSON(w, http.StatusOK, map[string]any{"reactions": reactions})
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			TargetType   string `json:"target_type"`
			TargetID     int64  `json:"target_id"`
			ReactionType string `json:"reaction_type"`
			// Note: We deliberately ignore any ActorUID sent by client
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Security: Always use claims.UID as the actor! Spoof rejection!
		reaction, err := svc.AddReaction(r.Context(), claims.FamilyID, claims.UID, req.TargetType, req.TargetID, req.ReactionType)
		if err != nil {
			// Do not leak internal error strings directly if they are generic, but for now we'll pass it to client for debugging
			shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		shared.WriteJSON(w, http.StatusOK, reaction)
		return
	}

	shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
}
