import subprocess
import json
import os

queries = [
    {
        "title": "1. All odyssey_* tables",
        "query": "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'odyssey_%' ORDER BY table_name;"
    },
    {
        "title": "2. All indexes for odyssey_* tables",
        "query": "SELECT tablename, indexname FROM pg_indexes WHERE schemaname = 'public' AND tablename LIKE 'odyssey_%' ORDER BY tablename, indexname;"
    },
    {
        "title": "3. All foreign keys for odyssey_* tables",
        "query": "SELECT tc.table_name, kcu.column_name, ccu.table_name AS foreign_table_name, ccu.column_name AS foreign_column_name FROM information_schema.table_constraints AS tc JOIN information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema JOIN information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name AND ccu.table_schema = tc.table_schema WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_name LIKE 'odyssey_%';"
    },
    {
        "title": "4. All RLS policies for odyssey_* tables",
        "query": "SELECT tablename, policyname, roles, cmd FROM pg_policies WHERE schemaname = 'public' AND tablename LIKE 'odyssey_%' ORDER BY tablename;"
    },
    {
        "title": "5. odyssey_schema_version",
        "query": "SELECT * FROM odyssey_schema_version;"
    },
    {
        "title": "6. Row counts for seeded tables",
        "query": "SELECT 'odyssey_realm_definitions' as table_name, count(*) FROM odyssey_realm_definitions UNION ALL SELECT 'odyssey_chapter_definitions', count(*) FROM odyssey_chapter_definitions UNION ALL SELECT 'odyssey_quest_definitions', count(*) FROM odyssey_quest_definitions UNION ALL SELECT 'odyssey_creative_prompt_definitions', count(*) FROM odyssey_creative_prompt_definitions UNION ALL SELECT 'odyssey_achievement_definitions', count(*) FROM odyssey_achievement_definitions UNION ALL SELECT 'odyssey_season_definitions', count(*) FROM odyssey_season_definitions UNION ALL SELECT 'odyssey_lore_definitions', count(*) FROM odyssey_lore_definitions UNION ALL SELECT 'odyssey_balance_configs', count(*) FROM odyssey_balance_configs;"
    }
]

os.makedirs("docs/reports", exist_ok=True)
with open("docs/reports/production-db-validation.md", "w") as f:
    f.write("# Production DB Post-Deployment Audit\n\n")

    for q in queries:
        f.write(f"## {q['title']}\n")
        print(f"Running: {q['title']}")
        result = subprocess.run(
            ["supabase", "db", "query", q["query"], "--linked"],
            capture_output=True,
            text=True,
            shell=True
        )
        
        # the output contains "Initialising login role...\n{...}"
        out = result.stdout
        if "{" in out:
            json_str = out[out.index("{"):]
            try:
                data = json.loads(json_str)
                rows = data.get("rows", [])
                
                if not rows:
                    f.write("*No rows returned.*\n\n")
                    continue
                
                # build markdown table
                headers = list(rows[0].keys())
                f.write("| " + " | ".join(headers) + " |\n")
                f.write("|" + "|".join(["---"] * len(headers)) + "|\n")
                
                for row in rows:
                    f.write("| " + " | ".join([str(row.get(h, "")) for h in headers]) + " |\n")
                
            except Exception as e:
                f.write(f"```text\nError parsing JSON: {e}\nOutput:\n{out}\n```\n")
        else:
            f.write(f"```text\n{out}\n```\n")
        
        f.write("\n")
        
print("Audit report generated.")
