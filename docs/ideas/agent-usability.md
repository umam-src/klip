# Rancangan: Usability Agen

> **Arsip.** Ditulis sebelum remodel model data v1. Istilah lama dipetakan ke istilah yang berlaku: `Sasaran` menjadi Goal dan `Pekerjaan` menjadi Proyek. Rujukan yang berlaku: [DOMAIN.md](../DOMAIN.md) dan [data-model.md](../data-model.md).

> **Status: Rancangan — belum disetujui.**
>
> Dokumen ini menyimpan arah desain pengalaman pengguna untuk Agen agar keputusan yang sedang dipikirkan tidak hilang dari percakapan. Dokumen ini belum menjadi spesifikasi implementasi, keputusan arsitektur, atau pekerjaan roadmap.
>
> Rancangan pada dokumen ini diperbarui berdasarkan draf visual alur pengembangan Agen. Draf tersebut dipakai sebagai alat berpikir tentang urutan pengalaman pengguna, bukan sebagai kewajiban bahwa seluruh langkah harus muncul dalam satu wizard.

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
- **Hasil** — keluaran kerja yang dapat dilihat atau digunakan kembali.

Untuk antarmuka, istilah **Ruang kerja** dapat digunakan bila terasa lebih jelas bagi pengguna. Ini hanya usulan bahasa antarmuka; istilah domain resmi tetap **Ruang** sampai ada keputusan lain.

Untuk rancangan AI yang lebih lengkap, istilah **Penyedia AI**, **Akun AI**, dan **Model** dipakai untuk membedakan layanan, konfigurasi akses, dan model. Hubungan beberapa penyedia dan akun belum menjadi fitur implementasi pada tahap Agen ini.

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

Agen berada di dalam Ruang kerja dan menjadi pelaksana pekerjaan. Sasaran memberi arah, Pekerjaan mengelompokkan hasil kerja, Tugas memecah langkah konkret, sedangkan Aktivitas dan Hasil membantu pengguna memahami perkembangan dan keluaran.

## Gambaran draf alur pengembangan

Draf visual menggunakan alur berikut sebagai gambaran produk secara keseluruhan:

```text
1. Buat Ruang kerja
        ↓
2. Siapkan Penyedia AI + Akun AI
        ↓
3. Siapkan Model
        ↓
4. Buat Agen
        ↓
5. Tetapkan Sasaran
        ↓
6. Buat Pekerjaan + Tugas
        ↓
7. Pantau progres, Aktivitas, dan Hasil
```

Langkah tersebut menggambarkan **alur pemahaman pengguna**, bukan berarti Klip harus memaksa pengguna melewati wizard tujuh langkah. Pengguna dapat masuk langsung ke konteks yang sudah tersedia dan hanya melakukan langkah yang dibutuhkan.

Urutan yang paling penting untuk fondasi Agen adalah:

```text
Ruang kerja
   ↓
Agen
   ↓
Identitas + peran + deskripsi
   ↓
Konfigurasi AI + model
   ↓
Atasan (bila ada)
   ↓
Pekerjaan + Tugas
   ↓
Aktivitas + Hasil
```

## Prinsip desain Agen

### Agen bukan model

Agen adalah identitas pelaksana kerja yang memiliki peran dan konteks. Model AI adalah salah satu bagian dari konfigurasi Agen.

Contoh:

```text
Agen: Peneliti Produk
Peran: Peneliti
Deskripsi: Mencari dan merangkum informasi untuk tim.
Akun AI: Claude Tim
Model: claude-sonnet-...
Atasan: Pemimpin
```

Pengguna seharusnya dapat memahami fungsi Agen tanpa harus mengetahui detail model terlebih dahulu.

### Agen menggunakan Akun AI, bukan menyimpan token sendiri

Kredensial AI sebaiknya berada pada **Akun/Connection AI** yang dikelola di Pengaturan. Agen hanya memilih akun yang akan digunakannya.

