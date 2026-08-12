-- Migration 031: Seed Phase 1 Quests
-- Adds learn_text, result_text, and converts challenges to MCQ for 12 important quests.

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Jadwal yang baik membantu kita mengatur waktu. Dengan jadwal, kita tahu apa yang harus diprioritaskan hari ini.',
  result_text = 'Luar biasa! Dengan jadwal yang jelas, kamu bisa lebih fokus dan mengurangi stres.',
  challenge_defs = '[{"slug":"buat-jadwal","description":"Mana dari pilihan berikut yang merupakan manfaat utama membuat jadwal?","type":"MCQ","question":"Apa manfaat utama membuat jadwal harian?","options":["Membuat kita lebih sibuk","Membantu mengatur waktu dan prioritas","Mengurangi waktu istirahat","Membuat tugas terasa lebih berat"],"correct_answer":"Membantu mengatur waktu dan prioritas"}]'
WHERE slug = 'ww-schedule';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Pesan yang sopan dan jelas mencegah salah paham. Saat meminta maaf, sebutkan alasannya dengan jujur.',
  result_text = 'Bagus! Komunikasi yang baik adalah kunci hubungan yang sehat.',
  challenge_defs = '[{"slug":"pesan-sopan","description":"Pilih kalimat yang paling sopan jika kamu terlambat.","type":"MCQ","question":"Bagaimana cara meminta maaf yang sopan karena terlambat?","options":["Maaf telat.","Woy, gw telat dikit ntar.","Maaf saya akan terlambat 15 menit karena jalanan macet. Terima kasih sudah menunggu.","Tunggu aja, bentar lagi sampe."],"correct_answer":"Maaf saya akan terlambat 15 menit karena jalanan macet. Terima kasih sudah menunggu."}]'
WHERE slug = 'ww-message';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Setiap produk makanan kemasan memiliki tanggal kedaluwarsa (Expired Date) untuk memastikan keamanan saat dikonsumsi.',
  result_text = 'Tepat sekali! Selalu periksa label sebelum mengonsumsi atau membeli makanan.',
  challenge_defs = '[{"slug":"baca-label","description":"Apa arti dari Expired Date (Tanggal Kedaluwarsa) pada kemasan makanan?","type":"MCQ","question":"Apa arti dari Expired Date (Tanggal Kedaluwarsa) pada kemasan makanan?","options":["Tanggal makanan tersebut dibuat","Batas waktu maksimal makanan aman untuk dikonsumsi","Tanggal makanan didiskon","Tanggal makanan dikirim ke toko"],"correct_answer":"Batas waktu maksimal makanan aman untuk dikonsumsi"}]'
WHERE slug = 'ww-reading';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Sopan santun bukan hanya tentang kata-kata, tapi juga menghargai orang lain dan menjaga perasaannya.',
  result_text = 'Kerja bagus! Kesopanan adalah cerminan diri yang baik.',
  challenge_defs = '[{"slug":"respons-baik","description":"Jika ada yang memuji karyamu, respons apa yang terbaik?","type":"MCQ","question":"Jika seseorang memuji karyamu, bagaimana sebaiknya merespons?","options":["Diam saja","Bilang ''Ah biasa saja''","Mengucapkan ''Terima kasih banyak!''","Menyombongkan diri"],"correct_answer":"Mengucapkan ''Terima kasih banyak!''"}]'
WHERE slug = 'ww-polite';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Mencatat pengeluaran harian membantu kita melacak ke mana perginya uang kita dan mencegah pemborosan.',
  result_text = 'Mantap! Mencatat keuangan adalah langkah pertama menjadi cerdas finansial.',
  challenge_defs = '[{"slug":"catat-uang","description":"Mengapa penting untuk mencatat pengeluaran?","type":"MCQ","question":"Mengapa penting untuk mencatat pengeluaran?","options":["Agar uang cepat habis","Untuk mengetahui ke mana uang kita dibelanjakan","Agar bisa pamer","Karena disuruh bank"],"correct_answer":"Untuk mengetahui ke mana uang kita dibelanjakan"}]'
WHERE slug = 'cc-budget';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Kebutuhan adalah sesuatu yang wajib dipenuhi untuk hidup, sedangkan keinginan adalah sesuatu yang kita mau tapi bisa ditunda.',
  result_text = 'Hebat! Mendahulukan kebutuhan daripada keinginan akan menyelamatkan keuanganmu.',
  challenge_defs = '[{"slug":"sebut-kebutuhan","description":"Manakah di bawah ini yang merupakan contoh KEBUTUHAN?","type":"MCQ","question":"Manakah di bawah ini yang merupakan contoh KEBUTUHAN?","options":["Makan bergizi setiap hari","Membeli skin game terbaru","Menonton bioskop setiap minggu","Membeli sepatu mahal yang sedang tren"],"correct_answer":"Makan bergizi setiap hari"}]'
