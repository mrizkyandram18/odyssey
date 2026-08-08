package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Mock implementations ---

type mockFirestoreReader struct {
	data     map[string]any
	err      error
	pathArgs []struct{ parentID, uid string }
}

func (m *mockFirestoreReader) GetDeviceDocument(ctx context.Context, parentID, uid string) (map[string]any, error) {
	m.pathArgs = append(m.pathArgs, struct{ parentID, uid string }{parentID, uid})
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

type mockProfileReader struct {
	passwordHash string
	hashErr      error
	boundDevice  string
	bindErr      error
}

func (m *mockProfileReader) GetPasswordHash(ctx context.Context, uid string) (string, error) {
	return m.passwordHash, m.hashErr
}

func (m *mockProfileReader) GetBoundDeviceID(ctx context.Context, uid string) (string, error) {
	return m.boundDevice, m.bindErr
}

type mockPasswordHasher struct {
	err error
}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	return password + "_hashed", nil
}

func (m *mockPasswordHasher) Verify(hashed, password string) error {
	if m.err != nil {
		return m.err
	}
	if hashed != password+"_hashed" {
		return errors.New("hash mismatch")
	}
	return nil
}

func newTestAuthenticator(t *testing.T, store FirestoreReader, profile ProfileReader, hasher PasswordHasher) *FirestoreAuthenticator {
	t.Helper()
	return NewFirestoreAuthenticator("test-parent", "49", hasher, store, profile)
}

// --- Build a valid Gatekeeper Firestore document ---

func validGatekeeperDoc() map[string]any {
	return map[string]any{
		"isOnline": true,
		"lastSeen": time.Now().UTC(),
		"details": map[string]any{
			"appBuildNumber": "49",
			"permissions": map[string]any{
				"battery_exemption":       true,
				"camera":                  true,
				"device_admin":            true,
				"exact_alarm":             true,
				"ignore_battery":          true,
				"microphone":              true,
				"oem_autostart_confirmed": true,
				"overlay":                 true,
				"usage_stats":             true,
				"notification":            false,
			},
		},
	}
}

// --- LoginMethod tests ---

func TestNormalizeLoginMethod(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  LoginMethod
	}{
		{"PASSWORD", LoginMethodPassword},
		{"password", LoginMethodPassword},
		{"Password", LoginMethodPassword},
		{"  PASSWORD  ", LoginMethodPassword},
		{"GATEKEEPER", LoginMethodGatekeeper},
		{"gatekeeper", LoginMethodGatekeeper},
		{"BOTH", LoginMethodBoth},
		{"both", LoginMethodBoth},
		{"", ""},
		{"INVALID", ""},
		{"oauth", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeLoginMethod(tt.input); got != tt.want {
				t.Errorf("NormalizeLoginMethod(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoginMethod_RequiresGatekeeperCompliance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		method LoginMethod
		want   bool
	}{
		{LoginMethodPassword, false},
		{LoginMethodGatekeeper, true},
		{LoginMethodBoth, true},
	}
	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			t.Parallel()
			if got := tt.method.RequiresGatekeeperCompliance(); got != tt.want {
				t.Errorf("RequiresGatekeeperCompliance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoginMethod_RequiresCredential(t *testing.T) {
	t.Parallel()
	tests := []struct {
		method LoginMethod
		want   bool
	}{
		{LoginMethodPassword, true},
		{LoginMethodGatekeeper, false},
		{LoginMethodBoth, true},
	}
	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			t.Parallel()
			if got := tt.method.RequiresCredential(); got != tt.want {
				t.Errorf("RequiresCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Input validation tests ---

func TestVerify_EmptyUID(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "", "secret", device)
	if !errors.Is(err, ErrUIDRequired) {
		t.Fatalf("expected ErrUIDRequired, got %v", err)
	}
}

func TestVerify_WhitespaceUID(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "   ", "secret", device)
	if !errors.Is(err, ErrUIDRequired) {
		t.Fatalf("expected ErrUIDRequired, got %v", err)
	}
}

func TestVerify_InvalidLoginMethod(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "OAUTH", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrLoginMethodInvalid) {
		t.Fatalf("expected ErrLoginMethodInvalid, got %v", err)
	}
}

func TestVerify_EmptyLoginMethod(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrLoginMethodInvalid) {
		t.Fatalf("expected ErrLoginMethodInvalid, got %v", err)
	}
}

func TestVerify_MissingDeviceID(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrDeviceRequired) {
		t.Fatalf("expected ErrDeviceRequired, got %v", err)
	}
}

func TestVerify_EmptyCredential_PASSWORD(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{passwordHash: "abc_hashed"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "PASSWORD", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrCredentialRequired) {
		t.Fatalf("expected ErrCredentialRequired, got %v", err)
	}
}

func TestVerify_EmptyCredential_BOTH(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{passwordHash: "abc_hashed"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrCredentialRequired) {
		t.Fatalf("expected ErrCredentialRequired, got %v", err)
	}
}

// --- PASSWORD mode tests ---

func TestVerify_PASSWORD_ValidCredential(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{passwordHash: "secret_hashed"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "PASSWORD", DeviceID: "dev-1"}
	_, newlyBound, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newlyBound {
		t.Fatal("expected newlyBound=true for unbound device")
	}
}

func TestVerify_PASSWORD_InvalidCredential(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{passwordHash: "secret_hashed"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "PASSWORD", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "wrong", device)
	if !errors.Is(err, ErrCredentialInvalid) {
		t.Fatalf("expected ErrCredentialInvalid, got %v", err)
	}
}

func TestVerify_PASSWORD_HasherError(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{passwordHash: "secret_hashed"}, &mockPasswordHasher{err: errors.New("internal")})
	device := DevicePayload{LoginMethod: "PASSWORD", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrCredentialInvalid) {
		t.Fatalf("expected ErrCredentialInvalid, got %v", err)
	}
}

func TestVerify_PASSWORD_NoStoredHash(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{passwordHash: ""}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "PASSWORD", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrCredentialNotSet) {
		t.Fatalf("expected ErrCredentialNotSet, got %v", err)
	}
}

