# Model Data Klip

Dokumen ini menjadi sumber rancangan model data Klip v1. Klip masih beta dan belum memiliki data pengguna yang harus dipertahankan kompatibilitasnya. Karena itu model baru menjadi sumber kebenaran dan schema lama tidak dipertahankan hanya demi kompatibilitas.

## Prinsip

- Goal adalah pusat arah kerja.
- Ruang Kerja adalah batas isolasi data.
- Proyek adalah wadah pekerjaan untuk membantu mencapai Goal.
- Tugas adalah pekerjaan konkret.
- Penugasan menentukan Agen yang menjalankan Tugas.
- Agen adalah pelaku kerja; CEO, Planner, Manager, dan Worker cukup menjadi role Agen.
- Board adalah fungsi pengarah dan koordinasi, bukan entitas Agen khusus.
- Eksekusi adalah kejadian pelaksanaan nyata.
- Hasil Kerja adalah keluaran atau bukti kerja.
- `progress` tidak disimpan sebagai angka pada schema inti v1.
- Tidak ada compatibility layer untuk `Sasaran` setelah remodel.
- Jangan membuat entity hanya untuk mengantisipasi fitur.

## Peta domain v1

```text
Ruang Kerja
│
├── Goal
│   ├── Goal turunan
│   └── Proyek
│       └── Tugas
│           └── Penugasan → Agen
│                         └── Eksekusi
│                              ├── Peristiwa
│                              └── Hasil Kerja
│
└── Agen
    └── Agen induk/anak
```

Board mengarahkan Proyek/Tugas dan melakukan Penugasan. Agen dengan role pengarah, misalnya CEO atau Planner, dapat menerima Penugasan lalu menunjuk Agen lain melalui Penugasan berikutnya.

## Bahasa domain

| Konsep | Istilah produk | Nama kode/storage yang disarankan |
|---|---|---|
| Workspace | Ruang Kerja | `ruang` / `ruang_id` |
| Goal | Goal | `goal` / `goal_id` |
| Project | Proyek | `proyek` / `proyek_id` |
| Task | Tugas | `tugas` / `tugas_id` |
| Assignment | Penugasan | `assignment` / `assignment_id` |
| Agent | Agen | `agen` / `agen_id` |
| Execution | Eksekusi | `execution` / `execution_id` |
| Result / Evidence | Hasil Kerja | `hasil` / `hasil_id` |
| Event | Peristiwa | `peristiwa` / `peristiwa_id` |

Istilah Inggris hanya dipertahankan pada nama kode jika membantu kestabilan implementasi atau merupakan istilah teknis yang sudah mapan. Bahasa produk menggunakan istilah pada kolom kedua.

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

Status: `active`, `archived`.

Aturan:

1. Semua data kerja yang terikat ruang harus memiliki `ruang_id` atau dapat ditelusuri secara jelas ke Ruang Kerja.
2. Ruang Kerja archived tidak menerima pekerjaan baru tanpa reaktivasi.
3. Pengarsipan lebih diutamakan daripada hard-delete.

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

Status: `active`, `completed`, `archived`.

Aturan:

1. Goal wajib memiliki Ruang Kerja.
2. `parent_goal_id` boleh kosong untuk Goal root.
3. Parent harus ada jika diisi.
4. Parent dan child harus berada pada Ruang Kerja yang sama.
5. Goal tidak boleh menjadi parent dirinya sendiri.
6. Hierarki Goal tidak boleh membentuk cycle.
7. Goal tidak berpindah Ruang Kerja melalui operasi biasa.
8. Goal tidak memiliki `progress`, `target`, `priority`, atau `deadline` sebagai field inti v1.

### Proyek

Proyek adalah wadah pekerjaan yang dibuat untuk membantu mencapai satu Goal utama.

```text
proyek
├── id
├── ruang_id
├── goal_id
├── title
├── description
├── status
├── created_at
└── updated_at
```

Relasi:

```text
Goal 1 ────< Proyek
```

Aturan:

1. `goal_id` wajib menunjuk Goal yang ada.
2. Proyek dan Goal harus berada pada Ruang Kerja yang sama.
3. Satu Proyek memiliki satu Goal utama.
4. Satu Goal dapat memiliki banyak Proyek.
5. Tidak ada many-to-many Goal ↔ Proyek pada v1.

Status Proyek/Tugas dapat menggunakan lifecycle operasional yang sudah dibuktikan oleh implementasi, misalnya `draft`, `ready`, `running`, `waiting`, `blocked`, `completed`, `failed`, dan `cancelled`, tanpa memaksakan seluruh status pada Goal.

### Tugas

Tugas adalah pekerjaan konkret yang dapat diberikan kepada Agen.

