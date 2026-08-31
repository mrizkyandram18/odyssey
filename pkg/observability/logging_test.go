package observability

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLogger_LogRequest(t *testing.T) {
	var buf testBuffer
	logger := NewLogger(&buf)
	logger.LogRequest(LogFields{
		RequestID: "req-123",
		UserID:    "user-1",
		FamilyID:  "fam-1",
		Endpoint:  "/api/tasks",
		Method:    "GET",
		Duration:  42 * time.Millisecond,
		Status:    200,
		RemoteIP:  "192.168.1.1",
	})
	logger.Flush()

	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["request_id"] != "req-123" {
		t.Errorf("expected request_id req-123, got %v", entry["request_id"])
	}
	if entry["user_id"] != "user-1" {
		t.Errorf("expected user_id user-1, got %v", entry["user_id"])
	}
	if entry["family_id"] != "fam-1" {
		t.Errorf("expected family_id fam-1, got %v", entry["family_id"])
	}
	if entry["endpoint"] != "/api/tasks" {
		t.Errorf("expected endpoint /api/tasks, got %v", entry["endpoint"])
	}
	if entry["method"] != "GET" {
		t.Errorf("expected method GET, got %v", entry["method"])
	}
	if entry["status"].(float64) != 200 {
		t.Errorf("expected status 200, got %v", entry["status"])
	}
	if entry["remote_ip"] != "192.168.1.1" {
		t.Errorf("expected remote_ip 192.168.1.1, got %v", entry["remote_ip"])
	}
	if entry["error"] != nil {
		t.Errorf("expected no error field, got %v", entry["error"])
	}
}

func TestLogger_LogRequest_WithAdmin(t *testing.T) {
	var buf testBuffer
	logger := NewLogger(&buf)
	logger.LogRequest(LogFields{
		RequestID: "req-456",
		AdminUID:  "admin-1",
		Endpoint:  "/api/admin/tasks",
		Method:    "GET",
		Duration:  10 * time.Millisecond,
		Status:    200,
		RemoteIP:  "10.0.0.1",
	})
	logger.Flush()

	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["admin_uid"] != "admin-1" {
		t.Errorf("expected admin_uid admin-1, got %v", entry["admin_uid"])
	}
}

func TestLogger_LogRequest_WithError(t *testing.T) {
	var buf testBuffer
	logger := NewLogger(&buf)
	logger.LogRequest(LogFields{
		RequestID: "req-789",
		UserID:    "user-2",
		Endpoint:  "/api/tasks",
		Method:    "POST",
		Duration:  5 * time.Millisecond,
		Status:    500,
		RemoteIP:  "192.168.1.2",
		Error:     "internal error",
	})
	logger.Flush()

	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["error"] != "internal error" {
		t.Errorf("expected error 'internal error', got %v", entry["error"])
	}
}

func TestLogger_Info(t *testing.T) {
	var buf testBuffer
	logger := NewLogger(&buf)
	logger.Info("startup", map[string]any{"component": "server"})
	logger.Flush()

	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["msg"] != "startup" {
		t.Errorf("expected msg 'startup', got %v", entry["msg"])
	}
	if entry["component"] != "server" {
		t.Errorf("expected component 'server', got %v", entry["component"])
	}
}

func TestLogger_SanitizesSensitiveFields(t *testing.T) {
	var buf testBuffer
	logger := NewLogger(&buf)
	logger.Info("event", map[string]any{
		"password": "secret123",
		"token":    "abc",
		"apikey":   "key",
		"safe_key": "safe_value",
	})
	logger.Flush()

	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["password"] != "[REDACTED]" {
		t.Errorf("expected password to be redacted, got %v", entry["password"])
	}
	if entry["token"] != "[REDACTED]" {
		t.Errorf("expected token to be redacted, got %v", entry["token"])
	}
	if entry["apikey"] != "[REDACTED]" {
		t.Errorf("expected apikey to be redacted, got %v", entry["apikey"])
	}
	if entry["safe_key"] != "safe_value" {
		t.Errorf("expected safe_key to be preserved, got %v", entry["safe_key"])
	}
}

