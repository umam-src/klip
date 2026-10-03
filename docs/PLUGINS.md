# Plugin

**Status: Rencana — belum diimplementasikan**

Dokumen ini mencatat arah arsitektur Plugin Klip secara abstrak. Dokumen ini bukan spesifikasi teknis final. Detail protokol, format berkas, keamanan, dan mekanisme distribusi akan diputuskan setelah fondasi Klip stabil.

## Tujuan

Plugin dimaksudkan sebagai komponen tambahan yang berdiri di luar inti Klip. Plugin dapat dikembangkan, diuji, dirilis, dan dipasang secara terpisah tanpa menjadikan dependensinya bagian dari Klip.

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

Tool calling adalah salah satu penggunaan Plugin, bukan definisi Plugin secara keseluruhan.

## Arah Komunikasi

Plugin direncanakan dapat berkomunikasi dengan Klip melalui kontrak yang sama untuk beberapa pola penggunaan:

```text
Klip ──────→ Plugin       permintaan atau pemanggilan kemampuan
Klip ←────── Plugin       event atau pemberitahuan
Klip ←────→ Plugin        komunikasi dua arah
```

Dengan demikian, Plugin tidak harus selalu pasif menunggu dipanggil Klip. Plugin dapat menjadi sumber event atau meneruskan event dari sistem lain kepada Klip.

Salah satu contoh adalah webhook:

```text
GitHub / layanan lain
          │
          │ webhook
          ▼
       Plugin
          │
          │ event
          ▼
         Klip
```

Webhook dapat menjadi pintu masuk event ke Klip melalui Plugin. Bentuk akhirnya belum ditentukan: Plugin dapat memiliki server sendiri atau Klip dapat menyediakan jalur penerimaan webhook dan meneruskannya ke Plugin.

## Arah Integrasi

Arah awal yang dipertimbangkan adalah Plugin berjalan sebagai proses terpisah dan berkomunikasi dengan Klip melalui kontrak yang sederhana. Salah satu bentuk yang dipertimbangkan adalah komunikasi melalui input/output proses dengan pesan terstruktur.

Contoh penggunaan yang direncanakan:

```text
klip --plugin ./plugins
```

Bentuk akhir opsi CLI belum dikunci. Nama `--plugin` digunakan sebagai arah awal; dukungan beberapa lokasi atau bentuk sumber Plugin akan ditentukan kemudian.

## Plugin dan Integrasi

Plugin dan integrasi bukan konsep yang sama.

**Plugin** adalah komponen eksternal yang berinteraksi dengan Klip melalui Plugin Protocol. Plugin dapat menyediakan kemampuan, menerima permintaan, mengirim event, atau melakukan keduanya.

**Integrasi** adalah hubungan Klip dengan sistem eksternal. Integrasi dapat berupa fitur bawaan Klip atau dapat menggunakan Plugin sebagai bagian dari implementasinya.

Contohnya:

```text
Git diff untuk AI
    → Plugin

GitHub webhook → Klip
    → Integrasi, dapat menggunakan Plugin

Kirim notifikasi ke layanan chat
    → Integrasi

AI membaca dan membalas layanan chat
    → Integrasi + kemampuan Plugin/tool bila diperlukan
```

Tidak semua integrasi perlu menjadi Plugin dan tidak semua Plugin harus merupakan integrasi. Pemisahan ini menjaga agar fitur bawaan dan ekosistem eksternal tidak dipaksa menggunakan satu bentuk implementasi.

## Contoh Kemampuan

Plugin nantinya dapat menyediakan kemampuan seperti:

- pencarian dan pengambilan informasi web;
- akses API eksternal;
- Git;
- membaca dan mengubah berkas lokal;
- basis data;
- sumber event atau webhook;
- kemampuan lokal lain yang memang dibutuhkan pengguna.

Contoh tersebut belum berarti bahwa kemampuan tersebut harus dibuat sebagai Plugin resmi Klip.

## Contoh Kontrak: Sumber Pengetahuan

Knowledge Base tidak menjadi bagian inti Klip. Kebutuhan ini dipenuhi Plugin yang menyediakan sumber pengetahuan: tempat teks rujukan yang dapat dicari dan dibaca Agen. Bagian ini hanya mencatat bentuk kemampuannya, bukan protokol atau format pesan, yang tetap mengikuti Batasan Rancangan Saat Ini.

