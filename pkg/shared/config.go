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
	RedemptionStartDay          int            `json:"redemption_start_day"`
	RedemptionEndDay            int            `json:"redemption_end_day"`
	PayoutDay                   int            `json:"payout_day"`
	EarningPeriodDays           int            `json:"earning_period_days"`
	IsOpen                      bool           `json:"is_open"`
	IsPayoutDay                 bool           `json:"is_payout_day"`
	CurrentDay                  int            `json:"current_day"`
	ConversionRate              int            `json:"conversion_rate"`
	PayoutTargetRupiah          int            `json:"payout_target_rupiah"`
	PayoutTargetCoins           int            `json:"payout_target_coins"`
	MaxPayoutCoins              int            `json:"max_payout_coins"`
	Timezone                    string         `json:"timezone"`
	DefaultMonthlyCoinTarget    int            `json:"default_monthly_coin_target"`
	DefaultMonthlyEarningCap    int            `json:"default_monthly_earning_cap,omitempty"`
	MaxMonthlyEarningCapCeiling int            `json:"max_monthly_earning_cap_ceiling,omitempty"`
	LevelCapBonus               map[string]int `json:"level_cap_bonus,omitempty"`
	TargetEarningStartDay       int            `json:"target_earning_start_day"`
	TargetEarningEndDay         int            `json:"target_earning_end_day"`
	AutoBlockInactivityDays     int            `json:"auto_block_inactivity_days"`
	// Announcement is the admin-configured system message surfaced to members.
	// All content originates from odyssey_system_config keys (announcement_*);
	// no title/body text is hardcoded here.
	Announcement *AnnouncementConfig `json:"announcement,omitempty"`
}

