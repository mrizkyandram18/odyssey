package rewards

import (
	"encoding/json"
	"net/http"

	"odyssey/pkg/auth"
	"odyssey/pkg/game/reward"
	"odyssey/pkg/shared"
)

var service *reward.Service

// Setup configures the rewards API dependencies.
func Setup(svc *reward.Service) {
	service = svc
}

// Handler serves the public rewards endpoint for Vercel.
func Handler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if r.Method == http.MethodGet {
		if service == nil {
			shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
			return
		}
		handleList(w, r, service, claims.UID)
		return
	}
	shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
}

func handleList(w http.ResponseWriter, r *http.Request, svc *reward.Service, uid string) {
	ctx := r.Context()
	ledgers, err := svc.GetLedger(ctx, uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ledgers)
}
