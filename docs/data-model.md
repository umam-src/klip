# Model Data Klip

Dokumen ini menjadi rancangan data keseluruhan sebelum remodel database. Klip masih dalam tahap beta dan belum memiliki database pengguna yang harus dipertahankan kompatibilitasnya. Karena itu model domain baru boleh menjadi sumber kebenaran dan schema lama tidak perlu dipertahankan hanya demi kompatibilitas.

## Prinsip

- Goal adalah pusat arah kerja.
- Ruang Kerja adalah batas isolasi data.
- Agen adalah pelaksana, bukan pemilik Goal.
- Runtime dan Execution adalah mekanisme pelaksanaan, bukan pusat domain.
- Result / Evidence adalah keluaran dan bukti yang dapat ditelusuri.
- Schema dibuat setelah hubungan dan invariant domain dikunci.
- Jangan membuat entity hanya untuk mengantisipasi fitur.
- Database beta boleh tidak kompatibel dengan schema sebelumnya.
- Remodel dilakukan sekali secara menyeluruh agar tidak membangun lapisan legacy yang akan dibawa sampai rilis.

## Peta domain v1

```text
Ruang Kerja
│
├── Goal
│   ├── Goal turunan
│   └── Pekerjaan
│       └── Tugas
│           └── Assignment → Agen
│                               └── Runtime
│                                   └── Execution
│                                       ├── Event
│                                       └── Result / Evidence
│
├── Agen
│   └── Agen induk/anak
│
├── Alur
├── Jadwal
├── Persetujuan
└── Aktivitas / Peristiwa
```

Peta ini adalah rancangan hubungan. Tidak semua kotak harus menjadi tabel pada migration pertama. Entity pendukung hanya dibuat ketika sudah memiliki kebutuhan operasional yang jelas.

## Entity inti

| Entity | Tujuan | Kepemilikan utama | Lifecycle awal |
|---|---|---|---|
| Ruang Kerja | Batas lingkungan kerja | mandiri | `active`, `archived` |
| Goal | Menentukan hasil/arah yang ingin dicapai | Ruang Kerja | `active`, `completed`, `archived` |
| Pekerjaan | Bagian kerja untuk membantu Goal | Ruang Kerja + Goal | mengikuti lifecycle kerja yang sudah ada |
| Tugas | Langkah konkret dalam Pekerjaan | Pekerjaan | mengikuti lifecycle tugas |
| Agen | Pelaksana kerja | Ruang Kerja | mengikuti lifecycle agen |
| Assignment | Menentukan Agen yang menjalankan Tugas | Tugas + Agen | aktif/nonaktif sesuai kebutuhan |
| Execution | Satu kejadian pelaksanaan | Pekerjaan + Agen | riwayat, tidak diedit menjadi objek baru |
| Event | Jejak kejadian pelaksanaan | Execution/konteks kerja | append-only secara konseptual |
| Result / Evidence | Keluaran atau bukti kerja | Execution/konteks kerja | tersimpan dan dapat ditelusuri |

## Entity pendukung

Entity berikut sudah ada atau diperlukan oleh fitur yang berjalan, tetapi bukan pusat model:

- **Sesi** — konteks percakapan/proses tertentu.
- **Alur** — urutan atau aturan yang menghubungkan pekerjaan/tugas.
- **Jadwal** — aturan waktu untuk menjalankan pekerjaan atau alur.
- **Persetujuan** — titik pengaman yang membutuhkan keputusan manusia.
- **Aktivitas / Peristiwa** — catatan perubahan dan kejadian yang dapat ditampilkan kepada pengguna.
- **Skill** — kemampuan/instruksi lokal yang dapat digunakan Agen.
- **Alat** — kemampuan yang dapat dipanggil Agen.
- **Penyedia AI** dan **Model AI** — konfigurasi kemampuan AI jika memang diperlukan sebagai entity persistence.

Entity seperti organisasi, perusahaan, tim, pengguna wajib, billing, marketplace, knowledge base, atau percakapan global tidak masuk schema inti v1 tanpa kebutuhan produk yang nyata.

## Schema konseptual v1

### Ruang Kerja

```text
ruang
├── id
├── nama
├── deskripsi
├── status
├── created_at
└── updated_at
```

`status` hanya `active` atau `archived` pada fondasi.

### Goal

```text
goal
├── id
├── ruang_id
├── parent_goal_id (nullable)
├── title
├── description
├── status
├── created_at
└── updated_at
```

`parent_goal_id` adalah self-reference ke `goal.id`.

Aturan:

1. Goal harus memiliki Ruang Kerja.
2. Parent boleh kosong untuk Goal root.
3. Jika parent ada, parent harus ada.
4. Parent dan child harus memiliki `ruang_id` yang sama.
5. Goal tidak boleh menjadi parent dirinya sendiri.
6. Tidak boleh ada cycle pada hierarki.
7. Goal tidak berpindah Ruang Kerja melalui operasi biasa.
8. Lifecycle Goal menggunakan `active`, `completed`, dan `archived`.

### Pekerjaan

```text
pekerjaan
├── id
├── ruang_id
├── goal_id
├── title
├── description
├── status
├── created_at
└── updated_at
```

Satu Pekerjaan memiliki satu Goal utama. Satu Goal dapat memiliki banyak Pekerjaan.

Aturan:

1. `goal_id` harus menunjuk Goal yang ada.
2. `pekerjaan.ruang_id == goal.ruang_id`.
3. Pekerjaan tidak boleh menunjuk Goal dari Ruang Kerja lain.
4. Many-to-many Goal ↔ Pekerjaan tidak dibuat pada v1.

### Tugas

```text
tugas
├── id
├── ruang_id
├── pekerjaan_id
├── parent_tugas_id (nullable)
├── title
├── description
├── status
├── created_at
└── updated_at
```

Jika field aktual yang sudah berjalan memiliki bentuk berbeda, remodel mengikuti kontrak domain yang sudah terbukti dan hanya menambah field yang dibutuhkan invariant.

Aturan:

- Tugas harus berada pada Pekerjaan yang sama.
- Parent Tugas, jika ada, harus berada pada Pekerjaan dan Ruang Kerja yang sama.
- Hierarki Tugas tidak boleh membentuk cycle.

### Agen

```text
agen
├── id
├── ruang_id
├── nama
├── peran
├── deskripsi
├── atasan_id (nullable)
├── status
├── created_at
└── updated_at
```

Aturan:

- Agen harus berada pada satu Ruang Kerja.
- `atasan_id`, jika ada, harus menunjuk Agen pada Ruang Kerja yang sama.
- Hierarki Agen tidak boleh membentuk cycle.
- Agen tidak memiliki `goal_id`.

### Assignment

Assignment memisahkan **siapa yang menjalankan** dari **apa yang ingin dicapai**.

```text
assignment
├── id
├── tugas_id
├── agen_id
├── created_at
└── updated_at
```

Relasi efektif:

```text
Goal → Pekerjaan → Tugas → Assignment → Agen
```

Assignment tidak menjadi kepemilikan Goal.

### Execution

Execution menyimpan satu kejadian pelaksanaan yang dapat ditelusuri.

```text
execution
├── id
├── ruang_id
├── pekerjaan_id
├── tugas_id (nullable)
├── agen_id
├── status
├── started_at
├── finished_at (nullable)
└── metadata aman yang memang diperlukan
```

Aturan minimum:

- Pekerjaan harus berada pada Ruang Kerja yang sama.
- Tugas, jika ada, harus berasal dari Pekerjaan tersebut.
- Agen harus berada pada Ruang Kerja yang sama.
- Execution tidak menyimpan credential atau secret.
- Execution adalah riwayat, bukan pemilik Goal.

Nama persistence `run` boleh diubah menjadi `execution` bila remodel dilakukan sekaligus. Tidak perlu mempertahankan dua konsep untuk makna yang sama.

### Event

```text
peristiwa
├── id
├── ruang_id
├── execution_id
├── pekerjaan_id
├── tugas_id (nullable)
├── agen_id
├── jenis
├── data aman
└── created_at
```

Event harus dapat ditelusuri ke Execution dan konteks Ruang Kerja. Data rahasia tidak boleh masuk event.

### Result / Evidence

```text
hasil
├── id
├── ruang_id
├── execution_id (nullable jika hasil berasal dari konteks lain yang sah)
├── pekerjaan_id
├── tugas_id (nullable)
├── kind
├── name
├── path / reference
└── created_at
```

Result harus memiliki jalur penelusuran ke konteks kerja. Path harus tetap dibatasi oleh root penyimpanan lokal yang ditentukan aplikasi.

## Relasi wajib

```text
Ruang Kerja 1 ────< Goal
Goal         1 ────< Goal
Goal         1 ────< Pekerjaan
Pekerjaan    1 ────< Tugas
Tugas        1 ────< Assignment >──── 1 Agen
Agen         1 ────< Agen
Pekerjaan    1 ────< Execution
Tugas        0..1 ─── Execution
Agen         1 ────< Execution
Execution   1 ────< Event
Execution   0..1 ───< Result / Evidence
```

