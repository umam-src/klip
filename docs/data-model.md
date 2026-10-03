# Model Data Klip

Model data Klip dibuat kecil agar alur kerja lokal mudah dipahami dan disimpan. Model ini membedakan arah kerja, pekerjaan, tindakan, pelaksana, dan pelaksanaan.

## Hubungan utama

```text
Ruang Kerja
 ├── Goal
 │    └── Goal turunan
 │         └── Pekerjaan
 │              └── Tugas
 │                   └── Assignment → Agen
 │                                      └── Runtime → Execution
 │                                                       └── Result / Evidence
 │
 ├── Agen
 │    └── Agen induk/anak
 │
 └── Alur / Jadwal / Aktivitas / Persetujuan
```

Hubungan di atas adalah model konseptual. Nama tabel dan kolom aktual dapat tetap menggunakan istilah lama sampai migrasi benar-benar diperlukan.

## Entitas inti

- **Ruang Kerja** — batas kerja yang independen; penyimpanan tetap dapat menggunakan `ruang` dan `ruang_id`.
- **Goal** — hasil atau arah yang ingin dicapai dalam Ruang Kerja. Implementasi lama dapat menggunakan istilah **Sasaran** sampai pemetaan domain disepakati.
- **Pekerjaan** — bagian kerja yang dilakukan untuk membantu mencapai satu Goal utama.
- **Tugas** — langkah konkret dalam Pekerjaan; dapat memiliki tugas induk.
- **Agen** — pelaksana kerja berbasis AI yang menerima tanggung jawab melalui assignment.
- **Runtime** — mekanisme pelaksanaan bagi Agen; bukan pemilik Goal.
- **Execution** — kejadian pelaksanaan yang menyimpan jejak proses.
- **Result / Evidence** — keluaran atau bukti yang dapat ditelusuri ke pelaksanaan dan konteks Goal.

Entitas pendukung seperti Sesi, Run, Alur, Jadwal, Aktivitas, dan Persetujuan tetap memiliki fungsi masing-masing dan tidak menggantikan Goal, Pekerjaan, atau Tugas.

## Goal dan Sasaran

Untuk bahasa produk, **Goal** adalah istilah yang dipakai sebagai pusat arah kerja. **Sasaran** adalah representasi lama yang tetap dipertahankan pada storage/API selama migrasi bertahap berlangsung.

Keputusan pemetaan adalah:

- `Sasaran.id` menjadi identitas Goal yang sama.
- `Sasaran.ruang_id` tetap menjadi batas Ruang Kerja Goal.
- `Sasaran.title` dipetakan ke `Goal.title`.
- `Sasaran.created_at` dan `Sasaran.updated_at` dipertahankan.
- `Sasaran` tidak mengarang `description` atau `parent_goal_id` karena data tersebut memang belum tersedia pada model lama.
- Status lama dipetakan satu arah untuk kompatibilitas: `completed` → `GoalStatusCompleted`; status lama lainnya → `GoalStatusActive`. Status `failed` dan `cancelled` tidak dianggap sebagai Goal selesai.
- Adapter domain `GoalFromSasaran` menjadi batas eksplisit pemetaan tersebut.
- Tidak ada tabel `goal` baru hanya untuk mengganti nama. Migrasi persistence ditunda sampai kontrak storage Goal benar-benar membutuhkan field yang belum tersedia.

Kontrak Goal v1:

```text
Goal
├── id
├── ruang_id
├── parent_goal_id (opsional)
├── title
├── description
├── status
├── created_at
└── updated_at
```

Atribut seperti `priority`, `progress`, `target`, dan `deadline` belum menjadi alasan untuk memperbesar model sebelum kebutuhan operasionalnya jelas.

### Hasil audit implementasi lama

Audit kode menunjukkan bahwa `Sasaran` adalah representasi lama yang paling dekat dengan Goal.

- Model `domain.Sasaran` saat ini memiliki `id`, `ruang_id`, `title`, `status`, `created_at`, dan `updated_at`.
- Model `Sasaran` belum memiliki `description` dan `parent_goal_id`, sehingga hierarki Goal belum tersedia pada model lama.
- Storage `sasaran` sudah dibatasi oleh `ruang_id` dan menyediakan pembuatan serta daftar Sasaran per Ruang Kerja.
- API lama masih menggunakan istilah `/sasaran` dan menerima `title` serta `status`.
- `Pekerjaan` saat ini sudah memiliki `sasaran_id` opsional. Saat dibuat, storage memeriksa bahwa Sasaran ada dan berada pada Ruang Kerja yang sama.
- Relasi `sasaran_id` menjadi jembatan kompatibilitas; tidak dibuat `goal_id` kedua sebelum migration persistence diperlukan.

Kesimpulan: **Sasaran dan Goal diperlakukan sebagai entitas yang sama secara semantik, dengan `Sasaran` sebagai representasi legacy sementara.** Perubahan bahasa produk tidak memicu migrasi kosmetik.

## Hierarki Goal

Goal dapat memiliki Goal turunan melalui `parent_goal_id`.

```text
Goal utama
├── Goal A
├── Goal B
└── Goal C
```

Aturan wajib:

- Goal induk dan Goal turunan harus berada pada `ruang_id` yang sama.
- Goal tidak berpindah Ruang Kerja melalui operasi biasa.
- Goal root memiliki `parent_goal_id` kosong.
- Referensi ke Goal induk yang tidak ada harus ditolak.
- Siklus hierarki Goal tidak boleh dibuat.