Satu Akun AI dapat digunakan oleh beberapa Agen:

```text
Akun Claude Tim
└── credential
    ├── Peneliti Produk
    ├── Reviewer
    └── Penulis
```

Beberapa Akun AI juga dapat berasal dari Penyedia yang sama:

```text
Anthropic
├── Claude Tim
│   ├── Peneliti Produk
│   └── Reviewer
│
└── Claude Pribadi
    └── Pengembang
```

Dengan pola ini, token/API key tidak perlu dimasukkan berulang kali pada setiap Agen. Pemisahan akun dapat digunakan bila kebutuhan akses, kuota, billing, atau identitas memang berbeda.

### Agen memiliki peran yang jelas

Nama dan peran harus menjawab:

- Agen ini siapa?
- Agen ini mengerjakan apa?
- Agen ini bekerja di bawah siapa?

Jangan menjadikan nama model atau nama credential sebagai nama utama Agen.

### Hubungan agen tetap sederhana

Hierarki Agen mengikuti aturan domain yang sudah ada:

```text
Pemimpin
├── Peneliti Produk
└── Pengembang
```

Agen anak berada pada Ruang yang sama dengan Agen induknya. Hubungan ini membantu pengguna memahami pembagian pekerjaan tanpa memaksa Klip memiliki konsep organisasi tambahan.

## Model data target

Draf visual memperlihatkan hubungan konseptual berikut:

```text
Ruang kerja
   │
   ├── Sasaran
   │     │
   │     └── Pekerjaan
   │            │
   │            └── Tugas ────── Agen
   │                              │
   │                              ├── Peran
   │                              ├── Deskripsi
   │                              ├── Atasan
   │                              ├── Akun AI
   │                              └── Model
   │
   └── Aktivitas / Hasil

Penyedia AI
   ↓
Akun AI
   ↓
Model
   ↓
Agen
```

Hubungan ini adalah **model target**, bukan skema SQLite final. Tahap implementasi Agen tetap harus mengikuti struktur data yang sudah ada dan hanya menambah entitas baru bila kebutuhan nyata sudah terbukti.

### Relasi AI yang diinginkan

```text
Penyedia AI
├── Akun AI A
│   ├── Agen Peneliti
│   └── Agen Reviewer
└── Akun AI B
    └── Agen Pengembang
```

Satu Akun AI dapat digunakan oleh banyak Agen. Satu Penyedia AI dapat memiliki beberapa Akun AI. Model dipilih untuk penggunaan Agen melalui Akun AI tersebut.

Rancangan beberapa Penyedia dan Akun AI disimpan terpisah dalam `docs/ideas/ai-multi-provider.md` dan belum disetujui sebagai fitur.

## Bentuk halaman Agen

### Daftar Agen

Halaman awal sebaiknya langsung memperlihatkan siapa saja yang tersedia.

```text
Agen                                      [ + Tambah agen ]

┌─────────────────────┐  ┌─────────────────────┐
│ Pemimpin            │  │ Peneliti Produk     │
│ Pemimpin            │  │ Peneliti            │
│                     │  │                     │
│ Mengarahkan kerja   │  │ Mencari informasi   │
│ utama.              │  │ untuk tim.          │
│                     │  │                     │
│ ● Aktif             │  │ ● Aktif             │
│ 1 akun · 1 model    │  │ 1 akun · 1 model    │
└─────────────────────┘  └─────────────────────┘
```

Informasi teknis seperti URL penyedia, credential, atau parameter model tidak perlu menjadi informasi utama pada daftar. Jika ringkasan AI ditampilkan, cukup tampilkan hubungan seperti **1 akun · 1 model** atau informasi setara yang mudah dipahami.

Kartu Agen sebaiknya menonjolkan:

- nama;
- peran;
- deskripsi singkat;
- status kerja;
- ringkasan konfigurasi AI seperlunya.

### Detail Agen

Detail Agen diarahkan menjadi pusat konteks untuk satu pelaksana kerja.

