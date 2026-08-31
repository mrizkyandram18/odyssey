import fetch from 'node-fetch';
import * as dotenv from 'dotenv';
dotenv.config();

async function run() {
  const url = process.env.SUPABASE_URL + '/rest/v1/';
  const headers = { 'apikey': process.env.SUPABASE_SERVICE_KEY };
  const res = await fetch(url, { headers });
  const data = await res.json();
  console.log(Object.keys(data.paths).filter(p => p.includes('odyssey')));
}
run();
