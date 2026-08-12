package gifts

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	gamechest "odyssey/pkg/game/gift"
	"odyssey/pkg/shared"
)

// Service is the interface the handler depends on.
type Service interface {
	ListChests(ctx context.Context, uid string) ([]game.Gift, error)
	GetChest(ctx context.Context, chestID int64, uid string) (*game.Gift, error)
	OpenChest(ctx context.Context, chestID int64, uid string) (*gamechest.OpenResult, error)
}

var svc Service

func Setup(s Service) {
	svc = s
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPost:
		handlePost(w, r)
	default:
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if svc == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/gifts")
	path = strings.TrimPrefix(path, "/")
	if path == "" || path == " " {
		gifts, err := svc.ListChests(r.Context(), claims.UID)
		if err != nil {
			shared.WriteJSONError(w, "failed to load gifts", http.StatusInternalServerError)
			return
		}
		views := make([]gamechest.ChestView, 0, len(gifts))
		for i := range gifts {
			views = append(views, gamechest.ToChestView(gifts[i]))
		}
		shared.WriteJSON(w, http.StatusOK, views)
		return
	}

	chestID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		shared.WriteJSONError(w, "invalid chest id", http.StatusBadRequest)
		return
	}

	ch, err := svc.GetChest(r.Context(), chestID, claims.UID)
	if err != nil {
		if err == gamechest.ErrChestNotFound {
			shared.WriteJSONError(w, "chest not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to load chest", http.StatusInternalServerError)
		return
	}
	view := gamechest.ToChestView(*ch)
	shared.WriteJSON(w, http.StatusOK, &view)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if svc == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/gifts")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/open")
	path = strings.Trim(path, "/")

	if path == "" {
		shared.WriteJSONError(w, "chest id required", http.StatusBadRequest)
		return
	}

	chestID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		shared.WriteJSONError(w, "invalid chest id", http.StatusBadRequest)
		return
	}

	result, err := svc.OpenChest(r.Context(), chestID, claims.UID)
	if err != nil {
		if err == gamechest.ErrChestNotFound {
			shared.WriteJSONError(w, "chest not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to open chest", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, result)
}
