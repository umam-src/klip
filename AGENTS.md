# AGENTS.md

Catatan pengembangan dan aturan untuk menjaga arah proyek Klip.

## Aturan utama

1. Jangan menjadikan dependency besar sebagai pilihan pertama.
2. Optimalkan ukuran dan penggunaan memori sejak awal.
3. Jangan memasukkan model AI ke binary atau Docker image Klip.
4. Utamakan AI lokal dan endpoint yang kompatibel dengan API umum.
5. Jangan mengikat inti aplikasi pada satu vendor AI.
6. Bahasa pengguna default adalah Bahasa Indonesia.
7. Data pengguna harus disimpan lokal secara aman dan tidak dikirim tanpa tindakan yang jelas dari pengguna.
8. Hindari logging secret, token, prompt sensitif, dan isi artifact secara tidak perlu.
9. Perubahan besar harus dicatat di dokumentasi dan changelog.
10. Jangan menambah fitur hanya karena tersedia di proyek sumber jika fitur tersebut memperbesar aplikasi tanpa manfaat inti.

## Penamaan Bahasa Indonesia

Prioritaskan Bahasa Indonesia untuk penamaan kode dan struktur proyek jika istilahnya jelas dan tidak bertentangan dengan standar ekosistem.

* Gunakan Bahasa Indonesia untuk nama konsep domain, tipe, struktur, fungsi, metode, variabel, file, dan direktori jika memungkinkan.
* Gunakan `snake_case` untuk nama skema atau kontrak data yang memang menggunakan bentuk tersebut.
* Pertahankan istilah Inggris jika merupakan istilah standar bahasa pemrograman, pustaka, protokol, API eksternal, nama resmi pihak ketiga, atau istilah yang jauh lebih jelas dalam bahasa aslinya.
* Jangan menerjemahkan istilah secara paksa jika hasilnya menjadi ambigu atau sulit dipahami.
* Jangan membuat dua nama untuk satu konsep hanya karena sebagian kode lama menggunakan Bahasa Inggris.
* Saat merombak kode lama, gunakan istilah Bahasa Indonesia yang sudah disepakati daripada menambah lapisan kompatibilitas permanen.
* Nama domain harus mengikuti kosakata yang sudah dikunci dalam `docs/data-model.md`.

Kosakata domain v1:

* `ruang` — Ruang Kerja
* `goal` — Goal
* `proyek` — Proyek
* `tugas` — Tugas
* `penugasan` — Penugasan
* `agen` — Agen
* `eksekusi` — Eksekusi
* `peristiwa` — Peristiwa
* `hasil_kerja` — Hasil Kerja
* `persetujuan` — Persetujuan
* `komentar` — Komentar
* `jadwal` — Jadwal
* `jadwal_eksekusi` — Jadwal Eksekusi
* `pengaturan_ai` — Pengaturan AI
* `saran` — Saran

## Prinsip belajar dari kesalahan

Setiap masalah yang berulang harus menghasilkan salah satu dari:

- perbaikan kode;
- test regresi;
- aturan dokumentasi;
- atau pemeriksaan CI.

Tujuannya agar masalah yang sama tidak perlu diselesaikan dua kali.

## Ukuran runtime

Ukuran yang diperhatikan:

- binary release;
- Docker image release;
- penggunaan memori;
- waktu mulai.

Ukuran source code dan ukuran clone repository bukan metrik utama untuk batas aplikasi.

Benchmark memori idle menggunakan heap sebagai metrik portabel. Jangan menyamakan heap dengan RSS dan jangan membuat batas keras dari satu hasil benchmark tanpa pembanding yang konsisten.

## AI provider

Core Klip harus menggunakan interface provider. Adapter provider berada di luar domain inti. Local AI harus dapat digunakan tanpa akun cloud.

## CI

CI harus hemat menit:

- cache dependency;
- hindari matrix berlebihan;
- test utama satu kali per pull request;
- build lintas platform terutama pada release;
- pemeriksaan ukuran dilakukan setelah build;
- E2E berat tidak dijalankan tanpa alasan pada setiap perubahan kecil;
- batalkan run lama pada ref yang sama ketika commit baru sudah memicu run yang lebih relevan;
- gunakan permission minimum untuk workflow.

## Pelajaran CI

- Jangan menganggap perubahan Go sudah terformat hanya karena perubahan kecil; sebelum push, jalankan `gofmt -w` pada file Go yang disentuh.
- Pertahankan pemeriksaan `gofmt -l .` di CI agar format tidak kembali rusak.
- Benchmark memori tidak dijalankan pada setiap CI karena hasilnya sensitif terhadap lingkungan runner; jalankan saat investigasi optimasi dan bandingkan pada lingkungan yang konsisten.
- Pada test table-driven, setiap kasus harus mengisi seluruh parameter yang memengaruhi perilaku yang diuji; jangan memakai nilai contoh dari kasus lain untuk skenario kosong/tidak terkonfigurasi.
- Linker flag release hanya boleh mengatur simbol yang benar-benar ada di binary; jangan menambahkan `-X` versi tanpa variable string target karena release build harus dapat diverifikasi lintas platform.
- CI dapat menghemat runner minutes dengan `concurrency.cancel-in-progress`, tetapi jangan menggantinya dengan filter path yang berisiko membuat required check tidak pernah muncul.
- Migrasi skema harus idempotent: skema instalasi baru dan database lama harus dapat melewati pemeriksaan kolom/index yang sama tanpa menjalankan `ALTER TABLE` dua kali.
- Relasi hierarki baru harus memvalidasi batas domain sebelum menulis foreign key; untuk hubungan induk-anak agen, parent harus berada di ruang yang sama dan operasi awal tidak boleh membuka jalur siklus.
- Streaming provider harus menjadi kemampuan opsional agar provider lama tidak dipaksa mengimplementasikan API baru; parsing SSE harus dibatasi ukurannya dan tetap menghormati pembatalan context.
- Pengaturan yang diubah dari UI harus disimpan di SQLite sebagai sumber utama; `config.json` hanya menjadi bootstrap/fallback untuk nilai yang belum tersimpan di database.