func TestVerify_PASSWORD_ProfileReaderError(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{hashErr: errors.New("db error")}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "PASSWORD", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrProfileUnavailable) {
		t.Fatalf("expected ErrProfileUnavailable, got %v", err)
	}
}

func TestVerify_PASSWORD_DoesNotReadFirestore(t *testing.T) {
	t.Parallel()
	store := &mockFirestoreReader{}
	auth := newTestAuthenticator(t, store, &mockProfileReader{passwordHash: "secret_hashed"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "PASSWORD", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.pathArgs) > 0 {
		t.Fatal("PASSWORD mode should not read Firestore")
	}
}

// --- PASSWORD mode: device binding ---

func TestVerify_PASSWORD_AlreadyBound(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{passwordHash: "secret_hashed", boundDevice: "dev-1"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "PASSWORD", DeviceID: "dev-1"}
	_, newlyBound, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newlyBound {
		t.Fatal("expected newlyBound=false for already-bound device")
	}
}

func TestVerify_PASSWORD_DeviceMismatch(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{}, &mockProfileReader{passwordHash: "secret_hashed", boundDevice: "dev-1"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "PASSWORD", DeviceID: "dev-2"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrDeviceMismatch) {
		t.Fatalf("expected ErrDeviceMismatch, got %v", err)
	}
}

// --- GATEKEEPER mode tests ---

func TestVerify_GATEKEEPER_ValidCompliance(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, newlyBound, err := auth.Verify(context.Background(), "user-1", "", device)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newlyBound {
		t.Fatal("expected newlyBound=true for new device")
	}
}

func TestVerify_GATEKEEPER_IgnoresCredential(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{boundDevice: "dev-1"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "ignored", device)
	if err != nil {
		t.Fatalf("GATEKEEPER mode should not verify credential: %v", err)
	}
}

func TestVerify_GATEKEEPER_DeviceNotFound(t *testing.T) {
	t.Parallel()
	store := &mockFirestoreReader{err: status.Error(codes.NotFound, "document not found")}
	auth := newTestAuthenticator(t, store, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrGatekeeperNotFound) {
		t.Fatalf("expected ErrGatekeeperNotFound, got %v", err)
	}
}

func TestVerify_GATEKEEPER_FirestoreUnavailable(t *testing.T) {
	t.Parallel()
	store := &mockFirestoreReader{err: status.Error(codes.ResourceExhausted, "quota exceeded")}
	auth := newTestAuthenticator(t, store, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrFirestoreUnavailable) {
		t.Fatalf("expected ErrFirestoreUnavailable, got %v", err)
	}
}

func TestVerify_GATEKEEPER_PermissionDenied(t *testing.T) {
	t.Parallel()
	store := &mockFirestoreReader{err: status.Error(codes.PermissionDenied, "denied")}
	auth := newTestAuthenticator(t, store, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrFirestoreUnavailable) {
		t.Fatalf("expected ErrFirestoreUnavailable, got %v", err)
	}
}

func TestVerify_GATEKEEPER_Unavailable(t *testing.T) {
	t.Parallel()
	store := &mockFirestoreReader{err: status.Error(codes.Unavailable, "server down")}
	auth := newTestAuthenticator(t, store, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrFirestoreUnavailable) {
		t.Fatalf("expected ErrFirestoreUnavailable, got %v", err)
	}
}

func TestVerify_GATEKEEPER_NonGRPCError(t *testing.T) {
	t.Parallel()
	store := &mockFirestoreReader{err: errors.New("some network error")}
	auth := newTestAuthenticator(t, store, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrFirestoreUnavailable) {
		t.Fatalf("expected ErrFirestoreUnavailable, got %v", err)
	}
}

func TestVerify_GATEKEEPER_DeviceOffline(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	doc["isOnline"] = false
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrDeviceOffline) {
		t.Fatalf("expected ErrDeviceOffline, got %v", err)
	}
}

