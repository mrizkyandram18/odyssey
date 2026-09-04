package login

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
)

func TestLoginHandler_DeviceBlocked_Returns403WithClearMessage(t *testing.T) {
	Setup(&mockAuthenticator{err: auth.ErrDeviceBlocked}, &mockIssuer{token: "tok", claims: &auth.SessionClaims{UID: "user-1", FamilyID: "fam1"}}, &mockProfileStore{profile: nil})
	body := `{"uid":"user-1","credential":"secret","device":{"device_id":"web_other","login_method":"BOTH"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "perangkat") {
		t.Fatalf("expected body to contain 'perangkat' for device blocked, got %s", w.Body.String())
	}
}