// AnnouncementConfig is the runtime view of the admin-configured system
// announcement. Storage lives in the existing odyssey_system_config table
// under announcement_* keys (no new table). Visible is server-computed from
// enabled + schedule window; the member UI must honor Visible and must never
// embed announcement text in source.
type AnnouncementConfig struct {
	Enabled  bool   `json:"enabled"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Audience string `json:"audience"`
	StartAt  string `json:"start_at,omitempty"`
	EndAt    string `json:"end_at,omitempty"`
	Priority string `json:"priority"`
	Visible  bool   `json:"visible"`
}

// Announcement key names in odyssey_system_config (existing table, existing pattern).
const (
	AnnouncementEnabledKey  = "announcement_enabled"
	AnnouncementTitleKey    = "announcement_title"
	AnnouncementBodyKey     = "announcement_body"
	AnnouncementAudienceKey = "announcement_audience"
	AnnouncementStartAtKey  = "announcement_start_at"
	AnnouncementEndAtKey    = "announcement_end_at"
	AnnouncementPriorityKey = "announcement_priority"
)

const (
	DefaultAnnouncementAudience = "ALL"
	DefaultAnnouncementPriority = "normal"
)

var validAnnouncementAudiences = map[string]bool{"ALL": true, "MEMBER": true, "ADMIN": true}
var validAnnouncementPriorities = map[string]bool{"low": true, "normal": true, "high": true, "urgent": true}

// ParseAnnouncementConfig builds the runtime announcement view from raw
// system-config key/values. It performs no I/O and contains no business
// content — only technical defaults and schedule evaluation.
func ParseAnnouncementConfig(values map[string]string, now time.Time) *AnnouncementConfig {
	enabled := false
	if v, ok := values[AnnouncementEnabledKey]; ok {
		s := strings.TrimSpace(strings.ToLower(v))
		enabled = s == "true" || s == "1" || s == "yes" || s == "on"
	}
	title := ""
	if v, ok := values[AnnouncementTitleKey]; ok {
		title = strings.TrimSpace(v)
	}
	body := ""
	if v, ok := values[AnnouncementBodyKey]; ok {
		body = strings.TrimSpace(v)
	}
	audience := DefaultAnnouncementAudience
	if v, ok := values[AnnouncementAudienceKey]; ok && strings.TrimSpace(v) != "" {
		audience = strings.ToUpper(strings.TrimSpace(v))
	}
	startAt := ""
	if v, ok := values[AnnouncementStartAtKey]; ok {
		startAt = strings.TrimSpace(v)
	}
	endAt := ""
	if v, ok := values[AnnouncementEndAtKey]; ok {
		endAt = strings.TrimSpace(v)
	}
	priority := DefaultAnnouncementPriority
	if v, ok := values[AnnouncementPriorityKey]; ok && strings.TrimSpace(v) != "" {
		priority = strings.ToLower(strings.TrimSpace(v))
	}
	visible := false
	if enabled && (title != "" || body != "") {
		visible = true
		if startAt != "" {
			if t, err := time.Parse(time.RFC3339, startAt); err == nil {
				if now.Before(t) {
					visible = false
				}
			}
		}
		if visible && endAt != "" {
			if t, err := time.Parse(time.RFC3339, endAt); err == nil {
				if now.After(t) {
					visible = false
				}
			}
		}
	}
	return &AnnouncementConfig{
		Enabled:  enabled,
		Title:    title,
		Body:     body,
		Audience: audience,
		StartAt:  startAt,
		EndAt:    endAt,
		Priority: priority,
		Visible:  visible,
	}
}

// ValidateAnnouncementInput validates admin-supplied announcement fields.
// Empty title+body with enabled=false is allowed (disabled state). When
// enabled, at least one of title/body must be present. Schedule bounds must
// be RFC3339 when provided, and start must not be after end.
func ValidateAnnouncementInput(enabled *bool, title, body, audience, startAt, endAt, priority *string) error {
	if audience != nil {
		a := strings.ToUpper(strings.TrimSpace(*audience))
		if a != "" && !validAnnouncementAudiences[a] {
			return fmt.Errorf("announcement_audience must be one of ALL, MEMBER, ADMIN")
		}
	}
	if priority != nil {
		p := strings.ToLower(strings.TrimSpace(*priority))
		if p != "" && !validAnnouncementPriorities[p] {
			return fmt.Errorf("announcement_priority must be one of low, normal, high, urgent")
		}
	}
	var startT, endT time.Time
	var hasStart, hasEnd bool
	if startAt != nil && strings.TrimSpace(*startAt) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*startAt))
		if err != nil {
			return fmt.Errorf("announcement_start_at must be RFC3339 (e.g. 2026-09-12T00:00:00+07:00)")
		}
		startT, hasStart = t, true
	}
	if endAt != nil && strings.TrimSpace(*endAt) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*endAt))
		if err != nil {
			return fmt.Errorf("announcement_end_at must be RFC3339 (e.g. 2026-09-30T23:59:59+07:00)")
		}
		endT, hasEnd = t, true
	}
	if hasStart && hasEnd && startT.After(endT) {
		return fmt.Errorf("announcement_start_at must not be after announcement_end_at")
	}
	if title != nil && len(*title) > 255 {
		return fmt.Errorf("announcement_title too long (max 255 characters)")
	}
	if body != nil && len(*body) > 5000 {
		return fmt.Errorf("announcement_body too long (max 5000 characters)")
	}
	if enabled != nil && *enabled {
		t := ""
		if title != nil {
			t = strings.TrimSpace(*title)
		}
		b := ""
		if body != nil {
			b = strings.TrimSpace(*body)
		}
		// When enabling via a partial update, the caller merges with stored
		// values before calling; here we only reject an explicit enable with
		// both fields blanked in the same payload.
		if title != nil && body != nil && t == "" && b == "" {
			return fmt.Errorf("announcement_title or announcement_body must be set when enabled")
		}
	}
	return nil
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

// DefaultMaxUploadBytes is the single request/file size limit for task proof
// uploads. It must stay in sync between the HTTP body-limit middleware and the
// multipart handler checks, hence one shared constant instead of literals.
const DefaultMaxUploadBytes = 10 << 20
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