WHERE slug = 'cc-needs';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Diskon adalah potongan harga yang diberikan penjual. Diskon 20% artinya kita membayar 80% dari harga asli.',
  result_text = 'Kamu jenius! Kemampuan menghitung diskon membantumu berbelanja lebih hemat.',
  challenge_defs = '[{"slug":"hitung-diskon-1","description":"Sebuah tas seharga Rp 100.000 diskon 20%. Berapa harganya sekarang?","type":"MCQ","question":"Sebuah tas seharga Rp 100.000 diskon 20%. Berapa harganya sekarang?","options":["Rp 20.000","Rp 80.000","Rp 100.000","Rp 120.000"],"correct_answer":"Rp 80.000"}]'
WHERE slug = 'cc-discount';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Menghitung waktu perjalanan penting agar kita tidak terlambat saat janjian atau pergi ke sekolah/kantor.',
  result_text = 'Tepat! Manajemen waktu yang baik menunjukkan bahwa kamu menghargai waktu orang lain.',
  challenge_defs = '[{"slug":"hitung-waktu","description":"Jika perjalanan butuh 30 menit dan kamu harus tiba jam 08:00, jam berapa harus berangkat?","type":"MCQ","question":"Jika perjalanan butuh 30 menit dan kamu harus tiba jam 08:00, jam berapa harus berangkat?","options":["07:00","07:30","08:00","08:30"],"correct_answer":"07:30"}]'
WHERE slug = 'cc-time';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Phishing adalah penipuan untuk mencuri password atau data pribadi melalui link palsu.',
  result_text = 'Bagus sekali! Jangan pernah klik sembarang link dari sumber yang tidak kamu kenal.',
  challenge_defs = '[{"slug":"cek-link","description":"Jika ada SMS hadiah berisi link mencurigakan, apa yang harus dilakukan?","type":"MCQ","question":"Jika ada SMS hadiah berisi link mencurigakan, apa yang harus dilakukan?","options":["Klik link untuk melihat hadiah","Abaikan dan hapus pesannya","Kirim balik pesan marah-marah","Sebarkan ke teman-teman"],"correct_answer":"Abaikan dan hapus pesannya"}]'
WHERE slug = 'sl-phishing';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Password yang kuat terdiri dari kombinasi huruf besar, huruf kecil, dan angka, serta tidak mudah ditebak (minimal 8 karakter).',
  result_text = 'Luar biasa! Keamanan akun sangat penting di era digital.',
  challenge_defs = '[{"slug":"buat-pass","description":"Manakah contoh password yang paling kuat?","type":"MCQ","question":"Manakah contoh password yang paling kuat?","options":["12345678","password","Budi123!","BukuMerahMilikSaya88"],"correct_answer":"BukuMerahMilikSaya88"}]'
WHERE slug = 'sl-password';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'OTP (One Time Password) adalah kode rahasia 6 angka yang dikirim lewat SMS/WA untuk verifikasi. Jangan pernah berikan pada siapa pun.',
  result_text = 'Hebat! Menyimpan rahasia OTP adalah kunci mencegah akunmu diretas.',
  challenge_defs = '[{"slug":"jaga-otp","description":"Siapa yang boleh mengetahui kode OTP milikmu?","type":"MCQ","question":"Siapa yang boleh mengetahui kode OTP milikmu?","options":["Pegawai Bank","Hanya diriku sendiri","Keluarga dekat","Teman baik"],"correct_answer":"Hanya diriku sendiri"}]'
WHERE slug = 'sl-otp';

UPDATE odyssey_quest_definitions
SET 
  learn_text = 'Hoax adalah informasi atau berita bohong. Selalu cek kebenarannya sebelum membagikannya ke orang lain.',
  result_text = 'Betul sekali! Saring sebelum sharing!',
  challenge_defs = '[{"slug":"cek-fakta","description":"Apakah boleh langsung membagikan pesan berantai di grup WA yang belum jelas kebenarannya?","type":"TRUE_FALSE","question":"Apakah boleh langsung membagikan pesan berantai di grup WA yang belum jelas kebenarannya?","options":["BENAR","SALAH"],"correct_answer":"SALAH"}]'
WHERE slug = 'sl-hoax';
