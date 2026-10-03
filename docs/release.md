# Rilis Klip

Rilis Klip mengikuti Semantic Versioning.

## Sebelum rilis

1. Pastikan perubahan utama tercatat di `CHANGELOG.md`, lalu pindahkan dari `[Unreleased]` ke bagian versi baru berikut tanggalnya (`## [X.Y.Z] - TAHUN-BULAN-TANGGAL`).
2. Jalankan test lokal.
3. Periksa ukuran binary hasil build.
4. Pastikan tidak ada secret, credential, database lokal, atau data pengguna yang ikut masuk ke repository.
5. Periksa perubahan migrasi database jika ada.
6. Pastikan dokumentasi perilaku yang berubah sudah diperbarui.

## Versi

- **MAJOR**: perubahan yang memutus kompatibilitas.
- **MINOR**: kemampuan baru yang kompatibel dengan rancangan yang ada.
- **PATCH**: perbaikan bug, keamanan, atau perubahan internal yang tidak mengubah kontrak pengguna.

Versi minor harus mengikuti rancangan produk Klip, bukan sekadar meniru versi proyek lain.

## Build

Build release mengutamakan binary yang kecil dan dapat direproduksi. Optimasi ukuran tidak boleh menghapus pemeriksaan keamanan, informasi yang dibutuhkan untuk diagnosis, atau membuat proses build rapuh.

Target ukuran runtime:

- binary di bawah 50 MiB;
- Docker image di bawah 100 MiB.

Batas keras release untuk binary adalah 100 MiB.

## Tag dan workflow

- `CI` berjalan otomatis pada pull request dan push ke `main`, kecuali perubahan yang hanya menyentuh dokumen. Isinya pemeriksaan format, `go vet`, dan test.
- `Build` hanya berjalan manual untuk mencoba build binary Linux, Windows, dan macOS beserta pemeriksaan ukurannya.
- `Release` berjalan saat tag `v*` didorong: memeriksa kode, membangun binary tiga platform, lalu membuat GitHub Release.

Versi dan catatan rilis dibaca dari bagian versi pertama di `CHANGELOG.md`, jadi tag `vX.Y.Z` harus sesuai dengan bagian `[X.Y.Z]` di sana dan menunjuk commit yang dirilis.

## Setelah rilis

- periksa artifact release;
- periksa checksum jika tersedia;
- catat perubahan penting;
- pastikan migrasi database tetap dapat digunakan pada versi berikutnya.
