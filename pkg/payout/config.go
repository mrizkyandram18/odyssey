package payout

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"odyssey/pkg/db"
	"odyssey/pkg/shared"
)

type Frequency string

const (
	FrequencyThreshold Frequency = "THRESHOLD"
	FrequencyWeekly    Frequency = "WEEKLY"
	FrequencyMonthly   Frequency = "MONTHLY"
)

const (
	DefaultFrequency         = FrequencyMonthly
	DefaultMinimumWithdrawal = 500
	DefaultWeeklyWeekday     = 1 // Monday
)

type UserPayoutConfig struct {
	UserUID               string    `json:"user_uid"`
	PayoutFrequency       Frequency `json:"payout_frequency"`
	MinimumWithdrawalCoins int       `json:"minimum_withdrawal_coins"`
	PayoutWeekday          *int      `json:"payout_weekday,omitempty"`
	PayoutMonthStartDay    *int      `json:"payout_month_start_day,omitempty"`
	PayoutMonthEndDay      *int      `json:"payout_month_end_day,omitempty"`
	Enabled               bool      `json:"enabled"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type EffectivePayoutConfig struct {
	Frequency             Frequency `json:"frequency"`
	MinimumWithdrawalCoins int       `json:"minimum_withdrawal_coins"`
	PayoutWeekday         int       `json:"payout_weekday"`
	PayoutMonthStartDay   int       `json:"payout_month_start_day"`
	PayoutMonthEndDay     int       `json:"payout_month_end_day"`
	Source                string    `json:"source"` // "user" or "system"
}

func IsValidFrequency(s string) bool {
	switch Frequency(strings.ToUpper(strings.TrimSpace(s))) {
	case FrequencyThreshold, FrequencyWeekly, FrequencyMonthly:
		return true
	}
	return false
}

func NormalizeFrequency(s string) Frequency {
	switch Frequency(strings.ToUpper(strings.TrimSpace(s))) {
	case FrequencyThreshold:
		return FrequencyThreshold
	case FrequencyWeekly:
		return FrequencyWeekly
	case FrequencyMonthly:
		return FrequencyMonthly
	default:
		return DefaultFrequency
	}
}

// GetEffectivePayoutConfig resolves per-user config with system defaults fallback.
// Single source of truth for payout policy resolution.
func GetEffectivePayoutConfig(ctx context.Context, client db.SupabaseClient, userUID string, now time.Time) (EffectivePayoutConfig, error) {
	// Fetch system defaults
	sysMin := DefaultMinimumWithdrawal
	sysFreq := DefaultFrequency
	sysWeekday := DefaultWeeklyWeekday
	sysMonthStart := shared.DefaultRedemptionStartDay
	sysMonthEnd := shared.DefaultRedemptionEndDay
	sysTimezone := shared.DefaultTimezone

	raw, err := client.Get(ctx, "odyssey_system_config", "key=in.(default_payout_frequency,default_minimum_withdrawal_coins,default_payout_weekday,redemption_start_day,redemption_end_day,payout_day,timezone)&select=key,value")
	if err == nil && len(raw) > 0 {
		var rows []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if jerr := json.Unmarshal(raw, &rows); jerr == nil {
			for _, r := range rows {
				switch r.Key {
				case "default_payout_frequency":
					if IsValidFrequency(r.Value) {
						sysFreq = NormalizeFrequency(r.Value)
					}
				case "default_minimum_withdrawal_coins":
					if v, err := strconv.Atoi(strings.TrimSpace(r.Value)); err == nil && v > 0 && v <= 100000 {
						sysMin = v
					}
				case "default_payout_weekday":
					if v, err := strconv.Atoi(strings.TrimSpace(r.Value)); err == nil && v >= 0 && v <= 6 {
						sysWeekday = v
					}
				case "redemption_start_day":
					if v, err := strconv.Atoi(strings.TrimSpace(r.Value)); err == nil && v >= 1 && v <= 31 {
						sysMonthStart = v
					}
				case "redemption_end_day":
					if v, err := strconv.Atoi(strings.TrimSpace(r.Value)); err == nil && v >= 1 && v <= 31 {
						sysMonthEnd = v
					}
				case "payout_day":
					// payout_day is single day monthly; if monthly start/end not set, use payout_day as window single day
					_ = r.Value
				case "timezone":
					if v := strings.TrimSpace(r.Value); v != "" {
						if _, err := time.LoadLocation(v); err == nil {
							sysTimezone = v
						}
					}
				}
			}
		}
	}
	// Also try dedicated month window overrides
	raw2, err := client.Get(ctx, "odyssey_system_config", "key=in.(default_payout_month_start_day,default_payout_month_end_day)&select=key,value")
	if err == nil && len(raw2) > 0 {
		var rows []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if jerr := json.Unmarshal(raw2, &rows); jerr == nil {
			for _, r := range rows {
				switch r.Key {
				case "default_payout_month_start_day":
					if v, err := strconv.Atoi(strings.TrimSpace(r.Value)); err == nil && v >= 1 && v <= 31 {
						sysMonthStart = v
					}
				case "default_payout_month_end_day":
					if v, err := strconv.Atoi(strings.TrimSpace(r.Value)); err == nil && v >= 1 && v <= 31 {
						sysMonthEnd = v
					}
				}
			}
		}
	}

	_ = sysTimezone
	_ = now

	effective := EffectivePayoutConfig{
		Frequency:             sysFreq,
		MinimumWithdrawalCoins: sysMin,
		PayoutWeekday:         sysWeekday,
		PayoutMonthStartDay:   sysMonthStart,
		PayoutMonthEndDay:     sysMonthEnd,
		Source:                "system",
	}

	// Try per-user override
	if userUID != "" {
		rawUser, err := client.Get(ctx, "odyssey_user_payout_config", "user_uid=eq."+userUID+"&select=user_uid,payout_frequency,minimum_withdrawal_coins,payout_weekday,payout_month_start_day,payout_month_end_day,enabled")
		if err == nil && len(rawUser) > 0 && strings.TrimSpace(string(rawUser)) != "[]" {
			var rows []struct {
				UserUID               string  `json:"user_uid"`
				PayoutFrequency       string  `json:"payout_frequency"`
				MinimumWithdrawalCoins int     `json:"minimum_withdrawal_coins"`
				PayoutWeekday          *int    `json:"payout_weekday"`
				PayoutMonthStartDay    *int    `json:"payout_month_start_day"`
				PayoutMonthEndDay      *int    `json:"payout_month_end_day"`
				Enabled               *bool   `json:"enabled"`
			}
			if jerr := json.Unmarshal(rawUser, &rows); jerr == nil && len(rows) > 0 {
				r := rows[0]
				// If disabled, fall back to system
				enabled := true
				if r.Enabled != nil {
					enabled = *r.Enabled
				}
				if enabled {
					if IsValidFrequency(r.PayoutFrequency) {
						effective.Frequency = NormalizeFrequency(r.PayoutFrequency)
					}
					if r.MinimumWithdrawalCoins > 0 {
						effective.MinimumWithdrawalCoins = r.MinimumWithdrawalCoins
					}
					if r.PayoutWeekday != nil && *r.PayoutWeekday >= 0 && *r.PayoutWeekday <= 6 {
						effective.PayoutWeekday = *r.PayoutWeekday
					}
					if r.PayoutMonthStartDay != nil && *r.PayoutMonthStartDay >= 1 && *r.PayoutMonthStartDay <= 31 {
						effective.PayoutMonthStartDay = *r.PayoutMonthStartDay
					}
					if r.PayoutMonthEndDay != nil && *r.PayoutMonthEndDay >= 1 && *r.PayoutMonthEndDay <= 31 {
						effective.PayoutMonthEndDay = *r.PayoutMonthEndDay
					}
					effective.Source = "user"
				}
			}
		}
	}

	// Validate month window ordering: if start > end, swap or fallback to system
	if effective.PayoutMonthStartDay > effective.PayoutMonthEndDay {
		// fallback to system window
		effective.PayoutMonthStartDay = sysMonthStart
		effective.PayoutMonthEndDay = sysMonthEnd
	}

	return effective, nil
}

// IsEligible checks if user can withdraw given balance and current time according to frequency + threshold.
func IsEligible(cfg EffectivePayoutConfig, balance int, now time.Time, timezone string) (bool, string) {
	if balance < cfg.MinimumWithdrawalCoins {
		return false, "balance below minimum withdrawal"
	}
	switch cfg.Frequency {
	case FrequencyThreshold:
		return true, ""
	case FrequencyWeekly:
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			loc = time.FixedZone("WIB", 7*3600)
		}
		wd := int(now.In(loc).Weekday()) // 0 Sunday
		if wd != cfg.PayoutWeekday {
			return false, "not weekly payout day"
		}
		return true, ""
	case FrequencyMonthly:
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			loc = time.FixedZone("WIB", 7*3600)
		}
		day := now.In(loc).Day()
		if day < cfg.PayoutMonthStartDay || day > cfg.PayoutMonthEndDay {
			return false, "not in monthly payout window"
		}
		return true, ""
	default:
		return true, ""
	}
}

// ValidateMinimumWithdrawal validates per-user threshold against system minimum.
func ValidateMinimumWithdrawal(userMin, systemMin int) error {
	if userMin <= 0 {
		return &ValidationError{Msg: "minimum_withdrawal_coins must be > 0"}
	}
	if systemMin > 0 && userMin < systemMin {
		return &ValidationError{Msg: "minimum withdrawal below system minimum"}
	}
	return nil
}

type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

// GetSystemMinimumWithdrawal resolves system minimum from DB or default.
func GetSystemMinimumWithdrawal(ctx context.Context, client db.SupabaseClient) int {
	raw, err := client.Get(ctx, "odyssey_system_config", "key=eq.default_minimum_withdrawal_coins&select=value")
	if err == nil && len(raw) > 0 {
		var rows []struct{ Value string `json:"value"` }
		if jerr := json.Unmarshal(raw, &rows); jerr == nil && len(rows) > 0 {
			if v, err := strconv.Atoi(strings.TrimSpace(rows[0].Value)); err == nil && v > 0 {
				return v
			}
		}
	}
	// fallback to odyssey_system_config min_withdrawal legacy or default
	raw2, err := client.Get(ctx, "odyssey_system_config", "key=eq.minimum_withdrawal_coins&select=value")
	if err == nil && len(raw2) > 0 {
		var rows []struct{ Value string `json:"value"` }
		if jerr := json.Unmarshal(raw2, &rows); jerr == nil && len(rows) > 0 {
			if v, err := strconv.Atoi(strings.TrimSpace(rows[0].Value)); err == nil && v > 0 {
				return v
			}
		}
	}
	return DefaultMinimumWithdrawal
}
