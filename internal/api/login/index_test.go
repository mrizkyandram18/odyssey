package login

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
)

type mockAuthenticator struct {
	err        error
	newlyBound bool
}

func (m *mockAuthenticator) Verify(ctx context.Context, uid, credential string, device auth.DevicePayload) (string, bool, error) {
	return uid, m.newlyBound, m.err
}

type mockIssuer struct {
	token  string
	claims *auth.SessionClaims
	err    error
}

func (m *mockIssuer) IssueSession(kind auth.SessionKind, uid string, cfg *auth.SessionConfig) (string, *auth.SessionClaims, error) {
	return m.token, m.claims, m.err
}

func (m *mockIssuer) ParseSession(token string) (*auth.SessionClaims, error) {
	return m.claims, m.err
}

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

func setupDeps(a auth.Authenticator, s auth.SessionIssuer, p db.ProfileStore) {
	Setup(a, s, p)
}

func makeValidClaims() *auth.SessionClaims {
	return &auth.SessionClaims{
		Version: 1,
		Kind:    "user",
		UID:     "user-1",
		Role:    "SEEKER",
		CrewID:  "crew-1",
		Issued:  1000,
		Expires: 9999999999,
	}
}

func makeValidProfile() *db.UserProfile {
	return &db.UserProfile{
		UID:          "user-1",
		CrewID:       "crew-1",
		ExplorerName: "Alice",
		Role:         "SEEKER",
		Level:        1,
		XP:           0,
	}
}

func TestLoginHandler_MethodNotAllowed(t *testing.T) {
	setupDeps(&mockAuthenticator{}, &mockIssuer{}, &mockProfileStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/login", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	setupDeps(
		&mockAuthenticator{newlyBound: true},
		&mockIssuer{token: "valid-token", claims: makeValidClaims()},
		&mockProfileStore{profile: makeValidProfile()},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp loginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Status != "success" {
		t.Errorf("expected status success, got %s", resp.Status)
	}
	if resp.UID != "user-1" {
		t.Errorf("expected uid user-1, got %s", resp.UID)
	}
	if resp.CrewID != "crew-1" {
		t.Errorf("expected crew_id crew-1, got %s", resp.CrewID)
	}
	if resp.Kind != "user" {
		t.Errorf("expected kind user, got %s", resp.Kind)
	}
	if resp.Role != "SEEKER" {
		t.Errorf("expected role SEEKER, got %s", resp.Role)
	}
}

func TestLoginHandler_PasswordRequired(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrCredentialRequired},
		&mockIssuer{claims: makeValidClaims()},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var resp loginResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Status != "password_required" {
		t.Errorf("expected password_required, got %s", resp.Status)
	}
	if resp.UID != "user-1" {
		t.Errorf("expected uid user-1, got %s", resp.UID)
	}
}

func TestLoginHandler_SetupNeeded(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrCredentialNotSet},
		&mockIssuer{token: "setup-token", claims: makeValidClaims()},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp loginResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Status != "setup_needed" {
		t.Errorf("expected setup_needed, got %s", resp.Status)
	}
	if resp.SetupToken != "setup-token" {
		t.Errorf("expected setup_token, got %s", resp.SetupToken)
	}
	if resp.Kind != "setup" {
		t.Errorf("expected kind setup, got %s", resp.Kind)
	}
}

func TestLoginHandler_InvalidCredential(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrCredentialInvalid},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"wrong","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLoginHandler_DeviceOffline(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrDeviceOffline},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLoginHandler_BuildTooOld(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrBuildTooOld},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLoginHandler_PermissionsMissing(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrPermissionsMissing},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLoginHandler_DeviceMismatch(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrDeviceMismatch},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLoginHandler_GatekeeperNotFound(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrGatekeeperNotFound},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLoginHandler_FirestoreUnavailable(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrFirestoreUnavailable},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestLoginHandler_ProfileUnavailable(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrProfileUnavailable},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestLoginHandler_InvalidLoginMethod(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrLoginMethodInvalid},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"OAUTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLoginHandler_UIDRequired(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrUIDRequired},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLoginHandler_DeviceRequired(t *testing.T) {
	setupDeps(
		&mockAuthenticator{err: auth.ErrDeviceRequired},
		&mockIssuer{},
		&mockProfileStore{},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLoginHandler_BadJSON(t *testing.T) {
	setupDeps(
		&mockAuthenticator{},
		&mockIssuer{},
		&mockProfileStore{},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLoginHandler_NotConfigured(t *testing.T) {
	Setup(nil, nil, nil)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestLoginHandler_SetsCookie(t *testing.T) {
	setupDeps(
		&mockAuthenticator{newlyBound: true},
		&mockIssuer{token: "valid-token", claims: makeValidClaims()},
		&mockProfileStore{profile: makeValidProfile()},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == auth.SessionCookieName {
			found = true
			if c.Value != "valid-token" {
				t.Errorf("expected cookie value valid-token, got %s", c.Value)
			}
		}
	}
	if !found {
		t.Error("expected odyssey_session cookie to be set")
	}
}

func TestLoginHandler_ProfileLookupFailureStillIssuesSession(t *testing.T) {
	setupDeps(
		&mockAuthenticator{newlyBound: true},
		&mockIssuer{token: "valid-token", claims: makeValidClaims()},
		&mockProfileStore{err: errors.New("db down")},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp loginResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Status != "success" {
		t.Errorf("expected status success, got %s", resp.Status)
	}
}

func TestLoginHandler_IssueSessionError(t *testing.T) {
	setupDeps(
		&mockAuthenticator{newlyBound: true},
		&mockIssuer{err: errors.New("issuer broken")},
		&mockProfileStore{profile: makeValidProfile()},
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestLoginHandler_SuccessNoProfile(t *testing.T) {
	setupDeps(
		&mockAuthenticator{newlyBound: true},
		&mockIssuer{token: "valid-token", claims: makeValidClaims()},
		&mockProfileStore{}, // nil profile, no error
	)

	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web-pwa","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp loginResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Status != "success" {
		t.Errorf("expected status success, got %s", resp.Status)
	}
	if resp.CrewID != "" && resp.CrewID != "crew-1" {
		t.Errorf("expected empty crew_id for no profile, got %s", resp.CrewID)
	}
}
