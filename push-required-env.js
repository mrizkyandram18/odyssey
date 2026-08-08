const { execSync } = require('child_process');

const envs = {
    "SUPABASE_URL": "https://hmrkssfhcxlvjzyigufd.supabase.co",
    "SUPABASE_SERVICE_KEY": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imhtcmtzc2ZoY3hsdmp6eWlndWZkIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc2ODU1MTc3NywiZXhwIjoyMDg0MTI3Nzc3fQ.OGA5s0RdtrVt0jKs99_-t9yE_DWH9YhHIUy14LUWlGg",
    "PARENT_ID": "demoparent",
    "SESSION_SIGNING_SECRET": "test-secret"
};

for (const [key, value] of Object.entries(envs)) {
    console.log(`Adding ${key}...`);
    try {
        execSync(`npx vercel env add ${key} production`, {
            input: value,
            stdio: ['pipe', 'inherit', 'inherit']
        });
    } catch (e) {
        console.error(`Failed to add ${key}`);
    }
}
