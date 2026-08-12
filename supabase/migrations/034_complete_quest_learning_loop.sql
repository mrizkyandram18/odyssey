-- Migration 034: Complete Mission Learning Loop (Slice 4.6 — Beta Readiness)
-- Adds learn_text, result_text, and properly-classified challenge_defs
-- to the 24 missions not covered by migration 031.
--
-- Classification rules applied (source: Slice 4.6 PRD):
--   MCQ / TRUE_FALSE  → factual knowledge with one objectively correct answer
--   WRITE             → reflective, personal, creative — auto-marked complete on submit
--   OBSERVATION       → real-world action — auto-marked complete on submit
--   DRAW              → creative drawing — auto-marked complete on submit
--   MOVEMENT          → physical action — auto-marked complete on submit
--
-- Backend grading invariant (pkg/game/mission/service.go L491-494):
--   MCQ/TRUE_FALSE: wrong answer → ErrIncorrectAnswer (not complete, no XP)
--   All other types: submitted → complete (no answer check, +XP awarded)
--
-- This migration uses only UPDATE statements.
-- No schema changes — learn_text / result_text added in migration 030.
-- Does not touch the 12 missions already updated in migration 031.

-- ============================================================
-- WHISPERING WOODS — 8 missions remaining
-- ============================================================

-- ww-creative-story: CREATIVE quest → WRITE (creative writing, no single correct answer)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Menulis cerita melatih imajinasi dan kemampuan mengungkapkan ide. Cerita pendek yang baik punya tokoh, situasi, dan penyelesaian — bahkan hanya dalam 2–3 kalimat.',
  result_text = 'Luar biasa! Imajinasi adalah otot — semakin sering dilatih, semakin kuat. Ceritamu tadi adalah karya nyata milikmu sendiri.',
  challenge_defs = '[{"slug":"tulis-cerita","description":"Buatlah cerita pendek 2–3 kalimat tentang hewan yang menemukan harta karun. Tidak ada jawaban salah — tuliskan apapun yang muncul di pikiranmu.","type":"WRITE","question":"Tulis ceritamu di sini:"}]'
WHERE slug = 'ww-creative-story';

-- ww-clean: SOLO quest → WRITE (reflective; planning + action report)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Mengerjakan tugas rumah secara terstruktur melatih tanggung jawab dan manajemen waktu. Membuat daftar tugas sebelum mulai bekerja membantu kita lebih fokus dan tidak ada yang terlewat.',
  result_text = 'Bagus sekali! Kemampuan membuat daftar prioritas dan menyelesaikannya adalah skill yang dipakai seumur hidup — dari rumah hingga kantor.',
  challenge_defs = '[{"slug":"checklist-rumah","description":"Buat daftar 3 pekerjaan rumah yang harus diselesaikan hari ini.","type":"WRITE","question":"Tuliskan 3 tugas rumah yang ingin kamu selesaikan hari ini:"},{"slug":"kerjakan-satu","description":"Selesaikan salah satu tugas dari daftarmu dan ceritakan hasilnya.","type":"WRITE","question":"Tugas mana yang sudah kamu selesaikan, dan bagaimana rasanya?"}]'
WHERE slug = 'ww-clean';

-- ww-photo-memory: SOLO quest → OBSERVATION (real-world device action)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Mengorganisir foto di HP melatih kebiasaan rapi dan hemat penyimpanan. HP yang penuh foto buram memperlambat kinerja perangkat dan menyulitkan kita menemukan foto penting di saat dibutuhkan.',
  result_text = 'Kerja bagus! Kebiasaan merapikan file digital mencerminkan kedisiplinan yang berguna di banyak aspek kehidupan — termasuk di lingkungan kerja.',
  challenge_defs = '[{"slug":"hapus-foto","description":"Buka galeri HP-mu sekarang. Hapus minimal 5 foto yang buram, duplikat, atau tidak terpakai.","type":"OBSERVATION","question":"Tulis ''Sudah'' setelah kamu berhasil menghapus minimal 5 foto:"},{"slug":"folder-foto","description":"Buat satu album atau folder baru di galeri HP-mu dan beri nama yang bermakna.","type":"WRITE","question":"Nama album baru yang kamu buat:"}]'
