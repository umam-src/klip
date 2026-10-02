# Runtime Klip

Dokumen ini menjelaskan alur kerja eksekusi lokal Klip.

## Alur eksekusi

```text
Tugas
  ↓
Sesi
  ↓
Run
  ↓
status akhir
```

Jika eksekusi terkait `Tugas`, Klip memastikan tugas berada di `Pekerjaan` yang sama. Saat proses dimulai, status tugas menjadi `running`. Setelah proses selesai, status tugas mengikuti hasil eksekusi.

`Run` menyimpan program, argumen, waktu, kode keluar, serta keluaran standar dan keluaran error. `Sesi` menyimpan hubungan eksekusi dengan pekerjaan dan agen.

## Keamanan proses

- Program dijalankan langsung sebagai argumen proses, bukan melalui shell.
- Konteks dapat membatalkan proses.
- Batas waktu dan ukuran keluaran diterapkan oleh runtime.
- Runner memberi tanda jika keluaran terpotong karena batas ukuran.
- Jumlah proses bersamaan dibatasi oleh runner.
- Model AI tidak disimpan di dalam binary Klip.

## Pengelolaan resource

- Database ditutup saat proses utama berhenti.
- Server HTTP memakai graceful shutdown dengan batas waktu.
- Respons HTTP dari provider AI selalu ditutup setelah selesai dibaca.
- Runner memakai `context.Context` untuk menghentikan proses dan semaphore untuk membatasi proses aktif.
- Transaksi penyimpanan memakai rollback saat terjadi kegagalan sebelum commit.
- Benchmark memory idle belum dianggap selesai karena belum ada pengukuran RSS/heap jangka panjang yang representatif.

## Pengukuran performa

Benchmark tersedia sebagai baseline lokal. Jalankan:

```bash
go test ./... -run '^$' -bench 'BenchmarkStartupComponents|BenchmarkOpenSQLite|BenchmarkRunnerRun' -benchmem
```

`BenchmarkStartupComponents` mengukur inisialisasi komponen aplikasi utama menggunakan database memori. `BenchmarkOpenSQLite` mengukur pembukaan dan penutupan database SQLite pada penyimpanan lokal. `BenchmarkRunnerRun` mengukur overhead menjalankan proses lokal melalui runner.

Benchmark digunakan untuk membandingkan perubahan runtime, bukan sebagai target angka tetap. Optimasi dilakukan setelah ada hasil pengukuran yang menunjukkan bagian yang memang perlu diperbaiki.

## Hasil

`Hasil` adalah artefak kerja yang memiliki identitas, jenis, nama, lokasi, dan hubungan opsional ke tugas. Keluaran proses tidak otomatis dianggap sebagai `Hasil`. Pemisahan ini menjaga riwayat proses dan artefak kerja tetap memiliki makna yang berbeda.

## Penyimpanan

Data runtime disimpan di SQLite lokal. Database menggunakan WAL dan satu koneksi aktif untuk menjaga perilaku lokal tetap sederhana dan dapat diprediksi.
