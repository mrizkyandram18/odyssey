package shared

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	SupabaseURL              string
	SupabaseServiceKey       string
	FirebaseCredentials      string
	ParentID                 string
	SessionSecret            string
	AdminSecret              string
	MinBuildNumber           string
	Timezone                 string
	DailyTurnXP              int64
	MaxDailyTurnsPerDay      int
	RealmProgressPerQuest    int
	RealmCompletionThreshold int
	AllowedOrigins           []string
	MaxBodyBytes             int64
	RateLimitWindowSec       int
	RateLimitMaxHits         int
	LoginRateLimitMax        int
	AdminRateLimitMax        int
	InternalMetricsToken     string
	VAPIDPublicKey           string
	VAPIDPrivateKey          string
	VAPIDSubject             string
	// System config table name (for admin service)
	SystemConfigTable string
}

func LoadConfig() Config {
	return Config{
		SupabaseURL:              os.Getenv("SUPABASE_URL"),
		SupabaseServiceKey:       os.Getenv("SUPABASE_SERVICE_KEY"),
		FirebaseCredentials:      os.Getenv("FIREBASE_CREDENTIALS"),
		ParentID:                 os.Getenv("PARENT_ID"),
		SessionSecret:            os.Getenv("SESSION_SIGNING_SECRET"),
		AdminSecret:              os.Getenv("ADMIN_SECRET"),
		MinBuildNumber:           getEnvDefault("GATEKEEPER_MIN_BUILD_NUMBER", "49"),
		Timezone:                 getEnvDefault("ODYSSEY_TIMEZONE", "Asia/Jakarta"),
		DailyTurnXP:              getEnvIntDefault("ODYSSEY_DAILY_TURN_XP", 30),
		MaxDailyTurnsPerDay:      int(getEnvIntDefault("ODYSSEY_MAX_DAILY_TURNS", 1)),
		RealmProgressPerQuest:    int(getEnvIntDefault("ODYSSEY_REALM_PROGRESS_PER_QUEST", 25)),
		RealmCompletionThreshold: int(getEnvIntDefault("ODYSSEY_REALM_COMPLETION_THRESHOLD", 100)),
		AllowedOrigins:           getEnvSlice("ODYSSEY_ALLOWED_ORIGINS"),
		MaxBodyBytes:             int64(getEnvIntDefault("ODYSSEY_MAX_BODY_BYTES", 1048576)),
		RateLimitWindowSec:       int(getEnvIntDefault("ODYSSEY_RATE_LIMIT_WINDOW_SEC", 60)),
		RateLimitMaxHits:         int(getEnvIntDefault("ODYSSEY_RATE_LIMIT_MAX_HITS", 100)),
		LoginRateLimitMax:        int(getEnvIntDefault("ODYSSEY_LOGIN_RATE_LIMIT_MAX", 5)),
		AdminRateLimitMax:        int(getEnvIntDefault("ODYSSEY_ADMIN_RATE_LIMIT_MAX", 30)),
		InternalMetricsToken:     os.Getenv("ODYSSEY_INTERNAL_METRICS_TOKEN"),
		VAPIDPublicKey:           os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey:          os.Getenv("VAPID_PRIVATE_KEY"),
		VAPIDSubject:             getEnvDefault("VAPID_SUBJECT", "mailto:admin@odyssey.example.com"),
		// System config table name (for admin service)
		SystemConfigTable: "odyssey_system_config",
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
	if c.MaxDailyTurnsPerDay < 1 {
		return fmt.Errorf("ODYSSEY_MAX_DAILY_TURNS must be >= 1, got %d", c.MaxDailyTurnsPerDay)
	}
	if c.RealmProgressPerQuest < 1 {
		return fmt.Errorf("ODYSSEY_REALM_PROGRESS_PER_QUEST must be >= 1, got %d", c.RealmProgressPerQuest)
	}
	if c.RealmCompletionThreshold < 1 {
		return fmt.Errorf("ODYSSEY_REALM_COMPLETION_THRESHOLD must be >= 1, got %d", c.RealmCompletionThreshold)
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
