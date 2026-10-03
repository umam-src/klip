# Progres UI Detail Agen

Status: **Implementasi bertahap — belum menjadi spesifikasi UI final.**

## Tahap 5 — Detail Agen

Detail Agen sekarang diarahkan untuk menjawab hal yang paling dibutuhkan pengguna tanpa membuka konfigurasi teknis terlalu dini.

Urutan informasi:

1. **Ringkasan** — nama, peran, status, dan deskripsi.
2. **AI yang digunakan** — penyedia dan model yang terhubung pada Agen.
3. **Struktur agen** — atasan dan identitas Agen.
4. **Pekerjaan** — sementara menampilkan keadaan belum ada pekerjaan yang terhubung langsung.
5. **Aktivitas** — sementara menampilkan keadaan belum ada aktivitas Agen yang ditampilkan.
6. **Pengaturan** — belum editable; pengaturan Agen akan datang setelah alur edit matang.

## Batas tahap ini

- Tidak menambahkan relasi Pekerjaan → Agen yang belum tersedia di domain.
- Tidak membuat data Aktivitas Agen buatan hanya agar tampilan terlihat penuh.
- Tidak menampilkan kredensial atau token.
- Tidak memasukkan banyak penyedia/Akun AI sebagai bagian dari tahap ini.
- Detail dibuat sebagai langkah UX terlebih dahulu; kemampuan edit dan hubungan kerja ditambahkan setelah kontrak domain siap.

## Catatan implementasi

Detail dimuat dari API Agen yang sudah ada. Modul UI terpisah menangani tampilan detail sehingga logika daftar Agen tetap sederhana.
