# Runtime Klip

Dokumen ini menjelaskan alur kerja eksekusi lokal Klip.

## Alur eksekusi

```text
Proyek
  ↓
Tugas (opsional)
  ↓
Agen
  ↓
Eksekusi
  ↓
status akhir
  ↓
Peristiwa / Hasil Kerja
```

Jika eksekusi terkait `Tugas`, Klip memastikan tugas berada di `Proyek` dan `Ruang` yang sama serta agen yang dipakai berada di ruang yang sama. Tugas hanya dapat dijalankan ketika statusnya mengizinkan transisi ke `running`.

`Eksekusi` menyimpan program, argumen, waktu mulai dan selesai, kode keluar, serta keluaran standar dan keluaran error. Konteks eksekusi menghubungkan proses dengan proyek, tugas, dan agen.

`Peristiwa` adalah riwayat append-only yang dapat menghubungkan kejadian dengan eksekusi, proyek, tugas, dan agen. `Hasil Kerja` adalah artefak yang memiliki identitas, jenis, nama, lokasi, serta hubungan opsional ke eksekusi dan tugas.

Keluaran proses tidak otomatis dianggap sebagai `Hasil Kerja`; artefak harus disimpan secara eksplisit agar dapat dilacak.

## Keamanan proses

- Program dijalankan langsung sebagai argumen proses, bukan melalui shell.
- Konteks dapat membatalkan proses.
- Runner menerapkan batas waktu dan ukuran keluaran.
- Runner memberi tanda jika keluaran terpotong karena batas ukuran.
- Jumlah proses bersamaan dibatasi oleh runner.
- Model AI tidak disimpan di dalam binary Klip.
- Nilai input yang mengandung karakter NUL ditolak sebelum proses dijalankan.

## Pengelolaan resource

- Database ditutup saat proses utama berhenti.
- Server HTTP memakai graceful shutdown dengan batas waktu.
- Respons HTTP dari provider AI selalu ditutup setelah selesai dibaca.
- Runner memakai `context.Context` untuk menghentikan proses dan semaphore untuk membatasi proses aktif.
- Transaksi penyimpanan memakai rollback saat terjadi kegagalan sebelum commit.
- Pengukuran heap memori idle dijelaskan di [size-budget.md](size-budget.md).

## CLI

Perintah dan opsi Klip dijelaskan di [penggunaan.md](penggunaan.md).

## Pengukuran performa

Benchmark tersedia sebagai baseline lokal. Jalankan:

```bash
go test ./... -run '^$' -bench 'BenchmarkStartupComponents|BenchmarkOpenMemory|BenchmarkOpenSQLite|BenchmarkSQLiteWriteRead' -benchmem
```

Benchmark digunakan untuk membandingkan perubahan runtime, bukan sebagai target angka tetap. Optimasi dilakukan setelah ada hasil pengukuran yang menunjukkan bagian yang memang perlu diperbaiki.

## Penyimpanan

Data runtime disimpan di SQLite lokal. Database menggunakan WAL dan satu koneksi aktif untuk menjaga perilaku local-first tetap sederhana dan dapat diprediksi.
