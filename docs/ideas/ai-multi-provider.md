# Ide: Beberapa Penyedia AI

> **Status: Ide — belum disetujui.**
>
> Dokumen ini hanya tempat menyimpan gagasan agar tidak hilang dari percakapan. Isinya belum menjadi keputusan arsitektur, belum masuk roadmap resmi, dan belum menjadi pekerjaan implementasi.

## Latar belakang

Saat ini Klip menyimpan satu konfigurasi penyedia AI aktif dan dapat mengambil daftar model dari penyedia tersebut. Pola ini cukup untuk tahap awal.

Gagasan berikutnya adalah memungkinkan satu instalasi Klip menyimpan beberapa konfigurasi penyedia AI, lalu memilih akun/connection dan model sesuai kebutuhan.

Contoh penggunaan:

- llama.cpp di komputer rumah untuk model lokal.
- Ollama di komputer lain.
- Penyedia OpenAI-compatible lain untuk kebutuhan tertentu.
- Beberapa akun atau alamat dari jenis penyedia yang sama.
- Satu akun AI dapat digunakan oleh beberapa Agen tanpa menyalin kredensial ke setiap Agen.

## Istilah yang perlu dibedakan

Jangan menyamakan **Penyedia**, **Akun**, **Kredensial**, dan **Agen**.

```text
Penyedia AI = layanan AI
Akun        = konfigurasi akses/identitas pada penyedia
Kredensial  = rahasia yang dipakai akun untuk mengakses penyedia
Model       = model AI yang tersedia dari penyedia
Agen        = pelaksana kerja yang menggunakan akun + model
```

Contoh:

```text
Anthropic
├── Akun Claude A
│   └── credential A
│       ├── Agen Peneliti
│       └── Agen Reviewer
│
└── Akun Claude B
    └── credential B
        └── Agen Developer
```

Satu credential tidak perlu disalin ke setiap Agen. Beberapa Agen dapat menggunakan akun/connection yang sama. Beberapa akun dari Penyedia yang sama juga dapat disimpan bila diperlukan untuk memisahkan akses, kuota, billing, atau identitas.

Untuk penyedia lokal tanpa autentikasi, akun dapat tetap direpresentasikan sebagai konfigurasi akses tanpa credential.

## Gagasan model data

Pisahkan **Penyedia**, **Akun/Connection**, dan **Model yang dipilih**.

```text
Penyedia AI
├── id
├── nama
└── jenis

Akun / Connection AI
├── id
├── penyedia_id
├── nama
├── alamat
├── status aktif/nonaktif
└── kredensial (bila diperlukan)

Konfigurasi penggunaan AI
├── akun_id
└── model_id
```

Satu Penyedia dapat memiliki banyak Akun/Connection. Satu Akun/Connection dapat digunakan oleh banyak Agen. Model tetap berasal dari kemampuan discovery penyedia jika tersedia; Klip tidak perlu membuat nama alias model hanya untuk memilih model.

Hubungan yang diinginkan:

```text
Penyedia
   │
   ├── Akun A ──────┬── Agen 1
   │                 └── Agen 2
   │
   └── Akun B ───────── Agen 3
```

Bentuk relasi ini **belum menjadi skema database final**.

## Gagasan alur pengguna

### Pengaturan

```text
Pengaturan
└── Penyedia AI
    ├── Anthropic
    │   ├── Akun Claude A
    │   └── Akun Claude B
    ├── llama.cpp rumah
    └── Ollama laptop
```

Setiap Akun/Connection dapat:

1. diberi nama agar mudah dikenali;
2. memilih jenis Penyedia;
3. menyimpan alamat penyedia;
4. memeriksa koneksi;
5. mengambil daftar model;
6. memilih model yang tersedia pada saat digunakan;
7. menyimpan kredensial bila memang diperlukan.

Kredensial dikelola pada tingkat Akun/Connection, bukan disalin ke setiap Agen.

### Agen

Agen nantinya dapat memilih Akun/Connection dan Model secara terpisah:

