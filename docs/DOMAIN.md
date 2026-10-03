# Domain Klip

Dokumen ini menetapkan bahasa dan batas domain Klip agar desain produk tetap mandiri, mudah dipahami, dan tidak menjadi salinan struktur produk lain.

## Prinsip

1. Gunakan istilah yang mudah dipahami pengguna Indonesia.
2. Pisahkan konsep produk dari detail implementasi.
3. Hindari membuat entitas hanya karena produk lain memilikinya.
4. Utamakan alur kerja lokal, sederhana, dan dapat ditelusuri.
5. Setiap entitas harus punya alasan operasional yang jelas.
6. Agen adalah pelaksana, bukan pemilik tujuan.
7. Goal menjadi pusat arah kerja; runtime hanya menjadi mekanisme pelaksanaan.
8. Batas Ruang Kerja harus dijaga pada seluruh hubungan domain.
9. Jangan mengunci Agen pada satu model, token, atau bentuk penyedia tertentu bila kebutuhan produk berkembang.

## Pusat domain

Klip berangkat dari kebutuhan untuk mencapai hasil. Karena itu **Goal** menjadi pusat semantik pekerjaan, sedangkan **Ruang Kerja** menjadi batas lingkungan tempat pekerjaan tersebut berlangsung.

```text
Ruang Kerja
    ↓
  Goal
    ↓
Pekerjaan
    ↓
  Tugas
    ↓
  Agen
    ↓
 Runtime
    ↓
Execution
    ↓
Result / Evidence
    ↓
Progress Goal
```

Makna tiap lapisan:

| Istilah | Arti di Klip |
|---|---|
| **Ruang Kerja** | Lingkungan mandiri tempat Goal, pekerjaan, tugas, agen, dan hasil kerja dikelola bersama. |
| **Goal** | Hasil atau arah yang ingin dicapai dan menjadi pusat pengelolaan kemajuan. Goal dapat memiliki Goal turunan. |
| **Pekerjaan** | Bagian pekerjaan yang dapat dikelola untuk membantu mencapai Goal. |
| **Tugas** | Langkah konkret yang dapat dikerjakan dan dilacak secara terpisah. |
| **Agen** | Pelaksana kerja berbasis AI yang memiliki identitas, peran, konteks, dan kemampuan. |
| **Skill** | Paket kemampuan atau instruksi yang dapat digunakan Agen. |
| **Alat** | Kemampuan yang dapat dipanggil Agen untuk melakukan pekerjaan di luar penalaran model. |
| **Runtime** | Mekanisme yang menyediakan kemampuan eksekusi bagi Agen. Runtime bukan pemilik Goal atau konteks kerja. |
| **Execution** | Satu kejadian pelaksanaan yang menjadi bukti bahwa pekerjaan dijalankan. |
| **Result / Evidence** | Keluaran atau bukti kerja yang dapat digunakan untuk memahami dan memperbarui kemajuan Goal. |
| **Progress** | Keadaan kemajuan Goal berdasarkan pekerjaan dan hasil yang tersedia. |
| **Alur** | Urutan atau aturan yang menghubungkan beberapa tugas. |
| **Sesi** | Konteks percakapan atau proses kerja tertentu. Sesi bukan pengganti Pekerjaan. |
| **Aktivitas** | Catatan kejadian dan perkembangan kerja. |
| **Persetujuan** | Titik pengaman yang membutuhkan keputusan manusia sebelum langkah tertentu diteruskan. |
| **Penyedia AI** | Layanan atau proses yang menyediakan kemampuan model AI. |
| **Akun AI** | Konfigurasi akses atau identitas untuk menggunakan Penyedia AI; belum menjadi entitas implementasi wajib. |
| **Model AI** | Model tertentu yang digunakan oleh penyedia. |
| **Jadwal** | Aturan waktu untuk menjalankan pekerjaan atau alur secara otomatis. |

## Ruang Kerja

### Definisi

**Ruang Kerja adalah lingkungan mandiri tempat Goal, pekerjaan, tugas, agen, dan hasil kerja dikelola bersama.**

