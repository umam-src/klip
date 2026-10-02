# Penyedia AI

Klip memakai kontrak provider agar domain inti tidak bergantung pada vendor tertentu.

## Urutan penggunaan

1. AI lokal.
2. Endpoint OpenAI-compatible lokal atau jaringan sendiri.
3. Penyedia gratis yang dipilih pengguna.
4. Penyedia berbayar sebagai pilihan tambahan.

Model tidak pernah menjadi bagian dari binary atau Docker image Klip.

## Provider awal

### Ollama

Klip dapat menggunakan endpoint OpenAI-compatible Ollama. Contoh alamat lokal:

```text
http://127.0.0.1:11434
```

### OpenAI-compatible

Adapter generik digunakan untuk server AI yang menyediakan endpoint:

```text
POST /v1/chat/completions
```

Ini memungkinkan Klip bekerja dengan berbagai server lokal tanpa membuat adapter khusus untuk setiap server.

## Streaming

Provider OpenAI-compatible mendukung streaming Server-Sent Events (SSE) dari endpoint chat. API Klip meneruskan potongan jawaban melalui:

```text
POST /api/v1/chat
```

dengan JSON:

```json
{"prompt":"Halo","stream":true}
```

Respons menggunakan `text/event-stream`. Setiap event berisi objek `model` dan `content`; aliran ditutup dengan event `data: [DONE]`.

Streaming adalah kemampuan opsional. Provider yang belum mendukungnya mengembalikan `501 Not Implemented` pada permintaan streaming, sementara mode chat biasa tetap tersedia.

Pembatalan request memakai `context.Context`, sehingga koneksi yang dihentikan klien dapat menghentikan pembacaan dari provider.

## Keamanan

- API key tidak boleh ditulis ke log.
- Endpoint lokal menjadi pilihan default.
- Timeout wajib digunakan pada request.
- Respons dibatasi ukurannya agar provider tidak dapat menghabiskan memori Klip secara tidak sengaja.
- Jangan menyimpan kredensial dalam source code atau pengujian nyata.

## Batas saat ini

Belum ada fallback multi-provider otomatis atau adapter khusus llama.cpp/llama-server. Adapter khusus hanya ditambahkan jika kebutuhan nyata tidak dapat dipenuhi oleh kontrak OpenAI-compatible.
