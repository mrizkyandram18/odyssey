package collections

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"odyssey/pkg/auth"
	gamerelic "odyssey/pkg/game/collection"
	"odyssey/pkg/shared"
)

// Service is the interface the handler depends on.
type Service interface {
	ListRelics(ctx context.Context) ([]gamerelic.RelicDefinition, error)
	GetRelic(ctx context.Context, slug string) (*gamerelic.RelicDefinition, error)
	ListInventory(ctx context.Context, uid string) ([]gamerelic.InventoryItem, error)
	GiftRelic(ctx context.Context, senderUID, recipientUID, relicSlug, crewID string) (*gamerelic.GiftResult, error)
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
		if strings.HasSuffix(r.URL.Path, "/gift") {
			handleGift(w, r)
		} else {
			shared.WriteJSONError(w, "not found", http.StatusNotFound)
		}
	default:
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type GiftRequest struct {
	RecipientUID string `json:"recipient_uid"`
	CollectionSlug    string `json:"collection_slug"`
}

func handleGift(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if svc == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	var req GiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONError(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	if req.RecipientUID == "" || req.CollectionSlug == "" {
		shared.WriteJSONError(w, "missing required fields", http.StatusBadRequest)
		return
	}

	res, err := svc.GiftRelic(r.Context(), claims.UID, req.RecipientUID, req.CollectionSlug, claims.FamilyID)
	if err != nil {
		switch err {
		case gamerelic.ErrSelfGift:
			shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		case gamerelic.ErrCrossCrewGift:
			shared.WriteJSONError(w, err.Error(), http.StatusForbidden)
		case gamerelic.ErrRecipientNotFound:
			shared.WriteJSONError(w, err.Error(), http.StatusNotFound)
		case gamerelic.ErrRelicNotOwned:
			shared.WriteJSONError(w, err.Error(), http.StatusConflict)
		case gamerelic.ErrRelicNotFound:
			shared.WriteJSONError(w, err.Error(), http.StatusNotFound)
		default:
			// Treat unexpected errors as 500, but in this simplified system we might get database errors
			shared.WriteJSONError(w, "failed to gift relic", http.StatusInternalServerError)
		}
		return
	}

	shared.WriteJSON(w, http.StatusOK, res)
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

	path := strings.TrimPrefix(r.URL.Path, "/api/collections")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	if path == "" || path == "inventory" {
		items, err := svc.ListInventory(r.Context(), claims.UID)
		if err != nil {
			shared.WriteJSONError(w, "failed to load inventory", http.StatusInternalServerError)
			return
		}
		shared.WriteJSON(w, http.StatusOK, items)
		return
	}

	if strings.Contains(path, "/") {
		shared.WriteJSONError(w, "not found", http.StatusNotFound)
		return
	}

	if !shared.ValidateSlug(path) || len(path) > 256 {
		shared.WriteJSONError(w, "invalid slug", http.StatusBadRequest)
		return
	}

	def, err := svc.GetRelic(r.Context(), path)
	if err != nil {
		shared.WriteJSONError(w, "relic not found", http.StatusNotFound)
		return
	}
	shared.WriteJSON(w, http.StatusOK, def)
}
