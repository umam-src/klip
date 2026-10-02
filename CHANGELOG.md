# Changelog

Semua perubahan penting pada Klip dicatat di berkas ini.

Format mengikuti Keep a Changelog dan versi mengikuti Semantic Versioning.

## [Unreleased]

### Added
- CLI dasar untuk menjalankan server, melihat versi, dan membuka bantuan.
- Fondasi i18n dengan `id-ID` sebagai locale default.
- Endpoint status koneksi provider AI lokal.
- Event runtime untuk mencatat awal dan akhir eksekusi secara lokal.
- Format `SKILL.md` sederhana untuk instruksi skill lokal.
- Loader skill lokal dengan batas ukuran, validasi nama, penolakan tautan simbolik, dan urutan deterministik.

### Changed
- Konfigurasi sekarang menyimpan locale secara eksplisit dan menormalkan locale yang belum didukung ke `id-ID`.
- State `Run`, `Sesi`, dan `Tugas` beserta event runtime diperbarui secara atomik saat memulai dan menyelesaikan eksekusi.

### Removed
- Belum ada.

## [0.1.0] - 2026-10-02

### Added
- Kerangka roadmap Klip.
- Target binary dan Docker image yang terukur.
- Prinsip local-first dan free-first untuk AI.
- Bahasa Indonesia sebagai bahasa default.
- Domain inti Klip: Ruang, Agen, Sasaran, Pekerjaan, Tugas, Sesi, dan Hasil.
- Model `Run` untuk riwayat eksekusi agen.
- Runner proses lokal berbasis argv tanpa interpolasi shell.
- Batas output proses default 1 MiB.
- Batas proses berjalan bersamaan pada runner.
- Klasifikasi kesalahan proses untuk timeout, pembatalan, kegagalan proses, dan kegagalan memulai.
- Penyimpanan riwayat `Run` di SQLite.
- Penyimpanan `Sesi` dan penyelesaian status eksekusi.
- Executor yang menghubungkan Pekerjaan, Sesi, Agen, dan Run.
- Migrasi skema SQLite versi 2 untuk riwayat eksekusi.
- Penyimpanan SQLite tanpa kebutuhan CGO.
- Bootstrap skema database dan pengujian dasar.
- Konfigurasi lokal JSON tanpa dependency parser tambahan.
- Factory provider untuk Ollama dan endpoint OpenAI-compatible.
- Endpoint lokal `POST /api/v1/chat` untuk percobaan AI.
- Test untuk konfigurasi dan pemilihan provider.
- CI hemat menit dengan cache Go, satu job test utama, build tag release, dan pemeriksaan ukuran binary.
- Dokumentasi batas keamanan local-first.

### Changed
- Fondasi aplikasi kini membuka database lokal saat dijalankan.
- TODO diselaraskan dengan domain Klip sendiri dan tidak lagi menganggap Organisasi/Pengguna/Proyek sebagai fondasi wajib.
- Konfigurasi runtime menggunakan `config.json` agar tetap memakai pustaka standar Go.
- Runtime eksekusi menolak input proses yang mengandung NUL dan direktori kerja relatif.
- API JSON menolak `Content-Type` non-JSON ketika header tersebut diberikan.
- Alur rilis mengambil versi dan catatan rilis langsung dari `CHANGELOG.md`.

### Removed
- Belum ada.
