# PWA Validation Checklist

Sesuai prinsip **Evidence First**, setiap baris pengujian PWA harus dilengkapi dengan bukti (evidence) yang diverifikasi saat pengujian runtime atau melalui *tools* khusus (misalnya Lighthouse atau devtools browser).

| Item | Expected | Actual | PASS/FAIL | Evidence |
|------|----------|--------|-----------|----------|
| **Install** | Muncul prompt instalasi PWA dan aplikasi dapat diinstal di homescreen (A2HS). | | | |
| **Offline** | Aplikasi dapat dibuka tanpa koneksi internet (menampilkan halaman yang di-*cache*). | | | |
| **Refresh (Offline)** | Saat *offline*, me-refresh halaman tidak memunculkan "No Internet" dino, melainkan memuat ulang UI dari *cache*. | | | |
| **Reconnect** | Saat koneksi kembali, aplikasi dapat melakukan fetch data terbaru dan me-reset status offline. | | | |
| **New Deployment** | Saat versi baru di-deploy ke server, *Service Worker* mendeteksi pembaruan di _background_. | | | |
| **Update Prompt** | Setelah mendeteksi pembaruan, *prompt* "Update Available" muncul, dan mengkliknya akan memuat *cache* baru. | | | |

*Catatan: Checklist ini harus diisi dan dilampirkan sebelum melakukan rilis tag.*