WHERE slug = 'ww-photo-memory';

-- ww-motivation: CREATIVE quest → DRAW (creative expression)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Pesan motivasi yang kita buat sendiri jauh lebih bermakna daripada yang kita baca dari orang lain. Menulis atau menggambar pesan untuk diri sendiri adalah bentuk self-care yang sederhana namun efektif.',
  result_text = 'Hebat! Kamu baru saja menciptakan sesuatu yang bisa kamu lihat lagi saat butuh semangat. Simpan karyamu di tempat yang mudah terlihat.',
  challenge_defs = '[{"slug":"buat-poster","description":"Gambarlah atau tuliskan pesan penyemangat untuk dirimu sendiri. Bisa berupa kata-kata, gambar sederhana, atau kombinasi keduanya.","type":"DRAW","question":"Gambar atau tulis pesan semangatmu di sini:"}]'
WHERE slug = 'ww-motivation';

-- ww-shopping: SOLO quest → WRITE (personal planning, no single correct answer)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Merencanakan belanja sebelum pergi ke toko mencegah pemborosan. Tanpa daftar, kita cenderung membeli barang yang tidak dibutuhkan karena tergiur impuls atau promo. Estimasi harga membantu kita membawa uang yang cukup.',
  result_text = 'Pintar! Dengan daftar belanja dan estimasi harga, kamu sudah berlatih budgeting dasar — skill penting untuk mengelola keuangan keluarga.',
  challenge_defs = '[{"slug":"daftar-belanja","description":"Tuliskan 5 barang kebutuhan pokok yang hampir habis atau perlu dibeli minggu ini.","type":"WRITE","question":"Tulis 5 barang kebutuhanmu:"},{"slug":"estimasi-harga","description":"Perkirakan total biaya dari daftar belanjamu tadi.","type":"WRITE","question":"Perkiraan total harga (boleh estimasi kasar):"}]'
WHERE slug = 'ww-shopping';

-- ww-instructions: SOLO quest → MOVEMENT (physical action instruction-following)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Mengikuti instruksi dengan tepat adalah skill penting di dunia kerja. Orang yang bisa memahami dan menjalankan arahan dengan benar lebih efisien dan jarang membuat kesalahan yang merugikan.',
  result_text = 'Berhasil! Kemampuan mengikuti instruksi terdengar sederhana, tapi ini adalah fondasi disiplin kerja dan kepatuhan prosedur yang sangat dihargai di tempat kerja manapun.',
  challenge_defs = '[{"slug":"ikuti-arah","description":"Ikuti instruksi berikut dengan tepat: Berdirilah dari tempat dudukmu, putar badan satu kali penuh ke kanan, lalu sentuh lutut kirimu.","type":"MOVEMENT","question":"Tulis ''Selesai'' setelah kamu berhasil melakukan gerakan di atas:"}]'
WHERE slug = 'ww-instructions';

-- ww-comic: CREATIVE quest → DRAW (multi-panel creative expression)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Komik adalah cara bercerita dengan gambar dan teks. Setiap panel mewakili satu momen cerita. Membuat komik melatih kreativitas, kemampuan berpikir berurutan, dan ekspresi diri.',
  result_text = 'Bravo! Kamu baru saja membuat komik — sesuatu yang dilakukan oleh seniman profesional di seluruh dunia. Tidak perlu gambar sempurna; yang penting ceritanya tersampaikan.',
  challenge_defs = '[{"slug":"buat-komik","description":"Gambarlah komik sederhana 2–4 panel tentang kejadian lucu, menarik, atau berkesan yang kamu alami minggu ini.","type":"DRAW","question":"Gambar komikmu di sini (boleh sketsa kasar):"}]'
WHERE slug = 'ww-comic';

