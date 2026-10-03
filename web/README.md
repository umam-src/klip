# Antarmuka Klip

Antarmuka web Klip sengaja dibuat kecil: HTML, CSS, dan JavaScript biasa tanpa CDN atau paket UI besar, sehingga tetap berjalan tanpa internet. Berkas di direktori ini dibundel ke dalam program melalui `embed`.

## Isi

- `index.html` — halaman tunggal untuk Ruang Kerja, agen, goal, proyek, tugas, komentar, dan pengaturan AI.
- `assets/app.js` — daftar dan form Ruang Kerja, agen, goal, proyek, tugas, serta komentar.
- `assets/agent-detail.js` — detail Agen.
- `assets/provider.js` dan `assets/settings.js` — status dan pilihan AI.
- `assets/app.css` — gaya tampilan.
- `handler.go` dan `embed.go` — penyajian berkas yang dibundel.

## Aturan

- Jangan memuat aset dari CDN atau alamat luar.
- Setiap skrip yang ditambahkan harus dimuat di `index.html`. Tes di `embed_test.go` memeriksa bahwa skrip dimuat dan elemen yang dibutuhkannya tersedia.
- Teks tampilan memakai Bahasa Indonesia.