## Goal dan Pekerjaan

Goal menjawab:

> **Apa yang ingin dicapai?**

Pekerjaan menjawab:

> **Bagian pekerjaan apa yang dilakukan untuk membantu mencapainya?**

Karena itu hubungan awal yang disiapkan adalah:

```text
Goal 1 ────────< Pekerjaan
```

Satu Pekerjaan memiliki satu Goal utama. Satu Goal dapat memiliki banyak Pekerjaan.

Untuk tahap fondasi, jangan membuat hubungan many-to-many tanpa kebutuhan nyata. Jika satu pekerjaan ternyata mendukung beberapa Goal, kasus tersebut perlu dibuktikan melalui kebutuhan produk sebelum model relasi diperluas.

Aturan wajib:

- `Pekerjaan.goal_id` harus menunjuk Goal yang ada jika hubungan Goal diwajibkan pada tahap implementasi.
- Goal dan Pekerjaan harus berada pada Ruang Kerja yang sama.
- Pekerjaan tidak boleh memakai Goal dari Ruang Kerja lain.
- Agen tidak menjadi pemilik Goal.
- Runtime tidak menjadi pemilik Goal.

## Pekerjaan dan Tugas

Pekerjaan adalah unit kerja yang dapat dikelola. Tugas adalah langkah konkret di dalamnya.

```text
Goal
  ↓
Pekerjaan
  ├── Tugas A
  ├── Tugas B
  └── Tugas C
```

Satu Pekerjaan dapat memiliki banyak Tugas. Tugas tetap berada pada Pekerjaan yang sama dan tidak boleh melampaui batas Ruang Kerja.

## Agen dan pelaksanaan

Agen menjalankan pekerjaan atau Tugas melalui assignment. Assignment menjawab siapa yang menjalankan, bukan apa tujuan kerja tersebut.

```text
Goal
  ↓
Pekerjaan
  ↓
Tugas
  ↓
Assignment
  ↓
Agen
  ↓
Runtime
  ↓
Execution
  ↓
Result / Evidence
```

Dengan bentuk ini, Goal tetap independen dari model AI, provider, runtime, dan kredensial.

## Progress

Progress adalah keadaan kemajuan Goal. Pada tahap matang, progress sebaiknya dapat dijelaskan oleh pekerjaan, status tugas, dan Result / Evidence yang tersedia.

Jangan menjadikan angka progress sebagai satu-satunya sumber kebenaran. Jika progress disimpan, sumber perubahan dan alasan perubahannya harus dapat ditelusuri.

## Aturan batas Ruang Kerja

Semua hubungan lintas entitas harus mempertahankan batas `ruang_id`.

Minimal:

1. Goal ↔ Goal induk: Ruang Kerja sama.
2. Goal ↔ Pekerjaan: Ruang Kerja sama.
3. Pekerjaan ↔ Tugas: Ruang Kerja sama.
4. Tugas ↔ Agen: Ruang Kerja sama.
5. Agen ↔ Agen induk: Ruang Kerja sama.
6. Execution ↔ konteks kerja: Ruang Kerja dapat ditelusuri.
7. Result / Evidence ↔ sumber pelaksanaan: konteks Ruang Kerja dapat ditelusuri.

Pemeriksaan ini harus berada pada service/storage yang menjadi batas mutasi data, bukan hanya pada UI.

## Status dan lifecycle

Status Goal awal sebaiknya kecil:

- `active`
- `completed`
- `archived`

Status seperti `blocked`, `paused`, atau `waiting` lebih cocok diekspresikan pada Pekerjaan/Tugas atau melalui aktivitas yang menjelaskan kondisi kerja.

Goal yang sudah memiliki riwayat tidak sebaiknya dihapus secara destruktif hanya untuk menghilangkan tampilan. Pengarsipan menjaga riwayat tetap dapat ditelusuri.

## Keputusan yang belum dikunci

Dokumen ini sengaja belum mengunci beberapa keputusan implementasi:

- bagaimana menambah `description` dan hierarki Goal pada persistence tanpa merusak data lama;
- kapan relasi Sasaran/Goal pada Pekerjaan harus menjadi wajib;
- bagaimana progress Goal dihitung dari Execution dan Result / Evidence;
- kapan `priority`, `target`, dan `deadline` benar-benar diperlukan;
- apakah kebutuhan many-to-many Goal ↔ Pekerjaan akan muncul.

Keputusan tersebut harus didasarkan pada model yang sudah ada dan kebutuhan nyata, bukan menambah struktur untuk mengantisipasi fitur.

## Prinsip implementasi

1. Matangkan semantik Goal sebelum migrasi schema.
2. Pertahankan kompatibilitas dengan model `Sasaran` selama migrasi bertahap.
3. Gunakan `GoalFromSasaran` sebagai batas mapping, bukan duplikasi relasi.
4. Jangan membuat relasi yang tidak dibutuhkan.
5. Setiap relasi baru harus memiliki pemeriksaan `ruang_id`.
6. Tambahkan regression test untuk setiap invariant baru.
7. Runtime hanya menerima konteks Goal; Runtime tidak memiliki Goal.
8. Result / Evidence menjadi dasar yang dapat ditelusuri untuk progress, bukan sumber arah kerja.
