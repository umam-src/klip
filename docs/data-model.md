# Model Data Klip

Model data Klip dibuat kecil agar alur kerja lokal mudah dipahami dan disimpan.

## Hubungan utama

```text
Ruang
 ├── Agen
 │    └── Agen induk/anak
 ├── Sasaran
 ├── Pekerjaan
 │    ├── Tugas
 │    ├── Sesi
 │    ├── Run
 │    └── Hasil
 └── Alur
```

## Entitas

- **Ruang** — batas kerja yang independen.
- **Agen** — pekerja AI yang berada dalam satu ruang dan dapat memiliki satu agen induk.
- **Sasaran** — hasil yang ingin dicapai dalam ruang.
- **Pekerjaan** — unit kerja utama dan pemilik sesi eksekusi.
- **Tugas** — langkah konkret dalam pekerjaan; dapat memiliki tugas induk.
- **Sesi** — konteks satu pelaksanaan pekerjaan oleh agen.
- **Run** — satu proses lokal yang dijalankan oleh runtime.
- **Hasil** — artefak kerja yang dihasilkan dan dicatat secara terpisah dari output proses.

## Aturan penting

`Tugas`, `Sesi`, dan `Run` tidak saling menggantikan. Tugas menyatakan pekerjaan yang harus dilakukan, sesi menyatakan konteks pelaksanaan, dan run menyatakan proses yang benar-benar dijalankan.

`Tugas` yang memiliki `ParentID` harus tetap berada pada `Pekerjaan` yang sama. `Agen` yang memiliki `ParentID` harus tetap berada pada `Ruang` yang sama dengan agen induknya. Hubungan induk agen hanya dibuat terhadap agen yang sudah ada, sehingga pembuatan hubungan awal tidak dapat membentuk siklus.

`Sesi`, `Run`, dan `Agen` juga diperiksa agar hubungan ruang dan pekerjaan tidak melampaui batas domainnya.

Status mengikuti aturan transisi domain. Penyimpanan tidak boleh membuat transisi yang tidak diizinkan hanya karena operasi SQL dapat melakukannya.

## Identitas

ID diberikan sebagai string agar model tetap sederhana dan tidak mengikat API pada format UUID tertentu. Executor menggunakan ID acak untuk sesi dan run baru.
