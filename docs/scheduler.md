# Scheduler Klip

Scheduler berjalan lokal di proses Klip dan memakai SQLite sebagai sumber jadwal serta riwayat.

## Jadwal

Jadwal memakai interval tetap dalam detik. Setiap jadwal menyimpan:

- pekerjaan dan tugas yang dijalankan;
- agen dan program yang digunakan;
- argumen proses;
- waktu eksekusi berikutnya;
- jumlah percobaan ulang maksimal.

Rentang interval yang diterima API adalah 1 detik sampai 7 hari.

## Antrean dan duplicate run

Scheduler melakukan pemeriksaan berkala terhadap jadwal yang sudah jatuh tempo. Jadwal yang sedang diproses ditandai aktif di memori sehingga satu proses Klip tidak memasukkan jadwal yang sama ke antrean lebih dari sekali.

Antrean memakai jumlah worker kecil dan tetap lokal. Tidak ada broker atau layanan eksternal.

## Retry

Retry berlaku per jadwal dan dibatasi maksimal 10 percobaan. Jeda bawaan adalah satu detik. Setiap percobaan memiliki catatan terpisah di `schedule_run`.

## Heartbeat

Heartbeat internal diperbarui secara berkala untuk pekerjaan scheduler yang sedang aktif. Nilainya hanya berada di memori proses dan tidak dianggap sebagai bukti bahwa proses agen eksternal masih hidup.

## Riwayat

Riwayat disimpan di SQLite dan dapat dibaca melalui:

`GET /api/v1/scheduler/{id}/riwayat?limit=50`

API pembuatan jadwal:

`POST /api/v1/scheduler`

Daftar jadwal:

`GET /api/v1/scheduler`

## Batasan

Versi ini belum menyediakan cron expression, kalender hari kerja, distributed lock, atau antrean lintas proses. Scheduler sengaja tetap sederhana agar sesuai dengan prinsip local-first dan hemat dependency.
