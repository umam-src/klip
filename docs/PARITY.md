# Matriks Inspirasi dan Independensi

Dokumen ini menjaga agar Klip mengambil **kebutuhan produk**, bukan menyalin implementasi, struktur kode, atau tampilan dari proyek lain.

| Kebutuhan | Klip | Prinsip independensi |
|---|---|---|
| Lingkup kerja | Ruang Kerja | Tidak memakai hierarki organisasi sebagai syarat dasar |
| Pelaksana AI | Agen | Model dan perilaku dipisahkan dari pekerjaan |
| Tujuan | Goal | Pusat arah kerja; dapat berjenjang dalam satu Ruang Kerja |
| Unit kerja | Proyek | Wadah kerja yang mengarah ke satu Goal |
| Langkah | Tugas | Fokus pada tindakan konkret |
| Urutan kerja | Alur | Model sederhana dan dapat diperluas |
| Eksekusi | Eksekusi | Tidak disamakan dengan Proyek atau Tugas |
| Keluaran | Hasil Kerja | Objek kelas satu |
| Kemampuan | Skill / Alat | Dipasang sesuai kebutuhan agen |
| Pengaman | Persetujuan | Hanya digunakan pada titik yang memang membutuhkan manusia |
| AI | Penyedia + Model | Provider-agnostic, lokal lebih dulu |

## Aturan anti-jiplak

- Jangan menyalin source code upstream.
- Jangan menyalin struktur direktori hanya karena namanya sama.
- Jangan menyalin desain UI, teks, ikon, atau komponen visual secara langsung.
- Jangan membuat domain model 1:1 tanpa kebutuhan Klip sendiri.
- Jangan menganggap setiap fitur upstream harus masuk Klip.
- Jika sebuah fitur tidak mendukung tujuan ringan, lokal, atau sederhana, fitur tersebut harus dipertimbangkan ulang.

## Cara melakukan perbandingan

Perbandingan dengan proyek inspirasi hanya digunakan untuk menemukan kebutuhan yang mungkin terlewat. Setelah ditemukan, fitur dirancang ulang berdasarkan prinsip Klip, lalu diuji dari sisi:

1. kebutuhan pengguna;
2. kesederhanaan;
3. ukuran runtime;
4. offline-first;
5. biaya operasional;
6. keamanan;
7. kemudahan pemeliharaan.

## Catatan lisensi

Jika pada masa depan Klip menggunakan kode pihak ketiga secara langsung, lisensi dan atribusi komponen tersebut harus diperiksa dan dipertahankan sesuai ketentuannya. Inspirasi konsep tidak sama dengan menyalin kode.
