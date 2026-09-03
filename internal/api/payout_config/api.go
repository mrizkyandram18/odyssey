package payout_config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/payout"
	"odyssey/pkg/shared"
)

type API struct {
	client db.SupabaseClient
}

func NewAPI(client db.SupabaseClient) *API { return &API{client: client} }

func (a *API) requireAdmin(w http.ResponseWriter, r *http.Request) (*auth.SessionClaims, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims == nil {
		shared.WriteUnauthorized(w)
		return nil, false
	}
	if auth.NormalizeRole(string(claims.Role)) != auth.RoleAdmin {
		shared.WriteJSONError(w, "akses ditolak: hanya untuk admin", http.StatusForbidden)
		return nil, false
	}
	return claims, true
}

// GET /api/admin/payout-config?user_uid=xxx
func (a *API) HandleGet(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	ctx := r.Context()
	userUID := strings.TrimSpace(r.URL.Query().Get("user_uid"))
	if userUID != "" {
		raw, err := a.client.Get(ctx, "odyssey_user_payout_config", "user_uid=eq."+userUID)
		if err != nil {
			shared.WriteJSONError(w, "gagal mengambil konfigurasi payout", http.StatusInternalServerError)
			return
		}
		if strings.TrimSpace(string(raw)) == "[]" || len(raw) == 0 {
			eff, _ := payout.GetEffectivePayoutConfig(ctx, a.client, userUID, time.Now())
			shared.WriteJSON(w, http.StatusOK, eff)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(raw)
		return
	}
	raw, _ := a.client.Get(ctx, "odyssey_system_config", "key=in.(default_payout_frequency,default_minimum_withdrawal_coins,default_payout_weekday,default_payout_month_start_day,default_payout_month_end_day,redemption_start_day,redemption_end_day,timezone)&select=key,value")
	w.Header().Set("Content-Type", "application/json")
	if raw == nil {
		raw = []byte("[]")
	}
	w.Write(raw)
}

// PUT /api/admin/payout-config
type UpsertRequest struct {
	UserUID               string  `json:"user_uid"`
	PayoutFrequency       string  `json:"payout_frequency"`
	MinimumWithdrawalCoins int     `json:"minimum_withdrawal_coins"`
	PayoutWeekday         *int    `json:"payout_weekday"`
	PayoutMonthStartDay   *int    `json:"payout_month_start_day"`
	PayoutMonthEndDay     *int    `json:"payout_month_end_day"`
	Enabled               *bool   `json:"enabled"`
}

func (a *API) HandleUpsert(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	var req UpsertRequest
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.UserUID = strings.TrimSpace(req.UserUID)
	if req.UserUID == "" {
		shared.WriteJSONError(w, "user_uid wajib diisi", http.StatusBadRequest)
		return
	}
	// Tenant check: target must be in admin's family
	profRaw, err := a.client.Get(ctx, "odyssey_user_profiles", "uid=eq."+req.UserUID+"&select=family_id")
	if err != nil || len(profRaw) == 0 {
		shared.WriteJSONError(w, "user tidak ditemukan", http.StatusNotFound)
		return
	}
	var profs []struct {
		FamilyID string `json:"family_id"`
	}
	_ = json.Unmarshal(profRaw, &profs)
	if len(profs) > 0 && claims.FamilyID != "" && profs[0].FamilyID != "" && profs[0].FamilyID != claims.FamilyID {
		shared.WriteJSONError(w, "akses ditolak: user bukan bagian dari keluarga Anda", http.StatusForbidden)
		return
	}
	freq := payout.NormalizeFrequency(req.PayoutFrequency)
	if !payout.IsValidFrequency(req.PayoutFrequency) {
		shared.WriteJSONError(w, "payout_frequency tidak valid (THRESHOLD|WEEKLY|MONTHLY)", http.StatusBadRequest)
		return
	}
	// Validate minimum against system minimum
	sysMin := payout.GetSystemMinimumWithdrawal(ctx, a.client)
	// Allow admin to set any >0, but if below system min, still allow? Spec says systemMinimumWithdrawal configurable then userMinimum >= systemMinimum if intended.
	// We enforce system min as floor.
	if req.MinimumWithdrawalCoins <= 0 {
		shared.WriteJSONError(w, "minimum_withdrawal_coins harus > 0", http.StatusBadRequest)
		return
	}
	if req.MinimumWithdrawalCoins < sysMin {
		shared.WriteJSONError(w, fmt.Sprintf("minimum_withdrawal_coins (%d) di bawah system minimum (%d)", req.MinimumWithdrawalCoins, sysMin), http.StatusBadRequest)
		return
	}
	if req.MinimumWithdrawalCoins > 100000 {
		shared.WriteJSONError(w, "minimum_withdrawal_coins terlalu besar (max 100000)", http.StatusBadRequest)
		return
	}
	// Validate schedule per frequency
	if freq == payout.FrequencyWeekly {
		if req.PayoutWeekday == nil {
			// default to system
			raw, _ := a.client.Get(ctx, "odyssey_system_config", "key=eq.default_payout_weekday&select=value")
			if len(raw) > 0 {
				var rows []struct{ Value string `json:"value"` }
				_ = json.Unmarshal(raw, &rows)
				if len(rows) > 0 {
					if v, err := strconv.Atoi(strings.TrimSpace(rows[0].Value)); err == nil {
						req.PayoutWeekday = &v
					}
				}
			}
			if req.PayoutWeekday == nil {
				v := payout.DefaultWeeklyWeekday
				req.PayoutWeekday = &v
			}
		}
		if *req.PayoutWeekday < 0 || *req.PayoutWeekday > 6 {
			shared.WriteJSONError(w, "payout_weekday harus 0..6 (0=Sunday)", http.StatusBadRequest)
			return
		}
	}
	if freq == payout.FrequencyMonthly {
		if req.PayoutMonthStartDay == nil || req.PayoutMonthEndDay == nil {
			// fallback to system redemption window
			raw, _ := a.client.Get(ctx, "odyssey_system_config", "key=in.(redemption_start_day,redemption_end_day)&select=key,value")
			var rows []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			_ = json.Unmarshal(raw, &rows)
			m := map[string]int{}
			for _, r := range rows {
				if v, err := strconv.Atoi(strings.TrimSpace(r.Value)); err == nil {
					m[r.Key] = v
				}
			}
			if req.PayoutMonthStartDay == nil {
				if v, ok := m["redemption_start_day"]; ok {
					req.PayoutMonthStartDay = &v
				} else {
					v := 24
					req.PayoutMonthStartDay = &v
				}
			}
			if req.PayoutMonthEndDay == nil {
				if v, ok := m["redemption_end_day"]; ok {
					req.PayoutMonthEndDay = &v
				} else {
					v := 26
					req.PayoutMonthEndDay = &v
				}
			}
		}
		if *req.PayoutMonthStartDay < 1 || *req.PayoutMonthStartDay > 31 || *req.PayoutMonthEndDay < 1 || *req.PayoutMonthEndDay > 31 {
			shared.WriteJSONError(w, "payout_month_start/end_day harus 1..31", http.StatusBadRequest)
			return
		}
		if *req.PayoutMonthStartDay > *req.PayoutMonthEndDay {
			shared.WriteJSONError(w, "payout_month_start_day tidak boleh lebih besar dari end_day", http.StatusBadRequest)
			return
		}
	}

	payload := map[string]any{
		"user_uid":                 req.UserUID,
		"payout_frequency":         string(freq),
		"minimum_withdrawal_coins": req.MinimumWithdrawalCoins,
		"enabled":                  true,
		"updated_at":               time.Now().UTC().Format(time.RFC3339),
	}
	if req.PayoutWeekday != nil {
		payload["payout_weekday"] = *req.PayoutWeekday
	}
	if req.PayoutMonthStartDay != nil {
		payload["payout_month_start_day"] = *req.PayoutMonthStartDay
	}
	if req.PayoutMonthEndDay != nil {
		payload["payout_month_end_day"] = *req.PayoutMonthEndDay
	}
	if req.Enabled != nil {
		payload["enabled"] = *req.Enabled
	}

	// Upsert via MutateAtomic merge-duplicates
	// Try insert with on conflict handling via POST with Prefer resolution=merge-duplicates
	_, err = a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_user_payout_config", payload, "", "resolution=merge-duplicates")
	if err != nil {
		// Fallback: try PATCH
		_, err = a.client.Mutate(ctx, http.MethodPatch, "odyssey_user_payout_config", payload, "user_uid=eq."+req.UserUID)
		if err != nil {
			// Try insert then update logic: check existence
			existing, _ := a.client.Get(ctx, "odyssey_user_payout_config", "user_uid=eq."+req.UserUID)
			if strings.TrimSpace(string(existing)) != "[]" && len(existing) > 2 {
				shared.WriteJSONError(w, "gagal update konfigurasi payout: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = a.client.Mutate(ctx, http.MethodPost, "odyssey_user_payout_config", payload, "")
			if err != nil {
				shared.WriteJSONError(w, "gagal menyimpan konfigurasi payout: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
	eff, _ := payout.GetEffectivePayoutConfig(ctx, a.client, req.UserUID, time.Now())
	shared.WriteJSON(w, http.StatusOK, eff)
}

// System config update for payout globals
type SystemConfigRequest struct {
	DefaultPayoutFrequency       *string `json:"default_payout_frequency"`
	DefaultMinimumWithdrawalCoins *int    `json:"default_minimum_withdrawal_coins"`
	DefaultPayoutWeekday         *int    `json:"default_payout_weekday"`
	DefaultPayoutMonthStartDay   *int    `json:"default_payout_month_start_day"`
	DefaultPayoutMonthEndDay     *int    `json:"default_payout_month_end_day"`
}

func (a *API) HandleUpdateSystemConfig(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	ctx := r.Context()
	var req SystemConfigRequest
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	upsert := func(key, value string) {
		payload := map[string]any{"key": key, "value": value}
		_, err := a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_system_config", payload, "", "resolution=merge-duplicates")
		if err != nil {
			_, _ = a.client.Mutate(ctx, http.MethodPatch, "odyssey_system_config", map[string]any{"value": value}, "key=eq."+key)
		}
	}
	if req.DefaultPayoutFrequency != nil {
		if !payout.IsValidFrequency(*req.DefaultPayoutFrequency) {
			shared.WriteJSONError(w, "default_payout_frequency tidak valid", http.StatusBadRequest)
			return
		}
		upsert("default_payout_frequency", strings.ToUpper(strings.TrimSpace(*req.DefaultPayoutFrequency)))
	}
	if req.DefaultMinimumWithdrawalCoins != nil {
		if *req.DefaultMinimumWithdrawalCoins <= 0 || *req.DefaultMinimumWithdrawalCoins > 100000 {
			shared.WriteJSONError(w, "default_minimum_withdrawal_coins harus 1..100000", http.StatusBadRequest)
			return
		}
		upsert("default_minimum_withdrawal_coins", strconv.Itoa(*req.DefaultMinimumWithdrawalCoins))
	}
	if req.DefaultPayoutWeekday != nil {
		if *req.DefaultPayoutWeekday < 0 || *req.DefaultPayoutWeekday > 6 {
			shared.WriteJSONError(w, "default_payout_weekday harus 0..6", http.StatusBadRequest)
			return
		}
		upsert("default_payout_weekday", strconv.Itoa(*req.DefaultPayoutWeekday))
	}
	if req.DefaultPayoutMonthStartDay != nil {
		if *req.DefaultPayoutMonthStartDay < 1 || *req.DefaultPayoutMonthStartDay > 31 {
			shared.WriteJSONError(w, "default_payout_month_start_day harus 1..31", http.StatusBadRequest)
			return
		}
		upsert("default_payout_month_start_day", strconv.Itoa(*req.DefaultPayoutMonthStartDay))
	}
	if req.DefaultPayoutMonthEndDay != nil {
		if *req.DefaultPayoutMonthEndDay < 1 || *req.DefaultPayoutMonthEndDay > 31 {
			shared.WriteJSONError(w, "default_payout_month_end_day harus 1..31", http.StatusBadRequest)
			return
		}
		upsert("default_payout_month_end_day", strconv.Itoa(*req.DefaultPayoutMonthEndDay))
	}
	// Validate window ordering if both provided or one updated
	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	if strings.HasPrefix(path, "/api/admin/payout-config") {
		if path == "/api/admin/payout-config" && r.Method == http.MethodGet {
			a.HandleGet(w, r)
			return
		}
		if path == "/api/admin/payout-config/user" && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
			a.HandleUpsert(w, r)
			return
		}
		if path == "/api/admin/payout-config/system" && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
			a.HandleUpdateSystemConfig(w, r)
			return
		}
		if path == "/api/admin/payout-config/system" && r.Method == http.MethodGet {
			a.HandleGet(w, r)
			return
		}
	}
	http.NotFound(w, r)
}
