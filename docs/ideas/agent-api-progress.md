# Progress API Agen

Status: Implementasi bertahap — bukan spesifikasi API final.

## Tahap 3 yang sudah tersedia

- daftar Agen tetap tersedia melalui konteks Ruang;
- detail Agen tersedia melalui `GET /api/v1/ruang/{ruang_id}/agen?agent_id={agen_id}`;
- perubahan Agen tersedia melalui `PUT /api/v1/ruang/{ruang_id}/agen?agent_id={agen_id}`;
- create Agen menerima peran dan status, dengan status default `active`;
- update menjaga parent tetap dalam Ruang yang sama;
- update menolak self-reference dan siklus hierarki;
- detail/update tidak mengekspos credential/token;
- error validasi parent dipetakan sebagai respons 400 dan Agen yang tidak ditemukan sebagai 404.

Endpoint ini sengaja mengikuti struktur routing yang sudah ada. Bentuk URL khusus Agen dapat dirapikan kemudian saat kebutuhan UI/detail Agen sudah stabil.