func TestVerify_GATEKEEPER_DeviceStale(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	doc["lastSeen"] = time.Now().Add(-10 * time.Minute)
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrDeviceOffline) {
		t.Fatalf("expected ErrDeviceOffline for stale device, got %v", err)
	}
}

func TestVerify_GATEKEEPER_MissingLastSeen(t *testing.T) {
	t.Parallel()
	doc := map[string]any{
		"isOnline": true,
		"details": map[string]any{
			"appBuildNumber": "49",
			"permissions":    map[string]any{"camera": true},
		},
	}
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrDeviceOffline) {
		t.Fatalf("expected ErrDeviceOffline for missing lastSeen, got %v", err)
	}
}

func TestVerify_GATEKEEPER_BuildTooOld(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	doc["details"].(map[string]any)["appBuildNumber"] = "47"
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrBuildTooOld) {
		t.Fatalf("expected ErrBuildTooOld, got %v", err)
	}
}

func TestVerify_GATEKEEPER_BuildPrefixV(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	doc["details"].(map[string]any)["appBuildNumber"] = "v49"
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if err != nil {
		t.Fatalf("expected no error for v49, got %v", err)
	}
}

func TestVerify_GATEKEEPER_BuildHigherThanMin(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	doc["details"].(map[string]any)["appBuildNumber"] = "50"
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if err != nil {
		t.Fatalf("expected no error for build 50 >= 49, got %v", err)
	}
}

func TestVerify_GATEKEEPER_BuildNonNumeric(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	doc["details"].(map[string]any)["appBuildNumber"] = "abc"
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrBuildTooOld) {
		t.Fatalf("expected ErrBuildTooOld for non-numeric build, got %v", err)
	}
}

func TestVerify_GATEKEEPER_PermissionsMissing(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	doc["details"].(map[string]any)["permissions"].(map[string]any)["camera"] = false
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrPermissionsMissing) {
		t.Fatalf("expected ErrPermissionsMissing, got %v", err)
	}
}

func TestVerify_GATEKEEPER_PermissionsMissingKey(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	delete(doc["details"].(map[string]any)["permissions"].(map[string]any), "camera")
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrPermissionsMissing) {
		t.Fatalf("expected ErrPermissionsMissing, got %v", err)
	}
}