```text
← Kembali

[ikon] Pengembang                         ● Aktif
       Pengembang                         [ Edit ] [ ⋮ ]

Ringkasan | Pekerjaan | Aktivitas | Pengaturan

Informasi dasar              Ringkasan aktivitas
Nama       Pengembang        Pekerjaan aktif       2
Peran      Pengembang        Tugas selesai (7 hari) 5
Deskripsi  ...               Terakhir aktif        ...

Agen ini melapor kepada     Penggunaan AI
Pemimpin                     ...

AI yang digunakan            Akun AI
Penyedia  ...                Claude Tim
Akun      ...
Model     ...
```

Halaman detail harus menjawab fungsi Agen terlebih dahulu, baru konfigurasi dan statusnya.

Tab awal yang dirancang:

- **Ringkasan** — identitas, peran, hubungan, AI yang digunakan, status, dan ringkasan aktivitas.
- **Pekerjaan** — pekerjaan dan tugas yang berkaitan dengan Agen.
- **Aktivitas** — jejak kejadian dan perkembangan kerja.
- **Pengaturan** — perubahan konfigurasi Agen yang memang relevan bagi pengguna.

Tab tersebut adalah rancangan informasi, bukan kewajiban implementasi awal. Bila jumlah informasi masih kecil, Klip boleh menyederhanakannya menjadi satu halaman.

## Formulir "Tambah agen" dan "Edit agen"

Draf visual menetapkan satu bentuk form yang dapat dipakai untuk tambah dan edit.

```text
Tambah agen

Nama        * [ Peneliti Produk                  ]
Peran       * [ Peneliti                       v ]
Deskripsi     [ Mencari dan merangkum informasi ]
              [ untuk tim.                      ]

Agen ini melapor kepada
              [ Pemimpin                       v ]

AI yang digunakan
Penyedia     * [ Anthropic                      v ]
Akun         * [ Claude Tim                     v ]
Model          [ claude-sonnet-...              v ]

                         [ Batal ] [ Simpan ]
```

Field inti:

1. **Nama** — identitas yang mudah dikenali.
2. **Peran** — fungsi kerja Agen.
3. **Deskripsi** — penjelasan singkat tentang tanggung jawab.
4. **Akun AI** — konfigurasi akses AI yang dipilih.
5. **Model** — model yang digunakan Agen.
6. **Atasan** — Agen induk bila memang diperlukan.

Penyedia dapat ditampilkan sebagai konteks pilihan Akun pada rancangan multi-provider. Pada tahap implementasi saat ini, jangan membuat field Penyedia baru bila arsitektur belum membutuhkannya.

### Urutan pengisian

1. Nama.
2. Peran.
3. Deskripsi.
4. Atasan bila ada.
5. Penyedia/Akun AI sesuai konfigurasi yang tersedia.
6. Model.

Urutan tersebut mengikuti cara pengguna memahami Agen: identitas → fungsi → hubungan kerja → kemampuan AI.

### Validasi form

Form perlu memberi umpan balik langsung untuk:

- nama kosong;
- peran kosong;
- deskripsi terlalu panjang;
- Akun AI yang tidak tersedia atau nonaktif;
- model yang tidak tersedia;
- atasan yang tidak valid;
- hubungan atasan yang menyebabkan siklus.

Validasi harus tetap dilakukan di server. Validasi antarmuka hanya membantu pengguna menemukan kesalahan lebih cepat.

### Hal yang tidak perlu ditampilkan di form awal

- URL endpoint mentah jika sudah tersedia melalui Pengaturan.
- Token/API key atau credential.
- Parameter teknis model yang tidak diperlukan untuk memilih model.
- Detail penyimpanan.
- Pengaturan internal eksekusi.
- Konfigurasi lanjutan sebelum ada kebutuhan nyata.

## Pemilihan AI

Agen memilih **Akun AI** dan **Model** sebagai pasangan penggunaan AI.

