# Progres Assignment Agen ke Tugas

Status: **Fondasi storage — implementasi bertahap.**

## Tahap 7 — Hubungkan Agen dengan pekerjaan

Fondasi awal assignment menggunakan relasi `tugas_agen` yang dipisahkan dari identitas Agen dan struktur Tugas.

### Aturan

- Satu Tugas memiliki paling banyak satu Agen pelaksana pada tahap awal.
- Assignment tidak mengubah hierarki Agen.
- Agen dan Pekerjaan Tugas harus berada dalam Ruang yang sama.
- Assignment dapat diganti atau dilepas tanpa mengubah status Agen maupun status Tugas.
- Data assignment menyimpan waktu penugasan.
- Credential dan konfigurasi AI tetap berada di luar relasi assignment.

### Storage yang tersedia

- `AssignTugasToAgen` untuk menetapkan atau mengganti pelaksana.
- `UnassignTugas` untuk melepas pelaksana.
- `GetTugasAssignment` untuk membaca assignment satu Tugas.
- `ListTugasByAgen` untuk menampilkan Tugas yang dikerjakan Agen.
- Pembuatan tabel assignment bersifat idempotent agar database lama tetap dapat digunakan.

### Belum selesai

- Endpoint HTTP assignment.
- Pilihan Agen pada form Tugas.
- Detail Agen menampilkan Tugas terkait.
- Detail Tugas menampilkan pelaksana.
- Integrasi assignment dengan runtime.
- Verifikasi UI end-to-end.