func TestVerify_GATEKEEPER_NilPermissions(t *testing.T) {
	t.Parallel()
	doc := map[string]any{
		"isOnline": true,
		"lastSeen": time.Now().UTC(),
		"details": map[string]any{
			"appBuildNumber": "49",
		},
	}
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrPermissionsMissing) {
		t.Fatalf("expected ErrPermissionsMissing, got %v", err)
	}
}

func TestVerify_GATEKEEPER_ExplicitFalseNonRequiredPermission(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	perms := doc["details"].(map[string]any)["permissions"].(map[string]any)
	perms["accessibility"] = false
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrPermissionsMissing) {
		t.Fatalf("expected ErrPermissionsMissing for explicit false non-required permission, got %v", err)
	}
}

func TestVerify_GATEKEEPER_FalseNotificationAllowed(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	doc["details"].(map[string]any)["permissions"].(map[string]any)["notification"] = false
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if err != nil {
		t.Fatalf("notification false should be allowed: %v", err)
	}
}

func TestVerify_GATEKEEPER_NilDocument(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: nil}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrGatekeeperNotFound) {
		t.Fatalf("expected ErrGatekeeperNotFound, got %v", err)
	}
}

// --- BOTH mode tests ---

func TestVerify_BOTH_ValidComplianceAndCredential(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{passwordHash: "secret_hashed"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, newlyBound, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newlyBound {
		t.Fatal("expected newlyBound=true for new device")
	}
}

func TestVerify_BOTH_ComplianceFails(t *testing.T) {
	t.Parallel()
	doc := validGatekeeperDoc()
	doc["isOnline"] = false
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{passwordHash: "secret_hashed"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrDeviceOffline) {
		t.Fatalf("expected ErrDeviceOffline, got %v", err)
	}
}

func TestVerify_BOTH_CredentialFails(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{passwordHash: "secret_hashed"}, &mockPasswordHasher{err: errors.New("hash mismatch")})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "wrong", device)
	if !errors.Is(err, ErrCredentialInvalid) {
		t.Fatalf("expected ErrCredentialInvalid, got %v", err)
	}
}

func TestVerify_BOTH_NoStoredHash(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{passwordHash: ""}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrCredentialNotSet) {
		t.Fatalf("expected ErrCredentialNotSet, got %v", err)
	}
}

func TestVerify_BOTH_CompliesBeforeCredential(t *testing.T) {
	t.Parallel()
	// If compliance fails, credential verification should not be attempted.
	store := &mockFirestoreReader{data: validGatekeeperDoc()}
	doc := validGatekeeperDoc()
	doc["isOnline"] = false
	store.data = doc
	profile := &mockProfileReader{passwordHash: "secret_hashed"}
	auth := newTestAuthenticator(t, store, profile, &mockPasswordHasher{err: errors.New("should not be called")})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrDeviceOffline) {
		t.Fatalf("expected ErrDeviceOffline, got %v", err)
	}
}

func TestVerify_BOTH_AlreadyBound(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{passwordHash: "secret_hashed", boundDevice: "dev-1"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, newlyBound, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newlyBound {
		t.Fatal("expected newlyBound=false for already-bound device")
	}
}

func TestVerify_BOTH_DeviceMismatch(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{passwordHash: "secret_hashed", boundDevice: "dev-1"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-2"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrDeviceMismatch) {
		t.Fatalf("expected ErrDeviceMismatch, got %v", err)
	}
}

func TestVerify_BOTH_FirestoreErrorMapping(t *testing.T) {
	t.Parallel()
	store := &mockFirestoreReader{err: status.Error(codes.NotFound, "not found")}
	auth := newTestAuthenticator(t, store, &mockProfileReader{passwordHash: "secret_hashed"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrGatekeeperNotFound) {
		t.Fatalf("expected ErrGatekeeperNotFound, got %v", err)
	}
}

// --- Case-insensitive login method ---

func TestVerify_LoginMethodCaseInsensitive(t *testing.T) {
	t.Parallel()
	for _, method := range []string{"password", "Password", "PASSWORD", "gatekeeper", "GATEKEEPER", "both", "BOTH"} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			upper := strings.ToUpper(strings.TrimSpace(method))
			// BOTH and GATEKEEPER require a valid Gatekeeper document so the
			// compliance check passes before the credential check.
			store := &mockFirestoreReader{data: validGatekeeperDoc()}
			profile := &mockProfileReader{}
			auth := newTestAuthenticator(t, store, profile, &mockPasswordHasher{})
			device := DevicePayload{LoginMethod: method, DeviceID: "dev-1"}
			_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
			switch upper {
			case "GATEKEEPER":
				// Compliance passes, credential is skipped, device is newly bound.
				if err != nil {
					t.Fatalf("expected no error for %s with valid doc, got %v", method, err)
				}
			case "PASSWORD", "BOTH":
				// Compliance (BOTH) or no Firestore (PASSWORD) passes, but no
				// hash is stored → credential not set.
				if !errors.Is(err, ErrCredentialNotSet) {
					t.Fatalf("expected ErrCredentialNotSet for %s with no hash, got %v", method, err)
				}
			}
		})
	}
}

