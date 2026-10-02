# Cadangan dan pemulihan data

Klip menyimpan data utama dalam satu database SQLite lokal dan berkas hasil pada direktori hasil lokal.

## Cadangan database

Gunakan helper `storage.BackupDatabase` dengan koneksi database yang sedang aktif. Klip memakai `VACUUM INTO` agar salinan konsisten dengan transaksi yang sudah tersimpan.

Tujuan cadangan tidak boleh sudah ada. Ini mencegah cadangan lama tertimpa tanpa tindakan yang disengaja.

## Cadangan hasil

Gunakan helper `storage.BackupResults` untuk membuat arsip `.tar.gz` dari direktori hasil. Arsip hanya memasukkan berkas biasa; tautan simbolik dan objek khusus tidak diikuti. Struktur direktori relatif terhadap root hasil dipertahankan.

Tujuan arsip tidak boleh sudah ada. Arsip dibuat melalui file sementara lalu dipindahkan secara atomik setelah penulisan selesai.

## Pemulihan database

Gunakan `storage.RestoreDatabase` setelah koneksi Klip ke database tujuan ditutup. Helper menyalin file ke file sementara lalu menggantinya dengan operasi rename agar kegagalan saat penyalinan tidak meninggalkan file tujuan setengah jadi.

Setelah pemulihan selesai, buka kembali database melalui `storage.Open` agar konfigurasi SQLite dan migrasi dijalankan seperti biasa.

Pemulihan berkas hasil dari arsip dilakukan secara manual setelah memeriksa isi arsip dan lokasi tujuan. Jangan mengekstrak arsip ke direktori yang digunakan proses lain tanpa pemeriksaan.

## Praktik aman

- Simpan cadangan di lokasi berbeda dari database dan direktori hasil utama.
- Jangan memasukkan database, arsip, atau isi rahasia ke log.
- Batasi akses file cadangan sesuai pengguna yang menjalankan Klip.
- Sebelum pemulihan database, buat cadangan database saat ini jika masih dapat dibuka.
- Jangan melakukan pemulihan ke database yang sedang digunakan oleh proses Klip.
