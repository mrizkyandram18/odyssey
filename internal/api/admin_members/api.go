package admin_members

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/shared"
)

type API struct {
	client db.SupabaseClient
	hasher auth.PasswordHasher
}

func NewAPI(client db.SupabaseClient) *API {
	return &API{
		client: client,
		hasher: auth.NewBcryptHasher(),
	}
}

type MemberView struct {
	UID                string    `json:"uid"`
	FamilyID           string    `json:"family_id"`
	ExplorerName       string    `json:"explorer_name"`
	Username           string    `json:"username"`
	Role               string    `json:"role"`
	IsActive           bool      `json:"is_active"`
	Level              int       `json:"level"`
	XP                 int64     `json:"xp"`
	Coins              int64     `json:"coins"`
	MonthlyCoinTarget  *int      `json:"monthly_coin_target,omitempty"`
	EarnedThisPeriod   int       `json:"earned_this_period,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type CreateMemberRequest struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	ExplorerName      string `json:"explorer_name"`
	Role              string `json:"role,omitempty"`
	MonthlyCoinTarget *int   `json:"monthly_coin_target,omitempty"`
}

type UpdateMemberRequest struct {
	ExplorerName      *string `json:"explorer_name,omitempty"`
	Role              *string `json:"role,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
	Password          *string `json:"password,omitempty"`
	ResetDevice       *bool   `json:"reset_device,omitempty"`
	MonthlyCoinTarget *int    `json:"monthly_coin_target,omitempty"`
}

func (a *API) requireAdmin(w http.ResponseWriter, r *http.Request) (*auth.SessionClaims, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims == nil {
		shared.WriteUnauthorized(w)
		return nil, false
	}
	role := auth.NormalizeRole(claims.Role)
	if role != auth.RoleAdmin {
		shared.WriteJSONError(w, "akses ditolak: hanya untuk admin", http.StatusForbidden)
		return nil, false
	}
	return claims, true
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().Unix(), hex.EncodeToString(b))
}

func getDefaultMonthlyTarget(ctx context.Context, client db.SupabaseClient) int {
	raw, err := client.Get(ctx, "odyssey_system_config", "key=eq.default_monthly_coin_target&select=value")
	if err == nil && len(raw) > 0 {
		var rows []struct{ Value string `json:"value"` }
		if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 {
			var n int
			if _, err := fmt.Sscanf(rows[0].Value, "%d", &n); err == nil && n >= 1 && n <= 10000 {
				return n
			}
		}
	}
	return shared.DefaultMonthlyCoinTarget
}