```text
Akun AI [ Claude Peneliti ]
Model   [ claude-sonnet-... ]
```

Dua Agen dapat menggunakan Akun yang sama:

```text
Peneliti ──┐
           ├── Claude Peneliti
Reviewer ──┘
```

Atau menggunakan akun berbeda dari Penyedia yang sama:

```text
Peneliti  → Anthropic / Claude A
Developer → Anthropic / Claude B
```

Ini membuka pemisahan akses tanpa membuat hubungan 1:1 antara Agen dan credential.

## Prinsip yang perlu dipertahankan

- **Offline-first** tetap menjadi prinsip arsitektur, bukan label yang perlu ditampilkan di antarmuka pengguna.
- Konfigurasi lokal yang sudah berjalan tidak boleh rusak ketika fitur ini diperkenalkan.
- Model menggunakan ID yang diberikan penyedia; hindari alias yang tidak diperlukan.
- Discovery model tetap opsional karena tidak semua penyedia harus memiliki endpoint discovery yang sama.
- Kredensial disimpan pada Akun/Connection dan tidak menjadi data milik Agen.
- Satu Akun/Connection boleh digunakan beberapa Agen.
- Beberapa Akun/Connection boleh berasal dari Penyedia yang sama.
- Perubahan konfigurasi harus tetap aman dan tidak menulis rahasia ke log atau respons API.
- Migrasi dari konfigurasi satu penyedia harus dapat dilakukan tanpa kehilangan pengaturan yang sudah ada.
- Tambahan fitur harus sejalan dengan prinsip optimasi dan tidak menambah kompleksitas tanpa kebutuhan nyata.

## Kemungkinan bentuk arsitektur

```text
AI Providers
    │
    ├── Anthropic
    │     ├── Connection A
    │     │     ├── Agent A
    │     │     └── Agent B
    │     └── Connection B
    │           └── Agent C
    │
    ├── llama.cpp rumah
    │     └── Connection lokal
    │           └── Agent D
    │
    └── Ollama laptop
          └── Connection lokal
                └── Agent E
```

Nama `AI Connection` di sini masih berupa gagasan dan belum menjadi nama domain resmi Klip. Istilah yang tampil kepada pengguna juga belum ditetapkan; **Akun AI** mungkin lebih mudah dipahami untuk layanan yang memang menggunakan akun, sedangkan **Connection** dapat tetap menjadi istilah internal bila diperlukan.

## Yang belum diputuskan

- Apakah `Penyedia AI` dan `Akun/Connection` menjadi entitas baru atau memperluas tabel pengaturan yang ada.
- Nama domain dan istilah antarmuka untuk `Akun/Connection`.
- Apakah semua jenis penyedia membutuhkan konsep Akun atau cukup Connection.
- Bagaimana satu Akun/Connection diberikan kepada beberapa Agen.
- Apakah ada satu penyedia/model default untuk seluruh ruang kerja.
- Apakah pemilihan Akun/Connection dan Model dilakukan pada tingkat Agen, Ruang, atau keduanya.
- Bagaimana menangani model yang sudah tidak tersedia ketika konfigurasi dibuka kembali.
- Bagaimana penyedia dengan API yang tidak memiliki discovery model ditampilkan.
- Apakah Akun/Connection dapat dinonaktifkan sementara tanpa dihapus.
- Bagaimana migrasi konfigurasi satu penyedia saat skema baru diperkenalkan.
- Apakah dan kapan fitur ini perlu masuk roadmap resmi.

## Batasan dokumen ini

Dokumen ini **bukan**:

- keputusan desain final;
- spesifikasi API;
- spesifikasi skema SQLite;
- tiket implementasi;
- persetujuan untuk mengubah kode saat ini.

Implementasi baru dilakukan setelah gagasan ini disetujui dan ruang lingkupnya ditetapkan.

## Catatan keputusan

**Belum disetujui.** Jangan menganggap isi dokumen ini sebagai kontrak arsitektur Klip sampai statusnya diubah secara eksplisit.
