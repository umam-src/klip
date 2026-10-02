# Roadmap Klip

Klip adalah orkestrator AI yang ringan, offline-first, dan mengutamakan AI lokal atau layanan AI gratis.

## Sasaran utama

- Binary runtime utama ditargetkan di bawah 50 MB.
- Docker image minimal ditargetkan di bawah 100 MB.
- Batas keras artefak runtime: 500 MB.
- Model AI tidak dibundel ke binary atau image Klip.
- Bahasa antarmuka default: Bahasa Indonesia.
- AI lokal menjadi pilihan utama; layanan gratis menjadi pilihan berikutnya.
- Fungsi utama Paperclip dipertahankan, tetapi implementasi dibuat lebih sederhana.

## Versi

### 0.1 — Fondasi

- [ ] Struktur proyek Go.
- [ ] Single binary.
- [ ] SQLite lokal.
- [ ] HTTP server dan UI dasar.
- [ ] CLI dasar.
- [ ] Konfigurasi lokal.
- [ ] Sistem i18n dengan `id-ID` sebagai default.
- [ ] Health check dan logging sederhana.
- [ ] CI dasar dengan cache dan pemeriksaan ukuran.

### 0.2 — Inti organisasi dan pekerjaan

- [ ] Organisasi.
- [ ] Pengguna.
- [ ] Agen dan struktur hierarki.
- [ ] Sasaran.
- [ ] Proyek.
- [ ] Tugas.
- [ ] Percakapan/thread tugas.
- [ ] Persetujuan.
- [ ] Riwayat aktivitas.

### 0.3 — Local AI

- [ ] Antarmuka provider AI.
- [ ] Ollama.
- [ ] OpenAI-compatible API.
- [ ] llama.cpp/llama-server.
- [ ] Streaming respons.
- [ ] Timeout dan pembatalan.
- [ ] Retry sederhana.
- [ ] Pemilihan model.
- [ ] Konfigurasi endpoint lokal.

### 0.4 — Runtime agen

- [ ] Eksekusi proses lokal.
- [ ] Tool execution.
- [ ] Skills.
- [ ] Heartbeat.
- [ ] Scheduler.
- [ ] Riwayat eksekusi.
- [ ] Batas waktu.
- [ ] Batas concurrency.
- [ ] Artifact.

### 0.5 — AI gratis dan provider tambahan

- [ ] Provider dengan free tier.
- [ ] Fallback provider.
- [ ] Routing local-first.
- [ ] Pemantauan kuota.
- [ ] Provider cloud opsional.

### 0.6 — UI lengkap

- [ ] Dasbor.
- [ ] Manajemen agen.
- [ ] Sasaran dan proyek.
- [ ] Daftar dan detail tugas.
- [ ] Thread tugas.
- [ ] Antrean persetujuan.
- [ ] Riwayat eksekusi.
- [ ] Model dan provider.
- [ ] Pengaturan.

### 0.7 — Keamanan dan ketahanan

- [ ] Manajemen secret.
- [ ] Permission dasar.
- [ ] Audit log.
- [ ] Validasi path artifact.
- [ ] Pembatasan proses.
- [ ] Backup.
- [ ] Restore.
- [ ] Pengujian migrasi database.

### 0.8 — Optimasi

- [ ] Ukur ukuran binary secara otomatis.
- [ ] Ukur ukuran Docker image.
- [ ] Optimasi dependency.
- [ ] Strip simbol debug untuk release.
- [ ] Optimasi asset UI.
- [ ] Uji startup time.
- [ ] Uji memory idle.
- [ ] Uji penggunaan SQLite.

### 0.9 — Kompatibilitas

- [ ] Matriks fungsi utama Paperclip.
- [ ] Import data yang diperlukan.
- [ ] Penyelarasan konsep domain.
- [ ] Stabilitas API lokal.
- [ ] Dokumentasi migrasi.

### 1.0 — Rilis stabil

- [ ] Regression test lengkap.
- [ ] Backup/restore teruji.
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
