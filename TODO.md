# TODO Klip

Daftar kerja pengembangan. Prioritas: P0 wajib, P1 penting, P2 setelah inti stabil, P3 opsional.

## P0 — Fondasi

- [x] Inisialisasi modul Go.
- [x] Tentukan struktur package internal.
- [x] Tambahkan konfigurasi lokal.
- [x] Tambahkan SQLite dan migrasi.
- [x] Tambahkan HTTP server.
- [x] Tambahkan template HTML.
- [x] Tambahkan CSS ringan tanpa framework UI besar.
- [x] Tambahkan CLI dasar.
- [x] Tambahkan logging.
- [x] Tambahkan health endpoint.
- [x] Tambahkan i18n dengan `id-ID` sebagai default.
- [x] Pastikan asset UI dibundel lokal, tanpa CDN wajib.
- [x] Tambahkan build reproducible.
- [x] Tambahkan pemeriksaan ukuran binary di CI.
- [ ] Tambahkan pemeriksaan ukuran Docker image.

## P0 — Domain inti

- [x] Relasi hierarki agen.
- [x] Thread dan komentar tugas.
- [x] Persetujuan.
- [x] Activity log.
- [x] Repository/service layer untuk domain.
- [x] Unit test domain.
- [x] Tetapkan Ruang Kerja sebagai batas domain.
- [x] Tetapkan Goal sebagai pusat arah kerja.
- [x] Audit model lama `Sasaran` dan relasi `Pekerjaan` terhadap kontrak Goal.
- [x] Tetapkan kontrak Goal v1: identitas, `ruang_id`, hierarki Goal, judul, deskripsi, lifecycle, dan timestamp.
- [x] Putuskan bahwa beta belum membutuhkan kompatibilitas database legacy.
- [x] Tetapkan peta entity dan relasi keseluruhan sebelum remodel schema.
- [ ] Audit seluruh entity dan kolom legacy sebelum migration baru.
- [ ] Finalisasi kontrak domain v1 berdasarkan model data keseluruhan.
- [ ] Remodel schema SQLite secara menyeluruh.
- [ ] Remodel repository/storage mengikuti schema v1.
- [ ] Remodel service dan API mengikuti Goal-centered model.
- [ ] Remodel UI dan bahasa produk agar konsisten dengan Goal.
- [ ] Hapus entity `Sasaran`, `sasaran_id`, dan adapter `GoalFromSasaran` setelah persistence baru aktif.
- [ ] Pastikan Goal parent/child satu Ruang Kerja dan tidak membentuk cycle.
- [ ] Pastikan Goal → Pekerjaan satu Goal utama dan satu Ruang Kerja.
- [ ] Pastikan seluruh relasi domain mempertahankan batas `ruang_id`.
- [ ] Pastikan Execution dan Result/Evidence dapat ditelusuri ke Ruang Kerja dan konteks Goal.
- [ ] Reset/migrate database development setelah schema v1 siap.
- [ ] Jalankan regression test dan CI penuh setelah remodel.

Catatan: `Sasaran` tidak lagi dipertahankan sebagai compatibility layer permanen. Karena Klip belum rilis, schema lama boleh tidak kompatibel. Prioritasnya adalah satu model data yang konsisten sebelum rilis.

## P0 — AI provider

- [x] Definisikan interface provider.
- [x] Provider HTTP generik melalui endpoint OpenAI-compatible.
- [x] Dukungan OpenAI-compatible API.
- [x] Adapter Ollama.
- [ ] Adapter llama.cpp/llama-server khusus jika kebutuhan native muncul.
- [x] Streaming.
- [x] Timeout.
- [x] Cancellation melalui `context.Context`.
- [x] Retry terbatas.
- [x] Pemilihan model melalui konfigurasi/request.
- [x] Status koneksi provider.
- [x] Jangan menyimpan API key ke log.

## P0 — Runtime agen

- [x] Model run.
- [x] Jalankan proses lokal melalui argv, bukan shell interpolation.
- [x] Capture stdout/stderr.
- [x] Timeout proses.
- [x] Cancellation.
- [x] Concurrency limit.
- [x] Run history.
- [x] Event runtime.
- [x] Error classification.
- [ ] Bawa konteks Goal ke Runtime tanpa menjadikan Runtime pemilik Goal.
- [ ] Hubungkan Execution dengan konteks Ruang Kerja, Pekerjaan, Tugas, dan Agen.
- [ ] Gunakan Result/Evidence sebagai input pembaruan progress Goal.

