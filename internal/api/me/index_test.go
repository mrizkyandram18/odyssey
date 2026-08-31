package me

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
)

type mockProfileStore struct {
	profile *db.UserProfile
	err     error
}

func (m *mockProfileStore) GetUserProfile(ctx context.Context, uid string) (*db.UserProfile, error) {
	return m.profile, m.err
}

func (m *mockProfileStore) GetPasswordHash(ctx context.Context, uid string) (string, error) {
	if m.profile != nil {
		return m.profile.PasswordHash, nil
	}
	return "", nil
}

func (m *mockProfileStore) GetBoundDeviceID(ctx context.Context, uid string) (string, error) {
	if m.profile != nil {
		return m.profile.DeviceID, nil
	}
	return "", nil
}

func (m *mockProfileStore) BindOrVerifyDevice(ctx context.Context, uid, deviceID string) (bool, error) {
	if m.profile != nil {
		if m.profile.DeviceID == "" {
			m.profile.DeviceID = deviceID
			return true, nil
		}
		if m.profile.DeviceID == deviceID {
			return false, nil
		}
		return false, auth.ErrDeviceBlocked
	}
	return false, nil
}

func (m *mockProfileStore) ResetDeviceBinding(ctx context.Context, uid string) error {
	if m.profile != nil {
		m.profile.DeviceID = ""
	}
	return nil
}

func (m *mockProfileStore) SetAvatarFrame(ctx context.Context, uid, frame string) error {
	return nil
}

func (m *mockProfileStore) SetExplorerEffect(ctx context.Context, uid, effect string) error {
	return nil
}

func (m *mockProfileStore) UpdateAvatar(ctx context.Context, uid string, style, seed string) error {
	if m.err != nil {
		return m.err
	}
	if m.profile != nil {
		m.profile.AvatarStyle = style
		m.profile.AvatarSeed = seed
	}
	return nil
}

func makeProfile() *db.UserProfile {
	return &db.UserProfile{
		UID:          "user-1",
		FamilyID:     "crew-1",
		ExplorerName: "Alice",
		Role:         "SEEKER",
		Level:        1,
		XP:           100,
	}
}

func makeUserToken(t *testing.T, issuer *auth.HMACSessionIssuer) string {
	t.Helper()
	token, _, err := issuer.IssueSession(auth.SessionKindUser, "user-1", &auth.SessionConfig{
		Role:     auth.RoleSeeker,
		FamilyID: "crew-1",
	})
	if err != nil {
		t.Fatalf("IssueSession: %v", err)
	}
	return token
}

func makeSetupToken(t *testing.T, issuer *auth.HMACSessionIssuer) string {
	t.Helper()
	token, _, err := issuer.IssueSession(auth.SessionKindSetup, "user-1", nil)
	if err != nil {
		t.Fatalf("IssueSession: %v", err)
	}
	return token
}

func TestMeHandler_MethodNotAllowed(t *testing.T) {
	Setup(&mockProfileStore{profile: makeProfile()})

	req := httptest.NewRequest(http.MethodPost, "/api/me", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestMeHandler_Success(t *testing.T) {
	Setup(&mockProfileStore{profile: makeProfile()})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var profile db.UserProfile
	if err := json.Unmarshal(w.Body.Bytes(), &profile); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if profile.UID != "user-1" {
		t.Errorf("expected UID user-1, got %s", profile.UID)
	}
	if profile.FamilyID != "crew-1" {
		t.Errorf("expected FamilyID crew-1, got %s", profile.FamilyID)
	}
	if profile.ExplorerName != "Alice" {
		t.Errorf("expected ExplorerName Alice, got %s", profile.ExplorerName)
	}
	if profile.Role != "SEEKER" {
		t.Errorf("expected Role SEEKER, got %s", profile.Role)
	}
	if profile.Level != 1 {
		t.Errorf("expected Level 1, got %d", profile.Level)
	}
	if profile.XP != 100 {
		t.Errorf("expected XP 100, got %d", profile.XP)
	}
}

func TestMeHandler_NoCookieAndPassword(t *testing.T) {
	Setup(&mockProfileStore{profile: makeProfile()})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMeHandler_InvalidToken(t *testing.T) {
	Setup(&mockProfileStore{profile: makeProfile()})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMeHandler_SetupTokenRejected(t *testing.T) {
	Setup(&mockProfileStore{profile: makeProfile()})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeSetupToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for setup token, got %d", w.Code)
	}
}

func TestMeHandler_ProfileNotFound(t *testing.T) {
	Setup(&mockProfileStore{err: db.ErrProfileNotFound})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestMeHandler_ProfileError(t *testing.T) {
	Setup(&mockProfileStore{err: errTestProfile})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestMeHandler_NotConfigured(t *testing.T) {
	Setup(nil)
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

var errTestProfile = errProfile{}

type errProfile struct{}

func (errProfile) Error() string { return "test profile error" }

func TestMeHandler_PatchAvatar(t *testing.T) {
	mockStore := &mockProfileStore{profile: makeProfile()}
	Setup(mockStore)
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	payload := `{"avatar_style": "adventurer", "avatar_seed": "my-seed"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/me/avatar", bytes.NewReader([]byte(payload)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if mockStore.profile.AvatarStyle != "adventurer" {
		t.Errorf("expected avatar_style adventurer, got %s", mockStore.profile.AvatarStyle)
	}
	if mockStore.profile.AvatarSeed != "my-seed" {
		t.Errorf("expected avatar_seed my-seed, got %s", mockStore.profile.AvatarSeed)
	}
}
