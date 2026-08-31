package push

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"

	"odyssey/pkg/db"
)

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

type Sender struct {
	publicKey  string
	privateKey string
	subject    string
	store      db.PushSubscriptionStore
}

type Config struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
}

var ErrVAPIDNotConfigured = errors.New("VAPID public/private key pair not configured")

func NewSender(cfg Config, store db.PushSubscriptionStore) (*Sender, error) {
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

func (s *Sender) sendToSubscription(ctx context.Context, sub db.PushSubscription, payload []byte) error {
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256DH,
			Auth:   sub.Auth,
		},
	}, &webpush.Options{
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		Subscriber:      s.subject,
		TTL:             60,
	})
	if err != nil {
		log.Printf("push: delivery error for uid %q: %v", sub.UID, err)
		return fmt.Errorf("push: send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone {
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
