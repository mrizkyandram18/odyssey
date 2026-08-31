package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestMiddleware(t *testing.T) *Middleware {
	t.Helper()
	return NewMiddleware(NewSessionIssuer("test-secret-key"))
}

func issueTestToken(t *testing.T, issuer *HMACSessionIssuer, kind SessionKind, uid string, cfg *SessionConfig) string {
	t.Helper()
	token, _, err := issuer.IssueSession(kind, uid, cfg)
	if err != nil {
		t.Fatalf("IssueSession: %v", err)
	}
	return token
}

func TestRequireAuth_ValidSession(t *testing.T) {
	mw := newTestMiddleware(t)
	issuer := NewSessionIssuer("test-secret-key")
	token := issueTestToken(t, issuer, SessionKindUser, "alice", &SessionConfig{
		Role: RoleMember, FamilyID: "crew-1",
	})

	called := false
	handler := mw.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			t.Fatal("expected claims in context")
		}
		if claims.UID != "alice" {
			t.Errorf("expected UID alice, got %s", claims.UID)
		}
		if claims.FamilyID != "crew-1" {
			t.Errorf("expected family_id crew-1, got %s", claims.FamilyID)
		}
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler(w, req)

	if !called {
		t.Fatal("handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireAuth_MissingToken(t *testing.T) {
	mw := newTestMiddleware(t)
	called := false
	handler := mw.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if called {
		t.Fatal("handler should not be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	mw := newTestMiddleware(t)
	called := false
	handler := mw.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	handler(w, req)

	if called {
		t.Fatal("handler should not be called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	mw := newTestMiddleware(t)
	issuer := NewSessionIssuer("test-secret-key")
	claims := &SessionClaims{
		Version: 1,
		Kind:    "user",
		UID:     "alice",
		Issued:  time.Now().Add(-2 * time.Hour).Unix(),
		Expires: time.Now().Add(-1 * time.Hour).Unix(),
	}
	token, err := signSession(claims, "test-secret-key")
	if err != nil {
		t.Fatalf("signSession: %v", err)
	}
	_ = issuer

	called := false
	handler := mw.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler(w, req)

	if called {
		t.Fatal("handler should not be called for expired token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_CookieToken(t *testing.T) {
	mw := newTestMiddleware(t)
	issuer := NewSessionIssuer("test-secret-key")
	token := issueTestToken(t, issuer, SessionKindUser, "alice", nil)

	called := false
	handler := mw.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	w := httptest.NewRecorder()
	handler(w, req)

	if !called {
		t.Fatal("handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireSessionKind_Matching(t *testing.T) {
	mw := newTestMiddleware(t)
	issuer := NewSessionIssuer("test-secret-key")
	token := issueTestToken(t, issuer, SessionKindUser, "alice", nil)

	innerCalled := false
	decorated := mw.RequireSessionKind(SessionKindUser)(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	decorated(w, req)

	if !innerCalled {
		t.Fatal("handler should be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireSessionKind_Mismatched(t *testing.T) {
	mw := newTestMiddleware(t)
	issuer := NewSessionIssuer("test-secret-key")
	token := issueTestToken(t, issuer, SessionKindSetup, "alice", nil)

	innerCalled := false
	decorated := mw.RequireSessionKind(SessionKindUser)(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	decorated(w, req)

	if innerCalled {
		t.Fatal("handler should not be called for mismatched kind")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireRole_Matching(t *testing.T) {
	mw := newTestMiddleware(t)
	issuer := NewSessionIssuer("test-secret-key")
	token := issueTestToken(t, issuer, SessionKindUser, "alice", &SessionConfig{Role: RoleAdmin})

	innerCalled := false
	decorated := mw.RequireRole(RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	decorated(w, req)

	if !innerCalled {
		t.Fatal("handler should be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireRole_Mismatched(t *testing.T) {
	mw := newTestMiddleware(t)
	issuer := NewSessionIssuer("test-secret-key")
	token := issueTestToken(t, issuer, SessionKindUser, "alice", &SessionConfig{Role: RoleMember})

	innerCalled := false
	decorated := mw.RequireRole(RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	decorated(w, req)

	if innerCalled {
		t.Fatal("handler should not be called for mismatched role")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireRole_NoRoleInSession(t *testing.T) {
	mw := newTestMiddleware(t)
	issuer := NewSessionIssuer("test-secret-key")
	token := issueTestToken(t, issuer, SessionKindSetup, "alice", nil)

	innerCalled := false
	decorated := mw.RequireRole(RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	decorated(w, req)

	if innerCalled {
		t.Fatal("handler should not be called when role missing from session")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireRole_Unauthenticated(t *testing.T) {
	mw := newTestMiddleware(t)
	innerCalled := false
	decorated := mw.RequireRole(RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	decorated(w, req)

	if innerCalled {
		t.Fatal("handler should not be called without token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestClaimsFromContext_Empty(t *testing.T) {
	claims, ok := ClaimsFromContext(context.Background())
	if ok {
		t.Fatalf("expected false, got true")
	}
	if claims != nil {
		t.Fatalf("expected nil claims, got %v", claims)
	}
}

func TestClaimsFromContext_Populated(t *testing.T) {
	original := &SessionClaims{Version: 1, Kind: "user", UID: "alice"}
	ctx := context.WithValue(context.Background(), sessionClaimsKey, original)
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		t.Fatal("expected true")
	}
	if claims.UID != "alice" {
		t.Errorf("expected UID alice, got %s", claims.UID)
	}
}
