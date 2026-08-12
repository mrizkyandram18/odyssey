package push

import (
	"context"
	"log"

	"odyssey/pkg/game/events"
)

// DailyTurnHandler is an event.Handler that sends a Web Push notification to
// the explorer who just received a new daily turn.
type DailyTurnHandler struct {
	sender *Sender
}

// NewDailyTurnHandler returns a Handler that pushes "Your Turn" to the
// explorer identified in the DailyTurnCompletedEvent.
func NewDailyTurnHandler(s *Sender) *DailyTurnHandler {
	return &DailyTurnHandler{sender: s}
}

// Handle implements events.Handler.
func (h *DailyTurnHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.DailyTurnCompletedEvent)
	if !ok {
		return nil
	}
	n := Notification{
		Type:  TypeDailyTurn,
		Title: "Your Turn",
		Body:  "It's your turn in Family Mission.",
		URL:   "/missions",
	}
	if err := h.sender.SendToUser(ctx, e.UID, n); err != nil {
		// Log but do not fail the event pipeline — push is best-effort.
		log.Printf("push: daily turn notification for uid %q: %v", e.UID, err)
	}
	return nil
}

// RelayHandoffHandler is an event.Handler that sends a Web Push notification
// to the next-assigned explorer when a relay quest leg is handed off.
type RelayHandoffHandler struct {
	sender *Sender
}

// NewRelayHandoffHandler returns a Handler that pushes "Your Turn" to the
// explorer who just received the next relay leg.
func NewRelayHandoffHandler(s *Sender) *RelayHandoffHandler {
	return &RelayHandoffHandler{sender: s}
}

// Handle implements events.Handler.
func (h *RelayHandoffHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.RelayHandoffEvent)
	if !ok {
		return nil
	}
	body := "It's your turn in Family Mission."
	if e.QuestTitle != "" {
		body = "It's your turn in \"" + e.QuestTitle + "\"."
	}
	n := Notification{
		Type:  TypeRelayHandoff,
		Title: "Your Turn",
		Body:  body,
		URL:   "/missions",
	}
	if err := h.sender.SendToUser(ctx, e.ToUID, n); err != nil {
		log.Printf("push: relay handoff notification for uid %q: %v", e.ToUID, err)
	}
	return nil
}
