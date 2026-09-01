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

	if r.Method != http.MethodGet && r.Method != http.MethodPatch && r.Method != http.MethodPost {
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

	if r.Method == http.MethodPost {
		if r.URL.Path != "/api/me/change-password" && r.URL.Path != "/api/me/change-password/" {
			shared.WriteJSONError(w, "not found", http.StatusNotFound)
			return
		}

		var req struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
			ConfirmPassword string `json:"confirm_password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(req.NewPassword) < 6 {
			shared.WriteJSONError(w, "kata sandi baru minimal 6 karakter", http.StatusBadRequest)
			return
		}
		if req.ConfirmPassword != "" && req.NewPassword != req.ConfirmPassword {
			shared.WriteJSONError(w, "konfirmasi kata sandi tidak cocok", http.StatusBadRequest)
			return
		}

		// For manual change (non-forced), current_password is required.
		// For forced first-login (must_change_password=true) we allow missing current for backward compat,
		// but if provided we verify it. For security, when current is provided we always verify.
		needsCurrentCheck := req.CurrentPassword != ""
		// Also enforce current check when user is not in forced state, to satisfy spec D.
		if !needsCurrentCheck {
			// Peek profile to see if forced change; if not forced, require current
			if prof, err := profiles.GetUserProfile(r.Context(), claims.UID); err == nil && prof != nil && !prof.MustChangePassword {
				shared.WriteJSONError(w, "kata sandi saat ini wajib diisi", http.StatusBadRequest)
				return
			}
		}
		if needsCurrentCheck {
			hash, err := profiles.GetPasswordHash(r.Context(), claims.UID)
			if err != nil {
				shared.WriteJSONError(w, "gagal memverifikasi kata sandi: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if hash == "" {
				shared.WriteJSONError(w, "kredensial tidak ditemukan", http.StatusNotFound)
				return
			}
			hasher := auth.NewBcryptHasher()
			if err := hasher.Verify(hash, req.CurrentPassword); err != nil {
				shared.WriteJSONError(w, "kata sandi saat ini tidak valid", http.StatusUnauthorized)
				return
			}
			if req.CurrentPassword == req.NewPassword {
				shared.WriteJSONError(w, "kata sandi baru tidak boleh sama dengan kata sandi saat ini", http.StatusBadRequest)
				return
			}
		} else {
			// Even without current, prevent reuse if we can fetch hash to compare (best effort)
			// Skip if hash fetch fails
			if hash, err := profiles.GetPasswordHash(r.Context(), claims.UID); err == nil && hash != "" {
				hasher := auth.NewBcryptHasher()
				if err := hasher.Verify(hash, req.NewPassword); err == nil {
					shared.WriteJSONError(w, "kata sandi baru tidak boleh sama dengan kata sandi saat ini", http.StatusBadRequest)
					return
				}
			}
		}

		if err := profiles.ChangePassword(r.Context(), claims.UID, req.NewPassword); err != nil {
			shared.WriteJSONError(w, "gagal mengubah kata sandi: "+err.Error(), http.StatusInternalServerError)
			return
		}

		shared.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "success",
			"message": "Kata sandi berhasil diubah",
		})
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
