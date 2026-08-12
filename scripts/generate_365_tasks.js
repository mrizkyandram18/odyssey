const fs = require('fs');

const templates = [
  { title: 'Keluarga: Waktu Berkualitas', q: 'Mana yang merupakan contoh waktu berkualitas bersama keluarga?', type: 'MCQ', opts: ['Makan malam bersama tanpa HP', 'Menonton TV masing-masing', 'Bermain game sendiri di kamar', 'Pergi ke mall tapi pisah'], ans: 'Makan malam bersama tanpa HP', exp: 'Waktu berkualitas membutuhkan interaksi penuh tanpa gangguan.' },
  { title: 'Finansial: Tabungan', q: 'Mengapa penting menyisihkan uang untuk ditabung di awal bulan?', type: 'MCQ', opts: ['Agar uang cepat habis', 'Agar tidak terpakai untuk keinginan sesaat', 'Untuk pamer ke teman', 'Tidak ada gunanya'], ans: 'Agar tidak terpakai untuk keinginan sesaat', exp: 'Menabung di awal mencegah uang habis untuk kebutuhan tidak penting.' },
  { title: 'Digital: Password', q: 'Password yang kuat sebaiknya mengandung huruf besar, huruf kecil, angka, dan simbol.', type: 'TRUE_FALSE', opts: ['Benar', 'Salah'], ans: 'Benar', exp: 'Kombinasi tersebut membuat password jauh lebih sulit ditebak.' },
  { title: 'Karier: Etika Email', q: 'Saat mengirim email profesional, kolom Subject/Judul boleh dikosongkan.', type: 'TRUE_FALSE', opts: ['Benar', 'Salah'], ans: 'Salah', exp: 'Subject sangat penting agar penerima tahu isi email sebelum membukanya.' },
  { title: 'Keluarga: Komunikasi', q: 'Apa yang sebaiknya dilakukan saat anggota keluarga sedang berbicara serius?', type: 'MCQ', opts: ['Menyela pembicaraan', 'Mendengarkan dengan penuh perhatian', 'Memeriksa handphone', 'Pergi begitu saja'], ans: 'Mendengarkan dengan penuh perhatian', exp: 'Mendengarkan aktif adalah bentuk penghargaan dan kasih sayang.' },
  { title: 'Finansial: Dana Darurat', q: 'Dana darurat sebaiknya digunakan untuk membeli HP seri terbaru.', type: 'TRUE_FALSE', opts: ['Benar', 'Salah'], ans: 'Salah', exp: 'Dana darurat hanya untuk kejadian tak terduga seperti sakit atau kehilangan pekerjaan.' },
  { title: 'Digital: Hoax', q: 'Semua berita yang dikirim melalui grup WhatsApp pasti benar.', type: 'TRUE_FALSE', opts: ['Benar', 'Salah'], ans: 'Salah', exp: 'Selalu verifikasi informasi dari sumber terpercaya sebelum mempercayainya.' },
  { title: 'Karier: Wawancara', q: 'Pakaian apa yang paling tepat untuk wawancara kerja kantoran?', type: 'MCQ', opts: ['Kaos oblong dan celana pendek', 'Kemeja rapi dan celana bahan', 'Pakaian tidur', 'Pakaian olahraga'], ans: 'Kemeja rapi dan celana bahan', exp: 'Berpakaian rapi menunjukkan profesionalisme dan rasa hormat.' },
  { title: 'Produktivitas: Fokus', q: 'Teknik Pomodoro adalah cara mengatur waktu dengan bekerja 25 menit lalu istirahat 5 menit.', type: 'TRUE_FALSE', opts: ['Benar', 'Salah'], ans: 'Benar', exp: 'Teknik ini terbukti ampuh menjaga fokus dan mencegah kelelahan mental.' },
  { title: 'Finansial: Cicilan', q: 'Total cicilan utang per bulan sebaiknya tidak melebihi berapa persen dari gaji?', type: 'MCQ', opts: ['30%', '50%', '80%', '100%'], ans: '30%', exp: 'Pakar keuangan menyarankan rasio utang tidak melebihi 30% dari pemasukan.' }
];

let sql = `-- Migration 039: Seed 365 Daily Activities
-- Generates a full year of daily tasks for Family, Financial, Digital, and Career learning.

INSERT INTO odyssey_daily_activities (slug, title, question, type, options, correct_answer, explanation, xp_reward, active) VALUES
`;

const rows = [];
for (let i = 1; i <= 365; i++) {
  const t = templates[(i - 1) % templates.length];
  const slug = `task-day-${i}`;
  const title = t.title;
  // Vary the question slightly by adding Day X prefix for clarity if we want, but keeping it as is is fine.
  const question = t.q;
  const options = JSON.stringify(t.opts).replace(/'/g, "''"); // escape SQL single quotes inside JSON
  
  rows.push(`('${slug}', '${title}', '${question}', '${t.type}', '${options}', '${t.ans}', '${t.exp}', 10, true)`);
}

sql += rows.join(',\n') + '\nON CONFLICT (slug) DO NOTHING;\n';

fs.writeFileSync('supabase/migrations/039_seed_365_daily_tasks.sql', sql);
console.log('Generated 039_seed_365_daily_tasks.sql with 365 tasks.');
