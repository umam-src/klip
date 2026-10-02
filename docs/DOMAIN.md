# Domain Klip

Dokumen ini menetapkan bahasa domain Klip agar desain produk tetap mandiri dan tidak menjadi salinan struktur produk lain.

## Prinsip

1. Gunakan istilah yang mudah dipahami pengguna Indonesia.
2. Pisahkan konsep produk dari detail implementasi.
3. Hindari membuat entitas hanya karena produk lain memilikinya.
4. Utamakan alur kerja lokal dan sederhana.
5. Setiap entitas harus punya alasan operasional yang jelas.

## Kosakata inti

| Istilah | Arti di Klip |
|---|---|
| **Ruang** | Lingkup kerja mandiri yang menyimpan agen, pekerjaan, pengaturan, dan hasil. |
| **Agen** | Pelaksana berbasis AI yang memiliki peran, aturan, kemampuan, dan model. |
| **Sasaran** | Hasil yang ingin dicapai oleh manusia atau agen. |
| **Pekerjaan** | Unit kerja utama yang memiliki konteks dan hasil yang diharapkan. |
| **Tugas** | Langkah konkret yang dapat dikerjakan dan dilacak secara terpisah. |
| **Alur** | Urutan atau aturan yang menghubungkan beberapa tugas. |
| **Sesi** | Konteks eksekusi percakapan atau proses kerja tertentu. |
| **Hasil** | Keluaran kerja seperti teks, berkas, laporan, atau data. |
| **Alat** | Kemampuan yang dapat dipanggil agen untuk melakukan pekerjaan di luar penalaran model. |
| **Skill** | Paket kemampuan atau instruksi yang dapat dipasang pada agen. |
| **Persetujuan** | Titik pengaman yang membutuhkan keputusan manusia sebelum langkah tertentu diteruskan. |
| **Penyedia** | Layanan atau proses yang menyediakan model AI. |
| **Model** | Model AI tertentu yang digunakan oleh penyedia. |
| **Jadwal** | Aturan waktu untuk menjalankan pekerjaan atau alur secara otomatis. |

## Hubungan inti

```text
Ruang
 ├── Agen
 │    ├── Skill
 │    └── Alat
 ├── Sasaran
 ├── Pekerjaan
 │    ├── Tugas
 │    ├── Sesi
 │    ├── Persetujuan
 │    └── Hasil
 └── Alur
      └── Tugas
```

## Keputusan desain

### Tidak ada organisasi sebagai konsep wajib

Klip dimulai dari **Ruang**, bukan hierarki organisasi. Pengguna tunggal dapat menjalankan Klip tanpa membuat organisasi, tim, atau struktur administratif tambahan.

Jika kebutuhan multi-pengguna berkembang, akses dapat ditambahkan tanpa menjadikan struktur organisasi sebagai syarat penggunaan dasar.

### Pekerjaan dan tugas berbeda

**Pekerjaan** menjawab "apa yang sedang ingin diselesaikan". **Tugas** menjawab "langkah konkret apa yang perlu dilakukan". Satu pekerjaan dapat memiliki satu atau banyak tugas.

### Sesi bukan pekerjaan

Sesi hanya menyimpan konteks eksekusi. Pekerjaan tetap menjadi objek yang dapat dilacak meskipun memiliki beberapa sesi atau percobaan.

### Hasil adalah objek kelas satu

Hasil tidak hanya dianggap sebagai log. Berkas, teks, data, dan keluaran lain dapat ditautkan ke pekerjaan dan digunakan kembali oleh tugas berikutnya.

## Bentuk data awal

Implementasi awal cukup menggunakan entitas berikut:

- `ruang`
- `agen`
- `sasaran`
- `pekerjaan`
- `tugas`
- `alur`
- `sesi`
- `hasil`
- `skill`
- `alat`
- `persetujuan`
- `penyedia_ai`
- `model_ai`
- `jadwal`
- `peristiwa`

Belum perlu membuat tabel khusus untuk fitur yang belum memiliki kebutuhan nyata.

## Status awal

Status umum pekerjaan dan tugas:

- `draft`
- `ready`
- `running`
- `waiting`
- `blocked`
- `completed`
- `failed`
- `cancelled`

Status harus diperlakukan sebagai bagian dari aturan domain, bukan sekadar label tampilan.

## Batasan v0.1

v0.1 hanya membutuhkan **Ruang, Agen, Pekerjaan, Tugas, Sesi, dan Hasil** sebagai fondasi. Sasaran, Alur, Persetujuan, Skill, Alat, dan Jadwal dapat hadir sebagai struktur data setelah fondasi stabil.
