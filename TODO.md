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

- [ ] Relasi hierarki agen.
- [x] Thread dan komentar tugas.
- [x] Persetujuan.
- [x] Activity log.
- [x] Repository/service layer untuk domain.
- [x] Unit test domain.

Catatan: `Ruang`, `Agen`, `Sasaran`, `Pekerjaan`, `Tugas`, `Sesi`, dan `Hasil` sudah memiliki dasar skema. Konsep `Organisasi`, `Pengguna`, dan `Proyek` tidak menjadi syarat fondasi Klip.

## P0 — AI provider

- [x] Definisikan interface provider.
- [x] Provider HTTP generik melalui endpoint OpenAI-compatible.
- [x] Dukungan OpenAI-compatible API.
- [x] Adapter Ollama.
- [ ] Adapter llama.cpp/llama-server khusus jika kebutuhan native muncul.
- [ ] Streaming.
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

## P1 — UI

- [x] Dasbor awal lokal.
- [x] CSS ringan tanpa framework besar.
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

## P1 — Skills dan hasil

- [x] Format skill sederhana.
- [x] Skill lokal.
- [x] Relasi skill ke agen.
- [x] Penyimpanan hasil.
- [x] Validasi nama dan path hasil.
- [x] Metadata hasil.
- [x] Download/view hasil melalui UI.

## P1 — Scheduler

- [x] Jadwal lokal.
- [x] Heartbeat agen.
- [x] Job queue sederhana.
- [x] Pencegahan duplicate run.
- [x] Retry policy.
- [x] Riwayat scheduler.

## P1 — Data dan keamanan

- [x] Backup database.
- [x] Restore database.
- [ ] Backup hasil.
- [x] Secret tidak masuk log.
- [x] Default bind hanya ke localhost.
- [x] Validasi dasar input HTTP.
- [ ] Proteksi CSRF jika diperlukan oleh mode UI.
- [ ] Audit aksi penting.
- [x] Dokumentasikan batas keamanan local-first.

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
- [ ] Benchmark memory idle.
- [x] Benchmark SQLite.
- [x] Benchmark concurrent runs.

## P2 — CI/CD hemat menit

- [x] Cache Go module dan build cache.
- [x] Satu job test utama untuk pull request/push utama.
- [x] Build release hanya pada tag.
- [ ] Browser E2E tidak dijalankan pada setiap commit jika tidak diperlukan.
- [x] Size check dilakukan setelah build.
- [x] Hindari matrix platform berlebihan pada setiap PR.
- [ ] Cross-platform build pada release.

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
- [x] docs/release.md.
- [x] docs/size-budget.md.
