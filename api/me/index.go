package me

import (
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

	if r.Method != http.MethodGet {
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
