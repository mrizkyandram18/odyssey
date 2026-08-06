package reactions

import (
	"context"
	"encoding/json"
	"net/http"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	"odyssey/pkg/shared"
)

// Service is the interface the handler depends on.
type Service interface {
	AddReaction(ctx context.Context, creatorID, targetUserID string, questID *string, emojiCode string) (*game.Reaction, error)
	ListReactionsForTarget(ctx context.Context, targetUserID string) ([]game.Reaction, error)
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
		target := r.URL.Query().Get("target")
		if target == "" {
			target = claims.UID // default to current user
		}
		reactions, err := svc.ListReactionsForTarget(r.Context(), target)
		if err != nil {
			shared.WriteJSONError(w, "failed to list reactions", http.StatusInternalServerError)
			return
		}
		shared.WriteJSON(w, http.StatusOK, map[string]any{"reactions": reactions})
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			TargetUserID string  `json:"target_user_id"`
			QuestID      *string `json:"quest_id,omitempty"`
			EmojiCode    string  `json:"emoji_code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		reaction, err := svc.AddReaction(r.Context(), claims.UID, req.TargetUserID, req.QuestID, req.EmojiCode)
		if err != nil {
			shared.WriteJSONError(w, "failed to add reaction", http.StatusInternalServerError)
			return
		}
		shared.WriteJSON(w, http.StatusOK, reaction)
		return
	}

	shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
}
