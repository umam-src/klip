# Progres Assignment Agen ke Tugas

Status: **UI assignment selesai — verifikasi end-to-end masih tersisa.**

## Tahap 7 — Hubungkan Agen dengan pekerjaan

Assignment menggunakan relasi `tugas_agen` yang dipisahkan dari identitas Agen dan struktur Tugas.

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

### API yang tersedia

- `GET /api/v1/tugas/{id}/agen` untuk membaca pelaksana.
- `PUT /api/v1/tugas/{id}/agen` untuk menetapkan atau mengganti pelaksana.
- `DELETE /api/v1/tugas/{id}/agen` untuk melepas pelaksana.

### UI yang tersedia

- Form Tugas menyediakan pilihan Agen pelaksana dari Ruang aktif.
- Tugas baru langsung menyimpan assignment bila Agen dipilih.
- Daftar Tugas menampilkan pelaksana setelah assignment tersedia.
- Detail Tugas menampilkan Agen pelaksana.
- Detail Agen menampilkan Tugas yang ditugaskan kepadanya.

### Belum selesai

- Integrasi assignment dengan runtime.
- Verifikasi UI end-to-end di browser.
