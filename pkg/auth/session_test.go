package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestIssuer(t *testing.T) *HMACSessionIssuer {
	t.Helper()
	return NewSessionIssuer("test-secret-key")
}

func TestIssueSession_UserCreatesTokenAndClaims(t *testing.T) {
	issuer := newTestIssuer(t)
	cfg := &SessionConfig{Role: RoleSeeker, FamilyID: "crew-1"}

	token, claims, err := issuer.IssueSession(SessionKindUser, "alice", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Fatalf("expected token format payload.sig, got %d parts", len(parts))
	}
	if claims.UID != "alice" {
		t.Errorf("expected UID alice, got %s", claims.UID)
	}
	if claims.Kind != "user" {
		t.Errorf("expected kind user, got %s", claims.Kind)
	}
	if claims.Role != "SEEKER" {
		t.Errorf("expected role SEEKER, got %s", claims.Role)
	}
	if claims.FamilyID != "crew-1" {
		t.Errorf("expected family_id crew-1, got %s", claims.FamilyID)
	}
	if claims.Version != 1 {
		t.Errorf("expected version 1, got %d", claims.Version)
	}
	if claims.Expires-claims.Issued != int64(UserSessionTTL.Seconds()) {
		t.Errorf("expected TTL %v, got %v", UserSessionTTL, time.Duration(claims.Expires-claims.Issued)*time.Second)
	}
}

func TestIssueSession_SetupUsesShorterTTL(t *testing.T) {
	issuer := newTestIssuer(t)
	token, claims, err := issuer.IssueSession(SessionKindSetup, "bob", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if claims.Kind != "setup" {
		t.Errorf("expected kind setup, got %s", claims.Kind)
	}
	if claims.Expires-claims.Issued != int64(SetupTokenTTL.Seconds()) {
		t.Errorf("expected setup TTL %v, got %v", SetupTokenTTL, time.Duration(claims.Expires-claims.Issued)*time.Second)
	}
	if claims.Role != "" || claims.FamilyID != "" {
		t.Errorf("setup session should not carry role/crew, got role=%s crew=%s", claims.Role, claims.FamilyID)
	}
}

func TestIssueSession_InvalidKind(t *testing.T) {
	issuer := newTestIssuer(t)
	_, _, err := issuer.IssueSession("invalid", "alice", nil)
	if err != ErrSessionInvalid {
		t.Fatalf("expected ErrSessionInvalid, got %v", err)
	}
}

func TestIssueSession_EmptyUID(t *testing.T) {
	issuer := newTestIssuer(t)
	_, _, err := issuer.IssueSession(SessionKindUser, "", nil)
	if err != ErrSessionInvalid {
		t.Fatalf("expected ErrSessionInvalid, got %v", err)
	}
}

func TestParseSession_ValidToken(t *testing.T) {
	issuer := newTestIssuer(t)
	token, _, err := issuer.IssueSession(SessionKindUser, "alice", &SessionConfig{Role: RoleBuilder, FamilyID: "crew-7"})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	claims, err := issuer.ParseSession(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UID != "alice" {
		t.Errorf("expected UID alice, got %s", claims.UID)
	}
	if claims.Kind != "user" {
		t.Errorf("expected kind user, got %s", claims.Kind)
	}
	if claims.Role != "BUILDER" {
		t.Errorf("expected role BUILDER, got %s", claims.Role)
	}
	if claims.FamilyID != "crew-7" {
		t.Errorf("expected family_id crew-7, got %s", claims.FamilyID)
	}
}

func TestParseSession_InvalidSignature(t *testing.T) {
	issuer := newTestIssuer(t)
	token, _, _ := issuer.IssueSession(SessionKindUser, "alice", nil)

	parts := strings.Split(token, ".")
	tampered := parts[0] + ".AAAA" + parts[1]

	_, err := issuer.ParseSession(tampered)
	if err != ErrSessionInvalid {
		t.Fatalf("expected ErrSessionInvalid, got %v", err)
	}
}

func TestParseSession_EmptyToken(t *testing.T) {
	issuer := newTestIssuer(t)
	_, err := issuer.ParseSession("")
	if err != ErrSessionInvalid {
		t.Fatalf("expected ErrSessionInvalid, got %v", err)
	}
}

func TestParseSession_ExpiredToken(t *testing.T) {
	issuer := newTestIssuer(t)
	// Manually craft an expired token using the sign helper.
	now := time.Now().UTC()
	claims := &SessionClaims{
		Version: 1,
		Kind:    "user",
		UID:     "alice",
		Issued:  now.Add(-2 * time.Hour).Unix(),
		Expires: now.Add(-1 * time.Hour).Unix(),
	}
	token, err := signSession(claims, "test-secret-key")
	if err != nil {
		t.Fatalf("signSession: %v", err)
	}

	_, err = issuer.ParseSession(token)
	if err != ErrSessionExpired {
		t.Fatalf("expected ErrSessionExpired, got %v", err)
	}
}

func TestParseSession_MalformedToken(t *testing.T) {
	issuer := newTestIssuer(t)
	_, err := issuer.ParseSession("not-a-valid-token")
	if err != ErrSessionInvalid {
		t.Fatalf("expected ErrSessionInvalid, got %v", err)
	}
}

func TestParseSession_DowngradedVersion(t *testing.T) {
	issuer := newTestIssuer(t)
	claims := &SessionClaims{
		Version: 2, // not version 1
		Kind:    "user",
		UID:     "alice",
		Issued:  time.Now().UTC().Unix(),
		Expires: time.Now().Add(time.Hour).Unix(),
	}
	token, err := signSession(claims, "test-secret-key")
	if err != nil {
		t.Fatalf("signSession: %v", err)
	}
	_, err = issuer.ParseSession(token)
	if err != ErrSessionInvalid {
		t.Fatalf("expected ErrSessionInvalid, got %v", err)
	}
}

func TestSessionValidator_InterfaceCompliance(t *testing.T) {
	var _ SessionValidator = (*HMACSessionIssuer)(nil)
	var _ SessionIssuer = (*HMACSessionIssuer)(nil)
}

func TestExtractBearerOrHeader_Bearer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer abc123")
	if got := ExtractBearerOrHeader(req, HeaderUserSession); got != "abc123" {
		t.Errorf("expected abc123, got %s", got)
	}
}

func TestExtractBearerOrHeader_FallbackHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(HeaderUserSession, "xyz789")
	if got := ExtractBearerOrHeader(req, HeaderUserSession); got != "xyz789" {
		t.Errorf("expected xyz789, got %s", got)
	}
}