-- ww-announcement: SOLO quest → WRITE (reading comprehension + summary, personal output)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Memahami isi pengumuman atau berita adalah skill literasi penting. Orang yang bisa membaca dan merangkum informasi dengan tepat mampu mengambil keputusan lebih baik dan tidak mudah salah paham.',
  result_text = 'Mantap! Kemampuan merangkum informasi dalam 1–2 kalimat yang tepat dipakai setiap hari — dari membaca berita, memahami instruksi atasan, hingga membaca perjanjian.',
  challenge_defs = '[{"slug":"cari-pengumuman","description":"Cari satu pengumuman, berita pendek, atau informasi penting (dari HP, papan pengumuman, atau media sosial). Tuliskan inti informasinya dalam 1–2 kalimat.","type":"WRITE","question":"Tulis inti pengumuman atau berita yang kamu temukan:"}]'
WHERE slug = 'ww-announcement';

-- ============================================================
-- CLOCKWORK CITY — 8 missions remaining
-- ============================================================

-- cc-change: PUZZLE → MCQ (math, one correct answer: Rp 50.000 - Rp 34.500 = Rp 15.500)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Menghitung uang kembalian adalah matematika dasar yang dipakai setiap hari. Rumusnya sederhana: Kembalian = Uang yang dibayar − Total belanja. Biasakan menghitung sebelum menerima kembalian untuk menghindari kesalahan.',
  result_text = 'Tepat! Kemampuan menghitung kembalian dengan cepat melindungimu dari kesalahan kasir. Biasakan selalu hitung dulu sebelum memasukkan kembalian ke dompet.',
  challenge_defs = '[{"slug":"hitung-kembalian","description":"Kamu belanja Rp 34.500 dan membayar dengan uang Rp 50.000.","type":"MCQ","question":"Berapa kembalian yang seharusnya kamu terima?","options":["Rp 14.500","Rp 15.500","Rp 16.500","Rp 25.500"],"correct_answer":"Rp 15.500","explanation":"Kembalian = Rp 50.000 − Rp 34.500 = Rp 15.500."}]'
WHERE slug = 'cc-change';

-- cc-installments: SOLO → MCQ (math challenge) + WRITE (reflection on interest)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Cicilan memungkinkan kita membeli barang mahal secara bertahap. Namun cicilan dengan bunga tinggi bisa membuat total pembayaran jauh melebihi harga asli. Selalu hitung total yang dibayarkan sebelum menyetujui cicilan apapun.',
  result_text = 'Bagus! Memahami cicilan sebelum menandatangani perjanjian adalah tanda kecerdasan finansial. Banyak orang terjebak utang karena tidak menghitung total pembayaran di awal.',
  challenge_defs = '[{"slug":"cicilan-dasar","description":"Harga HP Rp 1.200.000. Jika dicicil 12 bulan tanpa bunga, berapa yang harus dibayar per bulan?","type":"MCQ","question":"Berapa cicilan per bulan untuk HP seharga Rp 1.200.000 dalam 12 bulan tanpa bunga?","options":["Rp 80.000","Rp 100.000","Rp 120.000","Rp 150.000"],"correct_answer":"Rp 100.000","explanation":"Rp 1.200.000 dibagi 12 bulan = Rp 100.000 per bulan."},{"slug":"bahaya-bunga","description":"Jelaskan dengan kata-katamu sendiri: mengapa cicilan dengan bunga tinggi bisa berbahaya?","type":"WRITE","question":"Mengapa kita harus berhati-hati dengan cicilan berbunga tinggi?"}]'
WHERE slug = 'cc-installments';

-- cc-scam: SOLO → MCQ (factual decision, one correct answer)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Penipuan berkedok investasi biasanya menjanjikan keuntungan besar dalam waktu singkat tanpa risiko. Ini disebut skema Ponzi atau money game. Investasi yang sah selalu memiliki risiko dan tidak menjanjikan keuntungan pasti.',
  result_text = 'Benar! Tawaran keuntungan tidak masuk akal adalah tanda penipuan. Prinsip sederhana: jika terdengar terlalu bagus untuk menjadi kenyataan, hampir pasti itu penipuan.',
  challenge_defs = '[{"slug":"ciri-penipuan","description":"Ada yang menawarkan investasi dengan keuntungan 50% dalam sehari tanpa risiko sama sekali.","type":"MCQ","question":"Apakah tawaran investasi ''untung 50% per hari tanpa risiko'' itu wajar?","options":["Ya, asal dari teman terpercaya","Tidak, ini tanda penipuan","Ya, kalau ada bukti transferan","Tidak tahu, perlu dicoba dulu"],"correct_answer":"Tidak, ini tanda penipuan","explanation":"Tidak ada investasi legal yang menjanjikan keuntungan 50% per hari tanpa risiko. Ini adalah ciri khas skema Ponzi atau penipuan investasi."}]'