// --- fromFirestore mapping tests ---

func TestFromFirestore_TopLevelFields(t *testing.T) {
	t.Parallel()
	raw := map[string]any{
		"isOnline": true,
		"lastSeen": time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
		"details": map[string]any{
			"appBuildNumber": "49",
		},
	}
	id := fromFirestore(raw, "user-1", LoginMethodBoth)
	if id.UID != "user-1" {
		t.Errorf("expected UID user-1, got %s", id.UID)
	}
	if !id.IsOnline {
		t.Error("expected IsOnline=true")
	}
	if id.BuildNumber != "49" {
		t.Errorf("expected BuildNumber 49, got %s", id.BuildNumber)
	}
	if !id.LastSeen.Equal(time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected LastSeen match, got %v", id.LastSeen)
	}
}

func TestFromFirestore_NestedInDetails(t *testing.T) {
	t.Parallel()
	raw := map[string]any{
		"details": map[string]any{
			"isOnline":       true,
			"lastSeen":       time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
			"appBuildNumber": "49",
			"permissions": map[string]any{
				"camera":       true,
				"microphone":   true,
				"notification": false,
			},
		},
	}
	id := fromFirestore(raw, "user-1", LoginMethodGatekeeper)
	if !id.IsOnline {
		t.Error("expected IsOnline from details")
	}
	if id.BuildNumber != "49" {
		t.Errorf("expected BuildNumber from details, got %s", id.BuildNumber)
	}
	if !id.LastSeen.Equal(time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected LastSeen from details, got %v", id.LastSeen)
	}
	if !id.Permissions["camera"] {
		t.Error("expected camera permission true")
	}
	if !id.Permissions["microphone"] {
		t.Error("expected microphone permission true")
	}
	if id.Permissions["notification"] {
		t.Error("expected notification permission false")
	}
}

func TestFromFirestore_TopLevelPermissions(t *testing.T) {
	t.Parallel()
	raw := map[string]any{
		"isOnline": true,
		"lastSeen": time.Now().UTC(),
		"details": map[string]any{
			"appBuildNumber": "49",
		},
		"permissions": map[string]any{
			"camera": true,
		},
	}
	id := fromFirestore(raw, "user-1", LoginMethodBoth)
	if !id.Permissions["camera"] {
		t.Error("expected top-level permissions to be mapped")
	}
}

func TestFromFirestore_StringLastSeen(t *testing.T) {
	t.Parallel()
	raw := map[string]any{
		"details": map[string]any{
			"appBuildNumber": "49",
			"lastSeen":       "2026-08-03T00:00:00Z",
			"permissions":    map[string]any{"camera": true},
		},
	}
	id := fromFirestore(raw, "user-1", LoginMethodBoth)
	if id.LastSeen.IsZero() {
		t.Error("expected LastSeen to be parsed from RFC3339 string")
	}
	expected := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	if !id.LastSeen.Equal(expected) {
		t.Errorf("expected LastSeen %v, got %v", expected, id.LastSeen)
	}
}

