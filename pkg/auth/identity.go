package auth

import (
	"strconv"
	"strings"
	"time"
)

// LoginMethod represents the authentication path a user is attempting.
//
//   - PASSWORD:  credential (password) only — no Gatekeeper device compliance.
//   - GATEKEEPER: Gatekeeper device compliance only — no credential.
//   - BOTH:       Gatekeeper device compliance + credential (the supported
//     Odyssey path per docs/integrations.md).
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
// PASSWORD mode does not require Gatekeeper device compliance.
func (m LoginMethod) RequiresGatekeeperCompliance() bool {
	return m == LoginMethodGatekeeper || m == LoginMethodBoth
}

// RequiresCredential returns true for PASSWORD and BOTH methods.
// GATEKEEPER mode does not require a password credential.
func (m LoginMethod) RequiresCredential() bool {
	return m == LoginMethodPassword || m == LoginMethodBoth
}

// Identity is the domain model mapped from the Gatekeeper Firestore device
// document. It captures the device's auth-relevant state without exposing
// Firestore DTOs to the rest of the adapter or the domain layer.
//
// The Firestore document lives at users/{PARENT_ID}/children/{uid} and has
// the following shape (written by the Gatekeeper Android app):
//
//	isOnline:  bool
//	lastSeen:  timestamp
//	details: {
//	    appBuildNumber: string,
//	    permissions:    map[string]bool,
//	    isOnline:       bool,       // some devices nest isOnline in details
//	    lastSeen:       timestamp,  // some devices nest lastSeen in details
//	}
//	permissions: map[string]bool   // top-level sibling on some devices
//
// Fields may appear at the top level or nested inside "details"; fromFirestore
// resolves both locations, preferring "details" for appBuildNumber and
// permissions (matching Family Reward's resolution order).
type Identity struct {
	UID         string
	LoginMethod LoginMethod
	IsOnline    bool
	LastSeen    time.Time
	BuildNumber string
	Permissions map[string]bool
}

// deviceActivityThreshold is how recently the Gatekeeper device must have
// reported in to be considered active (5 minutes, matching Family Reward's
// CheckDeviceStatus rule).
const deviceActivityThreshold = 5 * time.Minute

// requiredPermissions lists the Gatekeeper permissions that must all be true.
// "notification" is optional and therefore excluded from this list.
var requiredPermissions = []string{
	"battery_exemption",
	"camera",
	"device_admin",
	"exact_alarm",
	"ignore_battery",
	"microphone",
	"oem_autostart_confirmed",
	"overlay",
	"usage_stats",
}

// fromFirestore maps a raw Firestore document (map[string]any as returned by
// DocumentSnapshot.Data()) into the domain Identity type.
func fromFirestore(raw map[string]any, uid string, method LoginMethod) *Identity {
	id := &Identity{
		UID:         uid,
		LoginMethod: method,
		Permissions: map[string]bool{},
	}

	var details map[string]any
	if d, ok := raw["details"].(map[string]any); ok {
		details = d
	}

	// isOnline — top-level takes precedence, then details
	if v, ok := asBool(raw["isOnline"]); ok {
		id.IsOnline = v
	} else if details != nil {
		if v, ok := asBool(details["isOnline"]); ok {
			id.IsOnline = v
		}
	}

	// lastSeen — top-level, then details
	id.LastSeen = asTime(raw["lastSeen"])
	if id.LastSeen.IsZero() && details != nil {
		id.LastSeen = asTime(details["lastSeen"])
	}

	// appBuildNumber — from details primarily, then top-level
	if details != nil {
		if s, ok := asString(details["appBuildNumber"]); ok {
			id.BuildNumber = s
		}
	}
	if id.BuildNumber == "" {
		if s, ok := asString(raw["appBuildNumber"]); ok {
			id.BuildNumber = s
		}
	}

	// permissions — details.permissions preferred, then top-level sibling
	var perms map[string]any
	if details != nil {
		if p, ok := details["permissions"].(map[string]any); ok {
			perms = p
		}
	}
	if perms == nil {
		if p, ok := raw["permissions"].(map[string]any); ok {
			perms = p
		}
	}
	for k, v := range perms {
		id.Permissions[k] = asBoolDefault(v, false)
	}

	return id
}

// --- Validation logic operating on Identity ---

// validateCompliance checks device compliance: online status, lastSeen
// recency, build number, and required permissions.
func validateCompliance(id *Identity, minBuildNumber string) error {
	if !id.IsOnline {
		return ErrDeviceOffline
	}
	if id.LastSeen.IsZero() || time.Since(id.LastSeen) > deviceActivityThreshold {
		return ErrDeviceOffline
	}
	if err := validateBuildNumber(id.BuildNumber, minBuildNumber); err != nil {
		return err
	}
	if err := validatePermissions(id.Permissions); err != nil {
		return err
	}
	return nil
}

// validateBuildNumber checks that the device's build number is >= the minimum.
// Both values are normalized (strip leading "v"/"V") and compared numerically.
func validateBuildNumber(actual, min string) error {
	actual = normalizeBuildNumber(actual)
	min = normalizeBuildNumber(min)
	actualNum, err1 := strconv.Atoi(actual)
	minNum, err2 := strconv.Atoi(min)
	if err1 != nil || err2 != nil {
		return ErrBuildTooOld
	}
	if actualNum < minNum {
		return ErrBuildTooOld
	}
	return nil
}

// normalizeBuildNumber strips a leading "v" or "V" and trims whitespace.
func normalizeBuildNumber(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 0 && (s[0] == 'v' || s[0] == 'V') {
		s = strings.TrimSpace(s[1:])
	}
	return s
}

// validatePermissions checks that all required permissions are true and that
// no non-required permission (except "notification") is explicitly set to false.
func validatePermissions(perms map[string]bool) error {
	if perms == nil {
		return ErrPermissionsMissing
	}
	for _, perm := range requiredPermissions {
		if !perms[perm] {
			return ErrPermissionsMissing
		}
	}
	for perm, val := range perms {
		if perm == "notification" {
			continue
		}
		if !val {
			return ErrPermissionsMissing
		}
	}
	return nil
}

// --- Firestore type extraction helpers ---
// These handle the untyped nature of Firestore document data (map[string]any)
// where nested values may arrive with various Go types depending on the SDK
// version and how the Gatekeeper app stored them.

func asBool(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}

func asBoolDefault(v any, def bool) bool {
	b, ok := v.(bool)
	if !ok {
		return def
	}
	return b
}

func asString(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

// asTime extracts a time.Time from a Firestore value. Firestore timestamps
// arrive as time.Time; some device documents store lastSeen as an RFC3339
// string. A zero time.Time is returned when the value cannot be parsed.
func asTime(v any) time.Time {
	switch val := v.(type) {
	case time.Time:
		return val.UTC()
	case string:
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
