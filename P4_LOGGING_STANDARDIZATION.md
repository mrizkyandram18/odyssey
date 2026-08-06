# P4 Deliverable 2 — Structured Logging Standardization

Milestone: P4 — Observability & Operations
Scope: `pkg/observability/logging.go` (+ `Logger.LogRequest`, correlation flow). No transport/auth changes.

## Standard

| Field | Source | JSON key |
|---|---|---|
| Timestamp | Logger.marshal (`ts`) | `ts` (RFC 3339 UTC) |
| Level | INFO / WARN / ERROR | `lvl` |
| Message type | `msg` | `msg` |
| Request correlation | `RequestIDFromContext` (request middleware sets `X-Request-ID`) | `request_id` |
| User / crew / admin identity | token claims, extracted per-request; `request_id` propagated into `odyssey_audit_logs` | `user_id`, `crew_id`, `admin_uid` |
| Endpoint + method | middleware | `endpoint`, `method` |
| Latency | `time.Since(start)` | `duration_ms`, `duration_s` |
| Outcome | business flow | `outcome` (ok \| conflict \| duplicate \| invalid \| error) |
| Retry | retry loops | `retried` (bool) |
| Conflict | optimistic-lock CAS miss | `conflict` (bool) |
| Idempotency skip | idempotency guard | `idempotency_skip` (bool) |
| Validation failure | content validator | `validation_failure` (bool) |

## Implementation
- `RequestID` middleware (`pkg/observability/middleware.go`, `requestid.go`) injects/preserves
  `X-Request-ID`; stored in context via `WithRequestID`.
- `pkg/game/audit.Logger.Log` already reads `RequestIDFromContext` and writes it to the
  `request_id` column (migration 009). All audit rows are now correlatable to a request log.
- **New**: `Logger.LogServiceCall(ServiceCallFields)` emits a single structured `service_call`
  line with `request_id`, `op`, `entity_id`, `duration_ms`, `outcome`, `retried`, `conflict`,
  `idempotency_skip`, `validation_failure`, and `error`. Sensitive keys are sanitized by the
  existing `sanitizeFields` redaction path.

## Where service-call logs are emitted
- `pkg/game/dailyturn/handler.go` `Consume` — `op=dailyturn.consume`, maps
  `ErrNoTurnsRemaining` to `outcome=duplicate` + `idempotency_skip=true`.
- `pkg/game/quest/handler.go` `CompleteChallenge` — `op=quest.complete_challenge`,
  logs `outcome=error` with the error on early return.

Identity fields (user/crew) are not repeated in service logs because `request_id`
already correlates to the request-level `LogRequest` line that carries them.

## Example
```json
{"ts":"2026-08-05T15:03:12Z","lvl":"INFO","msg":"service_call","request_id":"6ff2de0f3295600890fc1df33653d468","op":"dailyturn.consume","entity_id":"user-1","duration_ms":8.12,"outcome":"duplicate","retried":false,"conflict":false,"idempotency_skip":true,"validation_failure":false}
```

## Test
`pkg/observability/logging_service_test.go::TestLogger_LogServiceCall` verifies JSON shape,
`RequestIDFromContext` plumbing, and nil-safety.
