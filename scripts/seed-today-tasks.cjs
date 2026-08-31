const { config } = require('dotenv');
config();

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

async function seed() {
  console.log('--- Seeding 3-Step Linear Family Tasks for Today ---');

  // 1. Clear existing tasks
  await fetch(SUPABASE_URL + '/rest/v1/odyssey_tasks?id=gt.0', {
    method: 'DELETE',
    headers: { ...headers, Prefer: 'return=minimal' },
  });

  // 2. Insert 3 Linear Steps
  const tasks = [
    {
      title: 'Belajar Literasi Keuangan & Menabung',
      description: 'Tonton video singkat tentang cara mengelola uang saku, lalu jawab 2 pertanyaan kuis.',
      task_type: 'VIDEO_QUIZ',
      step_order: 1,
      reward_coins: 50,
      reward_xp: 100,
      is_active: true,
      config: {
        youtube_url: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
        questions: [
          {
            id: 1,
            question: 'Apa tujuan utama dari menyisihkan uang saku setiap hari?',
            options: [
              'A. Menabung untuk kebutuhan masa depan',
              'B. Menghabiskan uang secepat mungkin',
              'C. Membeli barang yang tidak dibutuhkan',
              'D. Meminjamkan tanpa dicatat',
            ],
            correct_answer: 'A',
          },
          {
            id: 2,
            question: 'Di mana tempat yang paling aman dan tepat untuk menyimpan tabungan?',
            options: [
              'A. Di bawah kasur kamar',
              'B. Rekening Bank / Celengan terkunci',
              'C. Ditinggal di saku celana',
              'D. Di atas meja makan',
            ],
            correct_answer: 'B',
          },
        ],
      },
    },
    {
      title: 'Rencana Anggaran & Catatan Tugas',
      description: 'Unggah file dokumen (PDF/Word/Excel) berisi catatan tugas sekolah, rencana belanja, atau ringkasan belajar hari ini.',
      task_type: 'DOCUMENT_UPLOAD',
      step_order: 2,
      reward_coins: 75,
      reward_xp: 150,
      is_active: true,
      config: {
        allowed_extensions: ['.pdf', '.xlsx', '.docx', '.doc', '.csv', '.txt'],
        prompt: 'Unggah file PDF / Excel / Word catatan belajar atau anggaran.',
      },
    },
    {
      title: 'Foto Meja Belajar / Aktivitas Produktif',
      description: 'Ambil foto bukti meja belajar yang rapi atau aktivitas produktifmu hari ini langsung menggunakan kamera HP.',
      task_type: 'PHOTO_PROOF',
      step_order: 3,
      reward_coins: 100,
      reward_xp: 200,
      is_active: true,
      config: {
        photo_prompt: 'Foto langsung meja belajar atau tugas yang sedang dikerjakan.',
        watermark: true,
      },
    },
  ];

  const res = await fetch(SUPABASE_URL + '/rest/v1/odyssey_tasks', {
    method: 'POST',
    headers,
    body: JSON.stringify(tasks),
  });

  const created = await res.json();
  console.log('Seeded tasks:', created.length, 'tasks successfully created!');
}

seed();
