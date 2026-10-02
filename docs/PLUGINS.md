# Plugin

**Status: Rencana — belum diimplementasikan**

Dokumen ini mencatat arah arsitektur Plugin Klip secara abstrak. Dokumen ini bukan spesifikasi teknis final. Detail protokol, format berkas, keamanan, dan mekanisme distribusi akan diputuskan setelah fondasi Klip stabil.

## Tujuan

Plugin dimaksudkan sebagai kemampuan tambahan yang berdiri di luar inti Klip. Plugin dapat dikembangkan, diuji, dirilis, dan dipasang secara terpisah tanpa menjadikan dependensinya bagian dari Klip.

Klip menyediakan satu cara umum untuk menemukan dan berkomunikasi dengan Plugin. Dengan demikian, inti Klip tetap kecil, lokal, dan tidak bergantung pada penyedia layanan tertentu.

## Bentuk Konsep

```text
Klip
 │
 ├── Inti Klip
 │
 └── Pengelola Plugin
       │
       ├── Plugin Web
       ├── Plugin Git
       ├── Plugin Berkas
       ├── Plugin API
       └── Plugin lain
```

Plugin bukan bagian dari domain inti Klip. Repository, siklus pengembangan, versi, dan rilis Plugin dapat berdiri sendiri.

## Hubungan dengan AI

AI dapat menggunakan kemampuan Plugin melalui mekanisme pemanggilan fungsi (tool calling). Tool bukan identitas Plugin; tool adalah operasi yang diberikan Plugin kepada Klip.

```text
AI
 │
 └── pemanggilan tool
          │
          ▼
    Registry kemampuan Klip
          │
          ▼
       Plugin
          │
          ▼
        hasil
```

Provider AI seperti Gemini, OpenAI-compatible, atau Ollama tidak menjadi bagian dari Plugin. Plugin juga tidak perlu bergantung pada provider AI tertentu.

## Arah Integrasi

Arah awal yang dipertimbangkan adalah Plugin berjalan sebagai proses terpisah dan berkomunikasi dengan Klip melalui kontrak yang sederhana. Salah satu bentuk yang dipertimbangkan adalah komunikasi melalui input/output proses dengan pesan terstruktur.

Contoh penggunaan yang direncanakan:

```text
klip --plugin ./plugins
```

Bentuk akhir opsi CLI belum dikunci. Nama `--plugin` digunakan sebagai arah awal; dukungan beberapa lokasi atau bentuk sumber Plugin akan ditentukan kemudian.

## Contoh Kemampuan

Plugin nantinya dapat menyediakan kemampuan seperti:

- pencarian dan pengambilan informasi web;
- akses API eksternal;
- Git;
- membaca dan mengubah berkas lokal;
- basis data;
- kemampuan lokal lain yang memang dibutuhkan pengguna.

Contoh tersebut belum berarti bahwa kemampuan tersebut harus dibuat sebagai Plugin resmi Klip.

## Keamanan

Plugin yang memiliki akses ke komputer lokal perlu diperlakukan sebagai komponen yang memiliki hak akses tersendiri. Rancangan akhir perlu mempertimbangkan izin, batas lokasi berkas, operasi yang diperbolehkan, persetujuan untuk tindakan berisiko, serta isolasi kegagalan Plugin.

Klip tidak seharusnya memberikan akses penuh kepada Plugin hanya karena Plugin ditemukan atau dipasang.

## Batasan Rancangan Saat Ini

Hal-hal berikut sengaja belum diputuskan:

- format manifest Plugin;
- protokol komunikasi;
- format pesan;
- lifecycle Plugin;
- versi protokol dan kompatibilitas;
- mekanisme izin;
- sandbox atau isolasi;
- pemasangan dan penghapusan Plugin;
- distribusi atau registry Plugin;
- pembaruan Plugin;
- penandatanganan dan verifikasi Plugin;
- apakah Plugin harus berupa satu berkas, direktori, atau bentuk lain;
- mekanisme pengembangan dan pengujian lintas bahasa.

Keputusan tersebut dibuat setelah kebutuhan nyata muncul, bukan sebagai prasyarat implementasi awal.

## Prinsip

1. **Plugin terpisah dari inti Klip.**
2. **Kontrak lebih penting daripada implementasi tertentu.**
3. **Tidak mengikat Plugin pada provider AI.**
4. **Tidak menambah dependency ke Klip tanpa kebutuhan nyata.**
5. **Kemampuan lokal tetap tunduk pada izin dan batas akses.**
6. **Offline-first tetap menjadi prinsip dasar.**
7. **Mulai dari protokol sederhana sebelum membangun ekosistem distribusi.**

## Tahap Lanjutan

Rancangan Plugin baru perlu dilanjutkan setelah fondasi Klip stabil. Urutan yang disarankan:

1. menetapkan kebutuhan Plugin pertama yang nyata;
2. merancang kontrak komunikasi minimal;
3. membuat Plugin contoh di repository terpisah;
4. menguji lifecycle dan kegagalan proses;
5. merancang permission;
6. menghubungkan tool calling AI;
7. baru mempertimbangkan distribusi dan pengelolaan Plugin.

Sampai tahap tersebut dimulai, dokumen ini menjadi catatan arah dan bukan komitmen terhadap API tertentu.
