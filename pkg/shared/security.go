package shared

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"
)

type SecurityConfig struct {
	AllowedOrigins []string
	MaxBodyBytes   int64
	// MaxBodyBytesByPath overrides MaxBodyBytes for a specific request path
	// (matched exactly or as a path-prefix subtree like "/api/creative/").
	MaxBodyBytesByPath map[string]int64
	MaxJSONDepth       int
	MaxStringLength    int
	RateLimitWindow    time.Duration
	RateLimitMaxHits   int
	LoginRateLimitMax  int
	AdminRateLimitMax  int
	CSRFHeaderName     string
	CSRFMaxAge         int
}

func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		AllowedOrigins:    []string{},
		MaxBodyBytes:      1 << 20,
		MaxJSONDepth:      32,
		MaxStringLength:   4096,
		RateLimitWindow:   time.Minute,
		RateLimitMaxHits:  100,
		LoginRateLimitMax: 5,
		AdminRateLimitMax: 30,
		CSRFHeaderName:    "X-CSRF-Token",
		CSRFMaxAge:        86400,
	}
}

func (c SecurityConfig) IsOriginAllowed(origin string) bool {
	if len(c.AllowedOrigins) == 0 {
		return true
	}
	for _, o := range c.AllowedOrigins {
		if o == "*" || o == origin {
			return true
		}
		if strings.HasPrefix(o, "https://") || strings.HasPrefix(o, "http://") {
			schemeEnd := strings.Index(o, "://")
			if schemeEnd > 0 {
				allowedHost := o[schemeEnd+3:]
				originHost := ""
				if strings.HasPrefix(origin, "https://") || strings.HasPrefix(origin, "http://") {
					originSchemeEnd := strings.Index(origin, "://")
					if originSchemeEnd > 0 {
						originHost = origin[originSchemeEnd+3:]
					}
				}
				if originHost == allowedHost || strings.HasSuffix(originHost, "."+allowedHost) {
					return true
				}
			}
		}
	}
	return false
}

// MaxBodyForPath resolves the request body size limit for a path, falling back
// to the global MaxBodyBytes when no per-path override is registered. A path
// prefix key (e.g. "/api/creative/") matches the subtree; an exact key
// (e.g. "/api/creative") matches that path precisely.
func (c SecurityConfig) MaxBodyForPath(path string) int64 {
	if exact, ok := c.MaxBodyBytesByPath[path]; ok && exact > 0 {
		return exact
	}
	for prefix, limit := range c.MaxBodyBytesByPath {
		if strings.HasSuffix(prefix, "/") && strings.HasPrefix(path, prefix) && limit > 0 {
			return limit
		}
	}
	return c.MaxBodyBytes
}

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
	window  time.Duration
	maxHits int
}

type rateEntry struct {
	hits    int
	expires time.Time
}

func NewRateLimiter(window time.Duration, maxHits int) *RateLimiter {
	return &RateLimiter{
		entries: make(map[string]*rateEntry),
		window:  window,
		maxHits: maxHits,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]
	if !exists || now.After(entry.expires) {
		rl.entries[key] = &rateEntry{hits: 1, expires: now.Add(rl.window)}
		return true
	}
	entry.hits++
	return entry.hits <= rl.maxHits
}

func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for k, v := range rl.entries {
		if now.After(v.expires) {
			delete(rl.entries, k)
		}
	}
}

func SecurityHeadersMiddleware(cfg SecurityConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'none'; object-src 'none'; base-uri 'self'; form-action 'self'")
		next.ServeHTTP(w, r)
	}
}

func CSRFMiddleware(cfg SecurityConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		token := r.Header.Get(cfg.CSRFHeaderName)
		if token == "" {
			cookie, err := r.Cookie("odyssey_csrf")
			if err == nil {
				token = cookie.Value
			}
		}
		if token == "" || len(token) < 32 {
			WriteJSONError(w, "missing or invalid CSRF token", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func CORSHeaderMiddleware(cfg SecurityConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && cfg.IsOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func RequestLimitMiddleware(cfg SecurityConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		maxBytes := cfg.MaxBodyForPath(r.URL.Path)
		if maxBytes > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}
		next.ServeHTTP(w, r)
	}
}

func GenerateCSRFToken() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
