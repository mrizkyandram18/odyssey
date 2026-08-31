import * as dotenv from 'dotenv';
dotenv.config();

const SUPABASE_URL = process.env.SUPABASE_URL;
const SUPABASE_KEY = process.env.SUPABASE_SERVICE_KEY;

// Journey 1: The Whispering Woods (Hutan Berbisik)
const topics = [
  { slug: 'hutan-berbisik', title: 'Hutan Berbisik' },
  { slug: 'kota-jam-mekanik', title: 'Kota Jam Mekanik' },
  { slug: 'perpustakaan-bintang', title: 'Perpustakaan Bintang' }
];

const templates = {
  'hutan-berbisik': [
    { title: 'Jejak Sang Penjaga', desc: 'Seseorang atau sesuatu telah melewati jalan ini. Tugas kita adalah menemukan jejak mereka.', type: 'SOLO_OBSERVATION', q: 'Carilah satu benda di sekitarmu yang bentuknya menyerupai daun, dan deskripsikan bentuknya.' },
    { title: 'Estafet Pesan Kuno', desc: 'Pesan kuno harus disampaikan secara estafet agar sihirnya tidak pudar.', type: 'RELAY', q: 'Teruskan pesan ini kepada anggota keluargamu: "Cahaya akan memandu langkah kita."' },
    { title: 'Legenda Penjaga Hutan', desc: 'Setiap kru memiliki kisah pahlawannya. Saatnya menulis bab pertama dari petualangan kita.', type: 'CREATIVE_STORY', q: 'Tuliskan satu kalimat yang menggambarkan keberanian kru kita hari ini.' },
    { title: 'Misteri Akar Bercahaya', desc: 'Ada akar aneh yang bersinar di malam hari. Kita perlu menelitinya.', type: 'SOLO_RESEARCH', q: 'Temukan fakta menarik tentang tanaman apa saja dan bagikan di sini.' },
    { title: 'Menyusun Peta Konstelasi', desc: 'Bintang-bintang di Hutan Berbisik tampak berbeda. Mari kita susun petanya bersama.', type: 'GROUP_PUZZLE', q: 'Diskusikan bersama: Jika keluarga kita adalah sebuah rasi bintang, apa namanya?' },
    { title: 'Obor Harapan', desc: 'Nyalakan obor untuk menerangi jalan bagi petualang lain.', type: 'SOLO_ACTION', q: 'Lakukan satu kebaikan kecil untuk anggota keluargamu hari ini, dan catat di sini.' },
  ],
  'kota-jam-mekanik': [
    { title: 'Roda Gigi yang Hilang', desc: 'Mesin utama kota ini terhenti. Cari tahu apa yang menyebabkannya.', type: 'SOLO_OBSERVATION', q: 'Temukan benda bundar di sekitarmu yang bisa berputar, dan jelaskan benda apa itu.' },
    { title: 'Teka-Teki Waktu', desc: 'Waktu berjalan mundur di area ini! Pecahkan teka-teki untuk mengembalikannya.', type: 'GROUP_PUZZLE', q: 'Apa kenangan paling berkesan keluargamu di tahun lalu?' },
    { title: 'Surat dari Masa Lalu', desc: 'Sebuah pesan tersangkut di mesin waktu.', type: 'CREATIVE_STORY', q: 'Tulis pesan singkat untuk dirimu di masa depan (1 tahun dari sekarang).' },
    { title: 'Menghitung Putaran', desc: 'Tugas ini membutuhkan presisi dan ketelitian tinggi.', type: 'SOLO_ACTION', q: 'Hitung berapa banyak pintu yang ada di rumahmu, dan catat jumlahnya.' },
  ],
  'perpustakaan-bintang': [
    { title: 'Buku yang Terlupakan', desc: 'Di rak tertinggi, terdapat buku yang bersinar.', type: 'SOLO_OBSERVATION', q: 'Sebutkan judul buku favoritmu dan alasannya.' },
    { title: 'Estafet Pengetahuan', desc: 'Pengetahuan harus dibagikan agar terus hidup.', type: 'RELAY', q: 'Ceritakan satu hal baru yang kamu pelajari hari ini kepada anggota keluargamu.' },
    { title: 'Mantra Perlindungan', desc: 'Lindungi kru kita dengan mantra magis.', type: 'CREATIVE_STORY', q: 'Ciptakan satu kata sandi rahasia untuk kru kita.' },
    { title: 'Menyusun Arsip', desc: 'Gulungan-gulungan ini berserakan di mana-mana.', type: 'GROUP_PUZZLE', q: 'Urutkan anggota keluargamu berdasarkan bulan lahir dari yang paling awal.' },
  ]
};

