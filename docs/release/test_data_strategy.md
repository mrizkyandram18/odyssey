# Test Data Strategy

Dokumen ini mendefinisikan strategi pengelolaan data untuk _Runtime Validation Automation_ (E2E Test) guna memastikan bahwa setiap tes bersifat deterministik, dapat diulang (idempotent), dan tidak bergantung pada *state* dari tes sebelumnya.

## Prinsip Utama
1. **Isolated State**: Setiap tes (*spec*) dimulai dari kondisi basis (baseline) yang sama.
2. **Deterministic Seed**: Data alam semesta (realm, chapter, quest, dll.) harus berasal dari _seed_ migrasi resmi tanpa intervensi data *random*.
3. **Automated Reset**: Skrip CI harus mengatur ulang (*reset*) database ke *baseline* setiap kali *suite* dijalankan.

## 1. Lingkungan Integrasi (Integration Environment)
Tes E2E tidak akan dijalankan pada lingkungan `staging` maupun `production`. E2E menggunakan:
- *Ephemeral Database* lokal (contoh: container PostgreSQL / Supabase Local).
- Skrip migrasi resmi (contoh: `scripts/migrations/010_seed_definitions.sql`) untuk mengisi struktur awal.

## 2. Test User
Untuk memastikan *login flow* dapat diprediksi:
- **Email**: `family-test@example.com`
- **Role/Level**: Level awal (kecuali ada tes khusus progresi).
- **Session/Token**: Dapat di-mock di lapisan adapter jika Gatekeeper tidak tersedia di lingkungan CI yang terisolasi, ATAU menggunakan *test account* resmi Gatekeeper.

## 3. Seed Data Basis (Phase 1.1)
Berdasarkan `010_seed_definitions.sql`, data berikut dijamin keberadaannya dalam setiap *run*:
- **Realm**: `whispering-woods` (ter-publish)
- **Chapter**: `the-awakening` dan `the-deep-woods`
- **Quest MVP (6 Quests)**: 
  - `morning-light`, `gather-herbs`, `riddle-of-the-stones` (Awakening)
  - `shadow-trail`, `the-old-growth`, `forest-riddle` (Deep Woods)
- **Challenge**: Setiap *quest* memiliki setidaknya 2 tipe tantangan (*Observation*, *Research*, dll.).

## 4. Expected State & Reset
- **Before Each Suite**: Database ephemeral di-drop dan di-recreate. Migrasi dan *seed* dijalankan ulang.
- **After Completion**: Jurnal dan progresi XP yang terjadi selama E2E (seperti penambahan poin XP di `morning-light`) hanya hidup di database ephemeral lokal dan otomatis lenyap bersama hancurnya container database pasca-CI.
