const { Client } = require('pg');
const client = new Client('postgresql://postgres.hmrkssfhcxlvjzyigufd:Odyssey2026!@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres');
client.connect().then(() => {
    return client.query("SELECT tablename FROM pg_tables WHERE schemaname = 'public'");
}).then(res => {
    console.log(res.rows.map(r => r.tablename));
    client.end();
}).catch(console.error);
