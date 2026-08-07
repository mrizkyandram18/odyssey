package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Domain error sentinels returned by the FirestoreAuthenticator.
// Callers use errors.Is to distinguish failure reasons (e.g. for user-facing
// error messages or HTTP status mapping).
var (
	ErrFirestoreUnavailable = errors.New("firestore unavailable")
	ErrGatekeeperNotFound   = errors.New("gatekeeper device document not found")
	ErrDeviceOffline        = errors.New("gatekeeper device is offline or inactive")
	ErrBuildTooOld          = errors.New("gatekeeper build number too old")
	ErrPermissionsMissing   = errors.New("required permissions missing")
	ErrDeviceMismatch       = errors.New("device id mismatch")
	ErrCredentialInvalid    = errors.New("credential invalid")
	ErrUIDRequired          = errors.New("uid required")
	ErrLoginMethodInvalid   = errors.New("login method invalid")
	ErrDeviceRequired       = errors.New("device id required")
	ErrCredentialRequired   = errors.New("credential required")
	ErrCredentialNotSet     = errors.New("credential not set")
	ErrProfileUnavailable   = errors.New("profile unavailable")
)

// FirestoreReader abstracts read-only Firestore access.
// The FirestoreAuthenticator depends on this interface — never on
// firestore.Client directly — so that Firestore stays out of the domain
// and the adapter is testable without a live connection.
type FirestoreReader interface {
	// GetDeviceDocument reads the Gatekeeper device document at
	// users/{parentID}/children/{uid} and returns its raw field map.
	GetDeviceDocument(ctx context.Context, parentID, uid string) (map[string]any, error)
}

// ProfileReader abstracts read-only access to auth-relevant fields stored in
// odyssey_user_profiles (Supabase). The concrete implementation is deferred to
// the login milestone; the adapter depends only on this interface.
type ProfileReader interface {
	// GetPasswordHash returns the stored credential hash for uid.
	// Returns ("", nil) when no credential has been set yet.
	GetPasswordHash(ctx context.Context, uid string) (string, error)
	// GetBoundDeviceID returns the device ID bound to the user's account.
	// Returns ("", nil) when no device is bound yet.
	GetBoundDeviceID(ctx context.Context, uid string) (string, error)
}

// FirestoreAuthenticator implements Authenticator by reading Gatekeeper's
// Firestore device document at users/{PARENT_ID}/children/{uid}.
//
// This is the SINGLE module that depends on Firestore (via FirestoreReader).
// The domain layer (pkg/game) never imports firestore or this package's
// Firestore-facing types directly — it depends only on the Authenticator port.
//
// Supported login methods:
//   - PASSWORD:  credential verification only (no Gatekeeper compliance).
//   - GATEKEEPER: Gatekeeper device compliance only (no credential).
//   - BOTH:       Gatekeeper compliance + credential (Odyssey's supported path).
type FirestoreAuthenticator struct {
	parentID       string
	minBuildNumber string
	hasher         PasswordHasher
	store          FirestoreReader
	profileReader  ProfileReader
}

// NewFirestoreAuthenticator constructs a FirestoreAuthenticator with all
// dependencies injected. parentID and minBuildNumber come from configuration
// (shared.LoadConfig / GATEKEEPER_MIN_BUILD_NUMBER env var).
func NewFirestoreAuthenticator(parentID, minBuildNumber string, hasher PasswordHasher, store FirestoreReader, profileReader ProfileReader) *FirestoreAuthenticator {
	return &FirestoreAuthenticator{
		parentID:       parentID,
		minBuildNumber: normalizeBuildNumber(minBuildNumber),
		hasher:         hasher,
		store:          store,
		profileReader:  profileReader,
	}
}