WHERE slug = 'cc-scam';

-- cc-expense: SOLO → WRITE (reflective; personal expense recall, no single correct answer)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Mencatat pengeluaran harian adalah langkah pertama menuju kebebasan finansial. Dengan mencatat, kita bisa melihat pola pengeluaran: ke mana uang pergi, apa yang bisa dikurangi, dan apa yang perlu diprioritaskan.',
  result_text = 'Luar biasa! Kamu baru saja melakukan sesuatu yang kebanyakan orang tidak lakukan: jujur dengan pengeluaran sendiri. Kebiasaan ini, jika dilakukan setiap hari, bisa mengubah kondisi finansialmu dalam 3 bulan.',
  challenge_defs = '[{"slug":"catat-kemarin","description":"Ingat-ingat semua yang kamu keluarkan kemarin. Tuliskan setiap pengeluaran semampu kamu ingat — estimasi boleh.","type":"WRITE","question":"Tuliskan pengeluaranmu kemarin:"}]'
WHERE slug = 'cc-expense';

-- cc-fraction: PUZZLE → MCQ (math, one correct answer: 8 - 2 = 6)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Pecahan dipakai setiap hari — membagi makanan, mengukur resep, menghitung sisa. Pecahan 2/8 artinya 2 dari 8 bagian. Sisa = total dikurangi yang diambil.',
  result_text = 'Tepat! Matematika pecahan tidak harus rumit. Dengan pemahaman dasar, kamu bisa menyesuaikan resep masakan, menghitung sisa bahan, atau membagi biaya patungan dengan tepat.',
  challenge_defs = '[{"slug":"potong-kue","description":"Sebuah kue dipotong menjadi 8 bagian yang sama. Kamu memakan 2 bagian.","type":"MCQ","question":"Berapa bagian kue yang tersisa?","options":["2 bagian","4 bagian","6 bagian","8 bagian"],"correct_answer":"6 bagian","explanation":"8 bagian dikurangi 2 bagian yang dimakan = 6 bagian tersisa."}]'
WHERE slug = 'cc-fraction';

-- cc-distance: SOLO → WRITE (estimation/reflection, personal answer — no single correct value)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Memperkirakan jarak dan waktu tempuh adalah skill navigasi dasar. Kecepatan berjalan normal manusia sekitar 4–5 km per jam. Dalam 15 menit, kita biasanya menempuh sekitar 1–1,25 km. Kemampuan estimasi ini berguna saat merencanakan perjalanan tanpa GPS.',
  result_text = 'Bagus! Kemampuan memperkirakan jarak tanpa alat bantu adalah skill yang berguna saat GPS tidak tersedia atau saat merencanakan aktivitas outdoor. Intuisi jarak bisa dilatih.',
  challenge_defs = '[{"slug":"estimasi-jarak","description":"Kamu berjalan dengan kecepatan normal selama 15 menit. Perkirakan seberapa jauh jarak yang sudah kamu tempuh.","type":"WRITE","question":"Menurutmu, berapa kira-kira jarak yang ditempuh dalam 15 menit berjalan kaki? Tuliskan perkiraanmu dan alasannya:"}]'
WHERE slug = 'cc-distance';