Ruang Kerja adalah batas, bukan tujuan. Ia memastikan data dan pekerjaan yang berbeda tidak tercampur, tetapi tidak mengambil peran Goal sebagai arah produk.

Dalam kode dan penyimpanan, istilah tetap menggunakan `ruang` dan `ruang_id` agar perubahan istilah pengguna tidak memaksa migrasi teknis yang tidak diperlukan.

### Tanggung jawab Ruang Kerja

Ruang Kerja menjadi batas untuk:

- data kerja;
- Goal dan Goal turunan;
- Pekerjaan dan Tugas;
- Agen dan hierarki Agen;
- konfigurasi yang memang khusus untuk ruang tersebut;
- Execution dan jejak hasil kerja;
- Aktivitas dan peristiwa lokal;
- backup, export, dan restore pada tingkat ruang bila kemampuan tersebut tersedia.

Ruang Kerja tidak menjadi tempat menumpuk semua konfigurasi global aplikasi.

### Bentuk data inti

```text
Ruang Kerja
├── id
├── nama
├── deskripsi
├── status
├── created_at
└── updated_at
```

Status inti cukup:

- `active`
- `archived`

Pengarsipan digunakan untuk menghentikan pekerjaan baru tanpa menghapus riwayat secara destruktif.

### Hal yang tidak menjadi bagian inti Ruang Kerja

Jangan menambahkan atribut atau entitas berikut hanya untuk memperluas makna Ruang Kerja:

- organisasi;
- perusahaan;
- tim;
- pengguna sebagai pemilik wajib;
- proyek sebagai lapisan tambahan;
- Goal default;
- Agen default;
- model default;
- API key;
- URL provider;
- konfigurasi runtime besar berbentuk JSON.

Kebutuhan tersebut dapat ditambahkan kemudian jika ada kebutuhan produk yang nyata.

### Aturan isolasi

Aturan berikut merupakan invariant domain dan harus dipertahankan oleh storage, service, API, serta runtime:

1. Setiap Ruang Kerja memiliki ID unik.
2. Objek yang terikat ruang tidak boleh diakses melalui ruang lain.
3. Goal induk dan Goal turunan harus berada pada Ruang Kerja yang sama.
4. Pekerjaan dan Goal yang dihubungkan harus berada pada Ruang Kerja yang sama.
5. Pekerjaan dan Tugas yang dihubungkan harus berada pada Ruang Kerja yang sama.
6. Tugas dan Agen yang dihubungkan harus berada pada Ruang Kerja yang sama.
7. Agen induk dan Agen anak harus berada pada Ruang Kerja yang sama.
8. Execution harus tetap dapat ditelusuri ke Ruang Kerja asalnya.
9. Ruang Kerja yang diarsipkan tidak menerima pekerjaan baru kecuali diaktifkan kembali.
10. API dan service tidak boleh mengabaikan batas `ruang_id` ketika mencari atau mengubah data domain.

### Lifecycle

Ruang Kerja boleh kosong. Pengguna tidak diwajibkan membuat Goal, Agen, atau Pekerjaan saat membuat ruang.

Satu Ruang Kerja dapat memiliki banyak Goal. Goal tidak boleh berpindah Ruang Kerja melalui operasi biasa. Jika suatu saat migrasi antar-ruang diperlukan, migrasi harus menjadi operasi eksplisit dengan validasi seluruh relasi turunannya.

Ruang Kerja yang sudah memiliki riwayat kerja sebaiknya diarsipkan daripada dihapus secara permanen.

## Goal

Goal adalah pusat arah kerja di dalam Ruang Kerja.

```text
Goal
├── id
├── ruang_id
├── parent_goal_id (opsional)
├── title
├── description
├── status
├── priority
├── progress
├── target
├── deadline
├── created_at
└── updated_at
```

Goal dapat memiliki Goal turunan. Goal induk dan seluruh turunannya harus berada pada Ruang Kerja yang sama.