## P1 — UI

- [x] Dasbor awal lokal.
- [x] CSS ringan tanpa framework UI besar.
- [x] Asset UI dibundel lokal.
- [x] Daftar ruang dasar.
- [x] Daftar agen.
- [x] Detail agen.
- [x] Struktur ruang.
- [x] Daftar sasaran.
- [x] Daftar pekerjaan.
- [x] Daftar tugas.
- [x] Detail tugas.
- [x] Thread tugas.
- [x] Antrean persetujuan.
- [x] Riwayat aktivitas.
- [x] Halaman provider/model.
- [x] Pengaturan.
- [ ] Selaraskan bahasa UI `Ruang` menjadi `Ruang Kerja` tanpa mengubah nama storage `ruang`.
- [ ] Selaraskan istilah `Sasaran` menjadi Goal setelah model Goal final ditetapkan.

## P1 — Skills dan hasil

- [x] Format skill sederhana.
- [x] Skill lokal.
- [x] Relasi skill ke agen.
- [x] Penyimpanan hasil.
- [x] Validasi nama dan path hasil.
- [x] Metadata hasil.
- [x] Download/view hasil melalui UI.
- [ ] Kaitkan hasil kerja dengan Execution dan konteks Goal.

## P1 — Scheduler

- [x] Jadwal lokal.
- [x] Heartbeat agen.
- [x] Job queue sederhana.
- [x] Pencegahan duplicate run.
- [x] Retry policy.
- [x] Riwayat scheduler.
- [ ] Pastikan scheduler tidak membuat eksekusi lintas Ruang Kerja.

## P1 — Data dan keamanan

- [x] Backup database.
- [x] Restore database.
- [x] Backup hasil.
- [x] Secret tidak masuk log.
- [x] Default bind hanya ke localhost.
- [x] Validasi dasar input HTTP.
- [x] Tinjau kebutuhan CSRF untuk mode UI saat ini; belum diperlukan karena API tidak memakai cookie sesi dan bind bawaan hanya localhost.
- [x] Audit aksi penting: approval dibuat/diputuskan dan jadwal dibuat dicatat sebagai event lokal.
- [x] Dokumentasikan batas keamanan local-first.
- [ ] Jadikan Ruang Kerja sebagai unit backup/export yang eksplisit bila export/import diperluas.
- [ ] Sesuaikan backup/restore dengan schema v1 setelah remodel.

## P2 — Optimasi

- [x] Audit dependency.
- [ ] Hapus dependency yang tidak diperlukan.
- [x] Optimasi binary release dasar dengan `-trimpath -ldflags '-s -w'` di CI.
- [ ] Optimasi static asset.
- [x] Uji target binary <50 MiB.
- [ ] Uji Docker image <100 MiB.
- [x] Hard fail jika binary >100 MiB.
- [ ] Hard fail Docker image >100 MiB.
- [x] Benchmark startup komponen utama.
- [x] Benchmark memory idle.
- [x] Benchmark SQLite.
- [x] Benchmark concurrent runs.

## P2 — CI/CD hemat menit

- [x] Cache Go module dan build cache.
- [x] Satu job test utama untuk pull request/push utama.
- [x] Build release hanya pada tag.
- [ ] Browser E2E tidak dijalankan pada setiap commit jika tidak diperlukan.
- [x] Size check dilakukan setelah build.
- [x] Hindari matrix platform berlebihan pada setiap PR.
- [x] Cross-platform build pada release.

## P3 — Integrasi tambahan

- [ ] Provider cloud opsional.
- [ ] Git integration.
- [ ] Connector tambahan berdasarkan kebutuhan nyata.
- [ ] Import/export tambahan.
- [ ] Integrasi chat.

## Dokumentasi

- [x] README.md Bahasa Indonesia.
- [x] CHANGELOG.md Keep a Changelog + SemVer.
- [x] CONTRIBUTING.md.
- [x] AGENTS.md.
- [x] LICENSE MIT.
- [x] docs/ARCHITECTURE.md.
- [x] docs/DOMAIN.md.
- [x] docs/PARITY.md.
- [x] docs/data-model.md.
- [x] docs/runtime.md.
- [x] docs/AI-PROVIDERS.md.
- [x] docs/security.md.
- [x] docs/storage.md.
- [x] docs/backup.md.
- [x] docs/size-budget.md.
