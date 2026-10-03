# Progres Form Agen

Status: **Implementasi bertahap — belum menjadi spesifikasi UI final.**

## Tahap 6 — Tambah/Edit Agen

Form Agen sekarang mengikuti urutan pengguna:

1. **Identitas** — nama.
2. **Fungsi** — peran dan deskripsi.
3. **Hubungan kerja** — atasan dari Ruang yang sama.
4. **AI** — penyedia dan model yang tersedia.

### Aturan

- Peran wajib diisi karena merupakan bagian dari kontrak Agen.
- Model dipilih dari daftar model provider yang tersedia.
- Atasan hanya berasal dari Agen dalam Ruang yang sama.
- Form tidak meminta credential atau token.
- Edit dilakukan dari detail Agen agar pengguna tidak kehilangan konteks.
- Penyedia dan model tetap menjadi konfigurasi Agen; desain Akun AI multi-provider belum diaktifkan pada tahap ini.
- Kegagalan daftar model tidak membuat form identitas kehilangan fungsi.

## Verifikasi UX

- Tambah Agen memakai form yang sama dengan field inti.
- Detail Agen menyediakan Edit.
- Simpan perubahan memakai endpoint update Agen yang sudah ada.
- Validasi server tetap menjadi sumber kebenaran untuk nama, peran, parent, dan status.
