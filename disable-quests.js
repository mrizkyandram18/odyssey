const SUPABASE_URL = "https://hmrkssfhcxlvjzyigufd.supabase.co";
const SUPABASE_KEY = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imhtcmtzc2ZoY3hsdmp6eWlndWZkIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc2ODU1MTc3NywiZXhwIjoyMDg0MTI3Nzc3fQ.OGA5s0RdtrVt0jKs99_-t9yE_DWH9YhHIUy14LUWlGg";

const slugsToDisable = [
  'morning-light', 'gather-herbs', 'riddle-of-the-stones', 'shadow-trail',
  'the-old-growth', 'forest-riddle', 'clockwork-intro', 'star-observation',
  'constellation-map', 'library-lore'
];

async function run() {
  const headers = {
    'apikey': SUPABASE_KEY,
    'Authorization': `Bearer ${SUPABASE_KEY}`,
    'Content-Type': 'application/json',
    'Prefer': 'return=representation'
  };

  console.log('Checking quests...');
  
  for (const slug of slugsToDisable) {
    const res = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_quest_definitions?slug=eq.${slug}`, { headers });
    const rows = await res.json();
    
    if (rows.length === 0) {
      console.log(`${slug} — NOT FOUND`);
    } else {
      const quest = rows[0];
      if (!quest.published) {
        console.log(`${slug} — ALREADY DISABLED`);
      } else {
        await fetch(`${SUPABASE_URL}/rest/v1/odyssey_quest_definitions?slug=eq.${slug}`, {
          method: 'PATCH',
          headers,
          body: JSON.stringify({ published: false, deleted_at: new Date().toISOString() })
        });
        console.log(`${slug} — disabled`);
      }
    }
  }
}

run().catch(console.error);
