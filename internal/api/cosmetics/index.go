package cosmetics

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"odyssey/pkg/auth"
	"odyssey/pkg/game/cosmetic"
	"odyssey/pkg/shared"
)

// Service is the handler dependency.
type Service interface {
	ListForUser(ctx context.Context, uid string) (*cosmetic.ListResult, error)
	Purchase(ctx context.Context, uid, cosmeticID string) (*cosmetic.PurchaseResult, error)
	Equip(ctx context.Context, uid, cosmeticID string) error
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

	path := r.URL.Path
	switch {
	case r.Method == http.MethodGet && (path == "/api/cosmetics" || path == "/api/cosmetics/"):
		handleList(w, r, claims.UID)
	case r.Method == http.MethodPost && (path == "/api/cosmetics/purchase" || path == "/api/cosmetics/purchase/"):
		handlePurchase(w, r, claims.UID)
	case r.Method == http.MethodPost && (path == "/api/cosmetics/equip" || path == "/api/cosmetics/equip/"):
		handleEquip(w, r, claims.UID)
	default:
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleList(w http.ResponseWriter, r *http.Request, uid string) {
	res, err := svc.ListForUser(r.Context(), uid)
	if err != nil {
		shared.WriteJSONError(w, "failed to list cosmetics", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, res)
}

func handlePurchase(w http.ResponseWriter, r *http.Request, uid string) {
	var req struct {
		CosmeticID string `json:"cosmetic_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CosmeticID == "" {
		shared.WriteJSONError(w, "cosmetic_id is required", http.StatusBadRequest)
		return
	}

	res, err := svc.Purchase(r.Context(), uid, req.CosmeticID)
	if err != nil {
		switch {
		case errors.Is(err, cosmetic.ErrUnknownCosmetic):
			shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, cosmetic.ErrInsufficientCoins):
			shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, cosmetic.ErrConcurrent):
			shared.WriteJSONError(w, err.Error(), http.StatusConflict)
		default:
			shared.WriteJSONError(w, "purchase failed", http.StatusInternalServerError)
		}
		return
	}
	shared.WriteJSON(w, http.StatusOK, res)
}

func handleEquip(w http.ResponseWriter, r *http.Request, uid string) {
	var req struct {
		CosmeticID string `json:"cosmetic_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CosmeticID == "" {
		shared.WriteJSONError(w, "cosmetic_id is required", http.StatusBadRequest)
		return
	}

	err := svc.Equip(r.Context(), uid, req.CosmeticID)
	if err != nil {
		switch {
		case errors.Is(err, cosmetic.ErrUnknownCosmetic):
			shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, cosmetic.ErrNotOwned):
			shared.WriteJSONError(w, err.Error(), http.StatusForbidden)
		default:
			shared.WriteJSONError(w, "equip failed", http.StatusInternalServerError)
		}
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "equipped"})
}