func getEarnedThisPeriod(ctx context.Context, client db.SupabaseClient, uids []string) map[string]int {
	if len(uids) == 0 {
		return map[string]int{}
	}
	tz := shared.LoadConfig().Timezone
	if tz == "" {
		tz = shared.DefaultTimezone
	}
	if raw, err := client.Get(ctx, "odyssey_system_config", "key=eq.timezone&select=value"); err == nil && len(raw) > 0 {
		var rows []struct{ Value string `json:"value"` }
		if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 && strings.TrimSpace(rows[0].Value) != "" {
			if _, err := time.LoadLocation(strings.TrimSpace(rows[0].Value)); err == nil {
				tz = strings.TrimSpace(rows[0].Value)
			}
		}
	}
	loc, _ := time.LoadLocation(tz)
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	nowInTz := time.Now().In(loc)
	// Cutover gate: before 2026-10-25 use legacy 1-24, after use rolling 25
	cutoverStr := "2026-10-25"
	if raw, err := client.Get(ctx, "odyssey_system_config", "key=eq.earning_cycle_cutover_date&select=value"); err == nil && len(raw) > 0 {
		var rows []struct{ Value string `json:"value"` }
		if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 && strings.TrimSpace(rows[0].Value) != "" {
			cutoverStr = strings.TrimSpace(rows[0].Value)
		}
	}
	cutoverDate, _ := time.Parse("2006-01-02", cutoverStr)
	nowDate := time.Date(nowInTz.Year(), nowInTz.Month(), nowInTz.Day(), 0, 0, 0, 0, loc)
	// If before cutover, use legacy 1-24
	if !cutoverDate.IsZero() && nowDate.Before(cutoverDate) {
		startDay, endDay := 1, 24
		if raw, err := client.Get(ctx, "odyssey_system_config", "key=in.(target_earning_start_day,target_earning_end_day)&select=key,value"); err == nil && len(raw) > 0 {
			var rows []struct{ Key string `json:"key"`; Value string `json:"value"` }
			if err := json.Unmarshal(raw, &rows); err == nil {
				for _, r := range rows {
					var n int
					if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil && n >= 1 && n <= 31 {
						if r.Key == "target_earning_start_day" {
							startDay = n
						} else if r.Key == "target_earning_end_day" {
							endDay = n
						}
					}
				}
			}
		}
		pStart := time.Date(nowInTz.Year(), nowInTz.Month(), startDay, 0, 0, 0, 0, loc)
		pEnd := time.Date(nowInTz.Year(), nowInTz.Month(), endDay+1, 0, 0, 0, 0, loc)
		periodStart := pStart.UTC().Format(time.RFC3339)
		periodEnd := pEnd.UTC().Format(time.RFC3339)
		params := fmt.Sprintf("user_uid=in.(%s)&type=eq.TASK_REWARD&created_at=gte.%s&created_at=lt.%s&select=user_uid,amount", strings.Join(uids, ","), periodStart, periodEnd)
		raw, err := client.Get(ctx, "odyssey_coin_transactions", params)
		if err != nil || len(raw) == 0 {
			return map[string]int{}
		}
		var rows []struct{ UserUID string `json:"user_uid"`; Amount int `json:"amount"` }
		if err := json.Unmarshal(raw, &rows); err != nil {
			return map[string]int{}
		}
		m := make(map[string]int, len(uids))
		for _, r := range rows {
			m[r.UserUID] += r.Amount
		}
		return m
	}
	// Post-cutover rolling 25->24
	anchor := 25
	if raw, err := client.Get(ctx, "odyssey_system_config", "key=in.(earning_cycle_anchor_day,target_earning_start_day)&select=key,value"); err == nil && len(raw) > 0 {
		var rows []struct{ Key string `json:"key"`; Value string `json:"value"` }
		if err := json.Unmarshal(raw, &rows); err == nil {
			for _, r := range rows {
				var n int
				if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil && n >= 1 && n <= 31 {
					if r.Key == "earning_cycle_anchor_day" {
						anchor = n
					} else if r.Key == "target_earning_start_day" && anchor == 25 {
						foundAnchor := false
						for _, rr := range rows {
							if rr.Key == "earning_cycle_anchor_day" {
								foundAnchor = true
								break
							}
						}
						if !foundAnchor {
							anchor = n
						}
					}
				}
			}
		}
	}
	var pStart, pEnd time.Time
	if nowInTz.Day() >= anchor {
		pStart = time.Date(nowInTz.Year(), nowInTz.Month(), anchor, 0, 0, 0, 0, loc)
		if nowInTz.Month() == 12 {
			pEnd = time.Date(nowInTz.Year()+1, 1, anchor, 0, 0, 0, 0, loc)
		} else {
			pEnd = time.Date(nowInTz.Year(), nowInTz.Month()+1, anchor, 0, 0, 0, 0, loc)
		}
	} else {
		if nowInTz.Month() == 1 {
			pStart = time.Date(nowInTz.Year()-1, 12, anchor, 0, 0, 0, 0, loc)
		} else {
			pStart = time.Date(nowInTz.Year(), nowInTz.Month()-1, anchor, 0, 0, 0, 0, loc)
		}
		pEnd = time.Date(nowInTz.Year(), nowInTz.Month(), anchor, 0, 0, 0, 0, loc)
	}
	periodStart := pStart.UTC().Format(time.RFC3339)
	periodEnd := pEnd.UTC().Format(time.RFC3339)
	params := fmt.Sprintf("user_uid=in.(%s)&type=eq.TASK_REWARD&created_at=gte.%s&created_at=lt.%s&select=user_uid,amount", strings.Join(uids, ","), periodStart, periodEnd)
	raw, err := client.Get(ctx, "odyssey_coin_transactions", params)
	if err != nil || len(raw) == 0 {
		return map[string]int{}
	}
	var rows []struct{ UserUID string `json:"user_uid"`; Amount int `json:"amount"` }
	if err := json.Unmarshal(raw, &rows); err != nil {
		return map[string]int{}
	}
	m := make(map[string]int, len(uids))
	for _, r := range rows {
		m[r.UserUID] += r.Amount
	}
	return m
}

