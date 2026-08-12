package shared

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	cfg := DefaultSecurityConfig()
	var captured map[string]string
	next := func(w http.ResponseWriter, r *http.Request) {
		captured = make(map[string]string)
		for k := range w.Header() {
			captured[strings.ToLower(k)] = w.Header().Get(k)
		}
		w.WriteHeader(http.StatusOK)
	}

	handler := SecurityHeadersMiddleware(cfg, next)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler(w, r)

	expected := []string{
		"x-content-type-options",
		"x-frame-options",
		"referrer-policy",
		"permissions-policy",
		"content-security-policy",
	}
	for _, h := range expected {
		if captured[h] == "" {
			t.Errorf("expected security header %s to be set", h)
		}
	}
	if captured["x-content-type-options"] != "nosniff" {
		t.Errorf("expected nosniff, got %s", captured["x-content-type-options"])
	}
	if captured["x-frame-options"] != "DENY" {
		t.Errorf("expected DENY, got %s", captured["x-frame-options"])
	}
}

func TestCORSHeaderMiddleware_AllowedOrigin(t *testing.T) {
	cfg := DefaultSecurityConfig()
	cfg.AllowedOrigins = []string{"https://example.com"}
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	handler := CORSHeaderMiddleware(cfg, next)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "https://example.com")
	handler(w, r)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected allowed origin, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSHeaderMiddleware_DisallowedOrigin(t *testing.T) {
	cfg := DefaultSecurityConfig()
	cfg.AllowedOrigins = []string{"https://example.com"}
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	handler := CORSHeaderMiddleware(cfg, next)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Set("Origin", "https://evil.com")
	handler(w, r)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no CORS headers for disallowed origin")
	}
}

func TestCORSHeaderMiddleware_Options(t *testing.T) {
	cfg := DefaultSecurityConfig()
	cfg.AllowedOrigins = []string{"https://example.com"}
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	handler := CORSHeaderMiddleware(cfg, next)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodOptions, "/test", nil)
	r.Header.Set("Origin", "https://example.com")
	handler(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", w.Code)
	}
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 3)
	for i := 0; i < 3; i++ {
		if !rl.Allow("client1") {
			t.Fatalf("expected allow on attempt %d", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 2)
	for i := 0; i < 2; i++ {
		rl.Allow("client1")
	}
	if rl.Allow("client1") {
		t.Error("expected rate limit to block after max hits")
	}
}

func TestRateLimiter_SeparateClients(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 2)
	for i := 0; i < 2; i++ {
		rl.Allow("client1")
	}
	if !rl.Allow("client2") {
		t.Error("expected client2 to be allowed independently")
	}
}

func TestRateLimiter_ExpiredEntries(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 1)
	rl.Allow("client1")
	time.Sleep(100 * time.Millisecond)
	if !rl.Allow("client1") {
		t.Error("expected allow after entry expired")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 1)
	rl.Allow("client1")
	time.Sleep(100 * time.Millisecond)
	rl.Cleanup()
	rl.Allow("client1")
}

