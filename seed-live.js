import fetch from 'node-fetch';
import * as dotenv from 'dotenv';
dotenv.config();

const SUPABASE_URL = process.env.SUPABASE_URL;
const SUPABASE_KEY = process.env.SUPABASE_SERVICE_KEY;

async function run() {
  const headers = {
    'apikey': SUPABASE_KEY,
    'Authorization': `Bearer ${SUPABASE_KEY}`,
    'Content-Type': 'application/json',
    'Prefer': 'return=representation'
  };

  console.log('Seeding demo crew...');

  // 1. Delete existing quests & challenges for demo crew
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_challenges`, {
    method: 'DELETE',
    headers: { ...headers, 'Prefer': 'return=minimal' }
  });
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_quests?crew_id=eq.demo-crew-1`, {
    method: 'DELETE',
    headers: { ...headers, 'Prefer': 'return=minimal' }
  });
  console.log('Deleted old quests & challenges');

  const quests = [
    // ARC 1: THE WHISPERING WOODS (Hutan Berbisik)
    { id: 101, title: 'Bab 1: Panggilan Pertama', slug: 'tutorial', type: 'SOLO' },
    { id: 102, title: 'Bab 2: Jejak Kaki Raksasa', slug: 'jejak-kaki', type: 'RELAY' },
    { id: 103, title: 'Bab 3: Melodi dari Dedaunan', slug: 'melodi-hutan', type: 'CREATIVE' },
    { id: 104, title: 'Bab 4: Teka-teki Sang Penjaga', slug: 'teka-teki-penjaga', type: 'PUZZLE' },
    { id: 105, title: 'Bab 5: Kunci Akar Berlian', slug: 'kunci-akar', type: 'GROUP' },
    // ARC 2: THE CLOCKWORK CITY
    { id: 106, title: 'Bab 6: Gerbang Roda Gigi', slug: 'gerbang-roda', type: 'RELAY' },
    { id: 107, title: 'Bab 7: Cetak Biru Keluarga', slug: 'cetak-biru', type: 'CREATIVE' },
    { id: 108, title: 'Bab 8: Mesin Waktu Mini', slug: 'mesin-waktu', type: 'SOLO' },
    { id: 109, title: 'Bab 9: Sandi Biner Kota', slug: 'sandi-biner', type: 'PUZZLE' },
    { id: 110, title: 'Bab 10: Menghentikan Sabotase', slug: 'sabotase-mekanik', type: 'GROUP' },
    // ARC 3: THE STARLIT LIBRARY
    { id: 111, title: 'Bab 11: Buku yang Terbang', slug: 'buku-terbang', type: 'SOLO' },
    { id: 112, title: 'Bab 12: Mitos Rasi Bintang', slug: 'mitos-rasi', type: 'CREATIVE' },
    { id: 113, title: 'Bab 13: Lorong Memori', slug: 'lorong-memori', type: 'RELAY' },
    { id: 114, title: 'Bab 14: Kaca Pembesar Ajaib', slug: 'kaca-pembesar', type: 'GROUP' },
    { id: 115, title: 'Bab 15: Ensiklopedia Akhir', slug: 'ensiklopedia-akhir', type: 'CREATIVE' },
  ];

  for (let i = 0; i < quests.length; i++) {
    const q = quests[i];
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_quests`, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        id: q.id,
        crew_id: 'demo-crew-1',
        template_slug: q.slug,
        title: q.title,
        status: i === 0 ? 'ACTIVE' : 'PENDING',
        started_at: i === 0 ? new Date().toISOString() : null
      })
    });
  }

  const challenges = [
    { quest_id: 101, slug: 'q1-1', description: 'Lihat ke luar jendela rumahmu. Amati langit, apakah ada awan yang berbentuk hewan?', status: 'PENDING' },
    { quest_id: 102, slug: 'q2-1', description: 'Peran 1 (Seeker): Ukur telapak kakimu dengan sebuah benda (seperti buku atau pensil).', status: 'PENDING' },
    { quest_id: 102, slug: 'q2-2', description: 'Peran 2 (Guide): Letakkan benda tersebut di ruang tengah sebagai "Jejak Raksasa".', status: 'PENDING' },
    { quest_id: 103, slug: 'q3-1', description: 'Ketikkan: Apa satu kalimat motivasi dari keluargamu yang paling kamu ingat?', status: 'PENDING' },
    { quest_id: 104, slug: 'q4-1', description: 'Pecahkan bersama: "Aku punya mulut tapi tak bicara, punya tempat tidur tapi tak tidur. Siapa aku?" (Jawab: Sungai)', status: 'PENDING' },
    { quest_id: 105, slug: 'q5-1', description: 'Sentuh tiga benda berbahan kayu yang ada di sekitarmu.', status: 'PENDING' },
    { quest_id: 106, slug: 'q6-1', description: 'Peran 1 (Builder): Gambar sebuah lingkaran roda gigi di kertas.', status: 'PENDING' },
    { quest_id: 106, slug: 'q6-2', description: 'Peran 2 (Guide): Tuliskan satu kemampuan khusus roda gigi tersebut.', status: 'PENDING' },
    { quest_id: 107, slug: 'q7-1', description: 'Ciptakan "Cetak Biru": Tuliskan satu aturan/kebiasaan unik rumah kalian.', status: 'PENDING' },
    { quest_id: 108, slug: 'q8-1', description: 'Tutup mata selama 10 detik dan hitung mundur. Rasakan berjalannya waktu.', status: 'PENDING' },
    { quest_id: 109, slug: 'q9-1', description: 'Tuliskan nama salah satu anggota keluarga secara terbalik!', status: 'PENDING' },
    { quest_id: 110, slug: 'q10-1', description: 'Kumpulkan lima barang sekecil mungkin di satu tempat (puing mekanik).', status: 'PENDING' },
    { quest_id: 111, slug: 'q11-1', description: 'Pilih buku paling tua di rumah, buka halaman acak, dan baca kalimat pertamanya.', status: 'PENDING' },
    { quest_id: 112, slug: 'q12-1', description: 'Beri nama rasi bintang baru yang hanya diketahui keluarga kalian.', status: 'PENDING' },
    { quest_id: 113, slug: 'q13-1', description: 'Peran 1 (Guide): Ceritakan satu kenangan lucu liburan masa lalu.', status: 'PENDING' },
    { quest_id: 113, slug: 'q13-2', description: 'Peran 2 (Builder): Tambahkan detail yang terlewat dari cerita itu!', status: 'PENDING' },
    { quest_id: 114, slug: 'q14-1', description: 'Bermain peran: Coba amati bagian belakang lehermu sendiri pakai kamera/cermin.', status: 'PENDING' },
    { quest_id: 115, slug: 'q15-1', description: 'Tuliskan pesan penutup untuk para penjelajah Odyssey lainnya!', status: 'PENDING' }
  ];

  for (const c of challenges) {
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_challenges`, {
      method: 'POST',
      headers,
      body: JSON.stringify(c)
    });
  }

  console.log('Seed 15 Epic Quests (1 Year Content) complete!');
}

run().catch(console.error);
