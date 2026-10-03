# Scheduler Klip

Scheduler berjalan lokal di proses Klip dan memakai SQLite sebagai sumber jadwal serta riwayat.

## Jadwal

Jadwal menyimpan:

- proyek dan tugas yang dijalankan;
- agen yang digunakan;
- program dan argumen proses;
- waktu eksekusi berikutnya;
- batas percobaan ulang.

Rentang interval yang diterima API adalah 1 detik sampai 7 hari.

Setiap jadwal berada dalam satu Ruang Kerja yang diambil dari Proyeknya, bukan dari input klien. Jadwal dengan agen atau tugas dari Ruang Kerja lain ditolak.

## Antrean dan pencegahan eksekusi ganda

Scheduler memeriksa jadwal yang sudah jatuh tempo secara berkala. Jadwal yang sedang diproses ditandai aktif di memori sehingga satu proses Klip tidak memasukkan jadwal yang sama ke antrean lebih dari sekali.

Antrean memakai jumlah worker kecil dan tetap lokal. Tidak ada broker atau layanan eksternal.

## Retry

Retry berlaku per jadwal dan dibatasi maksimal 10 percobaan. Jeda bawaan adalah satu detik. Setiap percobaan memiliki catatan terpisah pada `jadwal_eksekusi`.

`jadwal_eksekusi` adalah riwayat percobaan menjalankan sebuah `jadwal`. Eksekusi proses yang sebenarnya tetap dicatat sebagai `eksekusi` melalui runtime agen.

## Heartbeat

Heartbeat internal diperbarui secara berkala untuk jadwal yang sedang aktif. Nilainya hanya berada di memori proses dan tidak dianggap sebagai bukti bahwa proses agen eksternal masih hidup.

## Riwayat

Riwayat jadwal disimpan di SQLite dan dapat dibaca melalui:

`GET /api/v1/scheduler/{id}/riwayat?limit=50`

API pembuatan jadwal:

`POST /api/v1/scheduler`

Daftar jadwal:

`GET /api/v1/scheduler`

## Batasan

Versi ini belum menyediakan cron expression, kalender hari kerja, distributed lock, atau antrean lintas proses. Scheduler sengaja tetap sederhana agar sesuai dengan prinsip local-first dan hemat dependency.
