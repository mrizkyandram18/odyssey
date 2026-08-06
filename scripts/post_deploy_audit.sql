-- 1. Dump all odyssey_* tables
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' AND table_name LIKE 'odyssey_%'
ORDER BY table_name;

-- 2. Dump all indexes for odyssey_* tables
SELECT tablename, indexname, indexdef 
FROM pg_indexes 
WHERE schemaname = 'public' AND tablename LIKE 'odyssey_%'
ORDER BY tablename, indexname;

-- 3. Dump all foreign keys for odyssey_* tables
SELECT
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_name LIKE 'odyssey_%';

-- 4. Dump all RLS policies for odyssey_* tables
SELECT tablename, policyname, roles, cmd, qual, with_check 
FROM pg_policies 
WHERE schemaname = 'public' AND tablename LIKE 'odyssey_%'
ORDER BY tablename;

-- 5. Dump odyssey_schema_version
SELECT * FROM odyssey_schema_version;

-- 6. Row counts for seeded tables
SELECT 'odyssey_realm_definitions' as table_name, count(*) FROM odyssey_realm_definitions
UNION ALL
SELECT 'odyssey_chapter_definitions', count(*) FROM odyssey_chapter_definitions
UNION ALL
SELECT 'odyssey_quest_definitions', count(*) FROM odyssey_quest_definitions
UNION ALL
SELECT 'odyssey_creative_prompt_definitions', count(*) FROM odyssey_creative_prompt_definitions
UNION ALL
SELECT 'odyssey_achievement_definitions', count(*) FROM odyssey_achievement_definitions
UNION ALL
SELECT 'odyssey_season_definitions', count(*) FROM odyssey_season_definitions
UNION ALL
SELECT 'odyssey_lore_definitions', count(*) FROM odyssey_lore_definitions
UNION ALL
SELECT 'odyssey_balance_configs', count(*) FROM odyssey_balance_configs;
