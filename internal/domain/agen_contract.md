# Kontrak Domain Agen

Status: fondasi awal.

Aturan lokal yang tidak membutuhkan database:

- ID wajib.
- Ruang wajib.
- Nama wajib.
- Agen tidak boleh menjadi induknya sendiri.
- Peran wajib diisi.
- Status, bila diisi, harus berupa nilai yang dikenal domain.

Aturan lintas data tetap berada di storage/API:

- Ruang harus ada.
- Atasan harus ada.
- Atasan harus berada pada Ruang yang sama.
- Siklus hierarki harus ditolak.