func TestLogger_Error(t *testing.T) {
	var buf testBuffer
	logger := NewLogger(&buf)
	logger.Error("failed", map[string]any{"error": "something went wrong"})
	logger.Flush()

	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["error"] != "something went wrong" {
		t.Errorf("expected error, got %v", entry["error"])
	}
}

func TestLogger_ChannelFallback(t *testing.T) {
	var buf testBuffer
	logger := NewLoggerWithBuffer(&buf, 1)
	logger.Info("msg1", nil)
	logger.Info("msg2", nil)
	logger.Info("msg3", nil)
	logger.Flush()

	lines := strings.Count(buf.String(), "\n")
	if lines != 3 {
		t.Errorf("expected 3 log lines, got %d", lines)
	}
	logger.Close()
}

func TestExtractIdentityFromToken_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	uid, crewID, adminUID := extractIdentityFromToken(req)
	if uid != "" || crewID != "" || adminUID != "" {
		t.Errorf("expected empty values, got uid=%s crew=%s admin=%s", uid, crewID, adminUID)
	}
}

func TestExtractIdentityFromToken_Bearer(t *testing.T) {
	// Create a valid-looking token (we just need the payload format right)
	claims := `{"uid":"user-1","family_id":"crew-1","role":"SEEKER"}`
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	sig := "fakesig"
	token := payload + "." + sig

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	uid, crewID, adminUID := extractIdentityFromToken(req)
	if uid != "user-1" {
		t.Errorf("expected uid user-1, got %s", uid)
	}
	if crewID != "crew-1" {
		t.Errorf("expected family_id crew-1, got %s", crewID)
	}
	if adminUID != "" {
		t.Errorf("expected empty admin_uid, got %s", adminUID)
	}
}

func TestExtractIdentityFromToken_Admin(t *testing.T) {
	claims := `{"uid":"admin-1","family_id":"crew-1","role":"ADMIN"}`
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	token := payload + ".sig"

	req := httptest.NewRequest(http.MethodGet, "/api/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	_, _, adminUID := extractIdentityFromToken(req)
	if adminUID != "admin-1" {
		t.Errorf("expected admin_uid admin-1, got %s", adminUID)
	}
}

func TestExtractIdentityFromToken_XUserSession(t *testing.T) {
	claims := `{"uid":"user-2","family_id":"crew-2","role":"SEEKER"}`
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	token := payload + ".sig"

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("X-User-Session", token)
	uid, crewID, _ := extractIdentityFromToken(req)
	if uid != "user-2" {
		t.Errorf("expected uid user-2, got %s", uid)
	}
	if crewID != "crew-2" {
		t.Errorf("expected family_id crew-2, got %s", crewID)
	}
}

func TestExtractIdentityFromToken_Cookie(t *testing.T) {
	claims := `{"uid":"user-3","family_id":"crew-3","role":"GUIDE"}`
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	token := payload + ".sig"

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "odyssey_session", Value: token})
	uid, crewID, _ := extractIdentityFromToken(req)
	if uid != "user-3" {
		t.Errorf("expected uid user-3, got %s", uid)
	}
	if crewID != "crew-3" {
		t.Errorf("expected family_id crew-3, got %s", crewID)
	}
}

func TestExtractIdentityFromToken_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	uid, crewID, adminUID := extractIdentityFromToken(req)
	if uid != "" || crewID != "" || adminUID != "" {
		t.Errorf("expected empty values for invalid token, got uid=%s crew=%s admin=%s", uid, crewID, adminUID)
	}
}

func TestLogger_DefaultLogger(t *testing.T) {
	l := DefaultLogger()
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestLogger_NewLogger_NilWriter(t *testing.T) {
	l := NewLogger(nil)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
	l.Info("test", map[string]any{"msg": "ok"})
	l.Flush()
}

type testBuffer struct {
	data []byte
	mu   sync.Mutex
}

func (b *testBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *testBuffer) bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.data
}

func (b *testBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.data)
}

var _ = context.Background
