-- Schemas
SELECT schema_name FROM information_schema.schemata;

-- Tables in public
SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';

-- Extensions
SELECT extname, extversion FROM pg_extension;

-- RLS
SELECT relname, relrowsecurity FROM pg_class WHERE relnamespace = 'public'::regnamespace AND relkind = 'r';

-- Policies
SELECT * FROM pg_policies WHERE schemaname = 'public';

-- Triggers
SELECT trigger_name, event_manipulation, event_object_table FROM information_schema.triggers WHERE trigger_schema = 'public';

-- Indexes
SELECT indexname, tablename, indexdef FROM pg_indexes WHERE schemaname = 'public';

-- Functions
SELECT routine_name FROM information_schema.routines WHERE routine_schema = 'public';

-- Check if schema_version or schema_migrations exist and read them
SELECT EXISTS (
   SELECT FROM information_schema.tables 
   WHERE  table_schema = 'public'
   AND    table_name   = 'schema_migrations'
   ) as has_schema_migrations;

SELECT EXISTS (
   SELECT FROM information_schema.tables 
   WHERE  table_schema = 'public'
   AND    table_name   = 'schema_version'
   ) as has_schema_version;
