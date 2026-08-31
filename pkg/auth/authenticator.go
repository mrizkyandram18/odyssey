package auth

import (
	"context"
	"errors"
)

type Role string

const (
	RoleSeeker  Role = "SEEKER"
	RoleGuide   Role = "GUIDE"
	RoleBuilder Role = "BUILDER"
	RoleAdmin   Role = "ADMIN"
)

type DevicePayload struct {
	LoginMethod string `json:"login_method"`
	DeviceID    string `json:"device_id"`
	DeviceLabel string `json:"device_label"`
	DeviceName  string `json:"device_name"`
	DeviceModel string `json:"device_model"`
	Platform    string `json:"platform"`
	Browser     string `json:"browser"`
}

// Authenticator is the port (interface) for verifying user credentials
// against Gatekeeper device trust + a PIN/password. The domain layer
// depends only on this interface — it never imports Firestore.
//
// The bool return indicates whether the device was newly bound during
// this verification (true = newly bound, false = already bound).
type Authenticator interface {
	Verify(ctx context.Context, identifier, credential string, device DevicePayload) (string, bool, error)
}

// SessionConfig carries optional claims to embed in an issued session token.
// Pass nil for setup tokens or when role/crew information is unavailable.
type SessionConfig struct {
	Role     Role
	FamilyID string
}

type SessionClaims struct {
	Version  int    `json:"v"`
	Kind     string `json:"k"`
	UID      string `json:"uid"`
	Role     string `json:"role,omitempty"`
	FamilyID string `json:"family_id,omitempty"`
	Issued   int64  `json:"iat"`
	Expires  int64  `json:"exp"`
}

type SessionKind string

const (
	SessionKindUser  SessionKind = "user"
	SessionKindSetup SessionKind = "setup"
)

// SessionIssuer issues signed session tokens and validates them.
// It is the full-capability interface used during login flows.
type SessionIssuer interface {
	IssueSession(kind SessionKind, uid string, cfg *SessionConfig) (string, *SessionClaims, error)
	ParseSession(token string) (*SessionClaims, error)
}

// SessionValidator is the ISP-narrow interface consumed by auth middleware.
// A SessionIssuer always satisfies SessionValidator.
type SessionValidator interface {
	ParseSession(token string) (*SessionClaims, error)
}

// PasswordHasher abstracts password hashing so the auth adapter can be
// tested with a mock implementation. The concrete bcrypt/argon2
// implementation is wired in during the login milestone.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hashed, password string) error
}

var (
	ErrUIDRequired        = errors.New("uid required")
	ErrCredentialRequired = errors.New("credential required")
	ErrCredentialInvalid  = errors.New("credential invalid")
	ErrCredentialNotSet   = errors.New("credential not set")
	ErrDeviceRequired     = errors.New("device id required")
	ErrLoginMethodInvalid = errors.New("login method invalid")
	ErrProfileUnavailable = errors.New("profile unavailable")
	ErrSessionUID         = errors.New("session uid mismatch")
)

// SessionAuthErrorMessage translates session validation errors into
// user-facing messages suitable for API error responses.
func SessionAuthErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrSessionExpired):
		return "Session expired. Please log in again."
	case errors.Is(err, ErrSessionKind):
		return "Token is not valid for this action."
	case errors.Is(err, ErrSessionUID):
		return "Session UID mismatch."
	case errors.Is(err, ErrSessionInvalid):
		return "Invalid session token."
	default:
		return "Authentication required. Please log in again."
	}
}