func (a *API) HandleListMembers(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	page, limit, offset := shared.ParsePagination(r, 50, 100)

	familyID := claims.FamilyID
	if familyID == "" {
		familyID = "family_default"
	}

	// 1. Fetch user profiles strictly scoped to admin's family with deterministic pagination
	profParams := fmt.Sprintf("family_id=eq.%s&order=created_at.desc,uid.desc&limit=%d&offset=%d", familyID, limit, offset)

	profRaw, err := a.client.Get(ctx, "odyssey_user_profiles", profParams)
	if err != nil {
		shared.WriteJSONError(w, "gagal mengambil daftar anggota: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type ProfRow struct {
		UID               string    `json:"uid"`
		FamilyID          string    `json:"family_id"`
		ExplorerName      string    `json:"explorer_name"`
		Role              string    `json:"role"`
		IsActive          bool      `json:"is_active"`
		Level             int       `json:"level"`
		XP                int64     `json:"xp"`
		Coins             int64     `json:"coins"`
		MonthlyCoinTarget *int      `json:"monthly_coin_target"`
		CreatedAt         time.Time `json:"created_at"`
	}
	var profs []ProfRow
	_ = json.Unmarshal(profRaw, &profs)

	if len(profs) == 0 {
		shared.WriteJSON(w, http.StatusOK, shared.PaginatedResponse[MemberView]{
			Items: []MemberView{},
			Pagination: shared.PaginationMeta{
				Page:    page,
				Limit:   limit,
				Total:   0,
				HasNext: false,
			},
		})
		return
	}

	// 2. Fetch local user credentials ONLY for the retrieved profile UIDs (strictly targeted, no global dump)
	uids := make([]string, len(profs))
	for i, p := range profs {
		uids[i] = p.UID
	}
	localParams := fmt.Sprintf("profile_uid=in.(%s)&select=username,profile_uid", strings.Join(uids, ","))
	localRaw, _ := a.client.Get(ctx, "odyssey_local_users", localParams)
	type LocalRow struct {
		Username   string `json:"username"`
		ProfileUID string `json:"profile_uid"`
	}
	var locals []LocalRow
	_ = json.Unmarshal(localRaw, &locals)

	userMap := make(map[string]string, len(locals))
	for _, l := range locals {
		userMap[l.ProfileUID] = l.Username
	}

	earnedMap := getEarnedThisPeriod(ctx, a.client, uids)
	defaultTarget := getDefaultMonthlyTarget(ctx, a.client)
	items := make([]MemberView, len(profs))
	for i, p := range profs {
		username := userMap[p.UID]
		if username == "" {
			username = p.UID
		}
		tgt := p.MonthlyCoinTarget
		if tgt == nil {
			q := defaultTarget
			tgt = &q
		}
		earned := earnedMap[p.UID]
		items[i] = MemberView{
			UID:               p.UID,
			FamilyID:          p.FamilyID,
			ExplorerName:      p.ExplorerName,
			Username:          username,
			Role:              p.Role,
			IsActive:          p.IsActive,
			Level:             p.Level,
			XP:                p.XP,
			Coins:             p.Coins,
			MonthlyCoinTarget: tgt,
			EarnedThisPeriod:  earned,
			CreatedAt:         p.CreatedAt,
		}
	}

	hasNext := len(profs) == limit
	total := offset + len(items)
	if hasNext {
		total += 1 // indicate there are more records beyond current page
	}

	shared.WriteJSON(w, http.StatusOK, shared.PaginatedResponse[MemberView]{
		Items: items,
		Pagination: shared.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: hasNext,
		},
	})
}

