# Domain Klip

Dokumen ini menetapkan bahasa dan batas domain Klip agar desain produk tetap mandiri, mudah dipahami, dan tidak menjadi salinan struktur produk lain.

Rincian schema, kolom, dan relasi ada di [`data-model.md`](data-model.md), yang menjadi sumber kebenaran model data v1. Dokumen ini hanya menjelaskan makna istilah dan aturan domain, supaya keduanya tidak saling bertentangan.

## Prinsip

1. Gunakan istilah yang mudah dipahami pengguna Indonesia.
2. Pisahkan konsep produk dari detail implementasi.
3. Hindari membuat entitas hanya karena produk lain memilikinya.
4. Utamakan alur kerja lokal, sederhana, dan dapat ditelusuri.
5. Setiap entitas harus punya alasan operasional yang jelas.
6. Agen adalah pelaksana, bukan pemilik tujuan.
7. Goal menjadi pusat arah kerja; runtime hanya menjadi mekanisme pelaksanaan.
8. Batas Ruang Kerja harus dijaga pada seluruh hubungan domain.
9. Jangan mengunci Agen pada satu model, token, atau bentuk penyedia tertentu bila kebutuhan produk berkembang.

## Pusat domain

Klip berangkat dari kebutuhan untuk mencapai hasil. Karena itu **Goal** menjadi pusat semantik pekerjaan, sedangkan **Ruang Kerja** menjadi batas lingkungan tempat pekerjaan tersebut berlangsung.

```text
Ruang Kerja
    ↓
  Goal
    ↓
 Proyek
    ↓
  Tugas
    ↓
Penugasan
    ↓
  Agen
    ↓
 Eksekusi
    ↓
Hasil Kerja
```

Makna tiap lapisan:

| Istilah | Arti di Klip |
|---|---|
| **Ruang Kerja** | Lingkungan mandiri tempat Goal, Proyek, Tugas, Agen, dan Hasil Kerja dikelola bersama. |
| **Goal** | Hasil atau arah yang ingin dicapai dan menjadi pusat pengelolaan kemajuan. Goal dapat memiliki Goal turunan. |
| **Proyek** | Wadah pekerjaan yang dibuat untuk membantu mencapai satu Goal utama. |
| **Tugas** | Langkah konkret yang dapat dikerjakan dan dilacak secara terpisah. Tugas dapat memiliki Tugas turunan. |
| **Penugasan** | Hubungan yang menentukan Agen mana yang menjalankan suatu Tugas. |
| **Agen** | Pelaksana kerja berbasis AI yang memiliki identitas, peran, dan konteks. Peran seperti CEO, Planner, Manager, Worker, dan Reviewer cukup menjadi peran Agen. |
| **Eksekusi** | Satu kejadian pelaksanaan nyata yang menjadi bukti bahwa pekerjaan dijalankan. |
| **Hasil Kerja** | Keluaran atau bukti kerja yang dapat ditelusuri ke Eksekusi atau konteks kerja yang sah. |
| **Peristiwa** | Catatan kejadian lokal yang dapat ditelusuri ke Eksekusi. |
| **Skill** | Paket kemampuan atau instruksi yang dapat digunakan Agen. |
| **Runtime** | Mekanisme yang menyediakan kemampuan eksekusi bagi Agen. Runtime bukan pemilik Goal atau konteks kerja. |
| **Persetujuan** | Titik pengaman yang membutuhkan keputusan manusia sebelum langkah tertentu diteruskan. |
| **Jadwal** | Aturan waktu untuk menjalankan pekerjaan secara otomatis dalam satu Ruang Kerja. |
| **Penyedia AI** | Layanan atau proses yang menyediakan kemampuan model AI. |
| **Model AI** | Model tertentu yang digunakan oleh penyedia. |

**Board** adalah fungsi pengarah dan koordinasi, bukan entitas khusus. Board dapat membuat atau memecah Tugas dan melakukan Penugasan.

Istilah lama yang sudah tidak dipakai: `Sasaran` (kini Goal), `Pekerjaan` (kini Proyek), dan `Run` (kini Eksekusi). Klip tidak mempertahankan dua istilah untuk konsep yang sama.

## Ruang Kerja

### Definisi

**Ruang Kerja adalah lingkungan mandiri tempat Goal, Proyek, Tugas, Agen, dan Hasil Kerja dikelola bersama.**

Ruang Kerja adalah batas, bukan tujuan. Ia memastikan data dan pekerjaan yang berbeda tidak tercampur, tetapi tidak mengambil peran Goal sebagai arah produk.

Dalam kode dan penyimpanan, istilah tetap menggunakan `ruang` dan `ruang_id` agar perubahan istilah pengguna tidak memaksa migrasi teknis yang tidak diperlukan.

### Tanggung jawab Ruang Kerja

Ruang Kerja menjadi batas untuk:

- data kerja;
- Goal dan Goal turunan;
- Proyek dan Tugas;
- Agen dan hierarki Agen;
- Eksekusi, Peristiwa, dan Hasil Kerja;
- Jadwal dan Persetujuan;
- backup, export, dan restore pada tingkat ruang bila kemampuan tersebut tersedia.

