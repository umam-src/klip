# Cadangan dan pemulihan data

Klip menyimpan data utama dalam satu database SQLite lokal. Cadangan dibuat sebagai satu file SQLite baru sehingga tidak perlu menyalin file WAL dan file sementara secara terpisah.

## Cadangan

Gunakan helper `storage.BackupDatabase` dengan koneksi database yang sedang aktif. Klip memakai `VACUUM INTO` agar salinan konsisten dengan transaksi yang sudah tersimpan.

Tujuan cadangan tidak boleh sudah ada. Ini mencegah cadangan lama tertimpa tanpa tindakan yang disengaja.

## Pemulihan

Gunakan `storage.RestoreDatabase` setelah koneksi Klip ke database tujuan ditutup. Helper menyalin file ke file sementara lalu menggantinya dengan operasi rename agar kegagalan saat penyalinan tidak meninggalkan file tujuan setengah jadi.

Setelah pemulihan selesai, buka kembali database melalui `storage.Open` agar konfigurasi SQLite dan migrasi dijalankan seperti biasa.

## Praktik aman

- Simpan cadangan di lokasi berbeda dari database utama.
- Jangan memasukkan database atau cadangan ke log.
- Batasi akses file cadangan sesuai pengguna yang menjalankan Klip.
- Sebelum pemulihan, buat cadangan database saat ini jika masih dapat dibuka.
- Jangan melakukan pemulihan ke database yang sedang digunakan oleh proses Klip.
