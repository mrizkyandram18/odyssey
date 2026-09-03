package shared

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	SupabaseURL          string
	SupabaseServiceKey   string
	FirebaseCredentials  string
	ParentID             string
	SessionSecret        string
	AdminSecret          string
	MinBuildNumber       string
	Timezone             string
	AllowedOrigins       []string
	MaxBodyBytes         int64
	RateLimitWindowSec   int
	RateLimitMaxHits     int
	LoginRateLimitMax    int
	AdminRateLimitMax    int
	InternalMetricsToken string
	VAPIDPublicKey       string
	VAPIDPrivateKey      string
	VAPIDSubject         string
}

func LoadConfig() Config {
	return Config{
		SupabaseURL:          os.Getenv("SUPABASE_URL"),
		SupabaseServiceKey:   os.Getenv("SUPABASE_SERVICE_KEY"),
		FirebaseCredentials:  os.Getenv("FIREBASE_CREDENTIALS"),
		ParentID:             os.Getenv("PARENT_ID"),
		SessionSecret:        os.Getenv("SESSION_SIGNING_SECRET"),
		AdminSecret:          os.Getenv("ADMIN_SECRET"),
		MinBuildNumber:       getEnvDefault("GATEKEEPER_MIN_BUILD_NUMBER", "49"),
		Timezone:             getEnvDefault("ODYSSEY_TIMEZONE", "Asia/Jakarta"),
		AllowedOrigins:       getEnvSlice("ODYSSEY_ALLOWED_ORIGINS"),
		MaxBodyBytes:         int64(getEnvIntDefault("ODYSSEY_MAX_BODY_BYTES", 1048576)),
		RateLimitWindowSec:   int(getEnvIntDefault("ODYSSEY_RATE_LIMIT_WINDOW_SEC", 60)),
		RateLimitMaxHits:     int(getEnvIntDefault("ODYSSEY_RATE_LIMIT_MAX_HITS", 100)),
		LoginRateLimitMax:    int(getEnvIntDefault("ODYSSEY_LOGIN_RATE_LIMIT_MAX", 5)),
		AdminRateLimitMax:    int(getEnvIntDefault("ODYSSEY_ADMIN_RATE_LIMIT_MAX", 30)),
		InternalMetricsToken: os.Getenv("ODYSSEY_INTERNAL_METRICS_TOKEN"),
		VAPIDPublicKey:       os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey:      os.Getenv("VAPID_PRIVATE_KEY"),
		VAPIDSubject:         getEnvDefault("VAPID_SUBJECT", "mailto:admin@odyssey.example.com"),
	}
}

func (c Config) Validate() error {
	var missing []string
	if c.SupabaseURL == "" {
		missing = append(missing, "SUPABASE_URL")
	}
	if c.SupabaseServiceKey == "" {
		missing = append(missing, "SUPABASE_SERVICE_KEY")
	}
	if c.ParentID == "" {
		missing = append(missing, "PARENT_ID")
	}
	if c.SessionSecret == "" {
		missing = append(missing, "SESSION_SIGNING_SECRET")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	_, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", c.Timezone, err)
	}
	if c.MaxBodyBytes < 1 {
		return fmt.Errorf("ODYSSEY_MAX_BODY_BYTES must be >= 1, got %d", c.MaxBodyBytes)
	}
	if c.RateLimitWindowSec < 1 {
		return fmt.Errorf("ODYSSEY_RATE_LIMIT_WINDOW_SEC must be >= 1, got %d", c.RateLimitWindowSec)
	}
	if c.RateLimitMaxHits < 1 {
		return fmt.Errorf("ODYSSEY_RATE_LIMIT_MAX_HITS must be >= 1, got %d", c.RateLimitMaxHits)
	}
	if c.LoginRateLimitMax < 1 {
		return fmt.Errorf("ODYSSEY_LOGIN_RATE_LIMIT_MAX must be >= 1, got %d", c.LoginRateLimitMax)
	}
	if c.AdminRateLimitMax < 1 {
		return fmt.Errorf("ODYSSEY_ADMIN_RATE_LIMIT_MAX must be >= 1, got %d", c.AdminRateLimitMax)
	}
	return nil
}

func getEnvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvIntDefault(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var n int64
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil {
		return def
	}
	return n
}

func getEnvSlice(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

type RedemptionConfig struct {
	RedemptionStartDay       int    `json:"redemption_start_day"`
	RedemptionEndDay         int    `json:"redemption_end_day"`
	PayoutDay                int    `json:"payout_day"`
	EarningPeriodDays        int    `json:"earning_period_days"`
	IsOpen                   bool   `json:"is_open"`
	IsPayoutDay              bool   `json:"is_payout_day"`
	CurrentDay               int    `json:"current_day"`
	ConversionRate           int    `json:"conversion_rate"`
	PayoutTargetRupiah       int    `json:"payout_target_rupiah"`
	PayoutTargetCoins        int    `json:"payout_target_coins"`
	MaxPayoutCoins           int    `json:"max_payout_coins"`
	Timezone                 string `json:"timezone"`
	DefaultMonthlyCoinTarget int    `json:"default_monthly_coin_target"`
	TargetEarningStartDay    int    `json:"target_earning_start_day"`
	TargetEarningEndDay      int    `json:"target_earning_end_day"`
	AutoBlockInactivityDays  int    `json:"auto_block_inactivity_days"`
}

const DefaultRedemptionStartDay = 24
const DefaultRedemptionEndDay = 26
const DefaultPayoutDay = 24
const DefaultEarningPeriodDays = 30
const DefaultCoinConversionRate = 100
const DefaultPayoutTargetRupiah = 320000
const DefaultPayoutTargetCoins = 3200
const DefaultMaxPayoutCoins = 3200
const DefaultMonthlyCoinTarget = 0
const DefaultTargetEarningStartDay = 1
const DefaultTargetEarningEndDay = 24
const DefaultTimezone = "Asia/Jakarta"
const DefaultAutoBlockInactivityDays = 5

func ResolveRedemptionConfig(startDay, endDay int, tzName string, now time.Time) RedemptionConfig {
	return ResolveRedemptionConfigFull(ResolveRedemptionConfigParams{
		StartDay:           startDay,
		EndDay:             endDay,
		Timezone:           tzName,
		Now:                now,
		ConversionRate:     DefaultCoinConversionRate,
		PayoutTargetRupiah: DefaultPayoutTargetRupiah,
		PayoutTargetCoins:  DefaultPayoutTargetCoins,
		MaxPayoutCoins:     DefaultMaxPayoutCoins,
		PayoutDay:          DefaultPayoutDay,
		EarningPeriodDays:  DefaultEarningPeriodDays,
	})
}

type ResolveRedemptionConfigParams struct {
	StartDay                int
	EndDay                  int
	Timezone                string
	Now                     time.Time
	ConversionRate          int
	PayoutTargetRupiah      int
	PayoutTargetCoins       int
	MaxPayoutCoins          int
	PayoutDay               int
	EarningPeriodDays       int
	AutoBlockInactivityDays int
}

func ResolveRedemptionConfigWithRate(startDay, endDay int, tzName string, now time.Time, conversionRate int, maxPayoutCoins int) RedemptionConfig {
	return ResolveRedemptionConfigFull(ResolveRedemptionConfigParams{
		StartDay:           startDay,
		EndDay:             endDay,
		Timezone:           tzName,
		Now:                now,
		ConversionRate:     conversionRate,
		MaxPayoutCoins:     maxPayoutCoins,
		PayoutDay:          DefaultPayoutDay,
		EarningPeriodDays:  DefaultEarningPeriodDays,
		PayoutTargetRupiah: DefaultPayoutTargetRupiah,
		PayoutTargetCoins:  DefaultPayoutTargetCoins,
	})
}

func ResolveRedemptionConfigFull(p ResolveRedemptionConfigParams) RedemptionConfig {
	if p.StartDay <= 0 {
		p.StartDay = DefaultRedemptionStartDay
	}
	if p.EndDay <= 0 {
		p.EndDay = DefaultRedemptionEndDay
	}
	if p.PayoutDay <= 0 || p.PayoutDay > 31 {
		p.PayoutDay = DefaultPayoutDay
	}
	if p.EarningPeriodDays <= 0 {
		p.EarningPeriodDays = DefaultEarningPeriodDays
	}
	if p.ConversionRate <= 0 {
		p.ConversionRate = DefaultCoinConversionRate
	}
	if p.PayoutTargetRupiah < 0 {
		p.PayoutTargetRupiah = DefaultPayoutTargetRupiah
	}
	if p.MaxPayoutCoins <= 0 {
		p.MaxPayoutCoins = DefaultMaxPayoutCoins
	}
	if p.AutoBlockInactivityDays < 0 {
		p.AutoBlockInactivityDays = DefaultAutoBlockInactivityDays
	}
	if p.AutoBlockInactivityDays == 0 {
		// 0 means disabled; keep as 0, do not fallback to default
	}
	if p.AutoBlockInactivityDays > 365 {
		p.AutoBlockInactivityDays = DefaultAutoBlockInactivityDays
	}
	if p.Timezone == "" {
		p.Timezone = DefaultTimezone
	}
	// Derive target coins from rupiah/rate (single source of truth derivation)
	if p.ConversionRate > 0 && p.PayoutTargetRupiah >= 0 {
		p.PayoutTargetCoins = p.PayoutTargetRupiah / p.ConversionRate
	}
	if p.PayoutTargetCoins <= 0 {
		p.PayoutTargetCoins = p.MaxPayoutCoins
	}

	loc, err := time.LoadLocation(p.Timezone)
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	nowInTz := p.Now.In(loc)
	curDay := nowInTz.Day()
	isOpen := curDay >= p.StartDay && curDay <= p.EndDay
	isPayoutDay := curDay == p.PayoutDay

	return RedemptionConfig{
		RedemptionStartDay:       p.StartDay,
		RedemptionEndDay:         p.EndDay,
		PayoutDay:                p.PayoutDay,
		EarningPeriodDays:        p.EarningPeriodDays,
		IsOpen:                   isOpen,
		IsPayoutDay:              isPayoutDay,
		CurrentDay:               curDay,
		ConversionRate:           p.ConversionRate,
		PayoutTargetRupiah:       p.PayoutTargetRupiah,
		PayoutTargetCoins:        p.PayoutTargetCoins,
		MaxPayoutCoins:           p.MaxPayoutCoins,
		Timezone:                 p.Timezone,
		DefaultMonthlyCoinTarget: DefaultMonthlyCoinTarget,
		TargetEarningStartDay:    DefaultTargetEarningStartDay,
		TargetEarningEndDay:      DefaultTargetEarningEndDay,
		AutoBlockInactivityDays:  p.AutoBlockInactivityDays,
	}
}

func ValidateEconomyConfig(cfg RedemptionConfig) error {
	if cfg.ConversionRate <= 0 {
		return fmt.Errorf("coin_conversion_rate must be > 0")
	}
	if cfg.MaxPayoutCoins <= 0 {
		return fmt.Errorf("max_payout_coins must be > 0")
	}
	if cfg.PayoutTargetRupiah < 0 {
		return fmt.Errorf("payout_target_rupiah must be >= 0")
	}
	if cfg.EarningPeriodDays <= 0 {
		return fmt.Errorf("earning_period_days must be > 0")
	}
	if cfg.PayoutDay < 1 || cfg.PayoutDay > 31 {
		return fmt.Errorf("payout_day must be 1-31")
	}
	if cfg.RedemptionStartDay < 1 || cfg.RedemptionStartDay > 31 || cfg.RedemptionEndDay < 1 || cfg.RedemptionEndDay > 31 {
		return fmt.Errorf("redemption window days must be 1-31")
	}
	if cfg.PayoutTargetCoins > cfg.MaxPayoutCoins {
		return fmt.Errorf("payout_target_coins (%d) must not exceed max_payout_coins (%d)", cfg.PayoutTargetCoins, cfg.MaxPayoutCoins)
	}
	if _, err := time.LoadLocation(cfg.Timezone); err != nil {
		return fmt.Errorf("invalid timezone %q", cfg.Timezone)
	}
	return nil
}
