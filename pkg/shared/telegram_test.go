package shared

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSendTelegramMessage_Success(t *testing.T) {
	var capturedPayload TelegramPayload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"ok": true,
			"result": {
				"message_id": 456,
				"chat": {
					"id": "123456"
				}
			}
		}`))
	}))
	defer ts.Close()

	SetTelegramBaseURLForTest(ts.URL)
	defer SetTelegramBaseURLForTest("")

	t.Setenv("TELEGRAM_BOT_TOKEN", "fake-token-123")
	t.Setenv("TELEGRAM_CHAT_ID", "123456")

	msg := "<b>Test Header</b>\nContent"
	res, err := SendTelegramMessage(msg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if res.MessageID != 456 {
		t.Errorf("expected MessageID 456, got %d", res.MessageID)
	}
	if res.ChatID != "123456" {
		t.Errorf("expected ChatID 123456, got %s", res.ChatID)
	}
	if capturedPayload.ChatID != "123456" {
		t.Errorf("expected payload chat_id 123456, got %s", capturedPayload.ChatID)
	}
	if capturedPayload.ParseMode != "HTML" {
		t.Errorf("expected parse_mode HTML, got %s", capturedPayload.ParseMode)
	}
	if capturedPayload.Text != msg {
		t.Errorf("expected payload text %q, got %q", msg, capturedPayload.Text)
	}
}

func TestSendTelegramMessage_MissingConfig(t *testing.T) {
	_ = os.Unsetenv("TELEGRAM_BOT_TOKEN")
	_ = os.Unsetenv("TELEGRAM_CHAT_ID")

	res, err := SendTelegramMessage("Hello", nil)
	if err == nil {
		t.Fatalf("expected error when config is missing, got nil")
	}
	if res != nil {
		t.Fatalf("expected nil result on error")
	}
}

func TestSendTelegramMessage_ApiError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"ok": false,
			"error_code": 400,
			"description": "Bad Request: chat not found"
		}`))
	}))
	defer ts.Close()

	SetTelegramBaseURLForTest(ts.URL)
	defer SetTelegramBaseURLForTest("")

	t.Setenv("TELEGRAM_BOT_TOKEN", "fake-token-123")
	t.Setenv("TELEGRAM_CHAT_ID", "invalid-chat")

	res, err := SendTelegramMessage("Hello", nil)
	if err == nil {
		t.Fatalf("expected error from bad status, got nil")
	}
	if !strings.Contains(err.Error(), "chat not found") {
		t.Errorf("expected error message to contain 'chat not found', got %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result on error")
	}
}

func TestEscapeTelegramHTML(t *testing.T) {
	input := "A < B & C > D"
	expected := "A &lt; B &amp; C &gt; D"
	actual := EscapeTelegramHTML(input)
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}
