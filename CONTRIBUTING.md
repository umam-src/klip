# Kontribusi Klip

Terima kasih telah membantu mengembangkan Klip.

## Prinsip

- Utamakan kesederhanaan.
- Optimalkan sebelum menambah fitur baru.
- Utamakan penggunaan lokal dan offline.
- Hindari dependency yang tidak diperlukan.
- Jangan memasukkan data sensitif ke repository.
- Tulis kode, komentar, dan dokumen dalam Bahasa Indonesia.
- Jaga `README.md` tetap mudah dipahami orang awam; detail teknis masuk ke `docs/`.
- Jaga kompatibilitas data setelah versi mulai stabil.
- Dokumentasikan perubahan yang memengaruhi pengguna.

## Menyiapkan lingkungan

Pasang Go sesuai versi di `go.mod` (saat ini 1.26). Klip tidak memerlukan CGO, jadi tidak ada kompiler C yang perlu dipasang.

```text
go build ./cmd/klip
go run ./cmd/klip serve --data-dir ./data
```

Direktori `data/` dan berkas `config.json` sudah diabaikan Git agar data lokal tidak ikut ter-commit.

## Sebelum mengirim perubahan

Jalankan pemeriksaan yang sama dengan CI:

```text
gofmt -l .
go vet ./...
go test ./...
```

`gofmt -l .` tidak boleh menampilkan berkas apa pun.

## Perubahan

Sebelum membuat perubahan besar:

1. Periksa `docs/ROADMAP.md` dan `TODO.md`.
2. Jelaskan masalah yang ingin diselesaikan.
3. Pilih solusi dengan dependency dan kompleksitas serendah mungkin.
4. Tambahkan test yang relevan.
5. Periksa ukuran binary jika perubahan menyentuh runtime.

Setiap perubahan juga perlu:

- **Perbaikan bug disertai test regresi** yang gagal sebelum perbaikan dan lulus sesudahnya, agar kesalahan yang sama tidak terulang.
- **Catatan di `CHANGELOG.md`** pada bagian `[Unreleased]`, memakai kategori Keep a Changelog: Added, Changed, Deprecated, Removed, Fixed, atau Security.
- **Dokumentasi yang selaras**: perbarui `TODO.md` dan berkas di `docs/` bila perilaku atau keputusan berubah.

## CI

CI dirancang hemat menit:

- `CI` berjalan otomatis pada pull request dan push ke `main`, hanya menjalankan format, vet, dan test pada satu job Linux. Perubahan yang hanya menyentuh dokumen tidak memicunya, dan run lama dibatalkan saat ada commit baru.
- Build binary tidak dijalankan pada setiap commit. `Build` berjalan manual, dan `Release` hanya berjalan pada tag `v*`. Prosedurnya ada di `docs/release.md`.
- Jangan menambah job, matriks platform, atau pemeriksaan berat pada CI otomatis tanpa alasan yang kuat.

## Commit

Gunakan format `jenis: deskripsi`, atau `jenis(lingkup): deskripsi`, dengan deskripsi berbahasa Indonesia. Contoh:

- `feat: tambah scheduler`
- `fix: perbaiki timeout provider`
- `docs: perbarui panduan konfigurasi`
- `perf: kurangi penggunaan memori`
- `test: tambah pengujian runtime`
- `refactor(storage): sederhanakan repository`

## Pull request

Pull request sebaiknya kecil dan terfokus. Sertakan:

- masalah yang diselesaikan;
- perubahan utama;
- pengujian yang dilakukan;
- dampak terhadap ukuran atau penggunaan memori jika ada;
- perubahan dokumentasi jika diperlukan.
