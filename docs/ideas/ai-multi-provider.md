# Ide: Beberapa Penyedia AI

> **Status: Ide — belum disetujui.**
>
> Dokumen ini hanya tempat menyimpan gagasan agar tidak hilang dari percakapan. Isinya belum menjadi keputusan arsitektur, belum masuk roadmap resmi, dan belum menjadi pekerjaan implementasi.

## Latar belakang

Saat ini Klip menyimpan satu konfigurasi penyedia AI aktif dan dapat mengambil daftar model dari penyedia tersebut. Pola ini cukup untuk tahap awal.

Gagasan berikutnya adalah memungkinkan satu instalasi Klip menyimpan beberapa konfigurasi penyedia AI, lalu memilih pasangan penyedia dan model sesuai kebutuhan.

Contoh penggunaan:

- llama.cpp di komputer rumah untuk model lokal.
- Ollama di komputer lain.
- Penyedia OpenAI-compatible lain untuk kebutuhan tertentu.
- Beberapa alamat atau akun dari jenis penyedia yang sama.

## Gagasan model data

Pisahkan **konfigurasi penyedia** dari **model yang dipilih**.

```text
Penyedia AI
├── id
├── nama
├── jenis
├── alamat
├── status aktif/nonaktif
└── kredensial (bila diperlukan)

Konfigurasi penggunaan AI
├── penyedia_id
└── model_id
```

Satu penyedia dapat memiliki banyak model. Model tetap berasal dari kemampuan discovery penyedia jika tersedia; Klip tidak perlu membuat nama alias model hanya untuk memilih model.

## Gagasan alur pengguna

### Pengaturan

```text
Pengaturan
└── Penyedia AI
    ├── + Tambah penyedia
    ├── llama.cpp rumah
    ├── Ollama laptop
    └── Penyedia lain
```

Setiap konfigurasi dapat:

1. diberi nama agar mudah dikenali;
2. menyimpan jenis penyedia;
3. menyimpan alamat penyedia;
4. memeriksa koneksi;
5. mengambil daftar model;
6. memilih model yang tersedia;
7. menyimpan kredensial bila memang diperlukan.

### Agen

Agen nantinya dapat memilih penyedia dan model secara terpisah:

```text
Penyedia AI [ llama.cpp rumah ]
Model      [ qwen2.5-0.5b-instruct-q4_k_m ]
```

Ini membuka kemungkinan beberapa agen menggunakan penyedia berbeda tanpa mengubah pengaturan global.

## Prinsip yang perlu dipertahankan

- **Offline-first** tetap menjadi prinsip arsitektur, bukan label yang perlu ditampilkan di antarmuka pengguna.
- Konfigurasi lokal yang sudah berjalan tidak boleh rusak ketika fitur ini diperkenalkan.
- Model menggunakan ID yang diberikan penyedia; hindari alias yang tidak diperlukan.
- Discovery model tetap opsional karena tidak semua penyedia harus memiliki endpoint discovery yang sama.
- Kredensial sebaiknya dipisahkan dari data penyedia ketika kebutuhan penyedia cloud mulai muncul.
- Perubahan konfigurasi harus tetap aman dan tidak menulis rahasia ke log atau respons API.
- Migrasi dari konfigurasi satu penyedia harus dapat dilakukan tanpa kehilangan pengaturan yang sudah ada.
- Tambahan fitur harus sejalan dengan prinsip optimasi dan tidak menambah kompleksitas tanpa kebutuhan nyata.

## Kemungkinan bentuk arsitektur

```text
AI Provider Configurations
        │
        ├── Provider A
        │     ├── Model A1
        │     └── Model A2
        │
        ├── Provider B
        │     ├── Model B1
        │     └── Model B2
        │
        └── Provider C
              └── Model C1

Agent
  └── AI Connection
        ├── provider_id
        └── model_id
```

Nama `AI Connection` di sini masih berupa gagasan dan belum menjadi nama domain resmi Klip.

## Yang belum diputuskan

- Apakah konfigurasi penyedia disimpan sebagai entitas baru atau memperluas tabel pengaturan yang ada.
- Apakah satu penyedia dapat memiliki beberapa kredensial.
- Apakah ada satu penyedia/model default untuk seluruh ruang kerja.
- Apakah pemilihan penyedia/model dilakukan pada tingkat agen, ruang kerja, atau keduanya.
- Bagaimana menangani model yang sudah tidak tersedia ketika konfigurasi dibuka kembali.
- Bagaimana penyedia dengan API yang tidak memiliki discovery model ditampilkan.
- Apakah konfigurasi penyedia dapat dinonaktifkan sementara tanpa dihapus.
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