func (a *API) HandleCreateMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	var req CreateMemberRequest
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid json body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.ExplorerName = strings.TrimSpace(req.ExplorerName)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || len(req.Username) < 3 {
		shared.WriteJSONError(w, "username minimal 3 karakter", http.StatusBadRequest)
		return
	}
	if !usernameRegex.MatchString(req.Username) {
		shared.WriteJSONError(w, "username hanya boleh berisi huruf, angka, titik, strip, dan underscore", http.StatusBadRequest)
		return
	}
	if req.Password == "" || len(req.Password) < 6 {
		shared.WriteJSONError(w, "password minimal 6 karakter", http.StatusBadRequest)
		return
	}
	if req.ExplorerName == "" {
		shared.WriteJSONError(w, "nama anggota wajib diisi", http.StatusBadRequest)
		return
	}

	role := string(auth.NormalizeRole(req.Role))
	if req.MonthlyCoinTarget != nil {
		if *req.MonthlyCoinTarget < 1 || *req.MonthlyCoinTarget > 10000 {
			shared.WriteJSONError(w, "target koin bulanan harus 1..10000", http.StatusBadRequest)
			return
		}
	}

	// 1. Check username uniqueness
	existingRaw, err := a.client.Get(ctx, "odyssey_local_users", fmt.Sprintf("username=eq.%s", req.Username))
	if err == nil && len(existingRaw) > 2 && string(existingRaw) != "[]" {
		shared.WriteJSONError(w, "username sudah digunakan, pilih username lain", http.StatusBadRequest)
		return
	}

	// 2. Hash password server-side with bcrypt
	passHash, err := a.hasher.Hash(req.Password)
	if err != nil {
		shared.WriteJSONError(w, "gagal memproses enkripsi password", http.StatusInternalServerError)
		return
	}

	familyID := claims.FamilyID
	if familyID == "" {
		familyID = "family_default"
	}

	uid := generateID("usr")

	// 3. Resolve target (explicit or global default)
	target := getDefaultMonthlyTarget(ctx, a.client)
	if req.MonthlyCoinTarget != nil {
		target = *req.MonthlyCoinTarget
	}
	var targetVal *int
	if role == "MEMBER" || role == "SEEKER" {
		q := target
		targetVal = &q
	}

	// 4. Insert user profile
	now := time.Now().UTC()
	profPayload := map[string]any{
		"uid":                  uid,
		"family_id":            familyID,
		"explorer_name":        req.ExplorerName,
		"role":                 role,
		"level":                1,
		"xp":                   0,
		"coins":                0,
		"streak_days":          0,
		"is_active":            true,
		"must_change_password": true,
		"created_at":           now.Format(time.RFC3339),
		"updated_at":           now.Format(time.RFC3339),
	}
	if targetVal != nil {
		profPayload["monthly_coin_target"] = *targetVal
	}

	_, err = a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_user_profiles", profPayload, "", "return=representation")
	if err != nil {
		shared.WriteJSONError(w, "gagal membuat profil anggota: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Insert local auth user
	localID := generateID("loc")
	localPayload := map[string]any{
		"id":            localID,
		"username":      req.Username,
		"password_hash": passHash,
		"profile_uid":   uid,
		"created_at":    now.Format(time.RFC3339),
		"updated_at":    now.Format(time.RFC3339),
	}

	_, err = a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_local_users", localPayload, "", "return=representation")
	if err != nil {
		// Rollback profile on auth insertion failure
		_, _ = a.client.Mutate(ctx, http.MethodDelete, "odyssey_user_profiles", nil, fmt.Sprintf("uid=eq.%s", uid))
		shared.WriteJSONError(w, "gagal membuat kredensial login: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Create current-period target history (best-effort)
	if targetVal != nil {
		tz := shared.LoadConfig().Timezone
		if tz == "" {
			tz = shared.DefaultTimezone
		}
		if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=eq.timezone&select=value"); err == nil && len(raw) > 0 {
			var rows []struct{ Value string `json:"value"` }
			if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 && strings.TrimSpace(rows[0].Value) != "" {
				if _, err := time.LoadLocation(strings.TrimSpace(rows[0].Value)); err == nil {
					tz = strings.TrimSpace(rows[0].Value)
				}
			}
		}
		if loc, err := time.LoadLocation(tz); err == nil {
			nowInTz := time.Now().In(loc)
			// Cutover gate: before 2026-10-25 use legacy 1-24, after use rolling 25
			cutoverStr := "2026-10-25"
			if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=eq.earning_cycle_cutover_date&select=value"); err == nil && len(raw) > 0 {
				var rows []struct{ Value string `json:"value"` }
				if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 && strings.TrimSpace(rows[0].Value) != "" {
					cutoverStr = strings.TrimSpace(rows[0].Value)
				}
			}
			cutoverDate, _ := time.Parse("2006-01-02", cutoverStr)
			nowDate := time.Date(nowInTz.Year(), nowInTz.Month(), nowInTz.Day(), 0, 0, 0, 0, loc)
			var psTime, peTime time.Time
			if !cutoverDate.IsZero() && nowDate.Before(cutoverDate) {
				// Legacy 1-24
				startDay, endDay := 1, 24
				if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=in.(target_earning_start_day,target_earning_end_day)&select=key,value"); err == nil && len(raw) > 0 {
					var rows []struct{ Key string `json:"key"`; Value string `json:"value"` }
					if err := json.Unmarshal(raw, &rows); err == nil {
						for _, r := range rows {
							var n int
							if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil && n >= 1 && n <= 31 {
								if r.Key == "target_earning_start_day" {
									startDay = n
								} else if r.Key == "target_earning_end_day" {
									endDay = n
								}
							}
						}
					}
				}
				psTime = time.Date(nowInTz.Year(), nowInTz.Month(), startDay, 0, 0, 0, 0, loc)
				peTime = time.Date(nowInTz.Year(), nowInTz.Month(), endDay+1, 0, 0, 0, 0, loc)
			} else {
				anchor := 25
				if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=in.(earning_cycle_anchor_day,target_earning_start_day)&select=key,value"); err == nil && len(raw) > 0 {
					var rows []struct{ Key string `json:"key"`; Value string `json:"value"` }
					if err := json.Unmarshal(raw, &rows); err == nil {
						for _, r := range rows {
							var n int
							if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil && n >= 1 && n <= 31 {
								if r.Key == "earning_cycle_anchor_day" {
									anchor = n
								}
							}
						}
						found := false
						for _, r := range rows {
							if r.Key == "earning_cycle_anchor_day" {
								found = true
								break
							}
						}
						if !found {
							for _, r := range rows {
								var n int
								if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil && n >= 1 && n <= 31 && r.Key == "target_earning_start_day" {
									anchor = n
								}
							}
						}
					}
				}
				if nowInTz.Day() >= anchor {
					psTime = time.Date(nowInTz.Year(), nowInTz.Month(), anchor, 0, 0, 0, 0, loc)
					if nowInTz.Month() == 12 {
						peTime = time.Date(nowInTz.Year()+1, 1, anchor, 0, 0, 0, 0, loc)
					} else {
						peTime = time.Date(nowInTz.Year(), nowInTz.Month()+1, anchor, 0, 0, 0, 0, loc)
					}
				} else {
					if nowInTz.Month() == 1 {
						psTime = time.Date(nowInTz.Year()-1, 12, anchor, 0, 0, 0, 0, loc)
					} else {
						psTime = time.Date(nowInTz.Year(), nowInTz.Month()-1, anchor, 0, 0, 0, 0, loc)
					}
					peTime = time.Date(nowInTz.Year(), nowInTz.Month(), anchor, 0, 0, 0, 0, loc)
				}
			}
			ps := psTime.UTC().Format("2006-01-02")
			pe := peTime.UTC().Format("2006-01-02")
			_, _ = a.client.Mutate(ctx, http.MethodPost, "odyssey_member_monthly_targets", map[string]any{
				"user_uid":     uid,
				"period_start": ps,
				"period_end":   pe,
				"target":       *targetVal,
				"created_at":   now.Format(time.RFC3339),
			}, "")
		}
	}

	shared.WriteJSON(w, http.StatusCreated, MemberView{
		UID:               uid,
		FamilyID:          familyID,
		ExplorerName:      req.ExplorerName,
		Username:          req.Username,
		Role:              role,
		IsActive:          true,
		Level:             1,
		XP:                0,
		Coins:             0,
		MonthlyCoinTarget: targetVal,
		EarnedThisPeriod:  0,
		CreatedAt:         now,
	})
}

