package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

var (
	ErrSessionInvalid = errors.New("invalid session")
	ErrSessionExpired = errors.New("session expired")
	ErrSessionKind    = errors.New("wrong session kind")
)

const (
	UserSessionTTL    = 8 * time.Hour
	SetupTokenTTL     = 30 * time.Minute
	HeaderUserSession = "X-User-Session"
	SessionCookieName = "odyssey_session"
)

type HMACSessionIssuer struct {
	secret string
}

func NewSessionIssuer(secret string) *HMACSessionIssuer {
	return &HMACSessionIssuer{secret: strings.TrimSpace(secret)}
}

// IssueSession creates an HMAC-signed session token with the given kind,
// UID, and optional claims (role, family_id). The TTL is determined by
// the session kind (8h for user, 30m for setup).
func (s *HMACSessionIssuer) IssueSession(kind SessionKind, uid string, cfg *SessionConfig) (string, *SessionClaims, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return "", nil, ErrSessionInvalid
	}

	var ttl time.Duration
	switch kind {
	case SessionKindUser:
		ttl = UserSessionTTL
	case SessionKindSetup:
		ttl = SetupTokenTTL
	default:
		return "", nil, ErrSessionInvalid
	}

	now := time.Now().UTC()
	claims := &SessionClaims{
		Version: 1,
		Kind:    string(kind),
		UID:     uid,
		Issued:  now.Unix(),
		Expires: now.Add(ttl).Unix(),
	}
	if cfg != nil {
		claims.Role = string(cfg.Role)
		claims.FamilyID = cfg.FamilyID
	}

	token, err := signSession(claims, s.secret)
	if err != nil {
		return "", nil, err
	}
	return token, claims, nil
}

// ParseSession validates the HMAC signature, checks the version and
// required fields, and enforces expiry. Returns the claims on success.
func (s *HMACSessionIssuer) ParseSession(token string) (*SessionClaims, error) {
	claims, err := verifySessionSignature(token, s.secret)
	if err != nil {
		return nil, ErrSessionInvalid
	}
	if claims.Version != 1 {
		return nil, ErrSessionInvalid
	}
	if claims.UID == "" || claims.Kind == "" {
		return nil, ErrSessionInvalid
	}
	if time.Now().UTC().Unix() > claims.Expires {
		return nil, ErrSessionExpired
	}
	return claims, nil
}

// ExtractBearerOrHeader reads Authorization: Bearer, then fallback header.
func ExtractBearerOrHeader(r *http.Request, fallbackHeader string) string {
	if r == nil {
		return ""
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(auth) >= 7 && strings.EqualFold(auth[:7], "Bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	if fallbackHeader != "" {
		return strings.TrimSpace(r.Header.Get(fallbackHeader))
	}
	return ""
}

// ExtractToken reads the session token from, in order of precedence:
//  1. Authorization: Bearer header
//  2. X-User-Session header (fallback)
//  3. odyssey_session cookie
func ExtractToken(r *http.Request) string {
	if token := ExtractBearerOrHeader(r, HeaderUserSession); token != "" {
		return token
	}
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		return strings.TrimSpace(cookie.Value)
	}
	return ""
}

// SetSessionCookie writes the session token as an HTTP-only cookie
// with the same TTL used for the session.
func SetSessionCookie(w http.ResponseWriter, token string, ttl time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// signSession creates base64(JSON claims) + "." + base64(HMAC-SHA256 payload).
func signSession(claims *SessionClaims, secret string) (string, error) {
	raw, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

// verifySessionSignature checks the HMAC signature and returns the decoded
// claims. It does not validate semantic fields (version, UID, expiry) —
// callers are responsible for those checks.
func verifySessionSignature(token, secret string) (*SessionClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, ErrSessionInvalid
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return nil, ErrSessionInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrSessionInvalid
	}
	var claims SessionClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, ErrSessionInvalid
	}
	return &claims, nil
}
