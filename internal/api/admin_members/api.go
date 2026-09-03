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
	hasher auth.PasswordHasher
}

func NewAPI(client db.SupabaseClient) *API {
	return &API{
		client: client,
		hasher: auth.NewBcryptHasher(),
	}
}

type MemberView struct {
	UID                        string     `json:"uid"`
	FamilyID                   string     `json:"family_id"`
	ExplorerName               string     `json:"explorer_name"`
	Username                   string     `json:"username"`
	Role                       string     `json:"role"`
	IsActive                   bool       `json:"is_active"`
	Level                      int        `json:"level"`
	XP                         int64      `json:"xp"`
	Coins                      int64      `json:"coins"`
	MonthlyCoinTarget          *int       `json:"monthly_coin_target,omitempty"`
	EarnedThisPeriod           int        `json:"earned_this_period,omitempty"`
	BlockedAt                  *time.Time `json:"blocked_at,omitempty"`
	BlockedBy                  *string    `json:"blocked_by,omitempty"`
	BlockReason                *string    `json:"block_reason,omitempty"`
	CurrentCycleStart          string     `json:"current_cycle_start,omitempty"`
	CurrentCycleEnd            string     `json:"current_cycle_end,omitempty"`
	LastCompletedAt            *time.Time `json:"last_completed_at,omitempty"`
	LastCompletedDate          *string    `json:"last_completed_date,omitempty"`
	CompletedTasksCurrentCycle int        `json:"completed_tasks_current_cycle"`
	InactiveDays               *int       `json:"inactive_days,omitempty"`
	HasCurrentCycleActivity    bool       `json:"has_current_cycle_activity"`
	InactivityStatus           string     `json:"inactivity_status"`
	PayoutFrequency            string     `json:"payout_frequency,omitempty"`
	MinimumWithdrawalCoins     *int       `json:"minimum_withdrawal_coins,omitempty"`
	PayoutConfigSource         string     `json:"payout_config_source,omitempty"`
	CreatedAt                  time.Time  `json:"created_at"`
}

type CreateMemberRequest struct {
	Username               string  `json:"username"`
	Password               string  `json:"password"`
	ExplorerName           string  `json:"explorer_name"`
	Role                   string  `json:"role,omitempty"`
	MonthlyCoinTarget      *int    `json:"monthly_coin_target,omitempty"`
	PayoutFrequency        *string `json:"payout_frequency,omitempty"`
	MinimumWithdrawalCoins *int    `json:"minimum_withdrawal_coins,omitempty"`
	PayoutWeekday          *int    `json:"payout_weekday,omitempty"`
	PayoutMonthStartDay    *int    `json:"payout_month_start_day,omitempty"`
	PayoutMonthEndDay      *int    `json:"payout_month_end_day,omitempty"`
}

type UpdateMemberRequest struct {
	ExplorerName           *string `json:"explorer_name,omitempty"`
	Role                   *string `json:"role,omitempty"`
	IsActive               *bool   `json:"is_active,omitempty"`
	Password               *string `json:"password,omitempty"`
	ResetDevice            *bool   `json:"reset_device,omitempty"`
	MonthlyCoinTarget      *int    `json:"monthly_coin_target,omitempty"`
	PayoutFrequency        *string `json:"payout_frequency,omitempty"`
	MinimumWithdrawalCoins *int    `json:"minimum_withdrawal_coins,omitempty"`
	PayoutWeekday          *int    `json:"payout_weekday,omitempty"`
	PayoutMonthStartDay    *int    `json:"payout_month_start_day,omitempty"`
	PayoutMonthEndDay      *int    `json:"payout_month_end_day,omitempty"`
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
		var rows []struct {
			Value string `json:"value"`
		}
		if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 {
			var n int
			if _, err := fmt.Sscanf(rows[0].Value, "%d", &n); err == nil && n >= 0 && n <= 10000 {
				return n
			}
		}
	}
	return shared.DefaultMonthlyCoinTarget
}

func getAutoBlockThreshold(ctx context.Context, client db.SupabaseClient) int {
	for _, key := range []string{"auto_block_inactivity_days", "AUTO_BLOCK_INACTIVITY_DAYS"} {
		raw, err := client.Get(ctx, "odyssey_system_config", fmt.Sprintf("key=eq.%s&select=value", key))
		if err == nil && len(raw) > 0 {
			var rows []struct {
				Value string `json:"value"`
			}
			if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 {
				v := strings.TrimSpace(rows[0].Value)
				if v != "" {
					var n int
					if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
						if n <= 0 {
							return 0
						}
						if n > 365 {
							return shared.DefaultAutoBlockInactivityDays
						}
						return n
					}
				}
			}
		}
	}
	return shared.DefaultAutoBlockInactivityDays
}

