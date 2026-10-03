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
- [x] Audit seluruh entity dan kolom legacy sebelum migration baru.
- [ ] Finalisasi kontrak domain v1 berdasarkan model data keseluruhan.
- [x] Remodel schema SQLite secara menyeluruh.
- [x] Remodel repository/storage mengikuti schema v1.
- [x] Remodel service dan API mengikuti Goal-centered model.
- [x] Remodel UI dan bahasa produk agar konsisten dengan Goal.
- [x] Hapus entity `Sasaran`, `sasaran_id`, dan adapter `GoalFromSasaran` setelah persistence baru aktif.
- [x] Pastikan Goal parent/child satu Ruang Kerja dan tidak membentuk cycle.
- [x] Pastikan setiap Proyek mengacu ke satu Goal utama dan satu Ruang Kerja.
- [ ] Pastikan seluruh relasi domain mempertahankan batas `ruang_id`.
- [x] Pastikan Eksekusi dan Hasil Kerja dapat ditelusuri ke Ruang Kerja dan konteks Goal.
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

- [x] Model Eksekusi.
- [x] Jalankan proses lokal melalui argv, bukan shell interpolation.
- [x] Capture stdout/stderr.
- [x] Timeout proses.
- [x] Cancellation.
- [x] Concurrency limit.
- [x] Riwayat Eksekusi.
- [x] Event runtime.
- [x] Error classification.
- [ ] Bawa konteks Goal ke Runtime tanpa menjadikan Runtime pemilik Goal.
- [x] Hubungkan Eksekusi dengan konteks Ruang Kerja, Proyek, Tugas, dan Agen.
- [ ] Gunakan Hasil Kerja sebagai input pembaruan progress Goal.

## P1 — UI

- [x] Dasbor awal lokal.
- [x] CSS ringan tanpa framework UI besar.
- [x] Asset UI dibundel lokal.
- [x] Daftar ruang dasar.
- [x] Daftar agen.
- [x] Detail agen.
- [ ] Tampilkan daftar Tugas agen pada Detail Agen.
- [x] Struktur ruang.
- [x] Daftar goal.
- [x] Daftar proyek.
- [x] Daftar tugas.
- [x] Detail tugas.
- [x] Thread tugas.
- [x] Antrean persetujuan.
- [x] Riwayat aktivitas.
- [x] Halaman provider/model (status AI dan pilihan model).
- [x] Halaman pengaturan.
- [x] Selaraskan bahasa UI `Ruang` menjadi `Ruang Kerja` tanpa mengubah nama storage `ruang`.
- [x] Istilah `Sasaran` tidak lagi dipakai di UI; UI memakai Goal.

## P1 — Skills dan hasil

- [x] Format skill sederhana.
- [x] Skill lokal.
- [x] Relasi skill ke agen.
- [x] Penyimpanan hasil.
- [x] Validasi nama dan path hasil.
- [x] Metadata hasil.
- [x] Download/view hasil melalui UI.
- [ ] Kaitkan Hasil Kerja dengan Eksekusi dan konteks Goal.

## P1 — Scheduler

- [x] Jadwal lokal.
- [x] Heartbeat agen.
- [x] Job queue sederhana.
- [x] Pencegahan eksekusi ganda.
- [x] Retry policy.
- [x] Riwayat scheduler.
- [x] Pastikan scheduler tidak membuat eksekusi lintas Ruang Kerja.

## P1 — Data dan keamanan

- [x] Backup database.
- [x] Restore database.
- [x] Backup hasil.
- [x] Secret tidak masuk log.
- [x] Default bind hanya ke localhost.
- [x] Validasi dasar input HTTP.
- [x] Tinjau kebutuhan CSRF untuk mode UI saat ini; belum diperlukan karena API tidak memakai cookie sesi dan bind bawaan hanya localhost.
- [ ] Audit aksi penting di luar Eksekusi (persetujuan dibuat/diputuskan dan jadwal dibuat); Peristiwa saat ini hanya mencatat Eksekusi.
- [x] Dokumentasikan batas keamanan local-first.
- [ ] Jadikan Ruang Kerja sebagai unit backup/export yang eksplisit bila export/import diperluas.
- [x] Sesuaikan backup/restore dengan schema v1 setelah remodel.

## P1 — Setelah isu #28

Dikerjakan setelah remodel v1 selesai dan CI hijau. Tetap ringan: tanpa komponen dan dependency baru.

- [ ] Tambahkan deskripsi pada Tugas (satu kolom teks; prioritas tiga tingkat hanya bila terbukti perlu).
- [ ] Susun konteks Runtime dari Goal, Proyek, Tugas, dan komentar terbaru dengan batas panjang (merinci butir konteks Goal di P0 Runtime agen).
- [ ] Tambahkan layar Riwayat Eksekusi di UI: daftar dan detail Eksekusi, Peristiwa, dan Hasil Kerja per Proyek.

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

Catatan: belum ada Dockerfile. Butir ukuran Docker image di P0 Fondasi dan P2 Optimasi baru dapat dikerjakan setelah Dockerfile tersedia.

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
- [ ] Plugin sumber pengetahuan (folder Markdown, lalu DokuWiki baca-saja) setelah protokol Plugin ditetapkan; lihat docs/PLUGINS.md.

## Dokumentasi

- [x] README.md Bahasa Indonesia.
- [x] CHANGELOG.md Keep a Changelog + SemVer.
- [x] CONTRIBUTING.md.
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
- [x] docs/ROADMAP.md.
- [x] docs/penggunaan.md.
- [x] docs/scheduler.md.
- [x] docs/release.md.
- [x] docs/PLUGINS.md (rencana, belum diimplementasikan).
- [x] docs/README.md (indeks dokumentasi).
