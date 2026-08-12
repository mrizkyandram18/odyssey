const fs = require('fs');

const quests = [];

function createQuest(slug, realm, chapter, title, description, type, challenges, xp, chest) {
    quests.push({ slug, realm, chapter, title, description, type, challenges, xp, chest });
}

// Realm 1: whispering-woods (Creativity, Language, Daily Life) -> Chapter: forest-wisdom
// Realm 2: clockwork-city (Math, Finance, Productivity) -> Chapter: mechanical-efficiency
// Realm 3: starlit-library (Digital Literacy, Work World) -> Chapter: modern-knowledge

// Whispering Woods (12 quests)
createQuest('ww-schedule', 'whispering-woods', 'forest-wisdom', 'Jadwal Hutan', 'Buatlah rencana kegiatan sehari-hari agar waktumu lebih teratur.', 'SOLO', [
    { slug: 'buat-jadwal', description: 'Tuliskan daftar kegiatanmu dari pagi hingga malam hari ini.', type: 'WRITE' },
    { slug: 'prioritas', description: 'Pilih satu kegiatan yang paling penting dan jelaskan alasannya.', type: 'WRITE' }
], 100, 'wooden-chest');

createQuest('ww-message', 'whispering-woods', 'forest-wisdom', 'Pesan Angin', 'Belajar menulis pesan yang jelas dan sopan.', 'SOLO', [
    { slug: 'pesan-sopan', description: 'Tuliskan contoh pesan singkat untuk meminta maaf karena terlambat.', type: 'WRITE' },
    { slug: 'perbaiki-pesan', description: 'Perbaiki kalimat ini: "Bales woy, gw nungguin nih!"', type: 'WRITE' }
], 100, 'wooden-chest');

createQuest('ww-creative-story', 'whispering-woods', 'forest-wisdom', 'Cerita Pendek', 'Gunakan imajinasimu untuk membuat cerita pendek.', 'CREATIVE', [
    { slug: 'tulis-cerita', description: 'Buatlah cerita pendek 2-3 kalimat tentang hewan yang menemukan harta karun.', type: 'WRITE' }
], 120, 'bronze-chest');

createQuest('ww-reading', 'whispering-woods', 'forest-wisdom', 'Membaca Tanda', 'Latihan membaca dan memahami informasi.', 'SOLO', [
    { slug: 'baca-label', description: 'Ambil satu produk makanan di rumah. Tuliskan tanggal kedaluwarsanya.', type: 'OBSERVATION' },
    { slug: 'simpan-produk', description: 'Bagaimana cara menyimpan produk tersebut dengan benar menurut labelnya?', type: 'WRITE' }
], 80, 'wooden-chest');

createQuest('ww-clean', 'whispering-woods', 'forest-wisdom', 'Membersihkan Jalan', 'Membuat checklist pekerjaan rumah.', 'SOLO', [
    { slug: 'checklist-rumah', description: 'Buat daftar 3 pekerjaan rumah yang harus diselesaikan hari ini.', type: 'WRITE' },
    { slug: 'kerjakan-satu', description: 'Selesaikan salah satu tugas tersebut dan ceritakan perasaanmu.', type: 'WRITE' }
], 150, 'silver-chest');

createQuest('ww-polite', 'whispering-woods', 'forest-wisdom', 'Sopan Santun', 'Memilih respons yang baik dalam percakapan.', 'SOLO', [
    { slug: 'respons-baik', description: 'Seseorang memujimu. Apa balasan yang sopan?', type: 'WRITE' },
    { slug: 'respons-marah', description: 'Jika temanmu membatalkan janji, bagaimana merespons tanpa marah?', type: 'WRITE' }
], 100, 'bronze-chest');

createQuest('ww-photo-memory', 'whispering-woods', 'forest-wisdom', 'Mengabadikan Momen', 'Mengatur foto di HP.', 'SOLO', [
    { slug: 'hapus-foto', description: 'Hapus 5 foto yang buram atau tidak terpakai di HP-mu.', type: 'OBSERVATION' },
    { slug: 'folder-foto', description: 'Buat satu album/folder baru dan beri nama.', type: 'WRITE' }
], 100, 'wooden-chest');

createQuest('ww-motivation', 'whispering-woods', 'forest-wisdom', 'Pesan Motivasi', 'Membuat pesan semangat.', 'CREATIVE', [
    { slug: 'buat-poster', description: 'Buatlah gambar atau tulisan penyemangat untuk dirimu sendiri.', type: 'DRAW' }
], 120, 'bronze-chest');

