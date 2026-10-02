# TODO Klip

Daftar kerja awal untuk mencapai roadmap. Prioritas: P0 wajib, P1 penting, P2 setelah inti stabil, P3 opsional.

## P0 — Fondasi

- [ ] Inisialisasi modul Go.
- [ ] Tentukan struktur package internal.
- [ ] Tambahkan konfigurasi lokal.
- [ ] Tambahkan SQLite dan migrasi.
- [ ] Tambahkan HTTP server.
- [ ] Tambahkan template HTML.
- [ ] Tambahkan CSS ringan tanpa framework UI besar.
- [ ] Tambahkan CLI dasar.
- [ ] Tambahkan logging.
- [ ] Tambahkan health endpoint.
- [ ] Tambahkan i18n dengan `id-ID` sebagai default.
- [ ] Pastikan asset UI dibundel lokal, tanpa CDN wajib.
- [ ] Tambahkan build reproducible.
- [ ] Tambahkan pemeriksaan ukuran binary.
- [ ] Tambahkan pemeriksaan ukuran Docker image.

## P0 — Domain inti

- [ ] Model organisasi.
- [ ] Model pengguna.
- [ ] Model agen.
- [ ] Relasi hierarki agen.
- [ ] Model sasaran.
- [ ] Model proyek.
- [ ] Model tugas.
- [ ] Thread dan komentar tugas.
- [ ] Persetujuan.
- [ ] Activity log.
- [ ] Repository/service layer untuk setiap domain.
- [ ] Unit test domain.

## P0 — AI provider

- [ ] Definisikan interface provider.
- [ ] Provider HTTP generik.
- [ ] Dukungan OpenAI-compatible API.
- [ ] Adapter Ollama.
- [ ] Adapter llama.cpp/llama-server.
- [ ] Streaming.
- [ ] Timeout.
- [ ] Cancellation.
- [ ] Retry terbatas.
- [ ] Pemilihan model.
- [ ] Status koneksi provider.
- [ ] Jangan menyimpan API key ke log.

## P0 — Runtime agen

- [ ] Model run.
- [ ] Jalankan proses lokal melalui argv, bukan shell interpolation.
- [ ] Capture stdout/stderr.
- [ ] Timeout proses.
- [ ] Cancellation.
- [ ] Concurrency limit.
- [ ] Run history.
- [ ] Event runtime.
- [ ] Error classification.

## P1 — UI

- [ ] Dasbor.
- [ ] Daftar agen.
- [ ] Detail agen.
- [ ] Struktur organisasi.
- [ ] Daftar sasaran.
- [ ] Daftar proyek.
- [ ] Daftar tugas.
- [ ] Detail tugas.
- [ ] Thread tugas.
- [ ] Antrean persetujuan.
- [ ] Riwayat aktivitas.
- [ ] Halaman provider/model.
- [ ] Pengaturan.

## P1 — Skills dan artifact

- [ ] Format skill sederhana.
- [ ] Skill lokal.
- [ ] Relasi skill ke agen.
- [ ] Penyimpanan artifact.
- [ ] Validasi nama dan path artifact.
- [ ] Metadata artifact.
- [ ] Download/view artifact melalui UI.

## P1 — Scheduler

- [ ] Jadwal lokal.
- [ ] Heartbeat agen.
- [ ] Job queue sederhana.
- [ ] Pencegahan duplicate run.
- [ ] Retry policy.
- [ ] Riwayat scheduler.

## P1 — Data dan keamanan

- [ ] Backup database.
- [ ] Restore database.
- [ ] Backup artifact.
- [ ] Secret tidak masuk log.
- [ ] Default bind hanya ke localhost.
- [ ] Validasi input HTTP.
- [ ] Proteksi CSRF jika diperlukan oleh mode UI.
- [ ] Audit aksi penting.
- [ ] Dokumentasikan batas keamanan local-first.

## P2 — Optimasi

- [ ] Audit dependency.
- [ ] Hapus dependency yang tidak diperlukan.
- [ ] Optimasi binary release.
- [ ] Optimasi static asset.
- [ ] Uji binary <50 MB.
- [ ] Uji Docker image <100 MB.
- [ ] Hard fail jika artefak >500 MB.
- [ ] Benchmark startup.
- [ ] Benchmark memory idle.
- [ ] Benchmark SQLite.
- [ ] Benchmark concurrent runs.

## P2 — CI/CD hemat menit

- [ ] Cache Go module dan build cache.
- [ ] Satu job test utama untuk pull request.
- [ ] Build release hanya pada tag.
- [ ] Browser E2E tidak dijalankan pada setiap commit jika tidak diperlukan.
- [ ] Size check dilakukan setelah build.
- [ ] Hindari matrix platform berlebihan pada setiap PR.
- [ ] Cross-platform build pada release.

## P3 — Integrasi tambahan

- [ ] Provider cloud opsional.
- [ ] Git integration.
- [ ] Connector tambahan berdasarkan kebutuhan nyata.
- [ ] Import/export tambahan.
- [ ] Integrasi chat.

## Dokumentasi

- [ ] README.md Bahasa Indonesia.
- [ ] CHANGELOG.md Keep a Changelog + SemVer.
- [ ] CONTRIBUTING.md.
- [ ] AGENTS.md.
- [ ] LICENSE MIT.
- [ ] docs/architecture.md.
- [ ] docs/data-model.md.
- [ ] docs/runtime.md.
- [ ] docs/providers.md.
- [ ] docs/security.md.
- [ ] docs/storage.md.
- [ ] docs/backup.md.
- [ ] docs/release.md.
- [ ] docs/size-budget.md.
