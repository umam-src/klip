# Klip

<p align="center">
  <img src="assets/klip-icon.svg" alt="Ikon Klip" width="160">
</p>

Klip adalah aplikasi ringan untuk mengatur pekerjaan berbasis AI.

Klip dibuat dengan pendekatan **lokal terlebih dahulu**. AI yang berjalan di komputer sendiri menjadi pilihan utama. Layanan AI gratis dapat digunakan jika diperlukan, sedangkan layanan berbayar bersifat opsional.

## Tujuan

- Ringan dan cepat dijalankan.
- Dapat digunakan tanpa bergantung pada layanan cloud.
- Data utama disimpan secara lokal.
- Mendukung AI lokal seperti Ollama dan llama.cpp.
- Memiliki ruang kerja, agen, sasaran, pekerjaan, tugas, sesi, dan hasil.
- Mudah dipasang dan dipindahkan.
- Bahasa Indonesia sebagai bahasa bawaan.

## Ukuran

Ukuran yang menjadi sasaran adalah ukuran aplikasi yang digunakan, bukan ukuran kode sumber GitHub.

- Binary: target di bawah 50 MB.
- Docker image: target di bawah 100 MB.
- 500 MB adalah batas keras artefak runtime.
- Model AI tidak termasuk ke dalam binary atau Docker image Klip.

Model dapat dikelola secara terpisah oleh Ollama, llama.cpp, atau layanan AI lain.

## AI

Klip tidak mengharuskan pengguna membeli layanan AI tertentu.

Prioritas penggunaan:

1. AI lokal.
2. Layanan AI gratis.
3. Layanan AI berbayar sebagai pilihan pengguna.

Provider awal yang tersedia adalah Ollama dan endpoint yang kompatibel dengan OpenAI API.

## Menjalankan

Saat dijalankan tanpa perintah, Klip langsung menjalankan server lokal dan menggunakan data lokal. Server hanya tersedia dari komputer sendiri secara default.

```text
klip
klip serve --data-dir ./data
klip serve --listen 127.0.0.1:8788
klip version
klip help
```

Jika model belum dipilih, Klip tetap dapat berjalan untuk fungsi lokal. Untuk mengaktifkan AI, isi model pada `config.json` di direktori data Klip.

Contoh:

```json
{
  "listen": "127.0.0.1:8787",
  "ai": {
    "provider": "ollama",
    "base_url": "http://127.0.0.1:11434",
    "model": "nama-model-lokal"
  }
}
```

Endpoint percobaan AI lokal tersedia di `POST /api/v1/chat`.

## Status

Klip sudah memiliki rilis awal `0.1.0` dan masih dalam tahap pengembangan. Struktur dan fungsi dapat berubah sebelum versi 1.0.

## Dokumentasi

- `ROADMAP.md` — arah pengembangan.
- `TODO.md` — pekerjaan yang direncanakan.
- `CHANGELOG.md` — riwayat perubahan.
- `CONTRIBUTING.md` — panduan kontribusi.
- `AGENTS.md` — catatan dan aturan pengembangan.
- `docs/` — dokumentasi teknis.

## Lisensi

Klip menggunakan lisensi MIT.