Kemampuan yang dibayangkan:

- `cari`: mencari halaman dari kata kunci, mengembalikan judul, rujukan halaman, dan cuplikan;
- `baca`: mengambil isi satu halaman sebagai teks polos dalam batas panjang tertentu;
- `tulis`: membuat atau mengubah satu halaman (opsional).

Sumber yang dipertimbangkan:

- folder catatan Markdown lokal: tanpa server dan cocok untuk offline;
- DokuWiki melalui API jarak jauhnya: tanpa database, hanya perlu PHP dan web server; API perlu diaktifkan dan dibatasi untuk satu pengguna;
- MediaWiki melalui REST API atau Action API: hanya bila wiki tersebut sudah ada atau ada kebutuhan nyata, karena memerlukan PHP dan database.

Daftar tersebut belum berarti semuanya harus menjadi Plugin resmi Klip.

Aturan yang dipegang:

1. **Baca-saja secara bawaan.** `tulis` hanya setelah Persetujuan, dan sebaiknya ke ruang draf agar manusia meninjau.
2. **Isi halaman adalah data, bukan instruksi.** Klip memberi label sumber dan membatasi panjang yang disisipkan ke prompt.
3. **Satu Ruang Kerja, satu pemisah sumber** (folder, namespace, atau wiki) agar data antar Ruang Kerja tidak tercampur.
4. **Teks polos, bukan sintaks wiki.** Pemotongan per bagian mengikuti batas panjang prompt.
5. **Gagal dengan anggun.** Bila sumber tidak tersedia, Klip tetap berjalan dengan batas waktu pendek dan Agen bekerja tanpa konteks tambahan.
6. **Tanpa embedding atau vector database** pada tahap awal; pencarian memakai kata kunci dari sumbernya.
7. **Kredensial** akun sumber (bila ada) tidak ditulis ke log atau Hasil Kerja.

## Keamanan

Plugin yang memiliki akses ke komputer lokal perlu diperlakukan sebagai komponen yang memiliki hak akses tersendiri. Rancangan akhir perlu mempertimbangkan izin, batas lokasi berkas, operasi yang diperbolehkan, persetujuan untuk tindakan berisiko, serta isolasi kegagalan Plugin.

Klip tidak seharusnya memberikan akses penuh kepada Plugin hanya karena Plugin ditemukan atau dipasang.

## Batasan Rancangan Saat Ini

Hal-hal berikut sengaja belum diputuskan:

- format manifest Plugin;
- protokol komunikasi;
- format pesan;
- dukungan request/response dan event;
- mekanisme webhook;
- apakah server webhook dimiliki Klip atau Plugin;
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
3. **Plugin tidak terikat pada provider AI.**
4. **Tool calling adalah salah satu penggunaan Plugin, bukan satu-satunya.**
5. **Plugin dapat mendukung komunikasi satu arah maupun dua arah.**
6. **Event dan webhook dapat menjadi bagian dari arah Plugin tanpa mengunci implementasinya sekarang.**
7. **Tidak menambah dependency ke Klip tanpa kebutuhan nyata.**
8. **Kemampuan lokal tetap tunduk pada izin dan batas akses.**
9. **Offline-first tetap menjadi prinsip dasar.**
10. **Mulai dari protokol sederhana sebelum membangun ekosistem distribusi.**

## Tahap Lanjutan

Rancangan Plugin baru perlu dilanjutkan setelah fondasi Klip stabil. Urutan yang disarankan:

1. menetapkan kebutuhan Plugin pertama yang nyata;
2. merancang kontrak komunikasi minimal;
3. menentukan pola request/response dan event;
4. membuat Plugin contoh di repository terpisah;
5. menguji lifecycle dan kegagalan proses;
6. merancang permission dan batas akses;
7. menghubungkan tool calling AI;
8. menguji kebutuhan webhook bila ada kasus nyata;
9. baru mempertimbangkan distribusi dan pengelolaan Plugin.

Sampai tahap tersebut dimulai, dokumen ini menjadi catatan arah dan bukan komitmen terhadap API tertentu.
