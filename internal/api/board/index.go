package board

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	gameboard "odyssey/pkg/game/board"
	"odyssey/pkg/shared"
)

// Service is the HTTP handler dependency.
type Service interface {
	PostText(ctx context.Context, crewID, authorUID, content string) (*game.CreativeItem, error)
	ListForCrew(ctx context.Context, crewID string) ([]game.CreativeItem, error)
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
	if claims.Kind != string(auth.SessionKindUser) {
		shared.WriteForbidden(w)
		return
	}
	if svc == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleList(w, r, claims.CrewID)
	case http.MethodPost:
		handlePost(w, r, claims.UID, claims.CrewID)
	default:
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleList(w http.ResponseWriter, r *http.Request, crewID string) {
	items, err := svc.ListForCrew(r.Context(), crewID)
	if err != nil {
		shared.WriteJSONError(w, "failed to list board posts", http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []game.CreativeItem{}
	}
	shared.WriteJSON(w, http.StatusOK, map[string]any{"posts": items})
}

func handlePost(w http.ResponseWriter, r *http.Request, uid, crewID string) {
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	item, err := svc.PostText(r.Context(), crewID, uid, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, gameboard.ErrEmptyContent), errors.Is(err, gameboard.ErrContentTooLong):
			shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		default:
			shared.WriteJSONError(w, "failed to post", http.StatusInternalServerError)
		}
		return
	}
	shared.WriteJSON(w, http.StatusCreated, item)
}