Agen tidak memiliki Goal. Agen menjalankan pekerjaan dan tugas yang membantu mencapai Goal.

Progress Goal tidak harus menjadi angka manual saja. Pada rancangan matang, progress dapat diperbarui berdasarkan status pekerjaan, hasil, dan bukti eksekusi yang relevan.

## Hubungan inti

```text
Ruang Kerja
│
├── Goal
│    ├── Goal turunan
│    └── Progress
│
├── Pekerjaan
│    └── Tugas
│         └── Assignment → Agen
│                              ├── Skill
│                              └── Runtime
│                                   └── Execution
│                                        └── Result / Evidence
│
├── Alur
├── Jadwal
├── Persetujuan
└── Aktivitas / Peristiwa
```

Pekerjaan sebaiknya terhubung ke Goal ketika konteks Goal memang tersedia. Tugas dapat menjadi bagian dari Pekerjaan dan assignment menentukan Agen yang menjalankannya.

## Pekerjaan dan Tugas

**Pekerjaan** menjawab "bagian apa yang sedang diselesaikan untuk membantu Goal".

**Tugas** menjawab "langkah konkret apa yang perlu dilakukan".

Satu Pekerjaan dapat memiliki satu atau banyak Tugas. Jika `Pekerjaan` dan `Tugas` memiliki `ruang_id` tersendiri, service dan storage harus memastikan keduanya selalu sama.

## Agen

Agen adalah pelaksana kerja, bukan pusat domain.

Data konseptual Agen:

```text
Agen
├── id
├── ruang_id
├── nama
├── peran
├── deskripsi
├── atasan_id (opsional)
└── status
```

Satu Agen dapat memiliki paling banyak satu Agen induk. Agen induk dan Agen anak harus berada pada Ruang Kerja yang sama.

Agen dapat menggunakan Skill, Alat, model, dan Runtime. Hubungan tersebut menjelaskan kemampuan dan mekanisme kerja Agen; hubungan tersebut tidak menjadikan Agen pemilik Goal.

## Runtime dan Execution

Runtime adalah mekanisme, bukan entitas yang menentukan arah kerja.

Execution adalah kejadian pelaksanaan yang perlu dapat ditelusuri ke konteks kerja asalnya. Konteks tersebut secara konseptual mencakup Ruang Kerja, Pekerjaan, Tugas bila ada, dan Agen yang menjalankan.

```text
Goal
  ↓
Pekerjaan
  ↓
Tugas
  ↓
Agen
  ↓
Runtime
  ↓
Execution
  ↓
Result / Evidence
  ↓
Progress Goal
```

Desain ini memungkinkan Runtime AI berkembang tanpa mengubah Goal menjadi detail implementasi runtime.

## Result / Evidence

Result adalah keluaran kerja. Evidence adalah informasi yang membantu membuktikan atau memahami apa yang terjadi.

Keduanya dapat berupa:

- teks;
- berkas;
- laporan;
- data;
- status eksekusi;
- referensi ke artefak lokal;
- metadata pelaksanaan yang aman.

Result/Evidence harus dapat ditelusuri ke Execution atau pekerjaan yang menghasilkannya. Jangan menjadikan hasil sebagai pengganti Activity log.

## Aktivitas

Aktivitas mencatat kejadian dan perkembangan kerja agar pengguna dapat memahami perubahan yang terjadi.

Aktivitas berbeda dari Result:

- **Result / Evidence** menjawab "apa yang dihasilkan atau dibuktikan".
- **Aktivitas** menjawab "apa yang terjadi".

## AI provider

Hubungan AI tetap dipisahkan secara konseptual:

```text
Penyedia AI
   ↓
Akun AI
   ↓
Model AI
   ↓
Agen
```

Hubungan tersebut bukan berarti semua entitas harus langsung menjadi tabel. Implementasi hanya menambahkan entitas ketika kebutuhan nyata muncul.

Kredensial bukan milik Agen dan tidak boleh dicatat di log aktivitas atau Execution.

## Alur kerja produk

