# Keamanan Klip

Dokumen ini menjelaskan batas keamanan dasar Klip pada rilis awal.

## Batas proses

Klip dapat menjalankan proses lokal melalui runtime internal. Jalur ini tidak tersedia sebagai endpoint HTTP publik pada versi awal.

Alasan utamanya sederhana: menerima nama program dan argumen dari HTTP lalu menjalankannya langsung akan membuat API menjadi jalur eksekusi sistem operasi. Batas ini sengaja dipertahankan sampai mekanisme izin, persetujuan, dan pencatatan yang sesuai tersedia.

Runtime internal tetap menerapkan batas berikut:

- program dijalankan tanpa shell;
- input proses menolak karakter NUL;
- direktori kerja yang diberikan harus berupa path absolut;
- proses dapat dibatalkan melalui context;
- keluaran proses memiliki batas ukuran;
- jumlah proses bersamaan dibatasi runner;
- status `Run`, `Sesi`, dan `Tugas` diselesaikan secara atomik;
- keluaran proses tidak dicatat sebagai `Hasil` secara otomatis.

## HTTP lokal

Server menggunakan `127.0.0.1` sebagai alamat bawaan. Endpoint yang tersedia pada versi awal menangani kesehatan, percakapan AI, dan data domain yang diperlukan aplikasi.

Tidak ada endpoint untuk menjalankan program arbitrer. Ketidakhadiran endpoint ini adalah bagian dari desain keamanan, bukan keterbatasan yang boleh dilewati dengan meneruskan parameter proses ke endpoint lain.

## Data lokal

Database menggunakan SQLite lokal dengan foreign key aktif, WAL, dan satu koneksi aktif. Data konfigurasi disimpan dengan izin file yang membatasi akses pada pengguna lokal.

Model AI tidak dibundel ke binary Klip dan file model tetap berada di luar aplikasi.

## Batas yang belum tersedia

Beberapa perlindungan sengaja belum diaktifkan karena belum ada alur produk yang membutuhkannya:

- persetujuan manusia untuk eksekusi;
- audit aktivitas;
- isolasi proses atau sandbox;
- autentikasi dan otorisasi HTTP;
- CSRF untuk alur yang mengubah data melalui browser.

Sebelum eksekusi proses dibuka melalui HTTP, batas-batas tersebut perlu dirancang sebagai bagian dari alur eksekusi, bukan ditambahkan sebagai tambalan pada endpoint.
