# Domain Klip

Dokumen ini menetapkan bahasa domain Klip agar desain produk tetap mandiri dan tidak menjadi salinan struktur produk lain.

## Prinsip

1. Gunakan istilah yang mudah dipahami pengguna Indonesia.
2. Pisahkan konsep produk dari detail implementasi.
3. Hindari membuat entitas hanya karena produk lain memilikinya.
4. Utamakan alur kerja lokal dan sederhana.
5. Setiap entitas harus punya alasan operasional yang jelas.
6. Jangan mengunci Agen pada satu model, token, atau bentuk penyedia tertentu bila kebutuhan produk nantinya berkembang.

## Kosakata inti

| Istilah | Arti di Klip |
|---|---|
| **Ruang** | Lingkup kerja mandiri yang menyimpan agen, pekerjaan, pengaturan, dan hasil. |
| **Agen** | Pelaksana berbasis AI yang memiliki identitas kerja, peran, deskripsi, konteks, dan konfigurasi AI. Agen dapat berada di bawah agen induk dalam ruang yang sama. |
| **Sasaran** | Hasil yang ingin dicapai oleh manusia atau agen. |
| **Pekerjaan** | Unit kerja utama yang memiliki konteks dan hasil yang diharapkan. |
| **Tugas** | Langkah konkret yang dapat dikerjakan dan dilacak secara terpisah. |
| **Alur** | Urutan atau aturan yang menghubungkan beberapa tugas. |
| **Sesi** | Konteks eksekusi percakapan atau proses kerja tertentu. |
| **Hasil** | Keluaran kerja seperti teks, berkas, laporan, atau data. |
| **Aktivitas** | Catatan kejadian dan perkembangan kerja yang membantu pengguna memahami apa yang baru terjadi. |
| **Alat** | Kemampuan yang dapat dipanggil agen untuk melakukan pekerjaan di luar penalaran model. |
| **Skill** | Paket kemampuan atau instruksi yang dapat dipasang pada agen. |
| **Persetujuan** | Titik pengaman yang membutuhkan keputusan manusia sebelum langkah tertentu diteruskan. |
| **Penyedia** | Layanan atau proses yang menyediakan model AI. |
| **Akun AI** | Konfigurasi akses atau identitas untuk menggunakan Penyedia AI; konsep ini menjadi bagian dari rancangan AI berikutnya dan belum menjadi entitas implementasi wajib. |
| **Model** | Model AI tertentu yang digunakan oleh penyedia. |
| **Jadwal** | Aturan waktu untuk menjalankan pekerjaan atau alur secara otomatis. |

## Hubungan inti

Model domain dasar:

```text
Ruang
 ├── Agen
 │    ├── Agen induk
 │    │    └── Agen anak
 │    ├── Skill
 │    └── Alat
 ├── Sasaran
 ├── Pekerjaan
 │    ├── Tugas
 │    ├── Sesi
 │    ├── Persetujuan
 │    └── Hasil
 ├── Aktivitas
 └── Alur
      └── Tugas
```

Model hubungan AI yang menjadi arah rancangan:

```text
Penyedia AI
   ↓
Akun AI
   ↓
Model
   ↓
Agen
```

Hubungan AI di atas adalah **model target**, bukan skema SQLite final. Tahap Agen saat ini tidak mengharuskan implementasi multi-provider atau multi-account.

## Rancangan alur kerja Agen

Pengalaman pengguna yang diinginkan mengikuti urutan sederhana:

```text
Ruang kerja
   ↓
Penyedia AI + Akun AI
   ↓
Model
   ↓
Agen
   ↓
Sasaran
   ↓
Pekerjaan
   ↓
Tugas
   ↓
Agen bekerja
   ↓
Aktivitas + Hasil
```

Urutan tersebut adalah mental model, bukan wizard wajib. Pekerjaan sederhana boleh langsung dimulai dari Agen atau Tugas tanpa memaksa pengguna membuat semua tingkat di atasnya.

## Prinsip desain Agen

### Agen adalah pelaksana kerja

Agen bukan nama lain untuk Model. Agen memiliki identitas dan fungsi kerja sendiri.

Data konseptual Agen:

```text
Agen
├── nama
├── peran
├── deskripsi
├── ruang_id
├── atasan_id (opsional)
├── akun_ai_id (arah rancangan)
├── model_ai_id (arah rancangan)
└── status
```

Model AI adalah kemampuan yang dipilih Agen, sedangkan peran dan konteks menjelaskan mengapa Agen tersebut ada.

### Kredensial bukan milik Agen

Kredensial AI tidak disimpan berulang pada setiap Agen. Dalam rancangan Akun AI, kredensial berada pada konfigurasi akses yang dapat digunakan oleh beberapa Agen.

