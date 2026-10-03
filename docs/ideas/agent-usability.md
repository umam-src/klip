# Rancangan: Usability Agen

> **Status: Rancangan — belum disetujui.**
>
> Dokumen ini menyimpan arah desain pengalaman pengguna untuk Agen agar keputusan yang sedang dipikirkan tidak hilang dari percakapan. Dokumen ini belum menjadi spesifikasi implementasi, keputusan arsitektur, atau pekerjaan roadmap.

## Tujuan

Membuat konsep **Agen** mudah dipahami sebelum Klip menambah fitur kerja yang lebih kompleks.

Fokus awal adalah pertanyaan sederhana:

> "Siapa yang mengerjakan pekerjaan ini, apa perannya, dan AI apa yang digunakannya?"

Agen harus terasa sebagai pelaksana kerja, bukan sekadar pilihan model AI.

## Bahasa produk

Kosakata domain yang sudah ada tetap menjadi acuan:

- **Ruang** — lingkup kerja mandiri.
- **Agen** — pelaksana berbasis AI.
- **Sasaran** — hasil yang ingin dicapai.
- **Pekerjaan** — unit kerja utama.
- **Tugas** — langkah konkret yang dapat dikerjakan dan dilacak.
- **Aktivitas** — cara pengguna melihat kejadian dan perkembangan kerja.

Untuk antarmuka, istilah **Ruang kerja** dapat digunakan bila terasa lebih jelas bagi pengguna. Ini hanya usulan bahasa antarmuka; istilah domain resmi tetap **Ruang** sampai ada keputusan lain.

## Mental model

```text
Klip
└── Ruang kerja
    ├── Agen
    │   ├── Pemimpin
    │   ├── Peneliti Produk
    │   └── Pengembang
    ├── Sasaran
    ├── Pekerjaan
    │   ├── Tugas
    │   └── Subtugas
    └── Aktivitas
```

Urutan pengalaman yang diutamakan:

1. Pengguna masuk ke Ruang kerja.
2. Pengguna mengenali Agen yang tersedia.
3. Pengguna menetapkan Sasaran.
4. Sasaran menghasilkan atau mengelompokkan Pekerjaan.
5. Pekerjaan dipecah menjadi Tugas bila diperlukan.
6. Agen mengerjakan Tugas.
7. Pengguna melihat Hasil dan Aktivitas.

Urutan ini adalah model pengalaman, bukan keputusan bahwa semua entitas harus dibuat atau ditampilkan pada satu layar.

## Prinsip desain Agen

### Agen bukan model

Agen adalah identitas pelaksana kerja yang memiliki peran dan konteks. Model AI adalah salah satu bagian dari konfigurasi Agen.

Contoh:

```text
Agen: Peneliti Produk
Peran: Peneliti
AI: llama.cpp rumah
Model: qwen2.5-0.5b-instruct-q4_k_m
```

Pengguna seharusnya dapat memahami fungsi Agen tanpa harus mengetahui detail model terlebih dahulu.

### Agen memiliki peran yang jelas

Nama dan peran harus menjawab:

- Agen ini siapa?
- Agen ini mengerjakan apa?
- Agen ini bekerja di bawah siapa?

Jangan menjadikan nama model sebagai nama utama Agen.

### Hubungan agen tetap sederhana

Hierarki Agen mengikuti aturan domain yang sudah ada:

```text
Pemimpin
├── Peneliti Produk
└── Pengembang
```

Agen anak berada pada Ruang yang sama dengan Agen induknya. Hubungan ini membantu pengguna memahami pembagian pekerjaan tanpa memaksa Klip memiliki konsep organisasi tambahan.

## Bentuk halaman Agen

### Daftar Agen

Halaman awal sebaiknya langsung memperlihatkan siapa saja yang tersedia.

```text
Agen                              [ + Tambah agen ]

Pemimpin
  Mengarahkan pekerjaan utama

Peneliti Produk
  Mencari dan merangkum informasi

Pengembang
  Mengerjakan perubahan teknis
```

Informasi teknis seperti URL penyedia atau nama model tidak perlu menjadi informasi utama pada daftar.

### Detail Agen

```text
Peneliti Produk
Peneliti
Mencari dan merangkum informasi.

AI
llama.cpp rumah
qwen2.5-0.5b-instruct-q4_k_m

Melapor kepada
Pemimpin

Pekerjaan aktif
2

Aktivitas terakhir
Menunggu tugas berikutnya
```

Halaman detail harus menjawab fungsi Agen terlebih dahulu, baru konfigurasi dan statusnya.

## Formulir "Tambah agen"

Versi awal sebaiknya hanya meminta informasi yang benar-benar diperlukan:

```text
Nama              [ Peneliti Produk ]
Peran             [ Peneliti ]
Deskripsi         [ Mencari dan merangkum informasi. ]
Penyedia AI       [ llama.cpp rumah v ]
Model             [ qwen2.5-0.5b-instruct-q4_k_m v ]
Atasan            [ Pemimpin v ]

                         [ Batal ] [ Simpan agen ]
```

### Urutan pengisian

1. Nama.
2. Peran.
3. Deskripsi.
4. Penyedia AI.
5. Model.
6. Atasan.