Ruang Kerja tidak menjadi tempat menumpuk semua konfigurasi global aplikasi.

Status inti cukup `active` dan `archived`. Pengarsipan digunakan untuk menghentikan pekerjaan baru tanpa menghapus riwayat secara destruktif.

### Hal yang tidak menjadi bagian inti Ruang Kerja

Jangan menambahkan atribut atau entitas berikut hanya untuk memperluas makna Ruang Kerja:

- organisasi;
- perusahaan;
- tim;
- pengguna sebagai pemilik wajib;
- Goal default;
- Agen default;
- model default;
- API key;
- URL provider;
- konfigurasi runtime besar berbentuk JSON.

Kebutuhan tersebut dapat ditambahkan kemudian jika ada kebutuhan produk yang nyata.

### Aturan isolasi

Aturan berikut merupakan invarian domain dan harus dipertahankan oleh storage, service, API, scheduler, serta runtime. Daftar lengkapnya ada di bagian invarian `data-model.md`.

1. Setiap Ruang Kerja memiliki ID unik.
2. Objek yang terikat ruang tidak boleh diakses melalui ruang lain.
3. Goal induk dan Goal turunan harus berada pada Ruang Kerja yang sama.
4. Goal dan Proyek yang dihubungkan harus berada pada Ruang Kerja yang sama.
5. Proyek dan Tugas yang dihubungkan harus berada pada Ruang Kerja yang sama.
6. Tugas dan Agen yang dihubungkan melalui Penugasan harus berada pada Ruang Kerja yang sama.
7. Agen induk dan Agen anak harus berada pada Ruang Kerja yang sama.
8. Eksekusi harus dapat ditelusuri ke Ruang Kerja, Proyek, Tugas bila ada, dan Agen.
9. Jadwal hanya boleh menunjuk Proyek, Tugas, dan Agen dari Ruang Kerja yang sama; Ruang Kerja jadwal mengikuti Proyek.
10. Hierarki Goal, Tugas, dan Agen tidak boleh membentuk siklus.
11. Ruang Kerja yang diarsipkan tidak menerima pekerjaan baru kecuali diaktifkan kembali.
12. API dan service tidak boleh mengabaikan batas `ruang_id` ketika mencari atau mengubah data domain.

### Lifecycle

Ruang Kerja boleh kosong. Pengguna tidak diwajibkan membuat Goal, Agen, atau Proyek saat membuat ruang.

Satu Ruang Kerja dapat memiliki banyak Goal. Goal tidak boleh berpindah Ruang Kerja melalui operasi biasa. Jika suatu saat migrasi antar-ruang diperlukan, migrasi harus menjadi operasi eksplisit dengan validasi seluruh relasi turunannya.

Ruang Kerja yang sudah memiliki riwayat kerja sebaiknya diarsipkan daripada dihapus secara permanen.

## Goal

Goal adalah pusat arah kerja di dalam Ruang Kerja. Goal dapat memiliki Goal turunan, dan keduanya harus berada pada Ruang Kerja yang sama.

Agen tidak memiliki Goal. Agen menjalankan Tugas yang membantu mencapai Goal.

Progress Goal bukan angka yang diisi manual. Progress adalah kesimpulan keadaan yang dapat dihitung dari Proyek, Tugas, Eksekusi, dan Hasil Kerja di bawahnya. Karena itu angka `progress` tidak disimpan pada schema inti v1, dan belum ada rumus tunggal yang ditetapkan.

## Hubungan inti

```text
Ruang Kerja
│
├── Goal
│    ├── Goal turunan
│    └── Proyek
│         └── Tugas
│              └── Penugasan → Agen
│                                ├── Skill
│                                └── Eksekusi
│                                     ├── Peristiwa
│                                     └── Hasil Kerja
│
├── Jadwal
└── Persetujuan
```

## Proyek dan Tugas

**Proyek** menjawab "wadah pekerjaan apa yang dibuat untuk membantu satu Goal".

**Tugas** menjawab "langkah konkret apa yang perlu dilakukan".

Setiap Proyek mengacu langsung ke satu Goal utama, dan setiap Tugas mengacu langsung ke satu Proyek. Karena Proyek dan Tugas memiliki `ruang_id` tersendiri, service dan storage harus memastikan keduanya selalu sama dengan Ruang Kerja Goal-nya.

## Agen

Agen adalah pelaksana kerja, bukan pusat domain.

Satu Agen dapat memiliki paling banyak satu Agen induk. Agen induk dan Agen anak harus berada pada Ruang Kerja yang sama. Peran seperti CEO, Planner, Manager, Worker, dan Reviewer cukup menjadi peran pada Agen, bukan entitas tersendiri.

Agen dapat menggunakan Skill, model, dan Runtime. Hubungan tersebut menjelaskan kemampuan dan mekanisme kerja Agen; hubungan tersebut tidak menjadikan Agen pemilik Goal.