func TestFromFirestore_TopLevelLastSeenPrecedence(t *testing.T) {
	t.Parallel()
	raw := map[string]any{
		"lastSeen": time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
		"details": map[string]any{
			"lastSeen":       time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
			"appBuildNumber": "49",
		},
	}
	id := fromFirestore(raw, "user-1", LoginMethodBoth)
	if !id.LastSeen.Equal(time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("expected top-level LastSeen to take precedence, got %v", id.LastSeen)
	}
}

func TestFromFirestore_MissingFields(t *testing.T) {
	t.Parallel()
	raw := map[string]any{}
	id := fromFirestore(raw, "user-1", LoginMethodBoth)
	if id.IsOnline {
		t.Error("expected IsOnline=false for empty doc")
	}
	if !id.LastSeen.IsZero() {
		t.Error("expected LastSeen zero for empty doc")
	}
	if id.BuildNumber != "" {
		t.Errorf("expected empty BuildNumber, got %s", id.BuildNumber)
	}
	if len(id.Permissions) != 0 {
		t.Errorf("expected empty permissions, got %v", id.Permissions)
	}
}

func TestFromFirestore_DetailsNotMap(t *testing.T) {
	t.Parallel()
	raw := map[string]any{
		"isOnline": true,
		"details":  "not a map",
	}
	id := fromFirestore(raw, "user-1", LoginMethodBoth)
	if !id.IsOnline {
		t.Error("expected IsOnline from top-level")
	}
}

// --- Build number validation tests ---

func TestValidateBuildNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		actual  string
		min     string
		wantErr bool
	}{
		{"equal", "49", "49", false},
		{"higher", "50", "49", false},
		{"lower", "47", "49", true},
		{"v prefix equal", "v49", "49", false},
		{"V prefix equal", "V49", "49", false},
		{"min with v prefix", "49", "v49", false},
		{"non-numeric", "abc", "49", true},
		{"empty actual", "", "49", true},
		{"both empty", "", "", true},
		{"spaces", " 49 ", "49", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateBuildNumber(tt.actual, tt.min)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if err != nil && !errors.Is(err, ErrBuildTooOld) {
				t.Fatalf("expected ErrBuildTooOld, got %v", err)
			}
		})
	}
}

// --- Permissions validation tests ---