```text
Akun AI  [ Claude Tim ]
Model    [ claude-sonnet-... ]
```

Akun tersebut mengarah ke Penyedia AI dan menyimpan konfigurasi aksesnya. Beberapa Agen dapat memilih Akun yang sama.

Model yang ditampilkan sebaiknya berasal dari daftar model yang tersedia pada penyedia bila discovery didukung. Nama atau alias buatan Klip tidak diperlukan hanya untuk membuat model terlihat lebih ramah.

Detail konfigurasi Penyedia dan Akun tetap berada di Pengaturan. Agen hanya perlu memilih Akun yang tersedia.

## Sasaran, Pekerjaan, dan Tugas

Agen tidak berdiri sendiri. Setelah Agen tersedia, pengguna harus dapat menghubungkannya dengan pekerjaan nyata.

```text
Sasaran
   ↓
Pekerjaan
   ↓
Tugas
   ↓
Agen mengerjakan
   ↓
Hasil + Aktivitas
```

Peran masing-masing:

- **Sasaran** menetapkan hasil utama yang ingin dicapai.
- **Pekerjaan** mengelompokkan pekerjaan yang menuju Sasaran.
- **Tugas** adalah unit konkret yang dapat ditugaskan dan dilacak.
- **Agen** menjadi pelaksana Tugas.
- **Hasil** menyimpan keluaran kerja.
- **Aktivitas** menjelaskan perkembangan dan kejadian penting.

Pengguna tidak harus membuat semua tingkat tersebut untuk pekerjaan sederhana. Antarmuka sebaiknya menyediakan jalan singkat dari Agen ke Pekerjaan atau Tugas ketika konteksnya sudah jelas.

## Status Agen

Status yang terlihat pengguna sebaiknya menjawab keadaan kerja, bukan detail proses internal.

Contoh bahasa:

- Aktif / Siap
- Mengerjakan tugas
- Menunggu
- Berhenti
- Bermasalah

Draf visual menggunakan label **Aktif** sebagai contoh tampilan. Status domain final belum ditetapkan dalam dokumen ini dan perlu dibedakan dari status Pekerjaan dan Tugas.

## Aktivitas dan Hasil

Aktivitas dipakai untuk menjawab "apa yang baru saja terjadi?".

Contoh:

```text
10:42  Peneliti Produk mulai mengerjakan "Riset pasar lokal"
10:47  Peneliti Produk menghasilkan ringkasan
10:48  Tugas menunggu pemeriksaan
```

Aktivitas tidak menggantikan Hasil. Hasil adalah keluaran kerja; Aktivitas adalah jejak perkembangan kerja.

Pada halaman detail Agen, ringkasan aktivitas dapat menampilkan jumlah pekerjaan aktif, tugas selesai, waktu aktif terakhir, dan kejadian penting. Angka-angka tersebut hanya ditampilkan bila datanya memang tersedia dan tidak menambah beban penyimpanan yang tidak perlu.

## Alur penggunaan yang diharapkan

Skenario lengkap:

```text
Buat Ruang kerja
      ↓
Siapkan Penyedia AI + Akun AI
      ↓
Siapkan Model
      ↓
Buat Agen
      ↓
Tetapkan Sasaran
      ↓
Buat Pekerjaan
      ↓
Buat / pecah menjadi Tugas
      ↓
Tugaskan ke Agen
      ↓
Agen bekerja
      ↓
Pantau Aktivitas
      ↓
Lihat Hasil
```

Skenario cepat untuk pekerjaan sederhana:

```text
Ruang kerja
   ↓
Agen
   ↓
Buat Tugas
   ↓
Agen bekerja
   ↓
Hasil
```

Pengguna tidak harus memahami seluruh struktur tersebut untuk menjalankan tugas sederhana. Antarmuka sebaiknya memperlihatkan langkah berikutnya berdasarkan konteks.

## Prinsip desain visual

Draf visual memperkuat prinsip berikut:

- **Sederhana di awal** — pengguna memahami Agen sebelum melihat detail AI.
- **Fokus pada kebutuhan pengguna** — konfigurasi teknis muncul ketika memang diperlukan.
- **Konsisten** — daftar, detail, form, Pengaturan, dan halaman kerja menggunakan istilah yang sama.
- **Aman untuk kredensial** — token/API key tidak menjadi bagian dari data Agen.
- **Terhubung dengan pekerjaan** — Agen selalu punya jalan yang jelas menuju Sasaran, Pekerjaan, Tugas, Aktivitas, dan Hasil.
- **Fleksibel untuk berkembang** — rancangan tidak mengunci Agen pada satu token atau satu Penyedia selamanya.

## Batasan usability awal

Rancangan ini sengaja tidak memasukkan:

- marketplace Agen;
- karakter atau persona kompleks;
- konfigurasi model yang sangat rinci;
- sistem organisasi atau tim sebagai konsep wajib;
- izin pengguna yang kompleks;
- plugin sebagai prasyarat Agen;
- beberapa Agen bekerja paralel sebagai syarat awal;
- otomasi lanjutan sebelum alur dasar dapat dipahami;
- multi-provider sebagai pekerjaan implementasi tahap Agen saat ini;
- routing model otomatis;
- pemisahan billing atau kuota sebagai fitur Agen.

Tujuannya adalah memastikan alur **Agen → Pekerjaan → Tugas → Hasil** terasa jelas terlebih dahulu.

## Pertanyaan yang belum diputuskan

- Apakah halaman Agen menjadi halaman utama atau bagian dari Ruang.
- Apakah Agen pertama dibuat saat Ruang pertama dibuat atau melalui langkah terpisah.
- Apakah setiap Ruang wajib memiliki satu Agen pemimpin.
- Apakah Agen boleh tidak memiliki atasan.
- Apakah Akun AI dan Model ditampilkan di daftar Agen atau hanya di detail.
- Apakah tab Ringkasan, Pekerjaan, Aktivitas, dan Pengaturan perlu dipisahkan sejak versi awal.
- Bagaimana menampilkan Agen yang modelnya sudah tidak tersedia.
- Apakah Agen dapat diganti Akun atau Model setelah memiliki pekerjaan aktif.
- Bagaimana pengguna menghentikan Agen yang sedang bekerja.
- Status Agen mana yang benar-benar dibutuhkan oleh alur kerja awal.
- Kapan Sasaran perlu diperkenalkan dibanding langsung membuat Pekerjaan.
- Apakah penyedia perlu dipilih langsung pada form Agen atau cukup melalui Akun AI.

## Kriteria untuk implementasi nanti

Implementasi baru dipertimbangkan setelah rancangan ini disetujui dan minimal memenuhi prinsip berikut:

- Pengguna dapat mengenali fungsi Agen tanpa membaca dokumentasi teknis.
- Membuat Agen tidak membutuhkan konfigurasi yang tidak relevan.
- Pemilihan Akun AI dan Model jelas tetapi tidak mendominasi identitas Agen.
- Token/API key tidak perlu dimasukkan berulang kali untuk setiap Agen.
- Satu Akun AI dapat digunakan oleh beberapa Agen.
- Hubungan Agen induk dan anak mudah dipahami.
- Pekerjaan dan Tugas tetap terpisah secara konseptual.
- Aktivitas dan Hasil memiliki fungsi yang berbeda.
- Pengguna dapat melihat pekerjaan Agen dari halaman detail tanpa mencari ke seluruh aplikasi.
- Perubahan UX tidak merusak data dan alur Agen yang sudah ada.
- Kompleksitas tambahan dibuktikan oleh kebutuhan penggunaan nyata.

## Catatan keputusan

**Belum disetujui.** Jangan menganggap isi dokumen ini sebagai kontrak desain atau tiket implementasi sampai statusnya diubah secara eksplisit.
