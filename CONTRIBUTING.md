# Kontribusi Klip

Terima kasih telah membantu mengembangkan Klip.

## Prinsip

- Utamakan kesederhanaan.
- Optimalkan sebelum menambah fitur baru.
- Utamakan penggunaan lokal.
- Hindari dependency yang tidak diperlukan.
- Jangan memasukkan data sensitif ke repository.
- Jaga kompatibilitas data setelah versi mulai stabil.
- Dokumentasikan perubahan yang memengaruhi pengguna.

## Perubahan

Sebelum membuat perubahan besar:

1. Periksa `ROADMAP.md` dan `TODO.md`.
2. Jelaskan masalah yang ingin diselesaikan.
3. Pilih solusi dengan dependency dan kompleksitas serendah mungkin.
4. Tambahkan test yang relevan.
5. Periksa ukuran binary jika perubahan menyentuh runtime.

## Commit

Gunakan pesan commit yang jelas, misalnya:

- `feat: tambah scheduler`
- `fix: perbaiki timeout provider`
- `docs: perbarui panduan konfigurasi`
- `perf: kurangi penggunaan memori`
- `test: tambah pengujian runtime`

## Pull request

Pull request sebaiknya kecil dan terfokus. Sertakan:

- masalah yang diselesaikan;
- perubahan utama;
- pengujian yang dilakukan;
- dampak terhadap ukuran atau penggunaan memori jika ada;
- perubahan dokumentasi jika diperlukan.
