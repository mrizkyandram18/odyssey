package me

import (
	"encoding/json"
	"errors"
	"net/http"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/shared"
)

var profiles db.ProfileStore

func Setup(p db.ProfileStore) {
	profiles = p
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet && r.Method != http.MethodPatch {
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if claims.Kind != string(auth.SessionKindUser) {
		shared.WriteForbidden(w)
		return
	}

	if profiles == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	if r.Method == http.MethodPatch {
		if r.URL.Path != "/api/me/avatar" && r.URL.Path != "/api/me/avatar/" {
			shared.WriteJSONError(w, "not found", http.StatusNotFound)
			return
		}

		var req struct {
			Style string `json:"avatar_style"`
			Seed  string `json:"avatar_seed"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Style == "" || req.Seed == "" {
			shared.WriteJSONError(w, "style and seed are required", http.StatusBadRequest)
			return
		}

		if err := profiles.UpdateAvatar(r.Context(), claims.UID, req.Style, req.Seed); err != nil {
			shared.WriteJSONError(w, "failed to update avatar", http.StatusInternalServerError)
			return
		}

		shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
		return
	}

	profile, err := profiles.GetUserProfile(r.Context(), claims.UID)
	if err != nil {
		if errors.Is(err, db.ErrProfileNotFound) {
			shared.WriteJSONError(w, "profile not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to load profile", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, profile)
}
