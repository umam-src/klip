# Anggaran Ukuran Runtime

Ukuran yang diperhatikan Klip adalah ukuran artifact yang dijalankan, bukan ukuran source code di GitHub.

## Target

| Artifact | Target | Batas keras |
|---|---:|---:|
| Binary Klip | < 50 MiB | > 100 MiB gagal |
| Docker image | < 100 MiB | > 100 MiB perlu diperbaiki |

Target bukan alasan untuk menghapus fitur keamanan atau membuat kode sulit dipelihara.

## Prinsip optimasi

1. Utamakan library standar Go jika sudah mencukupi.
2. Jangan membawa model AI ke dalam binary atau image.
3. Asset web harus lokal dan secukupnya.
4. Hindari dependency besar untuk fungsi kecil.
5. Ukur artifact setelah build, bukan memperkirakan dari ukuran source.
6. Pantau startup dan penggunaan memori agar optimasi ukuran tidak memindahkan biaya ke runtime.

## Benchmark memori idle

Klip memiliki `BenchmarkAppMemoryIdle` untuk memantau heap yang masih dialokasikan setelah aplikasi dibuat dan garbage collection dijalankan.

Jalankan hanya saat investigasi optimasi, bukan pada setiap CI:

```text
go test ./internal/app -run '^$' -bench BenchmarkAppMemoryIdle -benchmem
```

Benchmark ini menggunakan heap sebagai metrik yang portabel antar sistem operasi. Heap bukan ukuran RSS proses secara keseluruhan; pengukuran RSS tetap dilakukan secara eksternal bila diperlukan untuk diagnosis runtime tertentu.

Hasil benchmark sebaiknya dibandingkan antar commit dengan lingkungan Go dan sistem yang sama. Jangan membuat batas keras dari satu hasil benchmark tanpa data pembanding.

## CI

CI mengukur ukuran binary setelah build release. Batas keras digunakan untuk mencegah regresi besar masuk tanpa sengaja.

Benchmark memori tidak dijalankan pada setiap perubahan karena hasilnya sensitif terhadap lingkungan runner dan tidak memberikan manfaat yang sebanding dengan menit CI tambahan.

Pemeriksaan Docker dilakukan terpisah ketika Docker build menjadi bagian release. Build Docker tidak perlu dijalankan pada setiap perubahan kecil jika tidak diperlukan.

## Jangan dilakukan

- menghapus pemeriksaan error demi beberapa byte;
- menyembunyikan dependency secara tidak aman;
- mengompres binary dengan cara yang membuat distribusi atau diagnosis sulit;
- membundel model AI besar;
- mengganti implementasi sederhana dengan optimasi kompleks tanpa pengukuran.
