# Model Inti: Goal sebagai Pusat Klip

> **Arsip.** Dokumen ini ditulis sebelum remodel model data v1 dan tidak lagi menjadi rujukan. Istilah lama dipetakan ke istilah yang berlaku: `Sasaran` menjadi Goal, `Pekerjaan` menjadi Proyek, `Run`/`Execution` menjadi Eksekusi, `Assignment`/`tugas_agen` menjadi Penugasan (`penugasan`), dan `Result/Evidence` menjadi Hasil Kerja. Field `progress` pada Goal tidak disimpan sebagai angka pada schema inti v1. Rujukan yang berlaku: [DOMAIN.md](../DOMAIN.md) dan [data-model.md](../data-model.md).

Status: **Tahap 7 selesai; desain berikutnya berpusat pada Goal sebelum Runtime AI diperluas.**

## Prinsip utama

Klip diarahkan oleh satu pertanyaan utama: **apa yang ingin dicapai?**

Karena itu `Goal` menjadi pusat model produk. Ruang berfungsi sebagai batas lingkungan, sedangkan Agen, Pekerjaan, Tugas, Skill, Runtime, dan Execution menjadi bagian dari mekanisme untuk mencapai Goal.

```text
Goal
 ├── Goal turunan
 ├── Pekerjaan
 │    └── Tugas
 │         └── Assignment → Agen
 │                              ├── Skill
 │                              └── Runtime
 │                                   └── Execution
 │                                        └── Result / Evidence
 └── Progress
```

## Peran entitas

| Entitas | Peran |
|---|---|
| Ruang | Batas lingkungan dan isolasi data. |
| Goal | Hasil yang ingin dicapai dan arah utama pekerjaan. |
| Pekerjaan | Bagian besar dari Goal yang dapat dikelola. |
| Tugas | Tindakan konkret yang dapat dieksekusi. |
| Agen | Aktor yang bertanggung jawab menjalankan Tugas. |
| Skill | Kemampuan yang tersedia bagi Agen. |
| Runtime | Mekanisme yang menjalankan Agen. |
| Execution | Bukti bahwa suatu tindakan benar-benar dijalankan. |
| Result / Evidence | Hasil yang dapat digunakan untuk memperbarui kemajuan Goal. |
| Progress | Gambaran kemajuan Goal berdasarkan pekerjaan dan hasilnya. |

## Relasi yang dituju

```text
Ruang
 └── Goal
      ├── Goal turunan
      ├── Pekerjaan
      │    └── Tugas
      │         └── Agen
      ├── Agen / responsibility
      └── Progress
```

Agen tidak menjadi pusat tujuan sistem. Agen memiliki `role`, kemampuan, dan runtime, tetapi tanggung jawabnya berasal dari Goal dan pekerjaan yang diturunkan darinya.

## Model Goal yang direncanakan

```text
Goal
├── id
├── ruang_id
├── parent_goal_id
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

Relasi Pekerjaan nantinya dapat mengarah langsung ke Goal:

```text
Goal
  └── Pekerjaan
       └── Tugas
            └── tugas_agen
                 └── Agen
```

## Prinsip feedback

Klip tidak hanya menghitung Tugas yang selesai. Alur yang dituju adalah:

```text
Goal
 ↓
Pekerjaan
 ↓
Tugas
 ↓
Execution
 ↓
Result / Evidence
 ↓
Progress Goal
```

Dengan model ini sistem dapat menjelaskan dasar kemajuan Goal dari pekerjaan dan hasil yang telah terjadi.

## Hubungan dengan desain Agen

Desain Agen yang dipertahankan untuk tahap berikutnya:

- identitas dan Ruang tetap terpisah dari assignment;
- hierarki Agen digunakan untuk struktur tanggung jawab;
- Skill menyatakan kemampuan Agen;
- Runtime menjadi mekanisme eksekusi;
- assignment `Tugas → Agen` menentukan pelaksana konkret;
- permission dan approval tetap menjadi batas governance;
- budget dan usage dapat ditambahkan ketika Runtime AI membutuhkan pengukuran sumber daya.

Credential dan konfigurasi rahasia tidak menjadi bagian dari assignment.

## Tahap 7 — Assignment Agen ke Tugas

Tahap 7 telah selesai secara fungsional:

- Storage mendukung assign, reassignment, unassign, lintas Ruang, dan daftar Tugas per Agen.
- API mendukung GET, PUT, dan DELETE assignment beserta validasi error utama.
- UI menyediakan pemilihan Agen dan menampilkan pelaksana.
- Kontrak UI dilindungi embedded asset test agar CI tetap ringan.
- Runtime menolak Agen yang berbeda dari assignment sebelum execution dibuat.
- Scheduler meneruskan `PekerjaanID`, `TugasID`, dan `AgenID` ke runtime.

Verifikasi browser penuh tidak ditambahkan agar CI tetap ringan; kontrak UI, API, storage, runtime, dan scheduler menjadi verifikasi otomatis tahap ini.

## Arah Tahap 8 — Runtime AI

Sebelum menambah provider atau fitur AI, Runtime harus menerima konteks Goal yang relevan. Prioritas desain:

1. definisikan kontrak Runtime AI tanpa mengunci provider tertentu;
2. hubungkan `Goal → Pekerjaan → Tugas → Agen → Runtime`;
3. pertahankan eksekusi lokal dan offline-first bila memungkinkan;
4. pisahkan konfigurasi runtime dari credential rahasia;
5. tambahkan usage/cost hanya ketika benar-benar diperlukan;
6. ukur hasil execution sebagai bahan progress Goal;
7. pertahankan test terfokus dan CI hemat menit.

## Keputusan desain

Klip **tidak menyalin Paperclip secara penuh**. Konsep yang relevan digunakan sebagai referensi desain, sementara `Ruang` tetap menjadi boundary lokal dan `Goal` menjadi pusat model produk.