-- cc-ratio: PUZZLE → MCQ (math, one correct answer: 4 gelas beras x (3/2) = 6 gelas air)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Perbandingan (rasio) dipakai dalam resep masakan, campuran cat, dan skala peta. Jika rasio beras:air = 2:3, artinya untuk 2 bagian beras dibutuhkan 3 bagian air. Jika beras dilipatduakan menjadi 4, maka air juga dilipatduakan menjadi 6.',
  result_text = 'Benar! Memahami rasio membuat kamu bisa menyesuaikan resep masakan, membuat campuran yang tepat, dan menghitung skala dengan mudah.',
  challenge_defs = '[{"slug":"resep-masakan","description":"Resep nasi membutuhkan 2 gelas beras untuk 3 gelas air. Kamu ingin memasak dengan 4 gelas beras.","type":"MCQ","question":"Berapa gelas air yang dibutuhkan untuk 4 gelas beras (dengan rasio beras:air = 2:3)?","options":["4 gelas","5 gelas","6 gelas","8 gelas"],"correct_answer":"6 gelas","explanation":"Rasio beras:air = 2:3. Beras naik dari 2 ke 4 (×2), jadi air juga ×2: 3 × 2 = 6 gelas."}]'
WHERE slug = 'cc-ratio';

-- cc-deadline: SOLO → MOVEMENT (physical timed action)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Teknik bekerja dengan batas waktu (time-boxing) meningkatkan fokus. Pasang timer dan kerjakan satu tugas hingga timer berbunyi — tidak perlu sempurna, yang penting mulai. Merapikan area kerja sebelum mulai terbukti meningkatkan konsentrasi.',
  result_text = 'Selesai! Kamu baru membuktikan bahwa pekerjaan yang terasa berat bisa diselesaikan dalam waktu singkat jika kita mulai. Rasa puas setelah merapikan ruangan juga terbukti meningkatkan mood dan produktivitas.',
  challenge_defs = '[{"slug":"set-timer","description":"Pasang timer 5 menit di HP-mu sekarang. Gunakan waktu itu untuk merapikan meja, tempat tidur, atau sudut ruanganmu. Berhenti ketika timer berbunyi.","type":"MOVEMENT","question":"Tulis ''Selesai'' setelah timermu berbunyi dan kamu sudah merapikan satu area:"}]'
WHERE slug = 'cc-deadline';

-- ============================================================
-- STARLIT LIBRARY — 8 missions remaining
-- ============================================================

-- sl-cv: SOLO → WRITE (personal CV knowledge, reflective output — no single correct answer)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'CV (Curriculum Vitae) adalah dokumen yang memperkenalkan dirimu kepada calon pemberi kerja. CV yang baik singkat (1–2 halaman), bersih, dan memuat: data diri, pendidikan terakhir, pengalaman kerja (jika ada), dan keahlian. Hindari informasi tidak relevan dan foto tidak profesional.',
  result_text = 'Kerja bagus! Memahami struktur CV adalah langkah pertama melamar kerja dengan percaya diri. CV yang baik membuka pintu wawancara — dari sanalah kesempatan karier dimulai.',
  challenge_defs = '[{"slug":"isi-cv","description":"Sebutkan 3 informasi paling penting yang wajib ada di dalam CV profesional.","type":"WRITE","question":"Tuliskan 3 informasi yang wajib ada dalam CV:"},{"slug":"hindari-cv","description":"Sebutkan 1 hal yang sebaiknya TIDAK dimasukkan ke dalam CV karena bisa mengurangi kesan profesional.","type":"WRITE","question":"Tuliskan 1 hal yang tidak perlu ada di CV:"}]'
WHERE slug = 'sl-cv';

-- sl-fake-job: SOLO → MCQ (factual safety decision, one correct answer)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Lowongan kerja palsu (job scam) sering meminta calon pelamar membayar biaya administrasi, seragam, atau pelatihan sebelum diterima. Perusahaan yang sah TIDAK PERNAH meminta bayaran apapun dari calon karyawan sebelum mereka mulai bekerja.',
  result_text = 'Benar! Tidak ada perusahaan legit yang meminta calon karyawan membayar biaya apapun sebelum diterima. Jika diminta membayar di awal, itu hampir pasti penipuan.',
  challenge_defs = '[{"slug":"ciri-palsu","description":"Sebuah lowongan kerja menawarkan gaji besar dan memintamu membayar Rp 500.000 untuk biaya seragam sebelum interview dilakukan.","type":"MCQ","question":"Apakah lowongan kerja yang meminta pembayaran sebelum interview itu wajar?","options":["Ya, wajar karena seragam mahal","Tidak, perusahaan sah tidak pernah minta bayaran di awal","Ya, asal ada tanda terimanya","Tidak tahu, perlu dikonfirmasi dulu"],"correct_answer":"Tidak, perusahaan sah tidak pernah minta bayaran di awal","explanation":"Perusahaan yang sah tidak pernah meminta pembayaran apapun dari calon karyawan. Permintaan uang di awal proses rekrutmen adalah tanda penipuan."}]'
