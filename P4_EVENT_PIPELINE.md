# P4 Deliverable 6 — Event Pipeline Diagnostics

Milestone: P4 — Observability & Operations
Scope: `pkg/game/events/events.go` `Dispatcher`. No new dependencies; in-process.

## Problem
`Dispatcher.Publish` previously iterated subscribed handlers and **discarded**
both the `Handler.Handle` return errors and any timing. Event-driven side
effects (chapter completion, achievements, chest-on-quest, lore unlocks,
level-up) were therefore invisible in production.

## Solution (additive, nil-safe)
- New optional `Recorder` interface:
  ```go
  type Recorder interface {
      RecordEventPublished(eventType string)
      RecordEventHandler(eventType string, duration time.Duration, err error)
  }
  ```
- `Dispatcher.SetRecorder(r Recorder)` — attaches a sink. `nil` = no-op.
- `Publish` now:
  - records published count per event type,
  - times each handler call,
  - forwards duration + error to each `Handler.Handle`'s result to the recorder,
  - accumulates per-type handled/errors/latency in the dispatcher itself.
- `Diagnostics() map[string]EventDiagnostics` returns, per event type:
  `published`, `handled`, `errors`, `handler_count`, `avg_handler_duration_ms`.
- `Close()` resets counters (consistent with clearing subscriptions).
- The `Publisher`/`Handler` interfaces and the synchronous, in-order dispatch
  contract are unchanged → no architectural redesign.

## Wiring
`api/dev/main.go` calls `dispatcher.SetRecorder(metrics)`, making
`*observability.Metrics` (which implements `RecordEventPublished` and
`RecordEventHandler`) the live metrics sink. Per-type counters surface in the
`/metrics` `event_types` map; aggregate counters appear as
`events_published`, `events_handled`, `events_handler_errors`,
`events_handler_avg_latency_ms`.

## Replay / idempotency signals in the pipeline
- `chapter.handleQuestCompleted` records `replay_ignored` when a chapter is
  already COMPLETE (replayed QuestCompleted event).
- `chest.QuestCompletedHandler` records `replay_ignored` when a quest-reward
  chest already exists for the player.
- `chest.OpenChest` records `replay_ignored` when a chest is already opened.
- `dailyturn.ConsumeDailyTurn` records `replay_ignored` on the same-quest slug
  idempotency guard.

## Test coverage
- `pkg/game/events/diagnostics_test.go`:
  - `TestDispatcher_Diagnostics` — 2 events × 2 handlers: published=2,
    handled=4, errors=2, handler_count=2, avg duration non-negative.
  - `TestDispatcher_Recorder` — Recorder receives published+handled callbacks.
  - `TestDispatcher_Concurrency` — 50 goroutines × 40 publishes = 2000 handled,
    counters consistent under concurrency.
