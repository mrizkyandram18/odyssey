require('dotenv').config();

const SUPABASE_URL = process.env.SUPABASE_URL;
const SUPABASE_KEY = process.env.SUPABASE_SERVICE_KEY;

async function fixHashes() {
  const hashVal = '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W'; // admin123
  const url = `${SUPABASE_URL}/rest/v1/odyssey_local_users?username=in.(admin,user_testing)`;
  const headers = {
    'apikey': SUPABASE_KEY,
    'Authorization': `Bearer ${SUPABASE_KEY}`,
    'Content-Type': 'application/json'
  };
  const res = await fetch(url, {
    method: 'PATCH',
    headers,
    body: JSON.stringify({ password_hash: hashVal })
  });
  console.log('Update status:', res.status);

  // Verify
  const checkRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_local_users?select=username,password_hash`, { headers });
  const checkData = await checkRes.json();
  console.log('Updated local user rows:', checkData);
}

fixHashes().catch(console.error);
