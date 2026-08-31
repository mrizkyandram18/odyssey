package auth

import (
	"strings"
)

// LoginMethod represents the authentication path a user is attempting.
// Supported methods:
//   - PASSWORD: password authentication
//   - BOTH: credential + device metadata
type LoginMethod string

const (
	LoginMethodPassword   LoginMethod = "PASSWORD"
	LoginMethodGatekeeper LoginMethod = "GATEKEEPER"
	LoginMethodBoth       LoginMethod = "BOTH"
)

// NormalizeLoginMethod validates and normalizes a login method string.
// Returns "" for unsupported or empty methods.
func NormalizeLoginMethod(method string) LoginMethod {
	switch LoginMethod(strings.ToUpper(strings.TrimSpace(method))) {
	case LoginMethodPassword:
		return LoginMethodPassword
	case LoginMethodGatekeeper:
		return LoginMethodGatekeeper
	case LoginMethodBoth:
		return LoginMethodBoth
	default:
		return ""
	}
}

// RequiresGatekeeperCompliance returns true for GATEKEEPER and BOTH methods.
func (m LoginMethod) RequiresGatekeeperCompliance() bool {
	return m == LoginMethodGatekeeper || m == LoginMethodBoth
}

// RequiresCredential returns true for PASSWORD and BOTH methods.
func (m LoginMethod) RequiresCredential() bool {
	return m == LoginMethodPassword || m == LoginMethodBoth
}
