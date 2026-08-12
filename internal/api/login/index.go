package login

import (
	"context"
	"errors"
	"net/http"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/shared"
)

type loginRequest struct {
	UID        string             `json:"uid"`
	Credential string             `json:"credential"`
	Device     auth.DevicePayload `json:"device"`
}

type loginResponse struct {
	Status     string `json:"status"`
	Session    string `json:"session,omitempty"`
	SetupToken string `json:"setup_token,omitempty"`
	UID        string `json:"uid,omitempty"`
	FamilyID     string `json:"family_id,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Role       string `json:"role,omitempty"`
	Expires    int64  `json:"expires,omitempty"`
	Message    string `json:"message,omitempty"`
}

var (
	authenticator auth.Authenticator
	sessionIssuer auth.SessionIssuer
	profiles      db.ProfileStore
)

func Setup(a auth.Authenticator, s auth.SessionIssuer, p db.ProfileStore) {
	authenticator = a
	sessionIssuer = s
	profiles = p
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if authenticator == nil || sessionIssuer == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	var req loginRequest
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resolvedUID, _, err := authenticator.Verify(ctx, req.UID, req.Credential, req.Device)
	if err != nil {
		writeError(w, r, req.UID, err)
		return
	}

	writeSuccess(w, r, ctx, resolvedUID)
}

func writeSuccess(w http.ResponseWriter, r *http.Request, ctx context.Context, uid string) {
	var cfg *auth.SessionConfig
	if profiles != nil {
		if profile, err := profiles.GetUserProfile(ctx, uid); err == nil && profile != nil {
			cfg = &auth.SessionConfig{
				Role:   auth.Role(profile.Role),
				FamilyID: profile.FamilyID,
			}
		}
	}

	token, claims, err := sessionIssuer.IssueSession(auth.SessionKindUser, uid, cfg)
	if err != nil {
		shared.WriteJSONError(w, "failed to issue session", http.StatusInternalServerError)
		return
	}

	auth.SetSessionCookie(w, token, auth.UserSessionTTL, r.TLS != nil)

	shared.WriteJSON(w, http.StatusOK, loginResponse{
		Status:  "success",
		Session: token,
		UID:     uid,
		FamilyID:  claims.FamilyID,
		Kind:    "user",
		Role:    claims.Role,
		Expires: claims.Expires,
	})
}

func writeError(w http.ResponseWriter, r *http.Request, uid string, err error) {
	switch {
	case errors.Is(err, auth.ErrCredentialNotSet):
		token, claims, issueErr := sessionIssuer.IssueSession(auth.SessionKindSetup, uid, nil)
		if issueErr != nil {
			shared.WriteJSONError(w, "authentication failed", http.StatusUnauthorized)
			return
		}
		auth.SetSessionCookie(w, token, auth.SetupTokenTTL, r.TLS != nil)
		shared.WriteJSON(w, http.StatusOK, loginResponse{
			Status:     "setup_needed",
			SetupToken: token,
			UID:        uid,
			Kind:       "setup",
			Expires:    claims.Expires,
			Message:    "Credential setup required.",
		})

	case errors.Is(err, auth.ErrCredentialRequired):
		shared.WriteJSON(w, http.StatusBadRequest, loginResponse{
			Status:  "password_required",
			UID:     uid,
			Message: "Password is required.",
		})

	case errors.Is(err, auth.ErrLoginMethodInvalid),
		errors.Is(err, auth.ErrUIDRequired),
		errors.Is(err, auth.ErrDeviceRequired):
		shared.WriteJSONError(w, "invalid request", http.StatusBadRequest)

	case errors.Is(err, auth.ErrCredentialInvalid),
		errors.Is(err, auth.ErrDeviceOffline),
		errors.Is(err, auth.ErrBuildTooOld),
		errors.Is(err, auth.ErrPermissionsMissing),
		errors.Is(err, auth.ErrDeviceMismatch),
		errors.Is(err, auth.ErrGatekeeperNotFound):
		shared.WriteJSONError(w, "authentication failed", http.StatusUnauthorized)

	case errors.Is(err, auth.ErrFirestoreUnavailable):
		shared.WriteJSONError(w, "auth service unavailable", http.StatusServiceUnavailable)

	case errors.Is(err, auth.ErrProfileUnavailable):
		shared.WriteJSONError(w, "profile unavailable", http.StatusInternalServerError)

	default:
		shared.WriteJSONError(w, "authentication failed", http.StatusUnauthorized)
	}
}