WHERE slug = 'sl-fake-job';

-- sl-interview: SOLO → WRITE (personal self-introduction, no single correct answer)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Perkenalan diri saat wawancara adalah kesempatan pertama untuk membuat kesan baik. Perkenalan yang efektif mencakup: nama, latar belakang pendidikan atau pengalaman singkat, keahlian utama, dan motivasi melamar. Latih sebelum wawancara agar terdengar natural dan percaya diri.',
  result_text = 'Bagus! Kamu sudah punya draf perkenalan diri. Berlatihlah mengucapkannya dengan suara keras — kepercayaan diri saat wawancara datang dari persiapan yang matang, bukan dari bakat.',
  challenge_defs = '[{"slug":"kenalan","description":"Tuliskan perkenalan dirimu untuk wawancara kerja. Sertakan: nama, pendidikan terakhir atau pengalaman singkat, satu keahlian utamamu, dan alasan kamu melamar.","type":"WRITE","question":"Tulis perkenalan dirimu untuk wawancara:"}]'
WHERE slug = 'sl-interview';

-- sl-outfit: SOLO → MCQ (factual professional appearance decision, one correct answer)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Penampilan saat wawancara kerja mencerminkan keseriusan dan rasa hormat kepada perusahaan. Untuk wawancara di kantoran formal: kemeja atau blus polos, celana bahan atau rok, sepatu tertutup. Hindari pakaian kasual, terlalu mencolok, atau berbau.',
  result_text = 'Tepat! Penampilan yang tepat meningkatkan kepercayaan diri dan memberi kesan positif sebelum kamu mengucapkan sepatah kata pun. Ini bukan soal mahal atau murahnya pakaian, tapi kerapian dan kesesuaiannya.',
  challenge_defs = '[{"slug":"pilih-baju","description":"Kamu akan wawancara kerja di perusahaan kantoran formal besok pagi.","type":"MCQ","question":"Pakaian mana yang paling tepat untuk wawancara di kantor formal?","options":["Kaos oblong dan sandal jepit","Kemeja rapi dan celana bahan atau rok","Baju olahraga yang bersih","Jaket kulit dan jeans sobek"],"correct_answer":"Kemeja rapi dan celana bahan atau rok","explanation":"Pakaian formal dan rapi (kemeja + celana bahan atau rok) adalah standar wawancara di kantor formal. Kesan pertama sangat menentukan peluangmu."}]'
WHERE slug = 'sl-outfit';

-- sl-work-ethics: SOLO → WRITE (professional scenario — personal response, no single correct phrasing)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Etika kerja profesional termasuk: hadir tepat waktu, memberi tahu atasan jika tidak bisa hadir, berkomunikasi dengan sopan, dan menyelesaikan tugas sesuai tenggat. Saat sakit dan tidak bisa masuk kerja, etikanya: kabarkan SESEGERA MUNGKIN, jelaskan kondisi, dan tawarkan solusi jika memungkinkan.',
  result_text = 'Benar sekali! Memberi tahu atasan dengan cara yang tepat saat tidak bisa hadir menunjukkan tanggung jawab dan profesionalisme — dua nilai yang sangat dihargai di lingkungan kerja manapun.',
  challenge_defs = '[{"slug":"izin-sakit","description":"Kamu sakit dan tidak bisa masuk kerja hari ini. Atasan belum tahu. Bagaimana cara memberitahunya yang paling profesional?","type":"WRITE","question":"Tulis pesan atau langkah yang akan kamu lakukan untuk memberi tahu atasan bahwa kamu sakit:"}]'
