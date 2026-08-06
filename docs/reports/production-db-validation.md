# Production DB Post-Deployment Audit

## 1. All odyssey_* tables
| table_name |
|---|
| odyssey_achievement_definitions |
| odyssey_achievements |
| odyssey_audit_logs |
| odyssey_balance_configs |
| odyssey_challenges |
| odyssey_chapter_definitions |
| odyssey_chapter_progress |
| odyssey_chest_definitions |
| odyssey_chests |
| odyssey_creative_items |
| odyssey_creative_prompt_definitions |
| odyssey_creative_submissions |
| odyssey_crews |
| odyssey_daily_turns |
| odyssey_drop_tables |
| odyssey_lore_definitions |
| odyssey_lore_unlocks |
| odyssey_player_relics |
| odyssey_quest_definitions |
| odyssey_quests |
| odyssey_realm_definitions |
| odyssey_realm_progress |
| odyssey_relic_definitions |
| odyssey_relics |
| odyssey_schema_version |
| odyssey_season_definitions |
| odyssey_system_config |
| odyssey_user_profiles |

## 2. All indexes for odyssey_* tables
| indexname | tablename |
|---|---|
| idx_odyssey_achievement_definitions_code | odyssey_achievement_definitions |
| idx_odyssey_achievement_definitions_deleted | odyssey_achievement_definitions |
| idx_odyssey_achievement_definitions_published | odyssey_achievement_definitions |
| idx_odyssey_achievement_definitions_season | odyssey_achievement_definitions |
| idx_odyssey_achievement_definitions_trigger | odyssey_achievement_definitions |
| odyssey_achievement_definitions_code_key | odyssey_achievement_definitions |
| odyssey_achievement_definitions_pkey | odyssey_achievement_definitions |
| idx_odyssey_achievements_crew_id | odyssey_achievements |
| idx_odyssey_achievements_trigger | odyssey_achievements |
| idx_odyssey_achievements_uid | odyssey_achievements |
| odyssey_achievements_pkey | odyssey_achievements |
| uniq_odyssey_achievements_crew_id_code | odyssey_achievements |
| uniq_odyssey_achievements_uid_code | odyssey_achievements |
| idx_odyssey_audit_logs_admin_uid | odyssey_audit_logs |
| idx_odyssey_audit_logs_created_at | odyssey_audit_logs |
| idx_odyssey_audit_logs_operation | odyssey_audit_logs |
| idx_odyssey_audit_logs_request_id | odyssey_audit_logs |
| idx_odyssey_audit_logs_resource | odyssey_audit_logs |
| odyssey_audit_logs_pkey | odyssey_audit_logs |
| idx_odyssey_balance_configs_key | odyssey_balance_configs |
| odyssey_balance_configs_key_key | odyssey_balance_configs |
| odyssey_balance_configs_pkey | odyssey_balance_configs |
| idx_odyssey_challenges_quest_id | odyssey_challenges |
| odyssey_challenges_pkey | odyssey_challenges |
| idx_odyssey_chapter_definitions_deleted | odyssey_chapter_definitions |
| idx_odyssey_chapter_definitions_published | odyssey_chapter_definitions |
| idx_odyssey_chapter_definitions_realm | odyssey_chapter_definitions |
| idx_odyssey_chapter_definitions_slug | odyssey_chapter_definitions |
| odyssey_chapter_definitions_pkey | odyssey_chapter_definitions |
| odyssey_chapter_definitions_slug_key | odyssey_chapter_definitions |
| idx_odyssey_chapter_progress_crew_id | odyssey_chapter_progress |
| idx_odyssey_chapter_progress_realm | odyssey_chapter_progress |
| idx_odyssey_chapter_progress_status | odyssey_chapter_progress |
| odyssey_chapter_progress_pkey | odyssey_chapter_progress |
| idx_odyssey_chest_definitions_deleted | odyssey_chest_definitions |
| idx_odyssey_chest_definitions_published | odyssey_chest_definitions |
| idx_odyssey_chest_definitions_season | odyssey_chest_definitions |
| idx_odyssey_chest_definitions_slug | odyssey_chest_definitions |
| odyssey_chest_definitions_pkey | odyssey_chest_definitions |
| odyssey_chest_definitions_slug_key | odyssey_chest_definitions |
| idx_odyssey_chests_uid | odyssey_chests |
| odyssey_chests_pkey | odyssey_chests |
| idx_odyssey_creative_items_author_uid | odyssey_creative_items |
| idx_odyssey_creative_items_crew_id | odyssey_creative_items |
| odyssey_creative_items_pkey | odyssey_creative_items |
| idx_odyssey_creative_prompt_definitions_deleted | odyssey_creative_prompt_definitions |
| idx_odyssey_creative_prompt_definitions_published | odyssey_creative_prompt_definitions |
| idx_odyssey_creative_prompt_definitions_realm | odyssey_creative_prompt_definitions |
| idx_odyssey_creative_prompt_definitions_season | odyssey_creative_prompt_definitions |
| idx_odyssey_creative_prompt_definitions_slug | odyssey_creative_prompt_definitions |
| odyssey_creative_prompt_definitions_pkey | odyssey_creative_prompt_definitions |
| odyssey_creative_prompt_definitions_slug_key | odyssey_creative_prompt_definitions |
| idx_odyssey_creative_submissions_author_uid | odyssey_creative_submissions |
| idx_odyssey_creative_submissions_crew_id | odyssey_creative_submissions |
| idx_odyssey_creative_submissions_quest_id | odyssey_creative_submissions |
| idx_odyssey_creative_submissions_status | odyssey_creative_submissions |
| odyssey_creative_submissions_pkey | odyssey_creative_submissions |
| odyssey_crews_pkey | odyssey_crews |
| odyssey_daily_turns_pkey | odyssey_daily_turns |
| idx_odyssey_drop_tables_deleted | odyssey_drop_tables |
| idx_odyssey_drop_tables_published | odyssey_drop_tables |
| odyssey_drop_tables_pkey | odyssey_drop_tables |
| uniq_odyssey_drop_tables_chest_rarity | odyssey_drop_tables |
| idx_odyssey_lore_definitions_chapter | odyssey_lore_definitions |
| idx_odyssey_lore_definitions_deleted | odyssey_lore_definitions |
| idx_odyssey_lore_definitions_published | odyssey_lore_definitions |
| idx_odyssey_lore_definitions_realm | odyssey_lore_definitions |
| idx_odyssey_lore_definitions_season | odyssey_lore_definitions |
| idx_odyssey_lore_definitions_slug | odyssey_lore_definitions |
| odyssey_lore_definitions_pkey | odyssey_lore_definitions |
| odyssey_lore_definitions_slug_key | odyssey_lore_definitions |
| idx_odyssey_lore_unlocks_crew_id | odyssey_lore_unlocks |
| idx_odyssey_lore_unlocks_lore_slug | odyssey_lore_unlocks |
| odyssey_lore_unlocks_pkey | odyssey_lore_unlocks |
| idx_odyssey_player_relics_uid | odyssey_player_relics |
| odyssey_player_relics_pkey | odyssey_player_relics |
| uniq_odyssey_player_relics_uid_slug | odyssey_player_relics |
| idx_odyssey_quest_definitions_chapter | odyssey_quest_definitions |
| idx_odyssey_quest_definitions_deleted | odyssey_quest_definitions |
| idx_odyssey_quest_definitions_published | odyssey_quest_definitions |
| idx_odyssey_quest_definitions_realm | odyssey_quest_definitions |
| idx_odyssey_quest_definitions_required_chapter | odyssey_quest_definitions |
| idx_odyssey_quest_definitions_season | odyssey_quest_definitions |
| idx_odyssey_quest_definitions_slug | odyssey_quest_definitions |
| odyssey_quest_definitions_pkey | odyssey_quest_definitions |
| odyssey_quest_definitions_slug_key | odyssey_quest_definitions |
| idx_odyssey_quests_chapter | odyssey_quests |
| idx_odyssey_quests_crew_id | odyssey_quests |
| idx_odyssey_quests_status | odyssey_quests |
| odyssey_quests_pkey | odyssey_quests |
| uniq_odyssey_quests_crew_id_slug | odyssey_quests |
| idx_odyssey_realm_definitions_deleted | odyssey_realm_definitions |
| idx_odyssey_realm_definitions_order | odyssey_realm_definitions |
| idx_odyssey_realm_definitions_published | odyssey_realm_definitions |
| idx_odyssey_realm_definitions_slug | odyssey_realm_definitions |
| odyssey_realm_definitions_pkey | odyssey_realm_definitions |
| odyssey_realm_definitions_slug_key | odyssey_realm_definitions |
| odyssey_realm_progress_pkey | odyssey_realm_progress |
| idx_odyssey_relic_definitions_deleted | odyssey_relic_definitions |
| idx_odyssey_relic_definitions_published | odyssey_relic_definitions |
| idx_odyssey_relic_definitions_realm | odyssey_relic_definitions |
| idx_odyssey_relic_definitions_slug | odyssey_relic_definitions |
| odyssey_relic_definitions_pkey | odyssey_relic_definitions |
| odyssey_relic_definitions_slug_key | odyssey_relic_definitions |
| idx_odyssey_relics_uid | odyssey_relics |
| odyssey_relics_pkey | odyssey_relics |
| odyssey_schema_version_pkey | odyssey_schema_version |
| idx_odyssey_season_definitions_deleted | odyssey_season_definitions |
| idx_odyssey_season_definitions_published | odyssey_season_definitions |
| idx_odyssey_season_definitions_slug | odyssey_season_definitions |
| odyssey_season_definitions_pkey | odyssey_season_definitions |
| odyssey_season_definitions_slug_key | odyssey_season_definitions |
| idx_odyssey_system_config_key | odyssey_system_config |
| odyssey_system_config_pkey | odyssey_system_config |
| idx_odyssey_user_profiles_crew_id | odyssey_user_profiles |
| odyssey_user_profiles_pkey | odyssey_user_profiles |