async function run() {
  const headers = {
    'apikey': SUPABASE_KEY,
    'Authorization': `Bearer ${SUPABASE_KEY}`,
    'Content-Type': 'application/json',
    'Prefer': 'return=representation'
  };

  console.log('Menyiapkan data petualangan (Seeding Odyssey tasks)...');

  console.log('Menghapus misi lama...');
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_exercises?id=gt.0`, {
    method: 'DELETE',
    headers: { ...headers, 'Prefer': 'return=minimal' }
  });
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_missions?id=gt.0`, {
    method: 'DELETE',
    headers: { ...headers, 'Prefer': 'return=minimal' }
  });
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_mission_definitions?slug=like.hutan-berbisik-*`, {
    method: 'DELETE',
    headers: { ...headers, 'Prefer': 'return=minimal' }
  });
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_mission_definitions?slug=like.kota-jam-mekanik-*`, {
    method: 'DELETE',
    headers: { ...headers, 'Prefer': 'return=minimal' }
  });
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_mission_definitions?slug=like.perpustakaan-bintang-*`, {
    method: 'DELETE',
    headers: { ...headers, 'Prefer': 'return=minimal' }
  });
  console.log('Misi dan latihan lama telah dihapus.');

  const definitionsToInsert = [];
  const questsToInsert = [];
  const challengesToInsert = [];

  // Asumsi 'demo-crew-1' adalah kru aktif
  let questIdCounter = 1000;

  for (let i = 1; i <= 365; i++) {
    let topicIndex = 0;
    // Mengatur fase perjalanan (Journey progression)
    if (i > 30) topicIndex = 1;
    if (i > 60) topicIndex = 2;
    
    let topic = topics[topicIndex];
    let tplList = templates[topic.slug];
    let tpl = tplList[Math.floor(Math.random() * tplList.length)];
    
    let xp = 100;
    if (i % 7 === 0) xp = 300; // Quest Mingguan / Boss
    else if (i % 3 === 0) xp = 150;

    const templateSlug = `${topic.slug}-${i}`;

    definitionsToInsert.push({
      slug: templateSlug,
      title: `Giliran ${i}: ${tpl.title}`,
      description: tpl.desc,
      mission_type: tpl.type,
      journey: 'whispering-woods',
      course: 'hutan-berbisik',
      reward_xp: xp,
      is_mandatory: false,
      published: true,
      exercise_defs: [
        {
          slug: `q-def-${i}-1`,
          description: tpl.q,
          type: 'FREE_TEXT'
        }
      ]
    });

    let questId = questIdCounter++;
    
    questsToInsert.push({
      id: questId,
      family_id: 'demo-crew-1',
      template_slug: templateSlug,
      title: `Giliran ${i}: ${tpl.title}`,
      status: i <= 5 ? 'ACTIVE' : 'PENDING',
      started_at: i <= 5 ? new Date().toISOString() : null
    });

    challengesToInsert.push({
      mission_id: questId,
      slug: `q-def-${i}-1`,
      description: tpl.q,
      status: 'PENDING'
    });
  }

  console.log(`Memasukkan ${definitionsToInsert.length} definisi misi...`);
  for (let i = 0; i < definitionsToInsert.length; i += 100) {
    const chunk = definitionsToInsert.slice(i, i + 100);
    const res = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_mission_definitions`, {
      method: 'POST',
      headers,
      body: JSON.stringify(chunk)
    });
    if (!res.ok) console.error('Error definisions:', await res.text());
  }

  console.log(`Memasukkan ${questsToInsert.length} misi petualangan (instance)...`);
  for (let i = 0; i < questsToInsert.length; i += 100) {
    const chunk = questsToInsert.slice(i, i + 100);
    const res = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_missions`, {
      method: 'POST',
      headers,
      body: JSON.stringify(chunk)
    });
    if (!res.ok) console.error('Error missions:', await res.text());
  }

  console.log(`Memasukkan ${challengesToInsert.length} latihan (exercises)...`);
  for (let i = 0; i < challengesToInsert.length; i += 100) {
    const chunk = challengesToInsert.slice(i, i + 100);
    const res = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_exercises`, {
      method: 'POST',
      headers,
      body: JSON.stringify(chunk)
    });
    if (!res.ok) console.error('Error exercises:', await res.text());
  }

  console.log('Seed Petualangan Odyssey (365 Giliran) Selesai!');
}

run().catch(console.error);
