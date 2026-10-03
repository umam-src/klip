# Progres Assignment Agen ke Tugas

Status: **Runtime terintegrasi — kontrak UI assignment teruji otomatis; verifikasi browser penuh masih tersisa.**

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
- Kontrak UI assignment diuji dari aset web yang di-embed agar endpoint dan elemen utama tidak terlepas saat refactor.

### Runtime

- `Executor.Execute` memeriksa assignment `Tugas → Agen` ketika `TugasID` diberikan.
- Jika tugas sudah memiliki assignment, Agen yang menjalankan harus sama dengan pelaksana yang tersimpan.
- Jika Agen tidak cocok, eksekusi dihentikan sebelum run dibuat.
- Assignment yang belum ada tetap kompatibel dengan eksekusi eksplisit lama.
- Test runtime mencakup assignment yang cocok dan tidak cocok.

### Belum selesai

- Verifikasi UI end-to-end nyata di browser.
