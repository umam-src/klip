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
- status akhir Eksekusi (selesai, gagal, atau dibatalkan) disimpan dengan batas waktu sendiri agar tetap tercatat meski konteks permintaan sudah dibatalkan;
- keluaran proses tidak dicatat sebagai `Hasil` secara otomatis.

## HTTP lokal

Server menggunakan `127.0.0.1` sebagai alamat bawaan. Endpoint yang tersedia pada versi awal menangani kesehatan, percakapan AI, dan data domain yang diperlukan aplikasi.

Tidak ada endpoint untuk menjalankan program arbitrer. Ketidakhadiran endpoint ini adalah bagian dari desain keamanan, bukan keterbatasan yang boleh dilewati dengan meneruskan parameter proses ke endpoint lain.

Endpoint yang menerima JSON hanya memproses `application/json` bila `Content-Type` diberikan. Ini membatasi beberapa permintaan lintas situs sederhana dengan tipe isi yang tidak cocok, tetapi bukan pengganti perlindungan CSRF.

### CSRF

Klip saat ini tidak menggunakan cookie sesi atau autentikasi browser untuk API lokal. Server juga hanya bind ke localhost secara bawaan. Karena tidak ada kredensial browser yang dikirim otomatis ke endpoint mutasi, perlindungan CSRF berbasis token belum diperlukan pada mode UI saat ini.

Jika autentikasi berbasis cookie atau mode bind non-localhost ditambahkan, CSRF harus ditinjau kembali sebelum mode tersebut diaktifkan.

## Jejak aktivitas

Klip mencatat Peristiwa lokal saat Eksekusi mulai, selesai, gagal, atau dibatalkan. Pencatatan bersifat best-effort: kegagalan menulis Peristiwa tidak menggagalkan Eksekusi. Argumen, keluaran proses, prompt, dan data rahasia tidak ikut dicatat.

Persetujuan dan pembuatan Jadwal belum dicatat sebagai Peristiwa karena keduanya bukan pelaksanaan Eksekusi. Audit untuk aksi tersebut belum tersedia dan perlu dirancang tersendiri; lihat `TODO.md`.

## Data lokal

Database menggunakan SQLite lokal dengan foreign key aktif, WAL, dan satu koneksi aktif. Data konfigurasi disimpan dengan izin file yang membatasi akses pada pengguna lokal.

Model AI tidak dibundel ke binary Klip dan file model tetap berada di luar aplikasi.

## Batas yang belum tersedia

Beberapa perlindungan tetap belum tersedia karena belum ada alur produk yang membutuhkannya:

- autentikasi dan otorisasi HTTP;
- isolasi proses atau sandbox;
- kebijakan `Origin` khusus untuk browser.

Sebelum eksekusi proses dibuka melalui HTTP atau server dipasang di luar localhost, batas-batas tersebut perlu dirancang sebagai bagian dari alur eksekusi, bukan ditambahkan sebagai tambalan pada endpoint.
