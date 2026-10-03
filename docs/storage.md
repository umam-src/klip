# Penyimpanan Klip

Klip menyimpan data utama secara lokal menggunakan SQLite. Penyimpanan dirancang sederhana agar mudah dicadangkan, dipulihkan, dan dipindahkan tanpa layanan tambahan.

## Lokasi dan koneksi

- Database dapat menggunakan path file lokal.
- Koneksi dibatasi agar SQLite tetap aman digunakan oleh runtime lokal Klip.
- Foreign key diaktifkan.
- WAL digunakan untuk menjaga perilaku baca/tulis yang baik pada penggunaan lokal.
- Batas waktu tunggu SQLite digunakan agar konflik singkat tidak langsung gagal.

## Pengaturan aplikasi

`klip.db` menjadi sumber utama pengaturan AI yang dapat diubah dari antarmuka, yaitu penyedia, alamat penyedia, dan model.

`config.json` tetap digunakan untuk bootstrap dan fallback. Urutannya:

1. `data_dir` dan `listen` dibaca dari `config.json`, lalu memakai nilai bawaan jika tidak tersedia.
2. Setelah database ditemukan, pengaturan AI dari database dipakai jika sudah tersimpan.
3. Jika database belum memiliki pengaturan AI, nilai AI dari `config.json` dipakai.
4. Jika keduanya tidak tersedia, nilai bawaan aplikasi dipakai.

Kunci API tidak dipindahkan oleh pengaturan AI ini; nilai tersebut tetap berasal dari konfigurasi yang dibaca saat mulai.

## Migrasi

Skema memiliki versi dan tabel `schema_migrations`. Perubahan skema harus dilakukan melalui migrasi yang dapat dijalankan berulang tanpa merusak database yang sudah ada.

Jangan mengubah skema produksi hanya dengan mengandalkan `CREATE TABLE IF NOT EXISTS`; perubahan kolom, indeks, atau aturan data harus memiliki langkah migrasi yang eksplisit.

Saat ini hanya ada satu versi skema, yaitu v1. `storage.Open` membuat skema pada database baru, menerima database v1 yang sudah ada tanpa mengubahnya, dan menolak database berversi lebih baru atau yang masih berisi tabel dari model lama (legacy).

## Integritas data

Lapisan penyimpanan memvalidasi hubungan penting sebelum menulis data. Contohnya:

- tugas induk harus berada pada proyek yang sama;
- agen yang menjalankan Eksekusi harus berasal dari Ruang Kerja yang sama;
- hasil yang ditautkan ke tugas harus berada pada proyek yang sama;
- perubahan status mengikuti aturan domain;
- konteks Eksekusi (Ruang Kerja, Proyek, Tugas, dan Agen) harus konsisten sebelum Eksekusi dibuat;
- Goal induk harus berada pada Ruang Kerja yang sama dan hierarki Goal tidak boleh membentuk siklus;
- Proyek harus mengacu ke satu Goal dalam Ruang Kerja yang sama;
- Hasil Kerja yang ditautkan ke Eksekusi harus berada pada Proyek dan Ruang Kerja yang sama, sehingga dapat ditelusuri sampai ke Goal.

Validasi di aplikasi melengkapi foreign key SQLite; keduanya tidak saling menggantikan.

## Backup dan restore

Gunakan API penyimpanan backup untuk membuat salinan database yang konsisten. Backup menggunakan `VACUUM INTO` sehingga database WAL dapat dipadatkan menjadi satu berkas.

Restore menulis salinan ke berkas sementara, melakukan sinkronisasi, lalu mengganti tujuan dengan operasi rename. Database tujuan tidak boleh sedang dibuka ketika restore dilakukan. Sebelum menimpa, restore memeriksa bahwa berkas cadangan adalah database Klip v1 yang utuh, lalu membersihkan berkas `-wal` dan `-shm` milik database lama.

Lihat [backup.md](backup.md) untuk prosedur dan batasannya.

## Prinsip perubahan

1. Setelah rilis stabil (1.0), utamakan kompatibilitas database lama. Selama beta, skema boleh berubah tanpa kompatibilitas ke database lama; database berskema legacy ditolak.
2. Gunakan transaksi untuk perubahan yang harus atomik.
3. Jangan menyimpan secret di database tanpa kebutuhan yang jelas.
4. Jangan menambah ORM hanya untuk mengurangi beberapa query SQL sederhana.
5. Uji migrasi dan aturan integritas sebelum perubahan dilepas.