```text
tugas
├── id
├── ruang_id
├── proyek_id
├── parent_tugas_id (nullable)
├── title
├── description
├── status
├── created_at
└── updated_at
```

Aturan:

1. Tugas wajib berada pada Proyek yang sama.
2. Proyek dan Tugas harus berada pada Ruang Kerja yang sama.
3. Parent Tugas, jika ada, harus berada pada Proyek dan Ruang Kerja yang sama.
4. Hierarki Tugas tidak boleh membentuk cycle.

### Penugasan

Penugasan adalah mekanisme penunjukan Agen untuk menjalankan Tugas. Penugasan bukan kepemilikan Goal.

```text
assignment
├── id
├── tugas_id
├── agen_id
├── created_at
└── updated_at
```

Relasi inti:

```text
Goal → Proyek → Tugas → Penugasan → Agen
```

Aturan:

1. Tugas dan Agen harus berada pada Ruang Kerja yang sama.
2. Penugasan harus dapat ditelusuri ke Ruang Kerja melalui Tugas dan Agen.
3. Penugasan dapat dilakukan oleh Board atau Agen pengarah sesuai aturan otorisasi produk.
4. CEO tidak membutuhkan entity khusus; CEO adalah role pada Agen.
5. Agen pengarah dapat menunjuk Agen lain dengan membuat Penugasan pada Tugas yang menjadi mandatnya.

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

`peran` adalah role produk, bukan entity terpisah. Contoh: `ceo`, `planner`, `manager`, `worker`, `reviewer`.

Aturan:

1. Agen wajib berada pada satu Ruang Kerja.
2. `atasan_id`, jika ada, harus menunjuk Agen pada Ruang Kerja yang sama.
3. Hierarki Agen tidak boleh membentuk cycle.
4. Agen tidak memiliki `goal_id`.
5. Agen dapat menerima dan membuat Penugasan sesuai kewenangan yang diberikan.

### Board

Board bukan tabel/entity inti v1.

Board diperlakukan sebagai **fungsi pengarah dan koordinasi** yang dapat:

- menetapkan atau mengarahkan Proyek;
- membuat atau memecah Tugas;
- menunjuk Agen melalui Penugasan;
- menunjuk Agen pengarah/perencana untuk membantu koordinasi;
- memantau keadaan pekerjaan melalui data operasional.

Kita tidak membuat `board`, `ceo`, atau `manager` entity hanya untuk merepresentasikan fungsi tersebut.

### Eksekusi

Eksekusi menyimpan satu kejadian pelaksanaan nyata.

```text
execution
├── id
├── ruang_id
├── proyek_id
├── tugas_id (nullable)
├── agen_id
├── status
├── started_at
├── finished_at (nullable)
└── metadata aman yang memang diperlukan
```

Aturan:

1. Proyek harus berada pada Ruang Kerja yang sama.
2. Tugas, jika ada, harus berasal dari Proyek tersebut.
3. Agen harus berada pada Ruang Kerja yang sama.
4. Eksekusi tidak menyimpan credential atau secret.
5. Eksekusi adalah riwayat pelaksanaan, bukan pemilik Goal.
6. Nama persistence `run` boleh diubah menjadi `execution` dalam remodel karena keduanya merepresentasikan satu konsep pelaksanaan.

### Peristiwa

```text
peristiwa
├── id
├── ruang_id
├── execution_id
├── proyek_id
├── tugas_id (nullable)
├── agen_id
├── jenis
├── data aman
└── created_at
```

Peristiwa harus dapat ditelusuri ke Eksekusi dan konteks Ruang Kerja. Data rahasia tidak boleh masuk Peristiwa.

### Hasil Kerja

Hasil Kerja adalah keluaran atau bukti yang dihasilkan dari pekerjaan. Ini **bukan pengganti Progress** dan bukan sekadar status berhasil/gagal.

```text
hasil
├── id
├── ruang_id
├── execution_id (nullable jika konteks kerja lain memang sah)
├── proyek_id
├── tugas_id (nullable)
├── kind
├── name
├── path / reference
└── created_at
```

Contoh Hasil Kerja:

- berkas yang dibuat;
- laporan;
- data keluaran;
- hasil pengujian;
- referensi artefak lokal;
- bukti bahwa suatu pekerjaan menghasilkan sesuatu.

Aturan:

1. Hasil Kerja harus memiliki konteks Ruang Kerja.
2. Jika berasal dari Eksekusi, `execution_id` harus menunjuk Eksekusi yang sesuai.
3. `proyek_id` dan `tugas_id`, jika ada, harus konsisten dengan Eksekusi/konteks kerja.
4. Path harus dibatasi oleh root penyimpanan lokal aplikasi.
5. Credential dan secret tidak boleh disimpan sebagai Hasil Kerja tanpa mekanisme keamanan khusus.

