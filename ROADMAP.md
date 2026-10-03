# Roadmap Klip

Klip adalah orkestrator AI yang ringan, offline-first, dan mengutamakan AI lokal atau layanan AI gratis.

## Sasaran utama

- Binary runtime utama ditargetkan di bawah 50 MB.
- Docker image minimal ditargetkan di bawah 100 MB.
- Batas keras artefak runtime: 500 MB.
- Model AI tidak dibundel ke binary atau image Klip.
- Bahasa antarmuka default: Bahasa Indonesia.
- AI lokal menjadi pilihan utama; layanan gratis menjadi pilihan berikutnya.
- Klip mengambil kebutuhan produk dari proyek inspirasi tanpa menyalin source code, UI, atau domain secara 1:1.

## Versi

### 0.1 — Fondasi

- [x] Struktur proyek Go.
- [x] Fondasi single binary.
- [x] SQLite lokal.
- [x] HTTP server lokal.
- [x] UI dasar.
- [x] CLI dasar.
- [x] Konfigurasi lokal.
- [x] Sistem i18n dengan `id-ID` sebagai default.
- [x] Health check dan logging sederhana.
- [x] CI dasar dengan cache dan pemeriksaan ukuran binary.

### 0.2 — Inti ruang dan pekerjaan

- [x] Ruang.
- [x] Agen.
- [x] Sasaran.
- [x] Pekerjaan.
- [x] Tugas.
- [x] Sesi.
- [x] Hasil.
- [x] Thread/komentar tugas.
- [x] Persetujuan.
- [x] Riwayat aktivitas.
- [x] Repository/service layer.

### 0.3 — Local AI

- [x] Antarmuka provider AI.
- [x] Ollama melalui endpoint OpenAI-compatible.
- [x] OpenAI-compatible API.
- [ ] llama.cpp/llama-server adapter khusus.
- [x] Streaming respons.
- [x] Timeout dan pembatalan.
- [x] Retry sederhana.
- [x] Pemilihan model.
- [x] Status koneksi provider.
- [x] Konfigurasi endpoint lokal.

### 0.4 — Runtime agen

- [x] Eksekusi proses lokal.
- [ ] Tool execution.
- [x] Skills.
- [x] Heartbeat.
- [x] Scheduler.
- [x] Riwayat eksekusi.
- [x] Batas waktu.
- [x] Batas concurrency.
- [x] Artifact.

### 0.5 — AI gratis dan provider tambahan

- [ ] Provider dengan free tier.
- [ ] Fallback provider.
- [ ] Routing local-first.
- [ ] Pemantauan kuota.
- [ ] Provider cloud opsional.

### 0.6 — UI lengkap

- [x] Dasbor.
- [x] Manajemen agen.
- [ ] Ruang dan sasaran.
- [ ] Daftar dan detail pekerjaan.
- [x] Daftar dan detail tugas.
- [x] Thread tugas.
- [x] Antrean persetujuan.
- [ ] Riwayat eksekusi.
- [x] Model dan provider.
- [x] Pengaturan.

### 0.7 — Keamanan dan ketahanan

- [ ] Manajemen secret.
- [ ] Permission dasar.
- [x] Audit log.
- [x] Validasi path artifact.
- [x] Pembatasan proses.
- [x] Backup.
- [x] Restore.
- [ ] Pengujian migrasi database.

### 0.8 — Optimasi

- [x] Ukur ukuran binary secara otomatis.
- [ ] Ukur ukuran Docker image.
- [x] Audit dependency.
- [x] Strip simbol debug untuk release.
- [ ] Optimasi asset UI.
- [x] Uji startup time.
- [x] Uji memory idle.
- [x] Uji penggunaan SQLite.

### 0.9 — Kompatibilitas

- [ ] Matriks kebutuhan produk dari proyek inspirasi.
- [ ] Import data yang benar-benar diperlukan.
- [ ] Penyelarasan konsep domain bila dibutuhkan.
- [ ] Stabilitas API lokal.
- [ ] Dokumentasi migrasi.

### 1.0 — Rilis stabil

- [ ] Regression test lengkap.
- [x] Backup/restore teruji.
- [ ] Upgrade database teruji.
- [ ] Release binary lintas platform.
- [ ] Docker image release.
- [ ] Dokumentasi pengguna lengkap.
- [ ] CHANGELOG final untuk rilis 1.0.

## Setelah 1.0

- [ ] Bahasa Inggris.
- [ ] Bahasa tambahan.
- [ ] Provider AI tambahan.
- [ ] Integrasi eksternal berdasarkan kebutuhan.
- [ ] Fitur lanjutan tanpa mengorbankan ukuran dan kesederhanaan.
