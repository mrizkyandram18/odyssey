package push

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/shared"
)

var store db.PushSubscriptionStore

func Setup(s db.PushSubscriptionStore) {
	store = s
}

type PushSubscribeRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

type PushUnsubscribeRequest struct {
	Endpoint string `json:"endpoint"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if store == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodPost:
		handleSubscribe(w, r, claims)
	case http.MethodDelete:
		handleUnsubscribe(w, r, claims)
	default:
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSubscribe(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	var req PushSubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Endpoint = strings.TrimSpace(req.Endpoint)
	req.Keys.P256dh = strings.TrimSpace(req.Keys.P256dh)
	req.Keys.Auth = strings.TrimSpace(req.Keys.Auth)

	if req.Endpoint == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		shared.WriteJSONError(w, "missing required subscription fields", http.StatusBadRequest)
		return
	}

	if len(req.Endpoint) > 2048 || len(req.Keys.P256dh) > 512 || len(req.Keys.Auth) > 512 {
		shared.WriteJSONError(w, "field length exceeds maximum limit", http.StatusBadRequest)
		return
	}

	u, err := url.Parse(req.Endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		shared.WriteJSONError(w, "endpoint must be a valid https URL", http.StatusBadRequest)
		return
	}

	sub := &db.PushSubscription{
		UID:      claims.UID,
		Endpoint: req.Endpoint,
		P256DH:   req.Keys.P256dh,
		Auth:     req.Keys.Auth,
	}

	_, err = store.UpsertSubscription(r.Context(), sub)
	if err != nil {
		shared.WriteJSONError(w, "failed to save push subscription", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, map[string]string{"status": "subscribed"})
}

func handleUnsubscribe(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	endpoint := strings.TrimSpace(r.URL.Query().Get("endpoint"))

	if endpoint == "" && r.Body != nil && r.ContentLength > 0 {
		var req PushUnsubscribeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			endpoint = strings.TrimSpace(req.Endpoint)
		}
	}

	if endpoint != "" {
		if len(endpoint) > 2048 {
			shared.WriteJSONError(w, "endpoint length exceeds maximum limit", http.StatusBadRequest)
			return
		}
		u, err := url.Parse(endpoint)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			shared.WriteJSONError(w, "endpoint must be a valid https URL", http.StatusBadRequest)
			return
		}
	}

	if err := store.DeleteSubscription(r.Context(), claims.UID, endpoint); err != nil {
		shared.WriteJSONError(w, "failed to delete push subscription", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "unsubscribed"})
}
