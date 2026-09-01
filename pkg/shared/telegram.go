package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var telegramBaseURL = "https://api.telegram.org"

// SetTelegramBaseURLForTest overrides the Telegram API host. Pass "" to restore default.
// Intended for unit tests only.
func SetTelegramBaseURLForTest(base string) {
	if base == "" {
		telegramBaseURL = "https://api.telegram.org"
		return
	}
	telegramBaseURL = base
}

type InlineButton struct {
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineButton `json:"inline_keyboard"`
}

type TelegramPayload struct {
	ChatID             string                 `json:"chat_id"`
	Text               string                 `json:"text"`
	ParseMode          string                 `json:"parse_mode,omitempty"`
	ReplyMarkup        *InlineKeyboardMarkup  `json:"reply_markup,omitempty"`
	LinkPreviewOptions map[string]interface{} `json:"link_preview_options,omitempty"`
}

type TelegramSendResult struct {
	ChatID    string
	MessageID int64
}

// SendTelegramMessage posts an HTML message to the configured Telegram chat.
// Reuses the implementation pattern from kuota-copy.
func SendTelegramMessage(message string, buttons []InlineButton) (*TelegramSendResult, error) {
	botToken := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	chatID := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID"))

	if botToken == "" || chatID == "" {
		log.Printf("Telegram Notice: Configuration missing (token present: %v, chatID present: %v)", botToken != "", chatID != "")
		return nil, fmt.Errorf("missing telegram config")
	}

	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", telegramBaseURL, botToken)

	payload := TelegramPayload{
		ChatID:             chatID,
		Text:               message,
		ParseMode:          "HTML",
		LinkPreviewOptions: map[string]interface{}{"is_disabled": true},
	}

	if len(buttons) > 0 {
		payload.ReplyMarkup = &InlineKeyboardMarkup{
			InlineKeyboard: [][]InlineButton{buttons},
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Telegram Error: Post failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var respBody struct {
			Ok          bool   `json:"ok"`
			Description string `json:"description"`
			ErrorCode   int    `json:"error_code"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&respBody)
		log.Printf("Telegram Error: API returned %d. Response: %s", resp.StatusCode, respBody.Description)
		return nil, fmt.Errorf("telegram request failed with status: %d, desc: %s", resp.StatusCode, respBody.Description)
	}

	result, err := decodeTelegramSendResult(resp, chatID)
	if err != nil {
		log.Printf("Telegram: sent but could not parse message id: %v", err)
		return &TelegramSendResult{ChatID: chatID}, nil
	}
	log.Printf("Telegram: Notification sent successfully to %s msg=%d", chatID, result.MessageID)
	return result, nil
}

func decodeTelegramSendResult(resp *http.Response, fallbackChat string) (*TelegramSendResult, error) {
	var body struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageID int64 `json:"message_id"`
			Chat      struct {
				ID interface{} `json:"id"`
			} `json:"chat"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	chatID := fallbackChat
	switch v := body.Result.Chat.ID.(type) {
	case string:
		if v != "" {
			chatID = v
		}
	case float64:
		chatID = strconv.FormatInt(int64(v), 10)
	case json.Number:
		chatID = v.String()
	}
	return &TelegramSendResult{
		ChatID:    chatID,
		MessageID: body.Result.MessageID,
	}, nil
}

// EscapeTelegramHTML escapes characters for Telegram HTML parse mode.
func EscapeTelegramHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