func upsertPayoutConfig(ctx context.Context, client db.SupabaseClient, uid string, freq *string, min *int, wd *int, ms *int, me *int) error {
	// Resolve effective values: use provided or system defaults for required fields
	effFreq := payout.DefaultFrequency
	if freq != nil && payout.IsValidFrequency(*freq) {
		effFreq = payout.NormalizeFrequency(*freq)
	} else if freq != nil {
		return fmt.Errorf("invalid frequency")
	} else {
		// If freq not provided, try to keep existing or use default
		if existing, err := client.Get(ctx, "odyssey_user_payout_config", "user_uid=eq."+uid+"&select=payout_frequency"); err == nil && len(existing) > 2 {
			var rows []struct {
				PayoutFrequency string `json:"payout_frequency"`
			}
			if err := json.Unmarshal(existing, &rows); err == nil && len(rows) > 0 && payout.IsValidFrequency(rows[0].PayoutFrequency) {
				effFreq = payout.NormalizeFrequency(rows[0].PayoutFrequency)
			}
		}
	}
	effMin := payout.DefaultMinimumWithdrawal
	if min != nil {
		effMin = *min
	} else {
		// try existing
		if existing, err := client.Get(ctx, "odyssey_user_payout_config", "user_uid=eq."+uid+"&select=minimum_withdrawal_coins"); err == nil && len(existing) > 2 {
			var rows []struct {
				Minimum int `json:"minimum_withdrawal_coins"`
			}
			if err := json.Unmarshal(existing, &rows); err == nil && len(rows) > 0 && rows[0].Minimum > 0 {
				effMin = rows[0].Minimum
			}
		} else {
			// system default
			effMin = payout.GetSystemMinimumWithdrawal(ctx, client)
		}
	}
	// For monthly, ensure start/end present if freq is monthly
	payload := map[string]any{
		"user_uid":                 uid,
		"payout_frequency":         string(effFreq),
		"minimum_withdrawal_coins": effMin,
		"enabled":                  true,
		"updated_at":               time.Now().UTC().Format(time.RFC3339),
	}
	// Weekday
	if effFreq == payout.FrequencyWeekly {
		effWd := payout.DefaultWeeklyWeekday
		if wd != nil {
			effWd = *wd
		} else {
			if existing, err := client.Get(ctx, "odyssey_user_payout_config", "user_uid=eq."+uid+"&select=payout_weekday"); err == nil && len(existing) > 2 {
				var rows []struct {
					Wd *int `json:"payout_weekday"`
				}
				if err := json.Unmarshal(existing, &rows); err == nil && len(rows) > 0 && rows[0].Wd != nil {
					effWd = *rows[0].Wd
				}
			} else {
				// try system default weekday
				if raw, err := client.Get(ctx, "odyssey_system_config", "key=eq.default_payout_weekday&select=value"); err == nil {
					var r2 []struct{ Value string `json:"value"` }
					if err := json.Unmarshal(raw, &r2); err == nil && len(r2) > 0 {
						if v, err := strconv.Atoi(strings.TrimSpace(r2[0].Value)); err == nil && v >= 0 && v <= 6 {
							effWd = v
						}
					}
				}
			}
		}
		payload["payout_weekday"] = effWd
	} else if wd != nil {
		payload["payout_weekday"] = *wd
	}
	// Monthly window
	if effFreq == payout.FrequencyMonthly {
		effMs := 24
		effMe := 26
		if ms != nil {
			effMs = *ms
		} else {
			if raw, err := client.Get(ctx, "odyssey_system_config", "key=eq.redemption_start_day&select=value"); err == nil {
				var r2 []struct{ Value string `json:"value"` }
				if err := json.Unmarshal(raw, &r2); err == nil && len(r2) > 0 {
					if v, err := strconv.Atoi(strings.TrimSpace(r2[0].Value)); err == nil {
						effMs = v
					}
				}
			}
		}
		if me != nil {
			effMe = *me
		} else {
			if raw, err := client.Get(ctx, "odyssey_system_config", "key=eq.redemption_end_day&select=value"); err == nil {
				var r2 []struct{ Value string `json:"value"` }
				if err := json.Unmarshal(raw, &r2); err == nil && len(r2) > 0 {
					if v, err := strconv.Atoi(strings.TrimSpace(r2[0].Value)); err == nil {
						effMe = v
					}
				}
			}
		}
		if ms != nil || me != nil || true {
			payload["payout_month_start_day"] = effMs
			payload["payout_month_end_day"] = effMe
		}
	} else {
		if ms != nil {
			payload["payout_month_start_day"] = *ms
		}
		if me != nil {
			payload["payout_month_end_day"] = *me
		}
	}
	_, err := client.MutateAtomic(ctx, http.MethodPost, "odyssey_user_payout_config", payload, "", "resolution=merge-duplicates")
	if err != nil {
		_, err = client.Mutate(ctx, http.MethodPatch, "odyssey_user_payout_config", payload, "user_uid=eq."+uid)
		if err != nil {
			// try plain insert
			_, err = client.Mutate(ctx, http.MethodPost, "odyssey_user_payout_config", payload, "")
		}
	}
	return err
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
		var rows []struct {
			Value string `json:"value"`
		}
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
	startDay := 1
	endDay := 24
	if raw, err := client.Get(ctx, "odyssey_system_config", "key=in.(target_earning_start_day,target_earning_end_day)&select=key,value"); err == nil && len(raw) > 0 {
		var rows []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal(raw, &rows); err == nil {
			for _, r := range rows {
				var n int
				if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil {
					if r.Key == "target_earning_start_day" && n >= 1 && n <= 31 {
						startDay = n
					}
					if r.Key == "target_earning_end_day" && n >= 1 && n <= 31 {
						endDay = n
					}
				}
			}
		}
	}
	periodStart := time.Date(nowInTz.Year(), nowInTz.Month(), startDay, 0, 0, 0, 0, loc).UTC().Format(time.RFC3339)
	periodEnd := time.Date(nowInTz.Year(), nowInTz.Month(), endDay+1, 0, 0, 0, 0, loc).UTC().Format(time.RFC3339)
	params := fmt.Sprintf("user_uid=in.(%s)&type=eq.TASK_REWARD&created_at=gte.%s&created_at=lt.%s&select=user_uid,amount", strings.Join(uids, ","), periodStart, periodEnd)
	raw, err := client.Get(ctx, "odyssey_coin_transactions", params)
	if err != nil || len(raw) == 0 {
		return map[string]int{}
	}
	var rows []struct {
		UserUID string `json:"user_uid"`
		Amount  int    `json:"amount"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return map[string]int{}
	}
	m := make(map[string]int, len(uids))
	for _, r := range rows {
		m[r.UserUID] += r.Amount
	}
	return m
}

func getCurrentCycleBounds(ctx context.Context, client db.SupabaseClient) (periodStart time.Time, periodEnd time.Time, loc *time.Location, today time.Time, periodStartStr, periodEndStr string) {
	tz := shared.LoadConfig().Timezone
	if tz == "" {
		tz = shared.DefaultTimezone
	}
	if raw, err := client.Get(ctx, "odyssey_system_config", "key=eq.timezone&select=value"); err == nil && len(raw) > 0 {
		var rows []struct {
			Value string `json:"value"`
		}
		if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 && strings.TrimSpace(rows[0].Value) != "" {
			if _, err := time.LoadLocation(strings.TrimSpace(rows[0].Value)); err == nil {
				tz = strings.TrimSpace(rows[0].Value)
			}
		}
	}
	loc, _ = time.LoadLocation(tz)
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
		tz = shared.DefaultTimezone
	}
	today = time.Now().In(loc)
	nowInTz := today
	startDay, endDay := 1, 24
	if raw, err := client.Get(ctx, "odyssey_system_config", "key=in.(target_earning_start_day,target_earning_end_day)&select=key,value"); err == nil && len(raw) > 0 {
		var rows []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal(raw, &rows); err == nil {
			for _, r := range rows {
				var n int
				if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil {
					if r.Key == "target_earning_start_day" && n >= 1 && n <= 31 {
						startDay = n
					}
					if r.Key == "target_earning_end_day" && n >= 1 && n <= 31 {
						endDay = n
					}
				}
			}
		}
	}
	periodStart = time.Date(nowInTz.Year(), nowInTz.Month(), startDay, 0, 0, 0, 0, loc)
	periodEnd = time.Date(nowInTz.Year(), nowInTz.Month(), endDay+1, 0, 0, 0, 0, loc)
	periodStartStr = periodStart.Format("2006-01-02")
	// periodEnd is exclusive, display inclusive end as periodEnd-1day
	periodEndStr = periodEnd.AddDate(0, 0, -1).Format("2006-01-02")
	return
}

type inactivityInfo struct {
	LastCompletedAt   *time.Time
	LastCompletedDate *string
	Count             int
	InactiveDays      *int
	HasActivity       bool
	Status            string
	CurrentCycleStart string
	CurrentCycleEnd   string
}

func getInactivityTracking(ctx context.Context, client db.SupabaseClient, uids []string, isActiveMap map[string]bool) map[string]inactivityInfo {
	result := make(map[string]inactivityInfo, len(uids))
	if len(uids) == 0 {
		return result
	}
	periodStart, periodEnd, loc, today, periodStartStr, periodEndStr := getCurrentCycleBounds(ctx, client)
	threshold := getAutoBlockThreshold(ctx, client)
	// 0 = disabled for display (never INACTIVE); keep as is, no fallback to 5
	if threshold < 0 {
		threshold = shared.DefaultAutoBlockInactivityDays
	}
	// Default for all uids
	for _, uid := range uids {
		result[uid] = inactivityInfo{
			CurrentCycleStart: periodStartStr,
			CurrentCycleEnd:   periodEndStr,
		}
	}
	// Fetch APPROVED submissions for these users
	params := fmt.Sprintf("user_uid=in.(%s)&status=eq.APPROVED&select=user_uid,created_at,reviewed_at", strings.Join(uids, ","))
	raw, err := client.Get(ctx, "odyssey_task_submissions", params)
	if err != nil || len(raw) == 0 || string(raw) == "[]" {
		// No completions in current cycle for anyone -> mark NO_ACTIVITY
		for uid, info := range result {
			if !isActiveMap[uid] {
				info.Status = "BLOCKED"
			} else {
				info.Status = "NO_ACTIVITY_THIS_CYCLE"
			}
			result[uid] = info
		}
		return result
	}
	var rows []struct {
		UserUID    string  `json:"user_uid"`
		CreatedAt  string  `json:"created_at"`
		ReviewedAt *string `json:"reviewed_at"`
	}
	_ = json.Unmarshal(raw, &rows)
	// Group by user, filter to current cycle
	tmp := make(map[string][]time.Time)
	countMap := make(map[string]int)
	lastMap := make(map[string]time.Time)
	for _, r := range rows {
		tsStr := r.CreatedAt
		if r.ReviewedAt != nil && strings.TrimSpace(*r.ReviewedAt) != "" {
			tsStr = *r.ReviewedAt
		}
		// Parse time (supabase returns UTC)
		var t time.Time
		var err error
		// Try RFC3339
		t, err = time.Parse(time.RFC3339, tsStr)
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05", tsStr)
			if err != nil {
				continue
			}
		}
		// Convert to loc date for cycle check
		localDateStr := t.In(loc).Format("2006-01-02")
		localDate, _ := time.ParseInLocation("2006-01-02", localDateStr, loc)
		psStr := periodStart.Format("2006-01-02")
		peStr := periodEnd.Format("2006-01-02")
		ps, _ := time.ParseInLocation("2006-01-02", psStr, loc)
		pe, _ := time.ParseInLocation("2006-01-02", peStr, loc)
		if localDate.Before(ps) || !localDate.Before(pe) {
			continue // outside current cycle
		}
		tmp[r.UserUID] = append(tmp[r.UserUID], t)
		countMap[r.UserUID]++
		if existing, ok := lastMap[r.UserUID]; !ok || t.After(existing) {
			lastMap[r.UserUID] = t
		}
	}
	todayDateStr := today.Format("2006-01-02")
	todayDate, _ := time.ParseInLocation("2006-01-02", todayDateStr, loc)
	for uid, info := range result {
		if !isActiveMap[uid] {
			info.Status = "BLOCKED"
			result[uid] = info
			continue
		}
		last, ok := lastMap[uid]
		if !ok {
			info.Status = "NO_ACTIVITY_THIS_CYCLE"
			info.HasActivity = false
			info.Count = 0
			result[uid] = info
			continue
		}
		// Has activity
		info.HasActivity = true
		info.Count = countMap[uid]
		// LastCompletedAt in UTC, LastCompletedDate in loc
		cpy := last
		info.LastCompletedAt = &cpy
		ldStr := last.In(loc).Format("2006-01-02")
		info.LastCompletedDate = &ldStr
		// Inactive days
		lastDateStr := last.In(loc).Format("2006-01-02")
		lastDate, _ := time.ParseInLocation("2006-01-02", lastDateStr, loc)
		days := int(todayDate.Sub(lastDate) / (24 * time.Hour))
		if days < 0 {
			days = 0
		}
		d := days
		info.InactiveDays = &d
		if threshold != 0 && days >= threshold {
			info.Status = "INACTIVE"
		} else {
			info.Status = "ACTIVE"
		}
		result[uid] = info
	}
	return result
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
		UID               string     `json:"uid"`
		FamilyID          string     `json:"family_id"`
		ExplorerName      string     `json:"explorer_name"`
		Role              string     `json:"role"`
		IsActive          bool       `json:"is_active"`
		Level             int        `json:"level"`
		XP                int64      `json:"xp"`
		Coins             int64      `json:"coins"`
		MonthlyCoinTarget *int       `json:"monthly_coin_target"`
		BlockedAt         *time.Time `json:"blocked_at"`
		BlockedBy         *string    `json:"blocked_by"`
		BlockReason       *string    `json:"block_reason"`
		CreatedAt         time.Time  `json:"created_at"`
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
	// Inactivity tracking (read-only, cycle-aware)
	isActiveMap := make(map[string]bool, len(profs))
	for _, p := range profs {
		isActiveMap[p.UID] = p.IsActive
	}
	trackingMap := getInactivityTracking(ctx, a.client, uids, isActiveMap)
	// Fetch per-user payout configs for list view (batch)
	payoutMap := make(map[string]payout.EffectivePayoutConfig)
	for _, uid := range uids {
		if eff, err := payout.GetEffectivePayoutConfig(ctx, a.client, uid, time.Now()); err == nil {
			payoutMap[uid] = eff
		}
	}
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
		tr := trackingMap[p.UID]
		pc := payoutMap[p.UID]
		minVal := pc.MinimumWithdrawalCoins
		pcMin := &minVal
		items[i] = MemberView{
			UID:                        p.UID,
			FamilyID:                   p.FamilyID,
			ExplorerName:               p.ExplorerName,
			Username:                   username,
			Role:                       p.Role,
			IsActive:                   p.IsActive,
			Level:                      p.Level,
			XP:                         p.XP,
			Coins:                      p.Coins,
			MonthlyCoinTarget:          tgt,
			EarnedThisPeriod:           earned,
			BlockedAt:                  p.BlockedAt,
			BlockedBy:                  p.BlockedBy,
			BlockReason:                p.BlockReason,
			CurrentCycleStart:          tr.CurrentCycleStart,
			CurrentCycleEnd:            tr.CurrentCycleEnd,
			LastCompletedAt:            tr.LastCompletedAt,
			LastCompletedDate:          tr.LastCompletedDate,
			CompletedTasksCurrentCycle: tr.Count,
			InactiveDays:               tr.InactiveDays,
			HasCurrentCycleActivity:    tr.HasActivity,
			InactivityStatus:           tr.Status,
			PayoutFrequency:            string(pc.Frequency),
			MinimumWithdrawalCoins:     pcMin,
			PayoutConfigSource:         pc.Source,
			CreatedAt:                  p.CreatedAt,
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
		if *req.MonthlyCoinTarget < 0 || *req.MonthlyCoinTarget > 10000 {
			shared.WriteJSONError(w, "target koin bulanan harus 0..10000", http.StatusBadRequest)
			return
		}
	}
	// Validate payout config fields if provided
	if req.PayoutFrequency != nil && !payout.IsValidFrequency(*req.PayoutFrequency) {
		shared.WriteJSONError(w, "payout_frequency tidak valid (THRESHOLD|WEEKLY|MONTHLY)", http.StatusBadRequest)
		return
	}
	if req.MinimumWithdrawalCoins != nil {
		if *req.MinimumWithdrawalCoins < 1 || *req.MinimumWithdrawalCoins > 100000 {
			shared.WriteJSONError(w, "minimum_withdrawal_coins harus 1..100000", http.StatusBadRequest)
			return
		}
		sysMin := payout.GetSystemMinimumWithdrawal(ctx, a.client)
		if *req.MinimumWithdrawalCoins < sysMin {
			shared.WriteJSONError(w, fmt.Sprintf("minimum_withdrawal_coins (%d) di bawah system minimum (%d)", *req.MinimumWithdrawalCoins, sysMin), http.StatusBadRequest)
			return
		}
	}
	if req.PayoutWeekday != nil && (*req.PayoutWeekday < 0 || *req.PayoutWeekday > 6) {
		shared.WriteJSONError(w, "payout_weekday harus 0..6", http.StatusBadRequest)
		return
	}
	if req.PayoutMonthStartDay != nil && (*req.PayoutMonthStartDay < 1 || *req.PayoutMonthStartDay > 31) {
		shared.WriteJSONError(w, "payout_month_start_day harus 1..31", http.StatusBadRequest)
		return
	}
	if req.PayoutMonthEndDay != nil && (*req.PayoutMonthEndDay < 1 || *req.PayoutMonthEndDay > 31) {
		shared.WriteJSONError(w, "payout_month_end_day harus 1..31", http.StatusBadRequest)
		return
	}
	if req.PayoutMonthStartDay != nil && req.PayoutMonthEndDay != nil && *req.PayoutMonthStartDay > *req.PayoutMonthEndDay {
		shared.WriteJSONError(w, "payout_month_start_day tidak boleh > end_day", http.StatusBadRequest)
		return
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
			var rows []struct {
				Value string `json:"value"`
			}
			if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 && strings.TrimSpace(rows[0].Value) != "" {
				if _, err := time.LoadLocation(strings.TrimSpace(rows[0].Value)); err == nil {
					tz = strings.TrimSpace(rows[0].Value)
				}
			}
		}
		if loc, err := time.LoadLocation(tz); err == nil {
			nowInTz := time.Now().In(loc)
			startDay := 1
			endDay := 24
			if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=in.(target_earning_start_day,target_earning_end_day)&select=key,value"); err == nil && len(raw) > 0 {
				var rows []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				}
				if err := json.Unmarshal(raw, &rows); err == nil {
					for _, r := range rows {
						var n int
						if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil {
							if r.Key == "target_earning_start_day" && n >= 1 && n <= 31 {
								startDay = n
							}
							if r.Key == "target_earning_end_day" && n >= 1 && n <= 31 {
								endDay = n
							}
						}
					}
				}
			}
			ps := time.Date(nowInTz.Year(), nowInTz.Month(), startDay, 0, 0, 0, 0, loc).UTC().Format("2006-01-02")
			pe := time.Date(nowInTz.Year(), nowInTz.Month(), endDay+1, 0, 0, 0, 0, loc).UTC().Format("2006-01-02")
			_, _ = a.client.Mutate(ctx, http.MethodPost, "odyssey_member_monthly_targets", map[string]any{
				"user_uid":     uid,
				"period_start": ps,
				"period_end":   pe,
				"target":       *targetVal,
				"created_at":   now.Format(time.RFC3339),
			}, "")
		}
	}
	// 7. Upsert payout config if any payout fields provided (best-effort)
	if req.PayoutFrequency != nil || req.MinimumWithdrawalCoins != nil || req.PayoutWeekday != nil || req.PayoutMonthStartDay != nil || req.PayoutMonthEndDay != nil {
		_ = upsertPayoutConfig(ctx, a.client, uid, req.PayoutFrequency, req.MinimumWithdrawalCoins, req.PayoutWeekday, req.PayoutMonthStartDay, req.PayoutMonthEndDay)
	}
	effPayout, _ := payout.GetEffectivePayoutConfig(ctx, a.client, uid, time.Now())
	minWd := effPayout.MinimumWithdrawalCoins

	shared.WriteJSON(w, http.StatusCreated, MemberView{
		UID:                    uid,
		FamilyID:               familyID,
		ExplorerName:           req.ExplorerName,
		Username:               req.Username,
		Role:                   role,
		IsActive:               true,
		Level:                  1,
		XP:                     0,
		Coins:                  0,
		MonthlyCoinTarget:      targetVal,
		EarnedThisPeriod:       0,
		PayoutFrequency:        string(effPayout.Frequency),
		MinimumWithdrawalCoins: &minWd,
		PayoutConfigSource:     effPayout.Source,
		CreatedAt:              now,
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
		if *req.MonthlyCoinTarget < 0 || *req.MonthlyCoinTarget > 10000 {
			shared.WriteJSONError(w, "target koin bulanan harus 0..10000", http.StatusBadRequest)
			return
		}
	}
	// Validate payout fields
	if req.PayoutFrequency != nil && !payout.IsValidFrequency(*req.PayoutFrequency) {
		shared.WriteJSONError(w, "payout_frequency tidak valid", http.StatusBadRequest)
		return
	}
	if req.MinimumWithdrawalCoins != nil {
		if *req.MinimumWithdrawalCoins < 1 || *req.MinimumWithdrawalCoins > 100000 {
			shared.WriteJSONError(w, "minimum_withdrawal_coins harus 1..100000", http.StatusBadRequest)
			return
		}
		sysMin := payout.GetSystemMinimumWithdrawal(ctx, a.client)
		if *req.MinimumWithdrawalCoins < sysMin {
			shared.WriteJSONError(w, fmt.Sprintf("minimum_withdrawal_coins (%d) di bawah system minimum (%d)", *req.MinimumWithdrawalCoins, sysMin), http.StatusBadRequest)
			return
		}
	}
	if req.PayoutWeekday != nil && (*req.PayoutWeekday < 0 || *req.PayoutWeekday > 6) {
		shared.WriteJSONError(w, "payout_weekday harus 0..6", http.StatusBadRequest)
		return
	}
	if req.PayoutMonthStartDay != nil && (*req.PayoutMonthStartDay < 1 || *req.PayoutMonthStartDay > 31) {
		shared.WriteJSONError(w, "payout_month_start_day harus 1..31", http.StatusBadRequest)
		return
	}
	if req.PayoutMonthEndDay != nil && (*req.PayoutMonthEndDay < 1 || *req.PayoutMonthEndDay > 31) {
		shared.WriteJSONError(w, "payout_month_end_day harus 1..31", http.StatusBadRequest)
		return
	}
	if req.PayoutMonthStartDay != nil && req.PayoutMonthEndDay != nil && *req.PayoutMonthStartDay > *req.PayoutMonthEndDay {
		shared.WriteJSONError(w, "payout_month_start_day tidak boleh > end_day", http.StatusBadRequest)
		return
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
		if !*req.IsActive {
			// Blocking via is_active toggle: set audit fields atomically if not already blocked
			profPatch["blocked_at"] = time.Now().UTC().Format(time.RFC3339)
			profPatch["blocked_by"] = claims.UID
			profPatch["block_reason"] = "manual block via admin"
		} else {
			profPatch["blocked_at"] = nil
			profPatch["blocked_by"] = nil
			profPatch["block_reason"] = nil
		}
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
			var rows []struct {
				Value string `json:"value"`
			}
			if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 && strings.TrimSpace(rows[0].Value) != "" {
				if _, err := time.LoadLocation(strings.TrimSpace(rows[0].Value)); err == nil {
					tz = strings.TrimSpace(rows[0].Value)
				}
			}
		}
		if loc, err := time.LoadLocation(tz); err == nil {
			now := time.Now().In(loc)
			startDay := 1
			endDay := 24
			if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=in.(target_earning_start_day,target_earning_end_day)&select=key,value"); err == nil && len(raw) > 0 {
				var rows []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				}
				if err := json.Unmarshal(raw, &rows); err == nil {
					for _, r := range rows {
						var n int
						if _, err := fmt.Sscanf(r.Value, "%d", &n); err == nil {
							if r.Key == "target_earning_start_day" && n >= 1 && n <= 31 {
								startDay = n
							}
							if r.Key == "target_earning_end_day" && n >= 1 && n <= 31 {
								endDay = n
							}
						}
					}
				}
			}
			ps := time.Date(now.Year(), now.Month(), startDay, 0, 0, 0, 0, loc).UTC().Format("2006-01-02")
			pe := time.Date(now.Year(), now.Month(), endDay+1, 0, 0, 0, 0, loc).UTC().Format("2006-01-02")
			_, _ = a.client.Mutate(ctx, http.MethodPost, "odyssey_member_monthly_targets", map[string]any{
				"user_uid": targetUID, "period_start": ps, "period_end": pe, "target": *req.MonthlyCoinTarget, "assigned_by": claims.UID, "created_at": time.Now().UTC().Format(time.RFC3339),
			}, "")
			_, _ = a.client.Mutate(ctx, http.MethodPatch, "odyssey_member_monthly_targets", map[string]any{"target": *req.MonthlyCoinTarget, "assigned_by": claims.UID}, fmt.Sprintf("user_uid=eq.%s&period_start=eq.%s", targetUID, ps))
		}
	}
	// Upsert payout config if any payout fields provided
	if req.PayoutFrequency != nil || req.MinimumWithdrawalCoins != nil || req.PayoutWeekday != nil || req.PayoutMonthStartDay != nil || req.PayoutMonthEndDay != nil {
		_ = upsertPayoutConfig(ctx, a.client, targetUID, req.PayoutFrequency, req.MinimumWithdrawalCoins, req.PayoutWeekday, req.PayoutMonthStartDay, req.PayoutMonthEndDay)
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
		effPayout, _ := payout.GetEffectivePayoutConfig(ctx, a.client, targetUID, time.Now())
		minWd := effPayout.MinimumWithdrawalCoins
		shared.WriteJSON(w, http.StatusOK, MemberView{
			UID:                    up.UID,
			FamilyID:               up.FamilyID,
			ExplorerName:           up.ExplorerName,
			Username:               username,
			Role:                   up.Role,
			IsActive:               up.IsActive,
			Level:                  up.Level,
			XP:                     up.XP,
			Coins:                  up.Coins,
			MonthlyCoinTarget:      tgt,
			EarnedThisPeriod:       earnedMap[targetUID],
			BlockedAt:              up.BlockedAt,
			BlockedBy:              up.BlockedBy,
			BlockReason:            up.BlockReason,
			PayoutFrequency:        string(effPayout.Frequency),
			MinimumWithdrawalCoins: &minWd,
			PayoutConfigSource:     effPayout.Source,
			CreatedAt:              up.CreatedAt,
		})
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *API) HandleBlockMember(w http.ResponseWriter, r *http.Request, targetUID string) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	targetRaw, err := a.client.Get(ctx, "odyssey_user_profiles", fmt.Sprintf("uid=eq.%s", targetUID))
	if err != nil || len(targetRaw) == 0 || strings.TrimSpace(string(targetRaw)) == "[]" {
		shared.WriteJSONError(w, "anggota tidak ditemukan", http.StatusNotFound)
		return
	}
	var checks []struct {
		UID      string `json:"uid"`
		FamilyID string `json:"family_id"`
		Role     string `json:"role"`
		IsActive bool   `json:"is_active"`
	}
	_ = json.Unmarshal(targetRaw, &checks)
	if len(checks) == 0 {
		shared.WriteJSONError(w, "anggota tidak ditemukan", http.StatusNotFound)
		return
	}
	if claims.FamilyID != "" && checks[0].FamilyID != "" && checks[0].FamilyID != claims.FamilyID {
		shared.WriteJSONError(w, "akses ditolak: anggota bukan bagian dari keluarga Anda", http.StatusForbidden)
		return
	}
	if checks[0].Role == "ADMIN" || checks[0].Role == "GUIDE" || checks[0].Role == "BUILDER" {
		shared.WriteJSONError(w, "akun admin tidak dapat diblokir", http.StatusBadRequest)
		return
	}
	if !checks[0].IsActive {
		shared.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "uid": targetUID, "already_blocked": true, "is_active": false})
		return
	}
	var req struct {
		Reason *string `json:"reason"`
	}
	_ = shared.ReadJSON(r, &req)
	reason := "manual block via admin"
	if req.Reason != nil && strings.TrimSpace(*req.Reason) != "" {
		reason = strings.TrimSpace(*req.Reason)
	}
	// Try RPC first for atomic audit
	if _, err := a.client.RPC(ctx, "odyssey_block_user", map[string]any{"p_target_uid": targetUID, "p_admin_uid": claims.UID, "p_reason": reason}); err == nil {
		shared.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "uid": targetUID, "is_active": false})
		return
	}
	// Fallback direct patch with conditional check is_active=true to avoid race
	patch := map[string]any{
		"is_active":    false,
		"blocked_at":   time.Now().UTC().Format(time.RFC3339),
		"blocked_by":   claims.UID,
		"block_reason": reason,
		"updated_at":   time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := a.client.Mutate(ctx, http.MethodPatch, "odyssey_user_profiles", patch, fmt.Sprintf("uid=eq.%s&is_active=eq.true", targetUID)); err != nil {
		shared.WriteJSONError(w, "gagal memblokir anggota: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "uid": targetUID, "is_active": false})
}

func (a *API) HandleUnblockMember(w http.ResponseWriter, r *http.Request, targetUID string) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	targetRaw, err := a.client.Get(ctx, "odyssey_user_profiles", fmt.Sprintf("uid=eq.%s", targetUID))
	if err != nil || len(targetRaw) == 0 || strings.TrimSpace(string(targetRaw)) == "[]" {
		shared.WriteJSONError(w, "anggota tidak ditemukan", http.StatusNotFound)
		return
	}
	var checks []struct {
		UID      string `json:"uid"`
		FamilyID string `json:"family_id"`
		IsActive bool   `json:"is_active"`
	}
	_ = json.Unmarshal(targetRaw, &checks)
	if len(checks) == 0 {
		shared.WriteJSONError(w, "anggota tidak ditemukan", http.StatusNotFound)
		return
	}
	if claims.FamilyID != "" && checks[0].FamilyID != "" && checks[0].FamilyID != claims.FamilyID {
		shared.WriteJSONError(w, "akses ditolak: anggota bukan bagian dari keluarga Anda", http.StatusForbidden)
		return
	}
	if checks[0].IsActive {
		shared.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "uid": targetUID, "already_active": true, "is_active": true})
		return
	}
	if _, err := a.client.RPC(ctx, "odyssey_unblock_user", map[string]any{"p_target_uid": targetUID, "p_admin_uid": claims.UID}); err == nil {
		shared.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "uid": targetUID, "is_active": true})
		return
	}
	patch := map[string]any{
		"is_active":    true,
		"blocked_at":   nil,
		"blocked_by":   nil,
		"block_reason": nil,
		"updated_at":   time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := a.client.Mutate(ctx, http.MethodPatch, "odyssey_user_profiles", patch, fmt.Sprintf("uid=eq.%s&is_active=eq.false", targetUID)); err != nil {
		shared.WriteJSONError(w, "gagal membuka blokir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "uid": targetUID, "is_active": true})
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
		// block / unblock sub-routes
		if strings.HasSuffix(trimmed, "/block") {
			targetUID := strings.TrimSuffix(trimmed, "/block")
			targetUID = strings.TrimSuffix(targetUID, "/")
			if targetUID != "" && r.Method == http.MethodPost {
				a.HandleBlockMember(w, r, targetUID)
				return
			}
		}
		if strings.HasSuffix(trimmed, "/unblock") {
			targetUID := strings.TrimSuffix(trimmed, "/unblock")
			targetUID = strings.TrimSuffix(targetUID, "/")
			if targetUID != "" && r.Method == http.MethodPost {
				a.HandleUnblockMember(w, r, targetUID)
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