func TestValidatePermissions(t *testing.T) {
	t.Parallel()
	validPerms := func() map[string]bool {
		return map[string]bool{
			"battery_exemption":       true,
			"camera":                  true,
			"device_admin":            true,
			"exact_alarm":             true,
			"ignore_battery":          true,
			"microphone":              true,
			"oem_autostart_confirmed": true,
			"overlay":                 true,
			"usage_stats":             true,
		}
	}

	t.Run("valid with notification false", func(t *testing.T) {
		t.Parallel()
		perms := validPerms()
		perms["notification"] = false
		if err := validatePermissions(perms); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("valid with notification true", func(t *testing.T) {
		t.Parallel()
		perms := validPerms()
		perms["notification"] = true
		if err := validatePermissions(perms); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("missing required permission", func(t *testing.T) {
		t.Parallel()
		perms := validPerms()
		delete(perms, "camera")
		if err := validatePermissions(perms); !errors.Is(err, ErrPermissionsMissing) {
			t.Fatalf("expected ErrPermissionsMissing, got %v", err)
		}
	})

	t.Run("false required permission", func(t *testing.T) {
		t.Parallel()
		perms := validPerms()
		perms["camera"] = false
		if err := validatePermissions(perms); !errors.Is(err, ErrPermissionsMissing) {
			t.Fatalf("expected ErrPermissionsMissing, got %v", err)
		}
	})

	t.Run("nil permissions", func(t *testing.T) {
		t.Parallel()
		if err := validatePermissions(nil); !errors.Is(err, ErrPermissionsMissing) {
			t.Fatalf("expected ErrPermissionsMissing, got %v", err)
		}
	})

	t.Run("explicit false non-required permission", func(t *testing.T) {
		t.Parallel()
		perms := validPerms()
		perms["accessibility"] = false
		if err := validatePermissions(perms); !errors.Is(err, ErrPermissionsMissing) {
			t.Fatalf("expected ErrPermissionsMissing, got %v", err)
		}
	})
}

// --- Compliance validation tests ---

func TestValidateCompliance(t *testing.T) {
	t.Parallel()

	validIdentity := func() *Identity {
		return &Identity{
			IsOnline:    true,
			LastSeen:    time.Now().UTC(),
			BuildNumber: "49",
			Permissions: map[string]bool{
				"battery_exemption":       true,
				"camera":                  true,
				"device_admin":            true,
				"exact_alarm":             true,
				"ignore_battery":          true,
				"microphone":              true,
				"oem_autostart_confirmed": true,
				"overlay":                 true,
				"usage_stats":             true,
			},
		}
	}

	t.Run("valid compliance", func(t *testing.T) {
		t.Parallel()
		if err := validateCompliance(validIdentity(), "49"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("offline", func(t *testing.T) {
		t.Parallel()
		id := validIdentity()
		id.IsOnline = false
		if err := validateCompliance(id, "49"); !errors.Is(err, ErrDeviceOffline) {
			t.Fatalf("expected ErrDeviceOffline, got %v", err)
		}
	})

	t.Run("stale lastSeen", func(t *testing.T) {
		t.Parallel()
		id := validIdentity()
		id.LastSeen = time.Now().Add(-10 * time.Minute)
		if err := validateCompliance(id, "49"); !errors.Is(err, ErrDeviceOffline) {
			t.Fatalf("expected ErrDeviceOffline, got %v", err)
		}
	})

	t.Run("zero lastSeen", func(t *testing.T) {
		t.Parallel()
		id := validIdentity()
		id.LastSeen = time.Time{}
		if err := validateCompliance(id, "49"); !errors.Is(err, ErrDeviceOffline) {
			t.Fatalf("expected ErrDeviceOffline, got %v", err)
		}
	})

	t.Run("build too old", func(t *testing.T) {
		t.Parallel()
		id := validIdentity()
		id.BuildNumber = "40"
		if err := validateCompliance(id, "49"); !errors.Is(err, ErrBuildTooOld) {
			t.Fatalf("expected ErrBuildTooOld, got %v", err)
		}
	})

	t.Run("permissions missing", func(t *testing.T) {
		t.Parallel()
		id := validIdentity()
		delete(id.Permissions, "camera")
		if err := validateCompliance(id, "49"); !errors.Is(err, ErrPermissionsMissing) {
			t.Fatalf("expected ErrPermissionsMissing, got %v", err)
		}
	})
}

// --- Error mapping tests ---

func TestMapFirestoreError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{"nil", nil, nil},
		{"not found", status.Error(codes.NotFound, "doc not found"), ErrGatekeeperNotFound},
		{"resource exhausted", status.Error(codes.ResourceExhausted, "quota"), ErrFirestoreUnavailable},
		{"permission denied", status.Error(codes.PermissionDenied, "denied"), ErrFirestoreUnavailable},
		{"unavailable", status.Error(codes.Unavailable, "down"), ErrFirestoreUnavailable},
		{"internal", status.Error(codes.Internal, "oops"), ErrFirestoreUnavailable},
		{"non-gRPC", errors.New("network"), ErrFirestoreUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := mapFirestoreError(tt.err)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

// --- Device payload validation integration ---

func TestVerify_GATEKEEPER_AlreadyBoundMatching(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{boundDevice: "dev-1"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, newlyBound, err := auth.Verify(context.Background(), "user-1", "", device)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newlyBound {
		t.Fatal("expected newlyBound=false for bound device")
	}
}

func TestVerify_GATEKEEPER_AlreadyBoundMismatch(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{boundDevice: "dev-1"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-2"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrDeviceMismatch) {
		t.Fatalf("expected ErrDeviceMismatch, got %v", err)
	}
}

func TestVerify_GATEKEEPER_ProfileReaderError(t *testing.T) {
	t.Parallel()
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: validGatekeeperDoc()}, &mockProfileReader{bindErr: errors.New("db down")}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrProfileUnavailable) {
		t.Fatalf("expected ErrProfileUnavailable, got %v", err)
	}
}

// --- Firestore path construction ---

func TestVerify_FirestorePathConstruction(t *testing.T) {
	t.Parallel()
	store := &mockFirestoreReader{data: validGatekeeperDoc()}
	auth := newTestAuthenticator(t, store, &mockProfileReader{passwordHash: "secret_hashed"}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, _, _ = auth.Verify(context.Background(), "user-1", "secret", device)

	if len(store.pathArgs) != 1 {
		t.Fatalf("expected 1 Firestore call, got %d", len(store.pathArgs))
	}
	call := store.pathArgs[0]
	if call.parentID != "test-parent" {
		t.Errorf("expected parentID test-parent, got %s", call.parentID)
	}
	if call.uid != "user-1" {
		t.Errorf("expected uid user-1, got %s", call.uid)
	}
}

// --- Interface compliance ---

func TestFirestoreAuthenticator_ImplementsAuthenticator(t *testing.T) {
	var _ Authenticator = (*FirestoreAuthenticator)(nil)
}

func TestFirestoreStore_ImplementsFirestoreReader(t *testing.T) {
	var _ FirestoreReader = (*firestoreStore)(nil)
}

func TestMockPasswordHasher_ImplementsPasswordHasher(t *testing.T) {
	var _ PasswordHasher = (*mockPasswordHasher)(nil)
}

// --- Configuration / constructor tests ---

func TestNewFirestoreAuthenticator_SetsFields(t *testing.T) {
	t.Parallel()
	store := &mockFirestoreReader{data: validGatekeeperDoc()}
	profile := &mockProfileReader{passwordHash: "secret_hashed"}
	hasher := &mockPasswordHasher{}
	auth := NewFirestoreAuthenticator("parent-1", "v49", hasher, store, profile)
	if auth.parentID != "parent-1" {
		t.Errorf("expected parentID parent-1, got %s", auth.parentID)
	}
	if auth.minBuildNumber != "49" {
		t.Errorf("expected minBuildNumber 49 (normalized), got %s", auth.minBuildNumber)
	}
}

func TestNewFirestoreAuthenticator_DefaultBuildNumber(t *testing.T) {
	t.Parallel()
	auth := NewFirestoreAuthenticator("parent-1", "", &mockPasswordHasher{}, &mockFirestoreReader{}, &mockProfileReader{})
	if auth.minBuildNumber != "" {
		t.Errorf("expected empty minBuildNumber when not provided, got %s", auth.minBuildNumber)
	}
}

func TestVerify_BOTH_CompliancePasses_BeforeBindingCheck(t *testing.T) {
	t.Parallel()
	// Device has valid compliance and valid credential, but profile reader
	// errors on binding lookup. Binding error should not mask compliance success.
	store := &mockFirestoreReader{data: validGatekeeperDoc()}
	profile := &mockProfileReader{
		passwordHash: "secret_hashed",
		bindErr:      errors.New("db down"),
	}
	auth := newTestAuthenticator(t, store, profile, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "BOTH", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "secret", device)
	if !errors.Is(err, ErrProfileUnavailable) {
		t.Fatalf("expected ErrProfileUnavailable from binding error, got %v", err)
	}
}

func TestVerify_MinuteBoundary(t *testing.T) {
	t.Parallel()
	// lastSeen just under 5 minutes ago should pass (boundary)
	doc := validGatekeeperDoc()
	doc["lastSeen"] = time.Now().Add(-(5*time.Minute - 2*time.Second))
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if err != nil {
		t.Fatalf("expected no error at 5min boundary, got %v", err)
	}
}

func TestVerify_JustPastBoundary(t *testing.T) {
	t.Parallel()
	// lastSeen 5min + 1s ago should fail
	doc := validGatekeeperDoc()
	doc["lastSeen"] = time.Now().Add(-(5*time.Minute + time.Second))
	auth := newTestAuthenticator(t, &mockFirestoreReader{data: doc}, &mockProfileReader{}, &mockPasswordHasher{})
	device := DevicePayload{LoginMethod: "GATEKEEPER", DeviceID: "dev-1"}
	_, _, err := auth.Verify(context.Background(), "user-1", "", device)
	if !errors.Is(err, ErrDeviceOffline) {
		t.Fatalf("expected ErrDeviceOffline past boundary, got %v", err)
	}
}
