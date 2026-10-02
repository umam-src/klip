# Changelog

Semua perubahan penting pada Klip dicatat di berkas ini.

Format mengikuti Keep a Changelog dan versi mengikuti Semantic Versioning.

## [Unreleased]

### Added
- Kerangka roadmap Klip.
- Daftar tugas awal pengembangan.
- Target binary dan Docker image yang terukur.
- Prinsip local-first dan free-first untuk AI.
- Bahasa Indonesia sebagai bahasa default.
- Domain inti Klip: Ruang, Agen, Sasaran, Pekerjaan, Tugas, Sesi, dan Hasil.
- Penyimpanan SQLite tanpa kebutuhan CGO.
- Bootstrap skema database dan pengujian dasar.

### Changed
- Fondasi aplikasi kini membuka database lokal saat dijalankan.

### Removed
- Belum ada.

### Fixed
- Konfigurasi SQLite dipisahkan dari transaksi migrasi agar WAL dapat diaktifkan dengan benar.

### Security
- Database lokal menggunakan direktori aplikasi dengan permission direktori privat.
- Server HTTP tetap hanya bind ke localhost.

## [0.1.0] - 2026-10-02

### Added
- Struktur awal dokumentasi proyek Klip.
- Roadmap pengembangan.
- TODO awal.
- Dasar kebijakan ukuran runtime.
- Dasar strategi local-first AI.

[Unreleased]: https://github.com/umam-src/klip/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/umam-src/klip/releases/tag/v0.1.0