Alur utama Klip sekarang dipahami sebagai:

```text
Ruang Kerja
   ↓
Goal
   ↓
Pekerjaan
   ↓
Tugas
   ↓
Assignment → Agen
   ↓
Runtime
   ↓
Execution
   ↓
Result / Evidence
   ↓
Progress Goal
```

Urutan ini adalah mental model domain, bukan wizard yang wajib diikuti. Pengguna tetap boleh melakukan pekerjaan sederhana tanpa membuat seluruh lapisan secara manual bila konteksnya belum diperlukan.

## Keputusan desain

### Ruang Kerja, bukan Proyek

Istilah **Ruang Kerja** dipakai pada bahasa produk karena lebih jelas sebagai batas lingkungan kerja. Istilah **Proyek** tidak digunakan sebagai lapisan domain tambahan karena mudah bertabrakan dengan makna Goal dan Pekerjaan.

### `ruang` tetap dipakai di kode

Perubahan istilah pengguna menjadi Ruang Kerja tidak memerlukan rename tabel atau kolom hanya demi kosmetik. Penyimpanan dan kode tetap dapat menggunakan `ruang` serta `ruang_id`.

### Tidak ada organisasi sebagai konsep wajib

Klip dapat digunakan tanpa organisasi, tim, atau struktur administratif. Jika kebutuhan multi-pengguna muncul, akses dapat ditambahkan sebagai lapisan terpisah tanpa menjadikan struktur organisasi syarat dasar Ruang Kerja.

### Tidak ada hubungan Agen → Goal

Goal menentukan arah. Pekerjaan dan Tugas membawa arah tersebut ke pekerjaan konkret. Agen menerima tanggung jawab melalui assignment, bukan dengan memiliki Goal.

### Tidak ada hard-delete sebagai lifecycle utama Ruang Kerja

Riwayat kerja perlu tetap dapat ditelusuri. Karena itu pengarsipan menjadi lifecycle utama; penghapusan permanen, jika dibutuhkan, harus menjadi operasi terkontrol dan mempertimbangkan seluruh data turunan.

## Bentuk data awal

Entitas yang sudah ada atau menjadi fondasi dapat tetap dipertahankan selama belum ada alasan migrasi besar. Perubahan terminologi Ruang → Ruang Kerja tidak dengan sendirinya membutuhkan perubahan tabel.

Entitas inti yang menjadi arah domain:

- `ruang`
- `goal` / sasaran
- `pekerjaan`
- `tugas`
- `agen`
- `execution` / riwayat eksekusi
- `hasil`
- `peristiwa`

Entitas pendukung dapat berkembang berdasarkan kebutuhan:

- `alur`
- `sesi`
- `skill`
- `alat`
- `persetujuan`
- `penyedia_ai`
- `model_ai`
- `jadwal`

Nama tabel dan skema aktual tetap mengikuti implementasi yang sudah ada sampai rancangan migrasi disetujui.

## Batas v0.1 dan arah berikutnya

Fondasi lama Klip sudah memiliki Ruang, Agen, Pekerjaan, Tugas, Sesi, dan Hasil. Tahap pematangan domain tidak dimaksudkan untuk menambah fitur sebanyak mungkin, tetapi untuk memperjelas hubungan antar-entitas sebelum Runtime AI diperluas.

Urutan pengembangan yang diutamakan:

1. tetapkan Ruang Kerja sebagai batas domain;
2. tetapkan Goal sebagai pusat arah kerja;
3. pastikan Pekerjaan dan Tugas dapat ditelusuri ke Goal;
4. pastikan assignment menghubungkan Tugas dengan Agen;
5. pastikan Execution dan Result/Evidence dapat ditelusuri kembali ke konteks tersebut;
6. baru perluas Runtime AI dengan konteks Goal yang jelas.

Prinsip akhirnya:

> **Ruang Kerja membatasi sistem. Goal mengarahkan sistem. Agen menjalankan sistem. Runtime mengeksekusi sistem. Result/Evidence membantu mengukur kemajuan Goal.**
