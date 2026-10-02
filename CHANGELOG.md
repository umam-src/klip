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
- Penyimpanan SQLite tanpa kebutuhan CGO.
- Bootstrap skema database dan pengujian dasar.
- Konfigurasi lokal JSON tanpa dependency parser tambahan.
- Factory provider untuk Ollama dan endpoint OpenAI-compatible.
- Endpoint lokal `POST /api/v1/chat` untuk percobaan AI.
- Test untuk konfigurasi dan pemilihan provider.
- CI hemat menit dengan cache Go, satu job test utama, build tag release, dan pemeriksaan ukuran binary.

### Changed
- Fondasi aplikasi kini membuka database lokal saat dijalankan.
- TODO diselaraskan dengan domain Klip sendiri dan tidak lagi menganggap Organisasi/Pengguna/Proyek sebagai fondasi wajib.
- Konfigurasi runtime menggunakan `config.json` agar tetap memakai pustaka standar Go.

### Removed
- Belum ada.

### Fixed
- Konfigurasi SQLite dipisahkan dari transaksi migrasi agar WAL dapat diaktifkan dengan benar.

### Security
- Database lokal menggunakan direktori aplikasi dengan permission direktori privat.
- File konfigurasi dibuat dengan permission `0600`.
- Server HTTP tetap hanya bind ke localhost secara default.
- Request chat dibatasi 64 KiB pada endpoint lokal.
- Respons provider dibatasi 8 MiB.
- API key tidak ditulis ke log.

## [0.1.0] - 2026-10-02

### Added
- Struktur awal dokumentasi proyek Klip.
- Roadmap pengembangan.
- TODO awal.
- Dasar kebijakan ukuran runtime.
- Dasar strategi local-first AI.

[Unreleased]: https://github.com/umam-src/klip/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/umam-src/klip/releases/tag/v0.1.0