createQuest('ww-shopping', 'whispering-woods', 'forest-wisdom', 'Rencana Belanja', 'Merencanakan belanja harian.', 'SOLO', [
    { slug: 'daftar-belanja', description: 'Tuliskan 5 barang kebutuhan pokok yang hampir habis di rumah.', type: 'WRITE' },
    { slug: 'estimasi-harga', description: 'Tuliskan perkiraan harga total dari barang-barang tersebut.', type: 'WRITE' }
], 100, 'wooden-chest');

createQuest('ww-instructions', 'whispering-woods', 'forest-wisdom', 'Memahami Instruksi', 'Latihan mengikuti arahan.', 'SOLO', [
    { slug: 'ikuti-arah', description: 'Berdirilah, putar badan sekali, lalu sentuh lututmu. Tulis "Sudah" jika berhasil.', type: 'MOVEMENT' }
], 50, 'wooden-chest');

createQuest('ww-comic', 'whispering-woods', 'forest-wisdom', 'Komik Kehidupan', 'Membuat komik 2-4 panel.', 'CREATIVE', [
    { slug: 'buat-komik', description: 'Gambarlah komik sederhana tentang kejadian lucumu minggu ini.', type: 'DRAW' }
], 150, 'silver-chest');

createQuest('ww-announcement', 'whispering-woods', 'forest-wisdom', 'Pengumuman Penting', 'Memahami isi pengumuman.', 'SOLO', [
    { slug: 'cari-pengumuman', description: 'Cari satu pengumuman atau berita pendek, lalu tuliskan intinya dalam 1 kalimat.', type: 'WRITE' }
], 100, 'bronze-chest');


// Clockwork City (12 quests)
createQuest('cc-budget', 'clockwork-city', 'mechanical-efficiency', 'Anggaran Mesin', 'Membuat anggaran sederhana.', 'SOLO', [
    { slug: 'catat-uang', description: 'Berapa uang yang kamu miliki saat ini? (Tulis secara perkiraan).', type: 'WRITE' },
    { slug: 'rencana-pengeluaran', description: 'Tuliskan maksimal 3 hal yang akan kamu beli minggu ini.', type: 'WRITE' }
], 100, 'bronze-chest');

createQuest('cc-needs', 'clockwork-city', 'mechanical-efficiency', 'Kebutuhan vs Keinginan', 'Membedakan hal yang penting.', 'SOLO', [
    { slug: 'sebut-kebutuhan', description: 'Sebutkan 2 hal yang benar-benar kamu butuhkan bulan ini.', type: 'WRITE' },
    { slug: 'sebut-keinginan', description: 'Sebutkan 1 hal yang hanya sekadar kamu inginkan, tapi tidak mendesak.', type: 'WRITE' }
], 80, 'wooden-chest');

createQuest('cc-discount', 'clockwork-city', 'mechanical-efficiency', 'Menghitung Diskon', 'Matematika praktis diskon.', 'SOLO', [
    { slug: 'hitung-diskon-1', description: 'Sebuah barang seharga Rp 100.000 diskon 20%. Berapa harga akhirnya?', type: 'PUZZLE' },
    { slug: 'hitung-diskon-2', description: 'Mana yang lebih murah: Diskon Rp 20.000 atau Diskon 15% dari Rp 100.000?', type: 'PUZZLE' }
], 120, 'silver-chest');

createQuest('cc-change', 'clockwork-city', 'mechanical-efficiency', 'Uang Kembalian', 'Menghitung kembalian.', 'SOLO', [
    { slug: 'hitung-kembalian', description: 'Kamu belanja Rp 34.500 dan membayar dengan uang Rp 50.000. Berapa kembaliannya?', type: 'PUZZLE' }
], 100, 'wooden-chest');

createQuest('cc-time', 'clockwork-city', 'mechanical-efficiency', 'Waktu Berjalan', 'Mengatur dan menghitung waktu.', 'SOLO', [
    { slug: 'hitung-waktu', description: 'Jika perjalanan memakan waktu 45 menit dan kamu harus tiba jam 10:00, jam berapa kamu harus berangkat?', type: 'PUZZLE' }
], 100, 'bronze-chest');

createQuest('cc-installments', 'clockwork-city', 'mechanical-efficiency', 'Memahami Cicilan', 'Keuangan dasar cicilan.', 'SOLO', [
    { slug: 'cicilan-dasar', description: 'Harga HP Rp 1.200.000. Jika dicicil 12 bulan tanpa bunga, berapa per bulannya?', type: 'PUZZLE' },
    { slug: 'bahaya-bunga', description: 'Kenapa kita harus berhati-hati jika mencicil dengan bunga tinggi?', type: 'WRITE' }
], 120, 'silver-chest');