Semua hubungan yang melibatkan entity ber-`ruang_id` harus mempertahankan Ruang Kerja yang sama.

## Invariant lintas domain

1. Tidak ada foreign key domain yang boleh melewati Ruang Kerja.
2. Goal parent/child selalu satu Ruang Kerja.
3. Goal → Pekerjaan selalu satu Ruang Kerja.
4. Pekerjaan → Tugas selalu satu Ruang Kerja.
5. Tugas → Agen melalui Assignment selalu satu Ruang Kerja.
6. Agen parent/child selalu satu Ruang Kerja.
7. Execution dapat ditelusuri ke Ruang Kerja, Pekerjaan, dan Agen.
8. Event dapat ditelusuri ke Execution.
9. Result/Evidence dapat ditelusuri ke Execution atau konteks kerja yang sah.
10. Ruang Kerja archived tidak menerima pekerjaan baru tanpa reaktivasi.

Invariant harus ditegakkan di database bila memungkinkan dan tetap divalidasi pada service/storage. UI hanya membantu pengguna dan bukan batas keamanan domain.

## Goal sebagai pusat model

Goal menjawab:

> Apa yang ingin dicapai?

Pekerjaan menjawab:

> Bagian pekerjaan apa yang dilakukan untuk membantu mencapainya?

Tugas menjawab:

> Langkah konkret apa yang perlu dilakukan?

Assignment menjawab:

> Agen mana yang menjalankannya?

Execution menjawab:

> Apa yang benar-benar dijalankan?

Result / Evidence menjawab:

> Apa yang dihasilkan atau dapat dibuktikan?

Dengan pembagian ini, Goal tidak bergantung pada provider AI, model, runtime, atau bentuk eksekusi tertentu.

## Progress Goal

Progress belum menjadi sumber kebenaran mandiri pada schema v1. Bila kelak disimpan, perubahan harus dapat dijelaskan oleh Pekerjaan, Tugas, Execution, dan Result/Evidence.

Jangan menambahkan `progress`, `target`, `priority`, atau `deadline` ke schema inti hanya untuk melengkapi tampilan Goal sebelum kebutuhan operasionalnya jelas.

## Status dan lifecycle

Status Goal tetap kecil:

- `active`
- `completed`
- `archived`

Status seperti `blocked`, `waiting`, dan `paused` lebih cocok pada Pekerjaan/Tugas atau aktivitas pelaksanaan.

Pengarsipan lebih diutamakan daripada hard-delete untuk entity yang sudah memiliki riwayat.

## Keputusan remodel

### `Sasaran` dihapus sebagai entity domain

`Sasaran` adalah nama lama untuk konsep yang sekarang disebut Goal. Karena Klip belum rilis, tidak ada alasan mempertahankan adapter `GoalFromSasaran` sebagai lapisan permanen.

Remodel v1 akan:

- mengganti `Sasaran` menjadi `Goal`;
- mengganti `sasaran_id` menjadi `goal_id`;
- menambahkan `parent_goal_id`;
- menyediakan `description`;
- menggunakan lifecycle Goal khusus;
- menghapus adapter `GoalFromSasaran` setelah persistence baru aktif;
- menghapus schema legacy `sasaran` daripada mempertahankannya sebagai compatibility layer.

### Database tidak harus kompatibel

Database beta boleh di-reset atau dimigrasikan dengan cara yang tidak mempertahankan bentuk lama. Prioritasnya adalah schema yang benar dan sederhana sebelum rilis, bukan kompatibilitas semu dengan database pengembangan lama.

Namun reset database tidak boleh menjadi alasan untuk menghapus regression test. Invariant domain tetap harus dibuktikan oleh test.

## Tahapan implementasi setelah blueprint

1. Audit seluruh entity dan kolom yang masih memakai istilah legacy.
2. Finalisasi kontrak domain berdasarkan dokumen ini.
3. Remodel schema SQLite sekali jalan.
4. Remodel repository/storage.
5. Perbarui service dan API.
6. Perbarui UI dan bahasa produk.
7. Hapus `Sasaran` dan adapter legacy.
8. Perbarui test dan fixture.
9. Jalankan seluruh CI dan perbaiki regresi.
10. Setelah schema v1 stabil, lanjutkan penguatan Runtime AI.

## Batas v1

Schema v1 tidak mencakup organisasi, tim, billing, marketplace, knowledge base, multi-provider orchestration kompleks, atau entity lain yang belum memiliki kebutuhan produk nyata.

Tujuan remodel adalah membuat **satu model data yang konsisten**, bukan membuat schema terbesar yang mungkin.