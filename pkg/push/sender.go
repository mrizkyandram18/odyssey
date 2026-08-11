// Package push provides a Web Push sender for Odyssey.
//
// Security: VAPID private key is loaded from config and never logged, returned
// to clients, or included in any payload. Subscription secrets (endpoint,
// p256dh, auth) are never logged.
package push

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"

	"odyssey/pkg/game"
)

// NotificationType is a typed identifier for push notification kinds.
type NotificationType string

const (
	TypeDailyTurn    NotificationType = "DAILY_TURN"
	TypeRelayHandoff NotificationType = "RELAY_HANDOFF"
)

// Notification is the typed payload delivered in a Web Push message.
// Keep it minimal — it must match sw.js showNotification usage.
type Notification struct {
	Type  NotificationType `json:"type"`
	Title string           `json:"title"`
	Body  string           `json:"body"`
	URL   string           `json:"url"`
}

// Sender delivers Web Push notifications to users.
// It requires VAPID credentials to sign outgoing requests.
type Sender struct {
	publicKey  string
	privateKey string
	subject    string
	store      game.PushSubscriptionStore
}

// Config holds VAPID credentials for the push sender.
type Config struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	// VAPIDSubject is an mailto: or https: URI identifying the sender.
	VAPIDSubject string
}

// ErrVAPIDNotConfigured is returned when VAPID keys are missing.
var ErrVAPIDNotConfigured = errors.New("VAPID public/private key pair not configured")

// NewSender constructs a push Sender. Returns ErrVAPIDNotConfigured if the
// public or private key is empty.
func NewSender(cfg Config, store game.PushSubscriptionStore) (*Sender, error) {
	if cfg.VAPIDPublicKey == "" || cfg.VAPIDPrivateKey == "" {
		return nil, ErrVAPIDNotConfigured
	}
	subject := cfg.VAPIDSubject
	if subject == "" {
		subject = "mailto:admin@odyssey.example.com"
	}
	return &Sender{
		publicKey:  cfg.VAPIDPublicKey,
		privateKey: cfg.VAPIDPrivateKey,
		subject:    subject,
		store:      store,
	}, nil
}

// SendToUser delivers n to every active push subscription registered for uid.
// HTTP 410 responses are treated as stale subscriptions and deleted silently.
// Other errors per subscription are logged (without revealing secrets) and do
// not abort delivery to remaining subscriptions.
// Returns an error only when no subscriptions exist or all deliveries fail.
func (s *Sender) SendToUser(ctx context.Context, uid string, n Notification) error {
	subs, err := s.store.ListSubscriptionsByUID(ctx, uid)
	if err != nil {
		return fmt.Errorf("push: list subscriptions for uid: %w", err)
	}
	if len(subs) == 0 {
		return nil
	}

	payload, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("push: marshal notification: %w", err)
	}

	var lastErr error
	for _, sub := range subs {
		if err := s.sendToSubscription(ctx, sub, payload); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// sendToSubscription sends to a single subscription, deleting it on 410.
func (s *Sender) sendToSubscription(ctx context.Context, sub game.PushSubscription, payload []byte) error {
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}, &webpush.Options{
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		Subscriber:      s.subject,
		TTL:             60,
	})
	if err != nil {
		// Log without any sub secrets.
		log.Printf("push: delivery error for uid %q: %v", sub.UID, err)
		return fmt.Errorf("push: send notification: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode == http.StatusGone {
		// Subscription is no longer valid; remove it silently.
		if delErr := s.store.DeleteSubscription(ctx, sub.UID, sub.Endpoint); delErr != nil {
			log.Printf("push: failed to delete stale subscription for uid %q: %v", sub.UID, delErr)
		}
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("push: unexpected status %d for uid %q", resp.StatusCode, sub.UID)
		return fmt.Errorf("push: unexpected HTTP %d from push service", resp.StatusCode)
	}
	return nil
}