WHERE slug = 'sl-work-ethics';

-- sl-whatsapp: SOLO → MCQ (digital safety, one correct answer)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'WhatsApp Web memungkinkan kamu mengakses WhatsApp di komputer. Tapi jika digunakan di komputer umum dan tidak logout, orang lain bisa membaca semua pesanmu. Menutup tab saja tidak cukup — sesi tetap aktif. Selalu logout setelah selesai menggunakan komputer bersama.',
  result_text = 'Benar! Lupa logout dari akun media sosial di komputer umum adalah salah satu penyebab terbesar peretasan akun. Biasakan selalu logout dan jangan simpan password di browser komputer orang lain.',
  challenge_defs = '[{"slug":"wa-web","description":"Kamu baru saja menggunakan WhatsApp Web di komputer umum untuk mengirim pesan penting. Sekarang kamu ingin meninggalkan komputer itu.","type":"MCQ","question":"Apa yang WAJIB kamu lakukan sebelum meninggalkan komputer umum setelah menggunakan WhatsApp Web?","options":["Minimize jendela browser","Logout dari WhatsApp Web","Tutup tab WhatsApp saja","Tidak perlu melakukan apa-apa"],"correct_answer":"Logout dari WhatsApp Web","explanation":"Menutup tab saja tidak cukup — sesi WhatsApp Web tetap aktif dan bisa diakses pengguna berikutnya. Wajib logout agar akun aman."}]'
WHERE slug = 'sl-whatsapp';

-- sl-typing: SOLO → WRITE (practical typing exercise, personal output)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Kemampuan mengetik dengan cepat dan tepat adalah skill produktivitas penting. Kecepatan mengetik rata-rata orang dewasa sekitar 40 kata per menit. Latihan rutin 15 menit sehari bisa meningkatkan kecepatan secara signifikan dan mengurangi kesalahan ketik.',
  result_text = 'Selesai! Latihan mengetik terdengar sederhana, tapi ini investasi yang menghasilkan. Setiap email, laporan, atau pesan yang harus ditulis akan terasa jauh lebih mudah seiring kemampuan mengetikmu meningkat.',
  challenge_defs = '[{"slug":"ketik-cepat","description":"Ketik ulang kalimat berikut dengan tepat. Perhatikan ejaan, tanda baca, dan huruf kapital.","type":"WRITE","question":"Ketik ulang kalimat ini dengan benar: ''Saya berjanji akan terus belajar dan meningkatkan kemampuan diri setiap hari.''"}]'
WHERE slug = 'sl-typing';

-- sl-note: SOLO → WRITE (professional skill reflection, personal output)
UPDATE odyssey_quest_definitions SET
  learn_text  = 'Mencatat poin penting saat rapat atau briefing adalah skill profesional yang sering diabaikan. Catatan yang baik tidak harus mencatat semua yang diucapkan — cukup: keputusan yang diambil, tugas yang perlu diselesaikan, tenggat waktu, dan hal yang membingungkan untuk ditanyakan.',
  result_text = 'Bagus! Orang yang terbiasa mencatat adalah orang yang bisa dipercaya karena tidak lupa arahan dan bisa membuktikan apa yang sudah disepakati. Di dunia kerja, catatan rapat bisa jadi bukti penting.',
  challenge_defs = '[{"slug":"catat-rapat","description":"Bayangkan kamu baru selesai menghadiri pengarahan dari atasanmu. Tuliskan 2 hal yang paling penting untuk selalu dicatat dari setiap pengarahan.","type":"WRITE","question":"Tuliskan 2 hal yang selalu harus dicatat setiap kali ada pengarahan dari atasan:"}]'
WHERE slug = 'sl-note';

-- ============================================================
-- Schema version update
-- ============================================================
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '034_complete_quest_learning_loop')
ON CONFLICT (key) DO UPDATE SET
  value      = EXCLUDED.value,
  updated_at = timezone('utc'::text, now());