// Verify implements the Authenticator port.
//
// The method validates the login method, performs device and Gatekeeper
// compliance checks (for GATEKEEPER/BOTH modes), verifies the credential
// (for PASSWORD/BOTH modes), and validates device binding.
//
// Returns (newlyBound, error): true if no prior device binding existed and
// the device is being registered; false if the device was already bound.
func (a *FirestoreAuthenticator) Verify(ctx context.Context, uid, credential string, device DevicePayload) (string, bool, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return "", false, ErrUIDRequired
	}

	method := NormalizeLoginMethod(device.LoginMethod)
	if method == "" {
		return "", false, ErrLoginMethodInvalid
	}

	// Device ID is required for all modes — it identifies the client device
	// for binding purposes.
	if strings.TrimSpace(device.DeviceID) == "" {
		return "", false, ErrDeviceRequired
	}

	// Gatekeeper compliance: GATEKEEPER and BOTH require a compliant device.
	if method.RequiresGatekeeperCompliance() {
		identity, err := a.readGatekeeperDevice(ctx, uid, method)
		if err != nil {
			return "", false, err
		}
		if err := validateCompliance(identity, a.minBuildNumber); err != nil {
			return "", false, err
		}
	}

	// Credential verification: PASSWORD and BOTH require a valid credential.
	if method.RequiresCredential() {
		if strings.TrimSpace(credential) == "" {
			return "", false, ErrCredentialRequired
		}
		if err := a.verifyCredential(ctx, uid, credential); err != nil {
			return "", false, err
		}
	}

	// Device binding: validate the incoming device against any existing binding.
	newlyBound, err := a.validateDeviceBinding(ctx, uid, device.DeviceID)
	if err != nil {
		return "", false, err
	}

	return uid, newlyBound, nil
}

// readGatekeeperDevice reads the Gatekeeper Firestore document and maps it
// into an Identity. Firestore/gRPC errors are mapped to domain sentinels.
func (a *FirestoreAuthenticator) readGatekeeperDevice(ctx context.Context, uid string, method LoginMethod) (*Identity, error) {
	raw, err := a.store.GetDeviceDocument(ctx, a.parentID, uid)
	if err != nil {
		return nil, mapFirestoreError(err)
	}
	if raw == nil {
		return nil, ErrGatekeeperNotFound
	}
	return fromFirestore(raw, uid, method), nil
}

// verifyCredential looks up the stored credential hash and verifies it
// against the supplied password using the PasswordHasher.
func (a *FirestoreAuthenticator) verifyCredential(ctx context.Context, uid, credential string) error {
	hashed, err := a.profileReader.GetPasswordHash(ctx, uid)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProfileUnavailable, err)
	}
	if hashed == "" {
		return ErrCredentialNotSet
	}
	if err := a.hasher.Verify(hashed, credential); err != nil {
		return ErrCredentialInvalid
	}
	return nil
}

// validateDeviceBinding checks the incoming device ID against any existing
// binding for this UID. Returns true when no binding exists (newly bound).
func (a *FirestoreAuthenticator) validateDeviceBinding(ctx context.Context, uid, deviceID string) (bool, error) {
	boundID, err := a.profileReader.GetBoundDeviceID(ctx, uid)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrProfileUnavailable, err)
	}
	boundID = strings.TrimSpace(boundID)
	deviceID = strings.TrimSpace(deviceID)
	if boundID == "" {
		return true, nil
	}
	if boundID != deviceID {
		return false, ErrDeviceMismatch
	}
	return false, nil
}

// mapFirestoreError translates Firestore/gRPC status errors into domain
// sentinels. NotFound becomes ErrGatekeeperNotFound; quota, permission, and
// availability errors become ErrFirestoreUnavailable (fail-closed).
func mapFirestoreError(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("%w: %v", ErrFirestoreUnavailable, err)
	}
	switch st.Code() {
	case codes.NotFound:
		return ErrGatekeeperNotFound
	case codes.ResourceExhausted, codes.PermissionDenied, codes.Unavailable:
		return ErrFirestoreUnavailable
	default:
		return fmt.Errorf("%w: %v", ErrFirestoreUnavailable, err)
	}
}