createQuest('cc-scam', 'clockwork-city', 'mechanical-efficiency', 'Modus Penipuan', 'Mengenali penipuan investasi.', 'SOLO', [
    { slug: 'ciri-penipuan', description: 'Ada yang menawarkan untung 50% dalam sehari tanpa risiko. Apakah ini wajar? Mengapa?', type: 'WRITE' }
], 150, 'golden-chest');

createQuest('cc-expense', 'clockwork-city', 'mechanical-efficiency', 'Mencatat Pengeluaran', 'Membiasakan mencatat uang.', 'SOLO', [
    { slug: 'catat-kemarin', description: 'Tuliskan semua pengeluaranmu kemarin (kira-kira saja jika lupa).', type: 'WRITE' }
], 100, 'bronze-chest');

createQuest('cc-fraction', 'clockwork-city', 'mechanical-efficiency', 'Pecahan Harian', 'Matematika praktis pecahan.', 'SOLO', [
    { slug: 'potong-kue', description: 'Jika 1 kue dipotong menjadi 8 bagian, dan kamu memakan 2 bagian, sisa berapa bagian?', type: 'PUZZLE' }
], 80, 'wooden-chest');

createQuest('cc-distance', 'clockwork-city', 'mechanical-efficiency', 'Menghitung Jarak', 'Matematika jarak dan kecepatan.', 'SOLO', [
    { slug: 'estimasi-jarak', description: 'Kamu berjalan dengan kecepatan biasa selama 15 menit. Kira-kira seberapa jauh jaraknya menurutmu?', type: 'WRITE' }
], 80, 'wooden-chest');

createQuest('cc-ratio', 'clockwork-city', 'mechanical-efficiency', 'Perbandingan', 'Menghitung perbandingan.', 'SOLO', [
    { slug: 'resep-masakan', description: 'Resep membutuhkan 2 gelas beras untuk 3 gelas air. Jika kamu pakai 4 gelas beras, butuh berapa gelas air?', type: 'PUZZLE' }
], 100, 'bronze-chest');

createQuest('cc-deadline', 'clockwork-city', 'mechanical-efficiency', 'Mengejar Deadline', 'Menyelesaikan tugas tepat waktu.', 'SOLO', [
    { slug: 'set-timer', description: 'Pasang timer 5 menit. Rapikan mejamu atau tempat tidurmu. Tulis "Selesai" jika berhasil sebelum timer berbunyi.', type: 'MOVEMENT' }
], 120, 'silver-chest');


// Starlit Library (12 quests)
createQuest('sl-phishing', 'starlit-library', 'modern-knowledge', 'Link Berbahaya', 'Mengenali link phishing.', 'SOLO', [
    { slug: 'cek-link', description: 'Jika kamu mendapat SMS menang hadiah dengan link "http://hadiah-gratis-banget.com", apa yang harus kamu lakukan?', type: 'WRITE' }
], 120, 'silver-chest');

createQuest('sl-password', 'starlit-library', 'modern-knowledge', 'Kunci Pengetahuan', 'Membuat password yang aman.', 'SOLO', [
    { slug: 'buat-pass', description: 'Tuliskan contoh password yang kuat (minimal 8 karakter, ada angka, dan huruf besar). Jangan gunakan password aslimu!', type: 'WRITE' }
], 100, 'bronze-chest');

createQuest('sl-hoax', 'starlit-library', 'modern-knowledge', 'Informasi Palsu', 'Mengenali hoax.', 'SOLO', [
    { slug: 'cek-fakta', description: 'Jika ada berita di grup WA yang belum jelas kebenarannya, apakah boleh langsung disebar?', type: 'WRITE' }
], 100, 'wooden-chest');

createQuest('sl-otp', 'starlit-library', 'modern-knowledge', 'Rahasia OTP', 'Memahami pentingnya OTP.', 'SOLO', [
    { slug: 'jaga-otp', description: 'Seseorang menelepon mengaku dari bank dan meminta kode 6 angka (OTP) di HP-mu. Apa tindakanmu?', type: 'WRITE' }
], 150, 'silver-chest');

createQuest('sl-cv', 'starlit-library', 'modern-knowledge', 'Membuka Gerbang Karier', 'Membuat CV sederhana.', 'SOLO', [
    { slug: 'isi-cv', description: 'Sebutkan 3 informasi paling penting yang wajib ada di dalam CV.', type: 'WRITE' },
    { slug: 'hindari-cv', description: 'Sebutkan 1 hal yang sebaiknya TIDAK dimasukkan ke dalam CV.', type: 'WRITE' }
], 150, 'golden-chest');

