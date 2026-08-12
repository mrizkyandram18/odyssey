package auth

import (
	"context"
	"net/http"

	"odyssey/pkg/shared"
)

type contextKey string

const (
	sessionClaimsKey contextKey = "session_claims"
	adminUIDKey      contextKey = "admin_uid"
)

// ClaimsFromContext extracts session claims from a context previously
// populated by RequireAuth middleware.
func ClaimsFromContext(ctx context.Context) (*SessionClaims, bool) {
	claims, ok := ctx.Value(sessionClaimsKey).(*SessionClaims)
	return claims, ok
}

// ContextWithClaims attaches session claims to a context for testing or internal dispatch.
func ContextWithClaims(ctx context.Context, claims *SessionClaims) context.Context {
	return context.WithValue(ctx, sessionClaimsKey, claims)
}

// ClaimsFromRequest extracts session claims from the request context.
func ClaimsFromRequest(r *http.Request) (*SessionClaims, bool) {
	return ClaimsFromContext(r.Context())
}

// Middleware wires a SessionValidator into HTTP auth/authorization handlers.
// Construct with NewMiddleware(issuer) where issuer is an HMACSessionIssuer.
type Middleware struct {
	validator SessionValidator
}

func NewMiddleware(v SessionValidator) *Middleware {
	return &Middleware{validator: v}
}

// RequireAuth is the auth middleware: it extracts the session token
// (Bearer header, X-User-Session header, or cookie), validates it,
// and attaches claims to the request context. Missing or invalid tokens
// produce a 401 Unauthorized.
func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := ExtractToken(r)
		if token == "" {
			shared.WriteUnauthorized(w)
			return
		}
		claims, err := m.validator.ParseSession(token)
		if err != nil {
			shared.WriteUnauthorized(w)
			return
		}
		ctx := context.WithValue(r.Context(), sessionClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireSessionKind is authorization middleware that builds on RequireAuth.
// After the session is validated it checks that the session kind matches
// (e.g. "user" vs "setup"). Mismatches return 403 Forbidden.
func (m *Middleware) RequireSessionKind(kind SessionKind) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok || claims.Kind != string(kind) {
				shared.WriteForbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole is authorization middleware that checks the Role claim
// embedded in the session token. Mismatches return 403 Forbidden.
// For admin role, also stores the UID in context under adminUIDKey.
func (m *Middleware) RequireRole(role Role) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok || claims.Role != string(role) {
				shared.WriteForbidden(w)
				return
			}
			ctx := r.Context()
			if role == RoleAdmin {
				ctx = context.WithValue(ctx, adminUIDKey, claims.UID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// IsAdmin checks whether the session claims contain an admin role.
func IsAdmin(r *http.Request) bool {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return false
	}
	return claims.Role == string(RoleAdmin)
}

// IsGuideOrAdmin checks whether the session claims contain an admin or guide role.
func IsGuideOrAdmin(r *http.Request) bool {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return false
	}
	return claims.Role == string(RoleAdmin) || claims.Role == string(RoleGuide)
}

// AdminUIDFromContext extracts the admin UID set by RequireRole(RoleAdmin).
func AdminUIDFromContext(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(adminUIDKey).(string)
	return uid, ok
}