## Runtime dan Eksekusi

Runtime adalah mekanisme, bukan entitas yang menentukan arah kerja.

Eksekusi adalah kejadian pelaksanaan yang perlu dapat ditelusuri ke konteks kerja asalnya: Ruang Kerja, Proyek, Tugas bila ada, dan Agen yang menjalankan. Runtime tidak pernah menjadi pemilik Goal; konteks Goal dibawa ke Runtime melalui Proyek dan Tugas.

Desain ini memungkinkan Runtime AI berkembang tanpa mengubah Goal menjadi detail implementasi runtime.

## Hasil Kerja

Hasil Kerja adalah keluaran atau bukti kerja. Bentuknya dapat berupa:

- teks;
- berkas;
- laporan;
- data;
- status eksekusi;
- referensi ke artefak lokal;
- metadata pelaksanaan yang aman.

Hasil Kerja harus dapat ditelusuri ke Eksekusi atau konteks kerja yang sah. Hasil Kerja bukan pengganti Peristiwa, dan bukan pengganti Progress.

## Peristiwa

Peristiwa mencatat kejadian pada Eksekusi agar pengguna dapat memahami apa yang terjadi. Pencatatan bersifat best-effort dan tidak menyimpan isi proses seperti argumen, stdout, atau stderr.

Peristiwa berbeda dari Hasil Kerja:

- **Hasil Kerja** menjawab "apa yang dihasilkan atau dibuktikan".
- **Peristiwa** menjawab "apa yang terjadi".

Persetujuan dan pembuatan Jadwal bukan Peristiwa Eksekusi karena belum merupakan pelaksanaan.

## AI provider

Hubungan AI tetap dipisahkan secara konseptual:

```text
Penyedia AI
   ↓
Model AI
   ↓
Agen
```

Hubungan tersebut bukan berarti semua entitas harus langsung menjadi tabel. Implementasi hanya menambahkan entitas ketika kebutuhan nyata muncul. Pengaturan penyedia dan model bawaan disimpan di SQLite.

Kredensial bukan milik Agen dan tidak boleh dicatat di log atau Eksekusi.

## Keputusan desain

### Ruang Kerja sebagai batas, Proyek sebagai wadah pekerjaan

**Ruang Kerja** adalah batas lingkungan kerja. **Proyek** adalah wadah pekerjaan di bawah Goal. Keduanya berbeda: Ruang Kerja tidak menggantikan Proyek, dan Proyek tidak menjadi batas isolasi data.

### `ruang` tetap dipakai di kode

Perubahan istilah pengguna menjadi Ruang Kerja tidak memerlukan rename tabel atau kolom hanya demi kosmetik. Penyimpanan dan kode tetap menggunakan `ruang` serta `ruang_id`.

### Tidak ada organisasi sebagai konsep wajib

Klip dapat digunakan tanpa organisasi, tim, atau struktur administratif. Jika kebutuhan multi-pengguna muncul, akses dapat ditambahkan sebagai lapisan terpisah tanpa menjadikan struktur organisasi syarat dasar Ruang Kerja.

### Tidak ada hubungan Agen → Goal

Goal menentukan arah. Proyek dan Tugas membawa arah tersebut ke pekerjaan konkret. Agen menerima tanggung jawab melalui Penugasan, bukan dengan memiliki Goal.

### Board dan role bukan entitas

Board adalah fungsi koordinasi, dan CEO, Planner, Manager, Worker, serta Reviewer adalah peran pada Agen. Menambah entitas khusus untuk keduanya hanya memperbesar model tanpa manfaat inti.

### Tidak ada hard-delete sebagai lifecycle utama Ruang Kerja

Riwayat kerja perlu tetap dapat ditelusuri. Karena itu pengarsipan menjadi lifecycle utama; penghapusan permanen, jika dibutuhkan, harus menjadi operasi terkontrol dan mempertimbangkan seluruh data turunan.

## Batas v1

Konsep berikut tidak menjadi schema inti v1: organisasi, tim, penagihan, marketplace, basis pengetahuan, percakapan global, orkestrasi multi-provider kompleks, entitas Board atau CEO khusus, dan penyimpanan progress numerik. Alur dan Sesi juga bukan entitas inti; keduanya baru dipertimbangkan bila kebutuhan operasionalnya terbukti. Daftar lengkap ada di `data-model.md`.

Urutan pengembangan yang diutamakan:

1. jaga Ruang Kerja sebagai batas domain pada seluruh relasi;
2. pastikan Proyek dan Tugas dapat ditelusuri ke Goal;
3. pastikan Penugasan menghubungkan Tugas dengan Agen;
4. pastikan Eksekusi dan Hasil Kerja dapat ditelusuri kembali ke konteks tersebut;
5. baru perluas Runtime AI dengan konteks Goal yang jelas.

Prinsip akhirnya:

> **Ruang Kerja membatasi sistem. Goal mengarahkan sistem. Agen menjalankan sistem. Runtime mengeksekusi sistem. Hasil Kerja membantu mengukur kemajuan Goal.**
