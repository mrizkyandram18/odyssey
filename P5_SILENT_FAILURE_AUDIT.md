# P5 Silent Failure & Error Propagation Audit

## 1. Executive Summary
This audit reviewed the entire repository for silent failure patterns, such as swallowed errors (`continue`, `return nil` in error checks), ignored returns (`_ =`), and unhandled panics. 

## 2. Number of Findings
Total occurrences found: 93

## 3. Risk Classification Table

| File | Function/Line | Current behavior | Why it exists | Risk level | Recommended action | Behavior impact |
|---|---|---|---|---|---|---|
| `api\achievements\index.go:42` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:467` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:476` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:507` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:519` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:531` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:583` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:592` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:605` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:622` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:635` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:644` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:657` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:665` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:678` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:691` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\admin\index.go:704` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chapters\index.go:61` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chapters\index.go:70` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chapters\index.go:83` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chapters\index.go:84` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chests\index.go:57` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chests\index.go:70` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chests\index.go:76` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chests\index.go:77` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chests\index.go:111` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chests\index.go:117` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\chests\index.go:118` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:101` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:114` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:145` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:151` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:160` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:166` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:167` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:190` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:196` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:197` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:215` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:223` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:233` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\creative\index.go:234` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\crews\index.go:67` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\crews\index.go:68` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\daily_turns\index.go:59` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\daily_turns\index.go:60` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\daily_turns\index.go:85` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\daily_turns\index.go:98` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\daily_turns\index.go:99` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\home\index.go:44` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\login\index.go:57` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\login\index.go:64` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\login\index.go:84` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\lore\index.go:63` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\lore\index.go:72` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\lore\index.go:81` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\me\index.go:43` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\me\index.go:44` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:106` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:115` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:121` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:122` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:135` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:140` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:141` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:154` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:159` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:165` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\quests\index.go:166` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\realm_progress\index.go:50` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\relics\index.go:55` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `api\relics\index.go:74` | `empty return in err check` | Writes HTTP error and returns without returning error to caller | Standard HTTP handler behavior (cannot return error from handler) | P3 | Ignore (documented) | Low |
| `pkg\auth\firestore.go:192` | `return nil in err check` | Swallows error | Unknown | P2 | Ignore (documented) | Low |
| `pkg\auth\middleware.go:51` | `empty return in err check` | Writes HTTP response or logs silently | Middleware/HTTP layer handling | P3 | Ignore (documented) | Low |
| `pkg\game\achievement\service.go:159` | `continue in err check` | Silently continues on DB failure during progress count or create | Loop over triggers; tries to avoid failing the whole batch | P0 | Fallback + warning log / Return error | Medium |
| `pkg\game\achievement\service.go:164` | `continue in err check` | Silently continues on DB failure during progress count or create | Loop over triggers; tries to avoid failing the whole batch | P0 | Fallback + warning log / Return error | Medium |
| `pkg\game\achievement\service.go:179` | `continue in err check` | Silently continues on DB failure during progress count or create | Loop over triggers; tries to avoid failing the whole batch | P0 | Fallback + warning log / Return error | Medium |
| `pkg\game\chapter\service.go:239` | `return nil in err check` | Returns nil on ErrNotFound | Expected graceful degradation when there is no next chapter | P3 | Ignore (documented) | Low |
| `pkg\game\chapter\service.go:249` | `return nil in err check` | Returns nil on ErrNotFound | Expected graceful degradation when there is no next chapter | P3 | Ignore (documented) | Low |
| `pkg\game\chapter\service.go:279` | `return nil in err check` | Returns nil on ErrNotFound | Expected graceful degradation when there is no next chapter | P3 | Ignore (documented) | Low |
| `pkg\game\chapter\service.go:380` | `return nil in err check` | Returns nil on ErrNotFound | Expected graceful degradation when there is no next chapter | P3 | Ignore (documented) | Low |
| `pkg\game\chest\catalog.go:87` | `continue in err check` | Silently continues if ListDropTableEntries fails | Tries to load other chests if one fails | P0 | Fallback + warning log | High |
| `pkg\game\chest\service.go:336` | `return nil in err check` | Returns nil (ACKs event) if GetQuest fails | Event handler signature requires error for nack, but developer probably didn't want to retry indefinitely | P0 | Return error | High |
| `pkg\game\lore\service.go:178` | `continue in err check` | Silently continues on CreateLoreUnlock DB failure | Tries to unlock remaining lore if one fails | P0 | Fallback + warning log / Return error | High |
| `pkg\game\lore\service.go:189` | `continue in err check` | Silently continues on CreateLoreUnlock DB failure | Tries to unlock remaining lore if one fails | P0 | Fallback + warning log / Return error | High |
| `pkg\game\quest\handler.go:208` | `empty return in err check` | Silently returns on DB failure in advanceRealm | Background task/handler avoiding panics | P0 | Return error or Fallback + warning log | High |
| `pkg\game\quest\handler.go:236` | `empty return in err check` | Silently returns on DB failure in advanceRealm | Background task/handler avoiding panics | P0 | Return error or Fallback + warning log | High |
| `pkg\game\quest\handler.go:264` | `empty return in err check` | Silently returns on DB failure in advanceRealm | Background task/handler avoiding panics | P0 | Return error or Fallback + warning log | High |
| `pkg\observability\logging.go:73` | `empty return in err check` | Writes HTTP response or logs silently | Middleware/HTTP layer handling | P3 | Ignore (documented) | Low |
| `pkg\observability\logging.go:104` | `empty return in err check` | Writes HTTP response or logs silently | Middleware/HTTP layer handling | P3 | Ignore (documented) | Low |
| `pkg\observability\logging.go:236` | `empty return in err check` | Writes HTTP response or logs silently | Middleware/HTTP layer handling | P3 | Ignore (documented) | Low |
| `pkg\observability\logging.go:244` | `empty return in err check` | Writes HTTP response or logs silently | Middleware/HTTP layer handling | P3 | Ignore (documented) | Low |
| `pkg\observability\middleware.go:47` | `recover` | Panic recovery | Prevents server crash on panic | P3 | Ignore (documented) | Low |

## 4. P0 Fixes Applied
(To be updated after remediation)

## 5. Remaining Findings Intentionally Deferred
All P3 findings (HTTP handlers returning early, middleware handling, missing chapters treated as end of content) have been intentionally deferred.

## 6. Verification Summary
(To be updated after tests pass)
