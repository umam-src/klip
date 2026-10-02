# Arsitektur Klip

## Tujuan

Klip dirancang sebagai aplikasi lokal yang dapat berjalan dari satu berkas program, memakai penyimpanan lokal, dan terhubung ke AI tanpa membawa model AI di dalam paket aplikasi.

## Bentuk runtime

```text
Klip
├── HTTP server + UI lokal
├── penyimpanan SQLite
├── runtime agen
├── scheduler sederhana
└── lapisan AI
     ├── Ollama
     ├── llama.cpp / llama-server
     └── endpoint OpenAI-compatible
```

## Batas ukuran

Ukuran diutamakan sekecil mungkin tanpa mengorbankan kualitas kode, keamanan, maintainability, atau performa.

- Berkas program: **target < 50 MiB**.
- 50–100 MiB: warning dan wajib ditinjau.
- > 100 MiB: gagal pada CI rilis.
- Docker image: target < 100 MiB.
- Model AI tidak dibundel.

Ukuran sumber kode GitHub bukan metrik ukuran produk.

## Struktur kode

```text
cmd/klip/             # pintu masuk program
internal/app/          # perakitan aplikasi dan lifecycle
internal/config/       # konfigurasi lokal
internal/storage/      # SQLite dan repository
internal/ai/           # kontrak AI dan routing
internal/agent/        # identitas dan perilaku agen
internal/work/         # pekerjaan dan tugas
internal/runtime/      # eksekusi sesi agen
internal/web/          # HTTP handler dan UI
internal/scheduler/    # jadwal lokal
internal/artifact/     # hasil dan berkas keluaran
web/templates/         # template UI
web/static/            # aset lokal tanpa CDN wajib
migrations/            # perubahan skema database
scripts/               # pemeriksaan dan build

docs/                  # keputusan dan detail teknis
```

## Lapisan AI

Kode inti tidak boleh bergantung langsung pada SDK vendor tertentu.

Kontrak awal:

```go
type AIProvider interface {
    Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
```

Klip saat ini menyediakan adapter `ollama` dan `openai-compatible`. Routing menentukan provider dan model tanpa mengubah domain pekerjaan.

## Konfigurasi lokal

Konfigurasi awal memakai JSON agar tidak menambah dependency parser. Berkasnya berada di direktori data Klip sebagai `config.json`.

Contoh minimal untuk Ollama:

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

API key hanya diperlukan untuk endpoint yang memang memerlukannya. Rahasia tidak ditulis ke log.

## Offline-first

Mode dasar harus tetap dapat digunakan tanpa internet untuk:

- membuka UI lokal;
- membaca dan menulis data;
- melihat riwayat;
- mengelola agen dan pekerjaan;
- menjalankan operasi yang tidak membutuhkan AI eksternal.

Koneksi internet hanya diperlukan jika pengguna memilih provider yang membutuhkannya. Provider lokal seperti Ollama tidak memerlukan koneksi internet setelah model tersedia di mesin pengguna.

## UI

UI v0.1 memakai HTML yang disajikan dari program dan JavaScript seminimal mungkin. Aset penting dibundel secara lokal. CDN bukan ketergantungan runtime.

## Penyimpanan

SQLite menjadi penyimpanan default. Data aplikasi berada di direktori pengguna. Nama lokasi mengikuti sistem operasi, dengan `config.json`, database, dan artefak lokal di dalam direktori data Klip.

Model AI tetap berada di luar paket Klip.

## API lokal awal

Klip menyediakan `POST /api/v1/chat` untuk percobaan AI lokal. Endpoint menerima prompt teks dan mengirimkannya ke provider yang dikonfigurasi. Endpoint hanya diikat ke alamat lokal secara default.

## Aturan ketergantungan

- Tambahkan dependency hanya jika manfaatnya jelas.
- Hindari dependency runtime besar untuk kebutuhan kecil.
- Prefer library standar Go jika cukup.
- Pisahkan adapter eksternal dari domain inti.
- Ukur ukuran binary dan image setelah perubahan yang memengaruhi runtime.
