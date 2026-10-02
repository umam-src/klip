# Changelog

Semua perubahan penting pada Klip dicatat di berkas ini.

Format mengikuti Keep a Changelog dan versi mengikuti Semantic Versioning.

## [Unreleased]

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

### Removed
- Belum ada.