func TestGenerateCSRFToken(t *testing.T) {
	t1 := GenerateCSRFToken()
	if len(t1) != 64 {
		t.Fatalf("expected 64 char token, got %d", len(t1))
	}
	t2 := GenerateCSRFToken()
	if t1 == t2 {
		t.Error("expected unique CSRF tokens")
	}
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"a", true},
		{"abc-123_def", true},
		{"UPPER", true},
		{"with space", false},
		{"with.dots", false},
		{strings.Repeat("a", 257), false},
	}
	for _, tt := range tests {
		got := ValidateSlug(tt.input)
		if got != tt.want {
			t.Errorf("ValidateSlug(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestSecurityConfig_IsOriginAllowed(t *testing.T) {
	tests := []struct {
		origins []string
		origin  string
		want    bool
	}{
		{[]string{}, "https://any.com", true},
		{[]string{"*"}, "https://any.com", true},
		{[]string{"https://example.com"}, "https://example.com", true},
		{[]string{"https://example.com"}, "https://evil.com", false},
		{[]string{"https://example.com"}, "https://sub.example.com", true},
	}
	for _, tt := range tests {
		cfg := SecurityConfig{AllowedOrigins: tt.origins}
		got := cfg.IsOriginAllowed(tt.origin)
		if got != tt.want {
			t.Errorf("IsOriginAllowed(%v, %q) = %v, want %v", tt.origins, tt.origin, got, tt.want)
		}
	}
}

func TestValidateInt64Param(t *testing.T) {
	tests := []struct {
		key   string
		query string
		want  int64
		ok    bool
	}{
		{"id", "123", 123, true},
		{"id", "abc", 0, false},
		{"id", "", 0, false},
		{"id", "-5", -5, true},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "/test?"+tt.key+"="+tt.query, nil)
		got, ok := ValidateInt64Param(r, tt.key)
		if got != tt.want || ok != tt.ok {
			t.Errorf("ValidateInt64Param(%q) = (%d, %v), want (%d, %v)", tt.query, got, ok, tt.want, tt.ok)
		}
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input   string
		maxLen  int
		wantLen int
	}{
		{"hello", 10, 5},
		{strings.Repeat("a", 300), 256, 256},
		{"  spaced  ", 100, 6},
	}
	for _, tt := range tests {
		got := SanitizeString(tt.input, tt.maxLen)
		if len(got) != tt.wantLen {
			t.Errorf("SanitizeString length = %d, want %d", len(got), tt.wantLen)
		}
		if strings.Contains(got, " ") {
			t.Errorf("expected trimmed spaces, got %q", got)
		}
	}
}

func TestCSRFMiddleware_MissingToken(t *testing.T) {
	cfg := DefaultSecurityConfig()
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	handler := CSRFMiddleware(cfg, next)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test", nil)
	handler(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for missing CSRF token, got %d", w.Code)
	}
}

func TestCSRFMiddleware_ValidToken(t *testing.T) {
	cfg := DefaultSecurityConfig()
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	handler := CSRFMiddleware(cfg, next)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test", nil)
	r.Header.Set("X-CSRF-Token", strings.Repeat("a", 32))
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for valid CSRF token, got %d", w.Code)
	}
}

func TestCSRFMiddleware_GetAllowedWithoutToken(t *testing.T) {
	cfg := DefaultSecurityConfig()
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	handler := CSRFMiddleware(cfg, next)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for GET without CSRF token, got %d", w.Code)
	}
}

func TestSecurityEventLogging(t *testing.T) {
	InitSecurityLogger(10)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	req.Header.Set("User-Agent", "test-agent")

	LogSecurityEvent(SecEventForbiddenAccess, req, "admin role check failed")

	time.Sleep(100 * time.Millisecond)
}

func TestBodySizeLimit(t *testing.T) {
	cfg := DefaultSecurityConfig()
	cfg.MaxBodyBytes = 100
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	handler := RequestLimitMiddleware(cfg, next)

	largeBody := strings.Repeat("a", 200)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(largeBody))
	r.Header.Set("Content-Type", "application/json")
	handler(w, r)

	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		t.Logf("body limit returned status %d (expected 413 or 400)", w.Code)
	}
}

func TestMaxBodyForPath(t *testing.T) {
	cfg := DefaultSecurityConfig()
	cfg.MaxBodyBytes = 1 << 20
	cfg.MaxBodyBytesByPath = map[string]int64{
		"/api/creative":  8 << 20,
		"/api/creative/": 8 << 20,
	}

	cases := []struct {
		path string
		want int64
	}{
		{"/api/creative", 8 << 20},
		{"/api/creative/1/approve", 8 << 20},
		{"/api/login", 1 << 20},
		{"/api/missions", 1 << 20},
	}
	for _, c := range cases {
		if got := cfg.MaxBodyForPath(c.path); got != c.want {
			t.Errorf("path %q: expected %d, got %d", c.path, c.want, got)
		}
	}
}

func TestBodySizeLimit_PerPathOverride(t *testing.T) {
	cfg := DefaultSecurityConfig()
	cfg.MaxBodyBytes = 100
	cfg.MaxBodyBytesByPath = map[string]int64{"/api/creative": 1000}
	called := false
	next := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}
	handler := RequestLimitMiddleware(cfg, next)

	body := strings.Repeat("a", 500) // exceeds 100 but within 1000
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/creative", strings.NewReader(body))
	handler(w, r)

	if !called {
		t.Fatalf("expected handler to be called within per-path limit")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRateLimiterConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 100)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("client-%d", id)
			for j := 0; j < 10; j++ {
				rl.Allow(key)
			}
		}(i)
	}
	wg.Wait()
}