## 3. All foreign keys for odyssey_* tables
| column_name | foreign_column_name | foreign_table_name | table_name |
|---|---|---|---|
| crew_id | id | odyssey_crews | odyssey_user_profiles |
| crew_id | id | odyssey_crews | odyssey_quests |
| quest_id | id | odyssey_quests | odyssey_challenges |
| completed_by | uid | odyssey_user_profiles | odyssey_challenges |
| crew_id | id | odyssey_crews | odyssey_creative_items |
| author_uid | uid | odyssey_user_profiles | odyssey_creative_items |
| uid | uid | odyssey_user_profiles | odyssey_daily_turns |
| crew_id | id | odyssey_crews | odyssey_achievements |
| uid | uid | odyssey_user_profiles | odyssey_relics |
| uid | uid | odyssey_user_profiles | odyssey_chests |
| quest_id | id | odyssey_quests | odyssey_creative_submissions |
| crew_id | id | odyssey_crews | odyssey_creative_submissions |
| chest_slug | slug | odyssey_chest_definitions | odyssey_drop_tables |

## 4. All RLS policies for odyssey_* tables
| cmd | policyname | roles | tablename |
|---|---|---|---|
| ALL | Allow service_role full access | {service_role} | odyssey_achievement_definitions |
| ALL | Allow service_role full access on achievements | {public} | odyssey_achievements |
| ALL | Allow service_role full access | {service_role} | odyssey_audit_logs |
| ALL | Allow service_role full access | {service_role} | odyssey_balance_configs |
| ALL | Allow service_role full access on challenges | {public} | odyssey_challenges |
| ALL | Allow service_role full access | {service_role} | odyssey_chapter_definitions |
| ALL | Allow service_role full access | {service_role} | odyssey_chapter_progress |
| ALL | Allow service_role full access | {service_role} | odyssey_chest_definitions |
| ALL | Allow service_role full access | {service_role} | odyssey_chests |
| ALL | Allow service_role full access on chests | {public} | odyssey_chests |
| ALL | Allow service_role full access on creative_items | {public} | odyssey_creative_items |
| ALL | Allow service_role full access | {service_role} | odyssey_creative_prompt_definitions |
| ALL | Allow service_role full access | {service_role} | odyssey_creative_submissions |
| ALL | Allow service_role full access | {public} | odyssey_crews |
| ALL | Allow service_role full access on daily_turns | {public} | odyssey_daily_turns |
| ALL | Allow service_role full access | {service_role} | odyssey_drop_tables |
| ALL | Allow service_role full access | {service_role} | odyssey_lore_definitions |
| ALL | Allow service_role full access | {service_role} | odyssey_lore_unlocks |
| ALL | Allow service_role full access | {service_role} | odyssey_player_relics |
| ALL | Allow service_role full access | {service_role} | odyssey_quest_definitions |
| ALL | Allow service_role full access on quests | {public} | odyssey_quests |
| ALL | Allow service_role full access | {service_role} | odyssey_realm_definitions |
| ALL | Allow service_role full access on realm_progress | {public} | odyssey_realm_progress |
| ALL | Allow service_role full access | {service_role} | odyssey_relic_definitions |
| ALL | Allow service_role full access | {service_role} | odyssey_relics |
| ALL | Allow service_role full access on relics | {public} | odyssey_relics |
| ALL | Allow service_role full access | {service_role} | odyssey_schema_version |
| ALL | Allow service_role full access | {service_role} | odyssey_season_definitions |
| ALL | Allow service_role full access | {service_role} | odyssey_system_config |
| ALL | Allow service_role full access on user_profiles | {public} | odyssey_user_profiles |

## 5. odyssey_schema_version
| key | updated_at | value |
|---|---|---|
| schema_version | 2026-08-06 06:12:55.228419+00 | 12 |

## 6. Row counts for seeded tables
| count | table_name |
|---|---|
| 3 | odyssey_realm_definitions |
| 4 | odyssey_chapter_definitions |
| 12 | odyssey_quest_definitions |
| 6 | odyssey_creative_prompt_definitions |
| 10 | odyssey_achievement_definitions |
| 1 | odyssey_season_definitions |
| 8 | odyssey_lore_definitions |
| 9 | odyssey_balance_configs |



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