func TestExtractBearerOrHeader_NoToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	if got := ExtractBearerOrHeader(req, HeaderUserSession); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestExtractToken_BearerHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer tok123")
	if got := ExtractToken(req); got != "tok123" {
		t.Errorf("expected tok123, got %s", got)
	}
}

func TestExtractToken_FallbackHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(HeaderUserSession, "hdr456")
	if got := ExtractToken(req); got != "hdr456" {
		t.Errorf("expected hdr456, got %s", got)
	}
}

func TestExtractToken_Cookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "cookie789"})
	if got := ExtractToken(req); got != "cookie789" {
		t.Errorf("expected cookie789, got %s", got)
	}
}

func TestExtractToken_PreferenceBearerOverCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer bearer-tok")
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "cookie-tok"})
	if got := ExtractToken(req); got != "bearer-tok" {
		t.Errorf("expected bearer-tok, got %s", got)
	}
}

func TestExtractToken_NonePresent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	if got := ExtractToken(req); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestSetSessionCookie(t *testing.T) {
	w := httptest.NewRecorder()
	SetSessionCookie(w, "mytoken", time.Hour, true)
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != SessionCookieName {
		t.Errorf("expected name %s, got %s", SessionCookieName, c.Name)
	}
	if c.Value != "mytoken" {
		t.Errorf("expected value mytoken, got %s", c.Value)
	}
	if !c.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if !c.Secure {
		t.Error("expected Secure to be true")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite Lax, got %d", c.SameSite)
	}
}

func TestSessionAuthErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"expired", ErrSessionExpired, "Session expired. Please log in again."},
		{"invalid", ErrSessionInvalid, "Invalid session token."},
		{"nil", nil, "Authentication required. Please log in again."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SessionAuthErrorMessage(tt.err); got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
