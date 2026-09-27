require('dotenv').config();

// Applies Migration 087: Replace tasks for 17 September 2026.
// 1. Re-verifies 544/545 carry 0 submissions + 0 ledger rows, then exact-ID
//    hard-deletes ONLY those two rows on 2026-09-17 (ABORTS otherwise).
// 2. Idempotent pre-clean: deletes ONLY the 3 exact new titles on
//    2026-09-17 (previous partial run of THIS migration).
// 3. Inserts the 3 canonical tasks (config mirrors
//    supabase/migrations/087_tasks_replace_2026_09_17.sql exactly).
// 4. Bumps odyssey_schema_version to 087_tasks_replace_2026_09_17.
// Does NOT touch any other date, task, or config.

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

const NEW_TITLES = [
  'Benerin Sesuatu yang Sering Bikin Ribet',
  'Bandingkan Sebelum Beli',
  'Jelaskan Sampai Orang Lain Paham',
];

async function get(path) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { headers });
  if (!r.ok) throw new Error(`GET ${path}: ${r.status} ${await r.text()}`);
  return r.json();
}

async function run() {
  console.log('=== Applying Migration 087: Replace tasks for 2026-09-17 ===');

  // 0. Re-audit 544/545 immediately before mutating.
  console.log('[0] Re-audit legacy tasks 544/545...');
  const legacy = await get('odyssey_tasks?id=in.(544,545)&select=id,title,active_date,is_active');
  console.log('   legacy rows: ' + JSON.stringify(legacy.map((t) => [t.id, t.title, t.active_date])));
  const subs544 = await get('odyssey_task_submissions?task_id=eq.544&select=id,status');
  const subs545 = await get('odyssey_task_submissions?task_id=eq.545&select=id,status');
  console.log(`   submissions: 544=${subs544.length} 545=${subs545.length}`);
  const allSubIds = [...subs544, ...subs545].map((s) => s.id);
  let ledger = [];
  if (allSubIds.length) {
    ledger = await get(`odyssey_coin_transactions?select=id&reference_id=in.(${allSubIds.join(',')})`);
  }
  console.log(`   linked ledger rows: ${ledger.length}`);
  if (subs544.length || subs545.length || ledger.length) {
    throw new Error('ABORT: legacy tasks carry submissions/ledger — hard-delete unsafe. Deactivate instead.');
  }
  const other = await get('odyssey_tasks?active_date=eq.2026-09-17&select=id');
  const otherIds = other.map((t) => t.id).sort((a, b) => a - b);
  console.log('   all rows on 2026-09-17: ' + JSON.stringify(otherIds));
  if (JSON.stringify(otherIds) !== '[544,545]') {
    throw new Error('ABORT: unexpected rows on 2026-09-17 — audit before proceeding.');
  }

  // 1. Exact-ID hard delete of legacy tasks.
  console.log('[1] Exact-ID delete of 544/545...');
  const del = await fetch(
    `${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(544,545)&active_date=eq.2026-09-17`,
    { method: 'DELETE', headers }
  );
  if (!del.ok) throw new Error(`DELETE legacy failed: ${del.status} ${await del.text()}`);
  console.log('   deleted 544/545 (verified in step [4]).');

  // 2. Idempotent pre-clean of THIS migration's titles only.
  console.log('[2] Idempotent pre-clean (exact new titles, exact date)...');
  const pre = await get(
    'odyssey_tasks?active_date=eq.2026-09-17&select=id,title&title=in.(Benerin%20Sesuatu%20yang%20Sering%20Bikin%20Ribet,Bandingkan%20Sebelum%20Beli,Jelaskan%20Sampai%20Orang%20Lain%20Paham)'
  );
  if (pre.length) {
    const ids = pre.map((t) => t.id).join(',');
    const d2 = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids})`, { method: 'DELETE', headers });
    if (!d2.ok) throw new Error(`Pre-clean DELETE failed: ${d2.status} ${await d2.text()}`);
    console.log(`   removed ${pre.length} partial-run row(s).`);
  } else {
    console.log('   no partial-run rows found.');
  }

  // 3. Insert 3 canonical tasks (payload mirrors 087 SQL migration).
  console.log('[3] Inserting 3 canonical tasks...');
  const tasks = [
    {
      family_id: 'demo-crew-1',
      title: 'Benerin Sesuatu yang Sering Bikin Ribet',
      description:
        'Temukan SATU hal kecil yang nyata di rumah, tempat kerja, atau lingkungan sekitarmu yang membuat aktivitas sehari-hari jadi kurang praktis atau kurang nyaman — misalnya tumpukan barang yang menghalangi jalan, kabel berbelit yang susah dirapikan, wadah tanpa label yang bikin salah ambil, atau sudut ruangan yang gelap dan rawan. Ambil 1 foto kondisi aslinya dan jelaskan singkat: apa yang membuatnya ribet, kenapa itu masalah bagimu, dan ide sederhana untuk memperbaikinya. Tidak perlu mengubah apa pun secara fisik dan tidak perlu membeli sesuatu — cukup amati, foto, dan jelaskan idemu.',
      task_type: 'PHOTO_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 1,
      active_date: '2026-09-17',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        max_files: 1,
        camera_only: false,
        instruction:
          '1) Temukan satu masalah NYATA yang bisa difoto di lingkungan sekitarmu — sesuatu yang benar-benar membuat aktivitas jadi ribet atau tidak nyaman (bukan sekadar foto benda biasa). 2) Ambil atau unggah 1 foto kondisi tersebut (boleh dari kamera langsung atau galeri). 3) Wajib tulis penjelasan singkat di kolom catatan dengan 3 bagian: (a) apa yang membuatnya ribet, (b) kenapa menurutmu itu masalah, (c) ide sederhana untuk memperbaikinya. Tidak wajib mengubah atau membeli apa pun. Admin akan menolak submission tanpa foto, tanpa penjelasan 3 bagian, atau foto yang tidak menunjukkan masalah yang bisa diamati.',
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Bandingkan Sebelum Beli',
      description:
        'Pilih SATU barang yang memang mungkin kamu beli dalam kehidupan nyata — misalnya charger, sepatu, alat kerja, perlengkapan rumah, aksesori HP, atau kebutuhan sehari-hari lainnya. Riset sendiri 2-3 pilihan NYATA barang tersebut (boleh dari toko, marketplace, atau katalog), lalu tulis hasil perbandinganmu: nama barang, setiap pilihan beserta harga dan fungsi/fitur utama, kelebihan dan kekurangan masing-masing, biaya tambahan jika ada, pilihan akhirmu, dan alasanmu berdasarkan kebutuhanmu. Tidak perlu membeli apa pun — cukup teliti sebelum memutuskan.',
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      active_date: '2026-09-17',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        prompt:
          'Tulis hasil perbandinganmu dengan format:\n1) Nama barang yang ingin dibeli:\n2) Pilihan 1 — nama/merk, harga, fungsi atau fitur utama, kelebihan, kekurangan:\n3) Pilihan 2 — nama/merk, harga, fungsi atau fitur utama, kelebihan, kekurangan:\n4) (Opsional) Pilihan 3 dengan format yang sama:\n5) Biaya tambahan jika ada (misalnya ongkir, garansi, aksesori):\n6) Pilihan akhirmu:\n7) Alasan pilihanmu berdasarkan kebutuhanmu (2-3 kalimat):\n\nWajib membandingkan minimal 2 pilihan nyata dengan harga masing-masing. Boleh mencantumkan link atau sumber sebagai teks biasa. Jangan cantumkan data pribadi. Admin akan menolak jawaban tanpa harga, dengan kurang dari 2 pilihan, tanpa pilihan akhir, atau tanpa alasan.',
        minimum_characters: 200,
        maximum_characters: 2000,
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Jelaskan Sampai Orang Lain Paham',
      description:
        'Pilih SATU hal yang benar-benar kamu pahami atau sering kamu lakukan — misalnya cara melipat baju dengan cepat, cara menyeduh minuman favoritmu, cara merapikan file di HP, tips mengatur uang belanja, cara melakukan bagian dari pekerjaanmu, atau keahlian praktis lainnya. Rekam video sekitar 60 detik seolah-olah kamu menjelaskannya kepada seseorang yang belum memahaminya sama sekali: tunjukkan atau jelaskan langkahnya satu per satu dengan bahasa sederhana. Wajah tidak wajib terlihat — boleh merekam tangan, benda, atau prosesnya sambil menjelaskan dengan suaramu.',
      task_type: 'VIDEO',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-09-17',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        recording: {
          enabled: true,
          max_duration_seconds: 60,
          camera_facing: 'user',
          instruction:
            'Rekam video maksimal 60 detik, santai seperti video story. Jelaskan SATU hal yang benar-benar kamu kuasai dari pengalamanmu sendiri (misalnya demonstrasi sederhana, tutorial singkat, penjelasan proses, tips praktis, atau cara melakukan sesuatu):\n1. Sebutkan dulu hal apa yang akan kamu jelaskan\n2. Jelaskan langkahnya satu per satu dengan bahasa sederhana, seolah pendengarmu belum tahu sama sekali\n3. Akhiri dengan satu tips agar berhasil\n\nWajah tidak wajib terlihat — boleh merekam tangan, benda, atau prosesnya sambil menjelaskan. Bukan kuis dan tidak perlu topik akademik. Jangan menyebut data pribadi sensitif. Bicaralah senatural mungkin!',
        },
      },
    },
  ];
  const postRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks`, {
    method: 'POST',
    headers,
    body: JSON.stringify(tasks),
  });
  if (!postRes.ok) throw new Error(`INSERT failed: ${postRes.status} ${await postRes.text()}`);
  const created = await postRes.json();
  console.log(`   inserted ${created.length} tasks:`);
  created.forEach((t) => {
    console.log(`   - [ID ${t.id}] Step #${t.step_order}: "${t.title}" (${t.task_type}, ${t.evaluation_type}, ${t.reward_coins}c/${t.reward_xp}xp)`);
  });

  // 4. Verify exactly 3 active tasks on 2026-09-17.
  console.log('[4] Verifying 2026-09-17 state...');
  const after = await get(
    'odyssey_tasks?active_date=eq.2026-09-17&select=id,title,task_type,evaluation_type,step_order,reward_coins,reward_xp,target_scope,family_id,is_active,config&order=step_order.asc'
  );
  const active = after.filter((t) => t.is_active);
  console.log(`   total rows: ${after.length}, active: ${active.length}`);
  active.forEach((t) => console.log(`   - [ID ${t.id}] #${t.step_order} "${t.title}" ${t.task_type}/${t.evaluation_type} ${t.reward_coins}c/${t.reward_xp}xp scope=${t.target_scope} fam=${t.family_id}`));
  if (active.length !== 3) throw new Error(`Expected exactly 3 active tasks, found ${active.length}`);
  const titles = active.map((t) => t.title);
  if (JSON.stringify(titles) !== JSON.stringify(NEW_TITLES)) throw new Error('Title/step mismatch: ' + JSON.stringify(titles));

  // 5. Bump schema version.
  console.log('[5] Updating odyssey_schema_version...');
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_schema_version`, {
    method: 'POST',
    headers: { ...headers, Prefer: 'resolution=merge-duplicates' },
    body: JSON.stringify([{ key: 'schema_version', value: '087_tasks_replace_2026_09_17', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 087 Applied Successfully ===');
  console.log('NEW_IDS=' + JSON.stringify(active.map((t) => t.id)));
}

run().catch((err) => {
  console.error('Migration failed:', err);
  process.exit(1);
});
