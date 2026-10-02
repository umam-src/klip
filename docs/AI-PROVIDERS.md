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

## Keamanan

- API key tidak boleh ditulis ke log.
- Endpoint lokal menjadi pilihan default.
- Timeout wajib digunakan pada request.
- Respons dibatasi ukurannya agar provider tidak dapat menghabiskan memori Klip secara tidak sengaja.
- Jangan menyimpan kredensial dalam source code atau pengujian nyata.

## Batas v0.1

Streaming, retry terukur, pemilihan provider otomatis, dan fallback multi-provider belum menjadi bagian dari adapter awal. Fitur tersebut ditambahkan setelah kontrak dasar stabil.