```text
Akun AI
   ├── credential
   ├── Agen A
   ├── Agen B
   └── Agen C
```

Dengan demikian, desain Agen tidak mengunci hubungan **satu Agen = satu token**.

### Hierarki Agen dibatasi oleh Ruang

Satu agen dapat memiliki paling banyak satu agen induk. Agen induk harus sudah ada dan berada pada **Ruang** yang sama. Hubungan awal dibuat saat agen anak dibuat; belum ada operasi pemindahan atau perubahan induk agar aturan siklus tetap sederhana dan aman.

Contoh:

```text
Pemimpin
├── Peneliti Produk
└── Pengembang
```

Tidak perlu membuat konsep organisasi atau tim baru hanya untuk merepresentasikan hubungan tersebut.

## Prinsip halaman Agen

Halaman Agen harus menjawab fungsi Agen sebelum detail teknis.

### Daftar Agen

Daftar menonjolkan:

- nama;
- peran;
- deskripsi singkat;
- status;
- ringkasan AI seperlunya.

URL penyedia, token, credential, dan parameter teknis tidak menjadi informasi utama.

### Detail Agen

Detail Agen secara konseptual dapat dibagi menjadi:

- **Ringkasan** — identitas, peran, hubungan, AI yang digunakan, status, dan ringkasan aktivitas.
- **Pekerjaan** — pekerjaan dan tugas yang berkaitan dengan Agen.
- **Aktivitas** — kejadian dan perkembangan kerja.
- **Pengaturan** — konfigurasi Agen yang memang relevan bagi pengguna.

Pembagian ini merupakan arah UX, bukan kewajiban bahwa semua tab harus ada sejak versi pertama.

### Form Agen

Field inti yang dirancang:

```text
Nama
Peran
Deskripsi
Atasan
Akun AI
Model
```

Form tidak meminta token/API key secara langsung. Validasi tetap dilakukan di server, termasuk validasi atasan, status akun, dan ketersediaan model bila data tersebut sudah tersedia.

## Keputusan desain

### Tidak ada organisasi sebagai konsep wajib

Klip dimulai dari **Ruang**, bukan hierarki organisasi. Pengguna tunggal dapat menjalankan Klip tanpa membuat organisasi, tim, atau struktur administratif tambahan.

Jika kebutuhan multi-pengguna berkembang, akses dapat ditambahkan tanpa menjadikan struktur organisasi sebagai syarat penggunaan dasar.

### Penyedia, Akun AI, dan Model dipisahkan secara konseptual

Jangan menyamakan:

```text
Penyedia AI = layanan AI
Akun AI    = konfigurasi akses/identitas
Kredensial = rahasia yang digunakan Akun AI
Model      = model AI yang tersedia
Agen       = pelaksana kerja yang menggunakan Akun AI + Model
```

Satu Penyedia dapat memiliki beberapa Akun AI. Satu Akun AI dapat digunakan oleh beberapa Agen. Rancangan ini belum menjadi skema database final.

### Pekerjaan dan tugas berbeda

**Pekerjaan** menjawab "apa yang sedang ingin diselesaikan". **Tugas** menjawab "langkah konkret apa yang perlu dilakukan". Satu pekerjaan dapat memiliki satu atau banyak tugas.

### Sesi bukan pekerjaan

Sesi hanya menyimpan konteks eksekusi. Pekerjaan tetap menjadi objek yang dapat dilacak meskipun memiliki beberapa sesi atau percobaan.

### Hasil dan aktivitas berbeda

**Hasil** adalah keluaran kerja seperti teks, berkas, laporan, atau data. **Aktivitas** adalah jejak kejadian dan perkembangan kerja. Aktivitas tidak menggantikan Hasil.

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

`akun_ai` belum ditambahkan sebagai tabel hanya karena rancangan visual memperlihatkannya. Entitas tersebut baru dibuat ketika desain multi-provider/account disetujui dan kebutuhan implementasinya nyata.

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

Status Agen masih dalam rancangan UX. Contoh label tampilan dapat berupa **Aktif**, **Menunggu**, **Berhenti**, atau **Bermasalah**, tetapi daftar status domain final belum ditetapkan.

Status Agen harus dibedakan dari status Pekerjaan dan Tugas.

## Batasan v0.1

v0.1 hanya membutuhkan **Ruang, Agen, Pekerjaan, Tugas, Sesi, dan Hasil** sebagai fondasi. Sasaran, Alur, Persetujuan, Skill, Alat, dan Jadwal dapat hadir sebagai struktur data setelah fondasi stabil.

Rancangan **Penyedia AI → Akun AI → Model → Agen** menjadi arah pengembangan berikutnya, tetapi multi-provider, multi-account, routing otomatis, dan pengelolaan credential per akun belum menjadi pekerjaan tahap Agen saat ini.