createQuest('sl-fake-job', 'starlit-library', 'modern-knowledge', 'Lowongan Palsu', 'Mengenali penipuan kerja.', 'SOLO', [
    { slug: 'ciri-palsu', description: 'Sebuah lowongan kerja memintamu membayar Rp 500.000 untuk biaya seragam sebelum interview. Apakah ini wajar?', type: 'WRITE' }
], 120, 'silver-chest');

createQuest('sl-interview', 'starlit-library', 'modern-knowledge', 'Latihan Wawancara', 'Persiapan interview kerja.', 'SOLO', [
    { slug: 'kenalan', description: 'Tuliskan perkenalan dirimu secara singkat (nama, pendidikan/pengalaman, dan keahlian).', type: 'WRITE' }
], 150, 'silver-chest');

createQuest('sl-outfit', 'starlit-library', 'modern-knowledge', 'Pakaian Profesional', 'Memilih pakaian yang tepat.', 'SOLO', [
    { slug: 'pilih-baju', description: 'Sebutkan pakaian apa yang paling pantas digunakan untuk wawancara kerja di kantoran.', type: 'WRITE' }
], 80, 'wooden-chest');

createQuest('sl-work-ethics', 'starlit-library', 'modern-knowledge', 'Etika Kerja', 'Memahami aturan profesional.', 'SOLO', [
    { slug: 'izin-sakit', description: 'Jika kamu sakit dan tidak bisa bekerja, bagaimana cara memberitahu atasan yang baik?', type: 'WRITE' }
], 100, 'bronze-chest');

createQuest('sl-whatsapp', 'starlit-library', 'modern-knowledge', 'Privasi Komunikasi', 'Menjaga keamanan akun WA.', 'SOLO', [
    { slug: 'wa-web', description: 'Setelah menggunakan WhatsApp Web di komputer umum, apa yang WAJIB kamu lakukan?', type: 'WRITE' }
], 120, 'silver-chest');

createQuest('sl-typing', 'starlit-library', 'modern-knowledge', 'Latihan Mengetik', 'Mengetik dengan cepat dan tepat.', 'SOLO', [
    { slug: 'ketik-cepat', description: 'Ketik ulang kalimat ini dengan benar tanpa salah eja: "Saya berjanji akan terus belajar dan meningkatkan kemampuan diri setiap hari."', type: 'WRITE' }
], 100, 'bronze-chest');

createQuest('sl-note', 'starlit-library', 'modern-knowledge', 'Catatan Penting', 'Membuat catatan dari informasi.', 'SOLO', [
    { slug: 'catat-rapat', description: 'Tuliskan 2 poin penting yang harus dicatat ketika ada pengarahan dari atasan.', type: 'WRITE' }
], 100, 'bronze-chest');


let sql = `
-- Migration 029: Seed Odyssey Learning Content
-- Seeds 36 quests covering Digital Literacy, Finance, Work, Productivity, Math, Creativity, and Daily Life.

INSERT INTO odyssey_chapter_definitions (slug, realm, title, description, "order", published, version)
VALUES
  ('forest-wisdom', 'whispering-woods', 'Kebijaksanaan Hutan', 'Pelajari kebijaksanaan dari alam untuk kehidupan sehari-harimu.', 3, true, 1),
  ('mechanical-efficiency', 'clockwork-city', 'Efisiensi Mekanis', 'Pahami angka, waktu, dan uang untuk menggerakkan mesin kehidupanmu.', 2, true, 1),
  ('modern-knowledge', 'starlit-library', 'Pengetahuan Modern', 'Buku-buku ini menyimpan rahasia dunia kerja dan digital.', 2, true, 1)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO odyssey_quest_definitions (slug, realm, chapter, title, description, quest_type, challenge_defs, reward_xp, reward_chest, is_mandatory, required_level, published, version)
VALUES
`;

quests.forEach((q, i) => {
    const isLast = i === quests.length - 1;
    sql += `  ('${q.slug}', '${q.realm}', '${q.chapter}', '${q.title}', '${q.description}', '${q.type}', '${JSON.stringify(q.challenges)}', ${q.xp}, '${q.chest}', true, 0, true, 1)${isLast ? '' : ','}\n`;
});

sql += `ON CONFLICT (slug) DO NOTHING;
`;

fs.writeFileSync('supabase/migrations/029_seed_odyssey_content.sql', sql);
console.log('Migration generated.');
