# Evidence-Based Seed Verification Report

**Date:** 2026-08-08  
**Method:** Supabase REST (`Prefer: count=exact`) + representative content selects (read-only)  
**Companion:** `docs/release/production_qa_evidence_report.md`

## Overview

This report verifies production Supabase `odyssey_*` tables against repository migrations **010, 017, 018, 019**.  
Complete prototype seed is **not** limited to local users.

## Canonical migration sources

| Path | Role |
| --- | --- |
| `supabase/migrations/` | Canonical migration tree |
| `scripts/migrations/` | Byte-identical mirror (verified this session) |

## Verification Matrix (Production — 2026-08-08)

| Entity | Migration | Expected | Actual | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `odyssey_schema_version` | 017 | 11 | 11 | **PASS** | Also reported by `/api/status` |
| `odyssey_local_users` | 019 | 3 | 3 | **PASS** | demo1/2/3 mapped; passwords not logged |
| `odyssey_families` | 017 | 1 | 1 | **PASS** | The Starseekers |
| `odyssey_user_profiles` | 017 | 3 | 3 | **PASS** | Leo/Maya/Sam (Leo XP evolved by QA) |
| `odyssey_missions` | 017 | 3 | 3 | **PASS** | 101, 102, 103 |
| `odyssey_exercises` | 017 | 6 | 6 | **PASS** | 103 exercises completed by QA |
| `odyssey_creative_items` | 017 | 5 | 5 | **PASS** | Journal content |
| `odyssey_daily_missions` | 017 | 6 | 6 | **PASS** | |
| `odyssey_collections` | 017 | 2 | 2 | **PASS** | |
| `odyssey_gifts` | 017 | 3 | 4 | **PASS** | +1 from gameplay |
| `odyssey_reactions` | 018 | 5 | 5 | **PASS** | seed UUID 000…001–005 |
| `odyssey_journey_definitions` | 010 | 3 | 3 | **PASS** | |
| `odyssey_course_definitions` | 010 | 4 | 4 | **PASS** | |
| `odyssey_quest_definitions` | 010 | 12 | 12 | **PASS** | includes riddle-of-the-stones |
| `odyssey_creative_prompt_definitions` | 010 | 6 | 6 | **PASS** | |
| `odyssey_achievement_definitions` | 010 | 10 | 10 | **PASS** | |
| `odyssey_concept_definitions` | 010 | 8 | 8 | **PASS** | |
| `odyssey_chest_definitions` | 010 | 5 | 5 | **PASS** | |
| `odyssey_drop_tables` | 010 | 22 | 22 | **PASS** | |
| `odyssey_relic_definitions` | 010 | 15 | 15 | **PASS** | |
| `odyssey_season_definitions` | 010 | 1 | 1 | **PASS** | |
| `odyssey_balance_configs` | 010 | 9 | 9 | **PASS** | |

## Critical content

```text
Mission 103: id=103 title=Riddle of the Stones template_slug=riddle-of-the-stones
Local users: demo1, demo2, demo3 → demo-uid-1/2/3
```

## Historical note

Earlier verification found `odyssey_local_users` missing when it lived only under `scripts/dev/`.  
Migration **019** and production apply resolved that gap. This matrix supersedes the prior FAIL for local users.

## Reseed decision

**DO NOT reseed.** Production already contains complete prototype content plus non-destructive QA progress.

## Conclusion

```text
SEED STATUS: COMPLETE
```
