# Progress UI Agen

Status: Implementasi bertahap — bukan spesifikasi UI final.

## Tahap 4 — Daftar Agen

Daftar Agen dalam Ruang kerja sekarang menampilkan informasi yang membantu pengguna mengenali pelaksana kerja tanpa membuka detail teknis:

- nama Agen;
- peran;
- deskripsi singkat bila tersedia;
- status Aktif/Nonaktif;
- hubungan induk sederhana melalui indentasi;
- tombol `Tambah agen` yang terlihat di panel Agen;
- klik item membuka detail Agen.

Daftar tidak menampilkan token, credential, atau alamat provider mentah. Provider/model tetap tersedia pada detail Agen karena informasi tersebut membantu memahami konfigurasi Agen tanpa menjadikannya fokus daftar.

Hierarki dirender dari `parent_id` yang sudah disimpan server. Jika data lama atau data rusak mengandung siklus, renderer tetap berhenti pada Agen yang sudah dikunjungi sehingga UI tidak masuk loop.
