import io

smoke_test_content = """

## 7. API Smoke Tests
- `/health`: `{"status":"ok","timestamp":"2026-08-06T06:22:42Z","checks":{"admin_store":{"status":"pass"},"audit_store":{"status":"pass"},"cache":{"status":"pass"},"configuration":{"status":"pass"},"content_generation":{"status":"pass"},"database":{"status":"pass"}}}`
- `/version`: `{"git_commit":"unknown","build_date":"unknown","semantic_version":"dev","schema_version":"0","content_generation":1}`
- `/api/quests`: `401 Unauthorized` (Expected for protected endpoint)
- `/api/status`: `{"app":"odyssey","boot_time":"2026-08-06T06:22:24Z","cache_hit_ratio":0,"content_generation":1,"uptime_seconds":18,"version":"dev"}`
- `/api/chests/catalog`: `401 Unauthorized` (Expected for protected endpoint)
- `/api/relics/catalog`: `401 Unauthorized` (Expected for protected endpoint)

## 8. Backend Startup Log
Backend started successfully without any fatal errors or panics.
```text
{"error":"schema version row not found","lvl":"WARN","msg":"schema_version_read_failed","ts":"2026-08-06T06:22:24Z"}
2026/08/06 13:22:24 Odyssey server starting on :8080
2026/08/06 13:22:24 Timezone: Asia/Jakarta | DailyTurnXP: 30 | MaxDailyTurns: 1 | RealmCatalog: 3 realms
2026/08/06 13:22:24 Routes: /api/login /api/me /api/status /api/quests /api/crews /api/realm_progress /api/daily_turns /api/creative /api/home /api/chests /api/relics /api/chapters /api/lore /api/achievements /api/admin /metrics /version /health /ready /live /debug/profile
{"lvl":"INFO","msg":"server_started","port":"8080","ts":"2026-08-06T06:22:24Z","version":"dev"}
```
"""

with io.open('docs/reports/production-db-validation.md', 'a', encoding='utf-8') as f:
    f.write(smoke_test_content)