## Relasi wajib

```text
Ruang Kerja 1 ────< Goal
Goal         1 ────< Goal
Goal         1 ────< Proyek
Proyek      1 ────< Tugas
Tugas       1 ────< Penugasan >──── 1 Agen
Agen        1 ────< Agen
Proyek      1 ────< Eksekusi
Tugas       0..1 ─── Eksekusi
Agen        1 ────< Eksekusi
Eksekusi    1 ────< Peristiwa
Eksekusi    0..1 ─── Hasil Kerja
```

Semua hubungan yang melibatkan entity ber-`ruang_id` harus mempertahankan Ruang Kerja yang sama.

## Progress Goal

Progress bukan sumber kebenaran mandiri pada schema v1.

Keadaan Goal nantinya dapat diturunkan dari kondisi pekerjaan di bawahnya, misalnya:

```text
Goal
 ↓
Proyek
 ↓
Tugas
 ↓
Penugasan
 ↓
Eksekusi
 ↓
Hasil Kerja
```

Hasil Kerja berperan sebagai **bukti/output**, sedangkan progress adalah **kesimpulan keadaan** yang dapat dihitung dari data kerja. Karena itu kita tidak menyimpan angka `progress` hanya agar tampilan memiliki persentase.

Belum ditetapkan satu rumus progress pada v1. Bobot Tugas, target numerik, prioritas, dan deadline hanya ditambahkan ketika kebutuhan operasionalnya benar-benar terbukti.

## Invariant lintas domain

1. Tidak ada foreign key domain yang boleh melewati Ruang Kerja.
2. Goal parent/child selalu satu Ruang Kerja.
3. Goal → Proyek selalu satu Ruang Kerja.
4. Proyek → Tugas selalu satu Ruang Kerja.
5. Tugas → Agen melalui Penugasan selalu satu Ruang Kerja.
6. Agen parent/child selalu satu Ruang Kerja.
7. Eksekusi dapat ditelusuri ke Ruang Kerja, Proyek, Tugas bila ada, dan Agen.
8. Peristiwa dapat ditelusuri ke Eksekusi.
9. Hasil Kerja dapat ditelusuri ke Eksekusi atau konteks kerja yang sah.
10. Ruang Kerja archived tidak menerima pekerjaan baru tanpa reaktivasi.
11. Hierarki Goal dan Tugas tidak boleh cycle.
12. Hierarki Agen tidak boleh cycle.

Invariant harus ditegakkan di database bila memungkinkan dan tetap divalidasi pada service/storage. UI bukan batas keamanan domain.

## Penghapusan model lama

### `Sasaran`

`Sasaran` adalah nama lama untuk konsep Goal. Remodel v1 menghapusnya sebagai entity domain.

Yang dihapus/diganti:

- `Sasaran` → `Goal`;
- `sasaran_id` → `goal_id`;
- adapter `GoalFromSasaran`;
- schema legacy `sasaran`;
- test compatibility yang hanya mempertahankan bentuk lama.

### `Pekerjaan`

`Pekerjaan` adalah istilah domain lama yang diganti menjadi **Proyek**. Remodel tidak mempertahankan dua istilah untuk konsep yang sama.

### `Run` dan `Execution`

`Run` merupakan nama lama untuk kejadian pelaksanaan. Persistence baru menggunakan `execution` dan kode domain menggunakan `Execution`/`Eksekusi` agar tidak ada dua konsep dengan makna sama.

## Batas v1

Tidak menjadi schema inti v1:

- Organization;
- Team;
- Billing;
- Marketplace;
- Knowledge Base;
- Conversation global;
- orkestrasi multi-provider kompleks;
- entity Board khusus;
- entity CEO khusus;
- penyimpanan progress numerik;
- fitur masa depan yang belum memiliki kebutuhan operasional nyata.

## Urutan implementasi

1. Audit seluruh entity, tabel, kolom, service, API, UI, scheduler, runtime, dan test.
2. Remodel schema SQLite menjadi schema native v1.
3. Remodel domain dan repository.
4. Perbarui service dan API.
5. Perbarui UI dan bahasa produk.
6. Perbarui scheduler dan runtime dari `Pekerjaan` ke `Proyek` serta `Run` ke `Eksekusi`.
7. Hapus `Sasaran`, `Pekerjaan`, dan adapter compatibility lama setelah seluruh pemakaian berpindah.
8. Perbarui test dan fixture.
9. Perbarui dokumentasi dan changelog.
10. Jalankan CI lengkap dan perbaiki regresi sebelum melanjutkan fitur Runtime AI.