Urutan tersebut mengikuti cara pengguna memahami Agen: identitas → fungsi → kemampuan AI → hubungan kerja.

### Hal yang tidak perlu ditampilkan di form awal

- URL endpoint mentah jika sudah tersedia melalui Pengaturan.
- Parameter teknis model yang tidak diperlukan untuk memilih model.
- Detail penyimpanan.
- Pengaturan internal eksekusi.
- Konfigurasi lanjutan sebelum ada kebutuhan nyata.

## Pemilihan AI

Agen memilih **Penyedia** dan **Model** sebagai pasangan penggunaan AI.

```text
Penyedia AI  [ llama.cpp rumah ]
Model        [ qwen2.5-0.5b-instruct-q4_k_m ]
```

Model yang ditampilkan sebaiknya berasal dari daftar model yang tersedia pada penyedia bila discovery didukung. Nama atau alias buatan Klip tidak diperlukan hanya untuk membuat model terlihat lebih ramah.

Detail konfigurasi penyedia tetap berada di Pengaturan. Agen hanya perlu memilih koneksi yang tersedia.

Rancangan beberapa konfigurasi penyedia AI disimpan terpisah dalam `docs/ideas/ai-multi-provider.md` dan belum disetujui.

## Status Agen

Status yang terlihat pengguna sebaiknya menjawab keadaan kerja, bukan detail proses internal.

Contoh bahasa:

- Siap
- Mengerjakan tugas
- Menunggu
- Berhenti
- Bermasalah

Daftar status final belum ditetapkan dalam dokumen ini. Status Agen perlu dibedakan dari status Pekerjaan dan Tugas.

## Aktivitas

Aktivitas dipakai untuk menjawab "apa yang baru saja terjadi?".

Contoh:

```text
10:42  Peneliti Produk mulai mengerjakan "Riset pasar lokal"
10:47  Peneliti Produk menghasilkan ringkasan
10:48  Tugas menunggu pemeriksaan
```

Aktivitas tidak menggantikan Hasil. Hasil adalah keluaran kerja; Aktivitas adalah jejak perkembangan kerja.

## Alur penggunaan yang diharapkan

Skenario sederhana:

```text
Buat Ruang kerja
      ↓
Buat Agen
      ↓
Pilih Penyedia AI + Model
      ↓
Buat Sasaran
      ↓
Buat Pekerjaan
      ↓
Buat Tugas
      ↓
Tugaskan ke Agen
      ↓
Agen bekerja
      ↓
Lihat Aktivitas
      ↓
Lihat Hasil
```

Pengguna tidak harus memahami seluruh struktur tersebut untuk menjalankan tugas sederhana. Antarmuka sebaiknya memperlihatkan langkah berikutnya berdasarkan konteks.

## Batasan usability awal

Rancangan ini sengaja tidak memasukkan:

- marketplace Agen;
- karakter atau persona kompleks;
- konfigurasi model yang sangat rinci;
- sistem organisasi atau tim sebagai konsep wajib;
- izin pengguna yang kompleks;
- plugin sebagai prasyarat Agen;
- beberapa Agen bekerja paralel sebagai syarat awal;
- otomasi lanjutan sebelum alur dasar dapat dipahami.

Tujuannya adalah memastikan alur **Agen → Pekerjaan → Tugas → Hasil** terasa jelas terlebih dahulu.

## Pertanyaan yang belum diputuskan

- Apakah halaman Agen menjadi halaman utama atau bagian dari Ruang.
- Apakah Agen pertama dibuat saat Ruang pertama dibuat atau melalui langkah terpisah.
- Apakah setiap Ruang wajib memiliki satu Agen pemimpin.
- Apakah Agen boleh tidak memiliki atasan.
- Apakah Penyedia dan Model ditampilkan di daftar Agen atau hanya di detail.
- Bagaimana menampilkan Agen yang modelnya sudah tidak tersedia.
- Apakah Agen dapat diganti Penyedia atau Model setelah memiliki pekerjaan aktif.
- Bagaimana pengguna menghentikan Agen yang sedang bekerja.
- Status Agen mana yang benar-benar dibutuhkan oleh alur kerja awal.
- Kapan Sasaran perlu diperkenalkan dibanding langsung membuat Pekerjaan.

## Kriteria untuk implementasi nanti

Implementasi baru dipertimbangkan setelah rancangan ini disetujui dan minimal memenuhi prinsip berikut:

- Pengguna dapat mengenali fungsi Agen tanpa membaca dokumentasi teknis.
- Membuat Agen tidak membutuhkan konfigurasi yang tidak relevan.
- Pemilihan Penyedia dan Model jelas tetapi tidak mendominasi identitas Agen.
- Hubungan Agen induk dan anak mudah dipahami.
- Pekerjaan dan Tugas tetap terpisah secara konseptual.
- Aktivitas dan Hasil memiliki fungsi yang berbeda.
- Perubahan UX tidak merusak data dan alur Agen yang sudah ada.
- Kompleksitas tambahan dibuktikan oleh kebutuhan penggunaan nyata.

## Catatan keputusan

**Belum disetujui.** Jangan menganggap isi dokumen ini sebagai kontrak desain atau tiket implementasi sampai statusnya diubah secara eksplisit.