func (a *API) HandleUpdateMember(w http.ResponseWriter, r *http.Request, targetUID string) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	// Tenant check: target profile must belong to admin's family
	targetRaw, err := a.client.Get(ctx, "odyssey_user_profiles", fmt.Sprintf("uid=eq.%s", targetUID))
	if err != nil || len(targetRaw) == 0 {
		shared.WriteJSONError(w, "anggota tidak ditemukan", http.StatusNotFound)
		return
	}

	type ProfileCheck struct {
		UID      string `json:"uid"`
		FamilyID string `json:"family_id"`
	}
	var checks []ProfileCheck
	_ = json.Unmarshal(targetRaw, &checks)
	if len(checks) > 0 && claims.FamilyID != "" && checks[0].FamilyID != "" && checks[0].FamilyID != claims.FamilyID {
		shared.WriteJSONError(w, "akses ditolak: anggota bukan bagian dari keluarga Anda", http.StatusForbidden)
		return
	}

	var req UpdateMemberRequest
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid json body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.MonthlyCoinTarget != nil {
		if *req.MonthlyCoinTarget < 1 || *req.MonthlyCoinTarget > 10000 {
			shared.WriteJSONError(w, "target koin bulanan harus 1..10000", http.StatusBadRequest)
			return
		}
	}
	profPatch := map[string]any{}
	if req.ExplorerName != nil {
		name := strings.TrimSpace(*req.ExplorerName)
		if name == "" {
			shared.WriteJSONError(w, "nama anggota tidak boleh kosong", http.StatusBadRequest)
			return
		}
		profPatch["explorer_name"] = name
	}
	if req.Role != nil {
		role := strings.ToUpper(strings.TrimSpace(*req.Role))
		if role != "MEMBER" && role != "ADMIN" && role != "SEEKER" && role != "GUIDE" {
			shared.WriteJSONError(w, "role tidak valid", http.StatusBadRequest)
			return
		}
		profPatch["role"] = role
	}
	if req.IsActive != nil {
		profPatch["is_active"] = *req.IsActive
	}
	if req.MonthlyCoinTarget != nil {
		profPatch["monthly_coin_target"] = *req.MonthlyCoinTarget
	}
	if req.ResetDevice != nil && *req.ResetDevice {
		profPatch["device_id"] = nil
		profPatch["device_bound_at"] = nil
		_, _ = a.client.RPC(ctx, "odyssey_admin_reset_device", map[string]any{"p_target_uid": targetUID})
	}

	// Handle password reset if provided
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		pass := strings.TrimSpace(*req.Password)
		if len(pass) < 6 {
			shared.WriteJSONError(w, "password baru minimal 6 karakter", http.StatusBadRequest)
			return
		}
		profPatch["must_change_password"] = true

		hash, err := a.hasher.Hash(pass)
		if err != nil {
			shared.WriteJSONError(w, "gagal memproses password", http.StatusInternalServerError)
			return
		}
		localPatch := map[string]any{
			"password_hash": hash,
			"updated_at":    time.Now().UTC().Format(time.RFC3339),
		}
		_, err = a.client.Mutate(ctx, http.MethodPatch, "odyssey_local_users", localPatch, fmt.Sprintf("profile_uid=eq.%s", targetUID))
		if err != nil {
			shared.WriteJSONError(w, "gagal update password anggota: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(profPatch) > 0 {
		profPatch["updated_at"] = time.Now().UTC().Format(time.RFC3339)
		_, err := a.client.Mutate(ctx, http.MethodPatch, "odyssey_user_profiles", profPatch, fmt.Sprintf("uid=eq.%s", targetUID))
		if err != nil {
			shared.WriteJSONError(w, "gagal update profil anggota: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if req.MonthlyCoinTarget != nil {
		tz := shared.LoadConfig().Timezone
		if tz == "" {
			tz = shared.DefaultTimezone
		}
		if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=eq.timezone&select=value"); err == nil && len(raw) > 0 {
			var rows []struct{ Value string `json:"value"` }
			if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 && strings.TrimSpace(rows[0].Value) != "" {
				if _, err := time.LoadLocation(strings.TrimSpace(rows[0].Value)); err == nil {
					tz = strings.TrimSpace(rows[0].Value)
				}
			}
		}
		if loc, err := time.LoadLocation(tz); err == nil {
			now := time.Now().In(loc)
			cutoverStr := "2026-10-25"
			if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=eq.earning_cycle_cutover_date&select=value"); err == nil && len(raw) > 0 {
				var rows []struct{ Value string `json:"value"` }
				if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 && strings.TrimSpace(rows[0].Value) != "" {
					cutoverStr = strings.TrimSpace(rows[0].Value)
				}
			}
			cutoverDate, _ := time.Parse("2006-01-02", cutoverStr)
			nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
			var psTime, peTime time.Time
			if !cutoverDate.IsZero() && nowDate.Before(cutoverDate) {
				startDay, endDay := 1, 24
				if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=in.(target_earning_start_day,target_earning_end_day)&select=key,value"); err == nil && len(raw) > 0 {
					var rows []struct{ Key string `json:"key"`; Value string `json:"value"` }
					if err := json.Unmarshal(raw, &rows); err == nil {
						for _, r := range rows {
							var n int
							if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil && n >= 1 && n <= 31 {
								if r.Key == "target_earning_start_day" {
									startDay = n
								} else if r.Key == "target_earning_end_day" {
									endDay = n
								}
							}
						}
					}
				}
				psTime = time.Date(now.Year(), now.Month(), startDay, 0, 0, 0, 0, loc)
				peTime = time.Date(now.Year(), now.Month(), endDay+1, 0, 0, 0, 0, loc)
			} else {
				anchor := 25
				if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=in.(earning_cycle_anchor_day,target_earning_start_day)&select=key,value"); err == nil && len(raw) > 0 {
					var rows []struct{ Key string `json:"key"`; Value string `json:"value"` }
					if err := json.Unmarshal(raw, &rows); err == nil {
						for _, r := range rows {
							var n int
							if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil && n >= 1 && n <= 31 {
								if r.Key == "earning_cycle_anchor_day" {
									anchor = n
								}
							}
						}
						found := false
						for _, r := range rows {
							if r.Key == "earning_cycle_anchor_day" {
								found = true
								break
							}
						}
						if !found {
							for _, r := range rows {
								var n int
								if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil && n >= 1 && n <= 31 && r.Key == "target_earning_start_day" {
									anchor = n
								}
							}
						}
					}
				}
				if now.Day() >= anchor {
					psTime = time.Date(now.Year(), now.Month(), anchor, 0, 0, 0, 0, loc)
					if now.Month() == 12 {
						peTime = time.Date(now.Year()+1, 1, anchor, 0, 0, 0, 0, loc)
					} else {
						peTime = time.Date(now.Year(), now.Month()+1, anchor, 0, 0, 0, 0, loc)
					}
				} else {
					if now.Month() == 1 {
						psTime = time.Date(now.Year()-1, 12, anchor, 0, 0, 0, 0, loc)
					} else {
						psTime = time.Date(now.Year(), now.Month()-1, anchor, 0, 0, 0, 0, loc)
					}
					peTime = time.Date(now.Year(), now.Month(), anchor, 0, 0, 0, 0, loc)
				}
			}
			ps := psTime.UTC().Format("2006-01-02")
			pe := peTime.UTC().Format("2006-01-02")
			_, _ = a.client.Mutate(ctx, http.MethodPost, "odyssey_member_monthly_targets", map[string]any{
				"user_uid": targetUID, "period_start": ps, "period_end": pe, "target": *req.MonthlyCoinTarget, "assigned_by": claims.UID, "created_at": time.Now().UTC().Format(time.RFC3339),
			}, "")
			_, _ = a.client.Mutate(ctx, http.MethodPatch, "odyssey_member_monthly_targets", map[string]any{"target": *req.MonthlyCoinTarget, "assigned_by": claims.UID}, fmt.Sprintf("user_uid=eq.%s&period_start=eq.%s", targetUID, ps))
		}
	}

	// Return reloaded member view
	updatedRaw, _ := a.client.Get(ctx, "odyssey_user_profiles", fmt.Sprintf("uid=eq.%s", targetUID))
	var updatedProfs []db.UserProfile
	_ = json.Unmarshal(updatedRaw, &updatedProfs)

	var username string
	localRaw, _ := a.client.Get(ctx, "odyssey_local_users", fmt.Sprintf("profile_uid=eq.%s", targetUID))
	var locals []struct {
		Username string `json:"username"`
	}
	_ = json.Unmarshal(localRaw, &locals)
	if len(locals) > 0 {
		username = locals[0].Username
	} else {
		username = targetUID
	}

	if len(updatedProfs) > 0 {
		up := updatedProfs[0]
		defaultTarget := getDefaultMonthlyTarget(ctx, a.client)
		tgt := up.MonthlyCoinTarget
		if tgt == nil {
			q := defaultTarget
			tgt = &q
		}
		earnedMap := getEarnedThisPeriod(ctx, a.client, []string{targetUID})
		shared.WriteJSON(w, http.StatusOK, MemberView{
			UID:               up.UID,
			FamilyID:          up.FamilyID,
			ExplorerName:      up.ExplorerName,
			Username:          username,
			Role:              up.Role,
			IsActive:          up.IsActive,
			Level:             up.Level,
			XP:                up.XP,
			Coins:             up.Coins,
			MonthlyCoinTarget: tgt,
			EarnedThisPeriod:  earnedMap[targetUID],
			CreatedAt:         up.CreatedAt,
		})
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func generateTemporaryPassword() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%*"
	const length = 14
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("generate temporary password: %w", err)
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

func (a *API) HandleResetPassword(w http.ResponseWriter, r *http.Request, targetUID string) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	// Tenant check: target profile must belong to admin's family
	targetRaw, err := a.client.Get(ctx, "odyssey_user_profiles", fmt.Sprintf("uid=eq.%s", targetUID))
	if err != nil || len(targetRaw) == 0 || strings.TrimSpace(string(targetRaw)) == "[]" {
		shared.WriteJSONError(w, "anggota tidak ditemukan", http.StatusNotFound)
		return
	}
	type ProfileCheck struct {
		UID      string `json:"uid"`
		FamilyID string `json:"family_id"`
	}
	var checks []ProfileCheck
	_ = json.Unmarshal(targetRaw, &checks)
	if len(checks) == 0 {
		shared.WriteJSONError(w, "anggota tidak ditemukan", http.StatusNotFound)
		return
	}
	if claims.FamilyID != "" && checks[0].FamilyID != "" && checks[0].FamilyID != claims.FamilyID {
		shared.WriteJSONError(w, "anggota tidak ditemukan", http.StatusNotFound)
		return
	}

	// Generate cryptographically secure temporary password
	tempPassword, err := generateTemporaryPassword()
	if err != nil {
		shared.WriteJSONError(w, "gagal membuat temporary password", http.StatusInternalServerError)
		return
	}

	hash, err := a.hasher.Hash(tempPassword)
	if err != nil {
		shared.WriteJSONError(w, "gagal memproses password", http.StatusInternalServerError)
		return
	}

	// Update password hash in local users table
	localPatch := map[string]any{
		"password_hash": hash,
		"updated_at":    time.Now().UTC().Format(time.RFC3339),
	}
	_, err = a.client.Mutate(ctx, http.MethodPatch, "odyssey_local_users", localPatch, fmt.Sprintf("profile_uid=eq.%s", targetUID))
	if err != nil {
		shared.WriteJSONError(w, "gagal reset password anggota: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set force change flag on profile
	profPatch := map[string]any{
		"must_change_password": true,
		"updated_at":           time.Now().UTC().Format(time.RFC3339),
	}
	_, err = a.client.Mutate(ctx, http.MethodPatch, "odyssey_user_profiles", profPatch, fmt.Sprintf("uid=eq.%s", targetUID))
	if err != nil {
		shared.WriteJSONError(w, "gagal memperbarui flag password: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]string{
		"temporary_password": tempPassword,
	})
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	if path == "/api/admin/members" {
		if r.Method == http.MethodGet {
			a.HandleListMembers(w, r)
			return
		}
		if r.Method == http.MethodPost {
			a.HandleCreateMember(w, r)
			return
		}
	}

	if strings.HasPrefix(path, "/api/admin/members/") {
		trimmed := strings.TrimPrefix(path, "/api/admin/members/")
		// Handle reset-password sub-route: /api/admin/members/{uid}/reset-password
		if strings.HasSuffix(trimmed, "/reset-password") {
			targetUID := strings.TrimSuffix(trimmed, "/reset-password")
			targetUID = strings.TrimSuffix(targetUID, "/")
			if targetUID != "" {
				a.HandleResetPassword(w, r, targetUID)
				return
			}
		}
		targetUID := trimmed
		if targetUID != "" && !strings.Contains(targetUID, "/") && r.Method == http.MethodPatch {
			a.HandleUpdateMember(w, r, targetUID)
			return
		}
	}

	http.NotFound(w, r)
}
