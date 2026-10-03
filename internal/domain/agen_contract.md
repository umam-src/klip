# Kontrak Domain Agen

Status: fondasi awal.

Aturan lokal yang tidak membutuhkan database:

- ID wajib.
- Ruang wajib.
- Nama wajib.
- Agen tidak boleh menjadi induknya sendiri.

Aturan lintas data tetap berada di storage/API:

- Ruang harus ada.
- Atasan harus ada.
- Atasan harus berada pada Ruang yang sama.
- Siklus hierarki harus ditolak.

Peran dan status tampilan Agen belum menjadi bagian kontrak domain final; keduanya akan ditetapkan setelah audit storage dan UX berikutnya.
