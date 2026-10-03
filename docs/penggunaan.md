# Penggunaan Klip

Panduan teknis untuk menjalankan dan mengonfigurasi Klip. Ringkasan untuk pengguna awam ada di `README.md`.

## Target ukuran

Ukuran yang diperhatikan adalah ukuran artefak runtime, bukan ukuran source code atau clone repository.

- Binary release: target di bawah 50 MiB.
- Docker image release: target di bawah 100 MiB.
- Batas keras artefak runtime: 100 MiB.
- Model AI tidak dibundel ke binary atau Docker image Klip.

Model dikelola terpisah oleh Ollama, llama.cpp, atau layanan AI lain. Detail anggaran ukuran ada di `size-budget.md`.

## Prioritas AI

1. AI lokal.
2. Layanan AI gratis.
3. Layanan AI berbayar sebagai pilihan pengguna.

Provider awal adalah Ollama dan endpoint yang kompatibel dengan OpenAI API. Detail provider ada di `AI-PROVIDERS.md`.

Untuk `llama.cpp`, gunakan provider `openai-compatible`. Nilai model harus mengikuti ID model dari `/v1/models` pada server, bukan nama berkas model.

## Menjalankan

Tanpa perintah, Klip menjalankan server lokal dengan data lokal. Secara default server hanya dapat diakses dari komputer yang sama.

```text
klip
klip serve --data-dir ./data
klip serve --listen 127.0.0.1:8788
klip version
klip help
```

Opsi `--data-dir` dan `--listen` hanya berlaku untuk proses tersebut dan tidak menulis ulang `config.json`.

Jika model belum dipilih, Klip tetap berjalan untuk fungsi lokal. Untuk mengaktifkan AI, isi model pada `config.json` di direktori data, atau ubah dari halaman pengaturan (disimpan di SQLite dan menjadi sumber utama).

## Contoh konfigurasi

Ollama:

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

Server `llama.cpp` (API OpenAI-compatible):

```json
{
  "ai": {
    "provider": "openai-compatible",
    "base_url": "http://127.0.0.1:8080",
    "model": "id-model-dari-v1-models"
  }
}
```

Endpoint percobaan AI lokal: `POST /api/v1/chat`.
