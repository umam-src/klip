# Dokumentasi Klip

Ringkasan untuk pengguna awam ada di [README.md](../README.md). Berikut dokumentasi lengkapnya.

## Memakai Klip

- [penggunaan.md](penggunaan.md) — menjalankan dan mengonfigurasi Klip.
- [backup.md](backup.md) — cadangan dan pemulihan data.
- [AI-PROVIDERS.md](AI-PROVIDERS.md) — penyedia dan model AI.

## Rancangan dan cara kerja

- [ARCHITECTURE.md](ARCHITECTURE.md) — bentuk aplikasi, struktur kode, dan API lokal.
- [DOMAIN.md](DOMAIN.md) — istilah dan batas domain; rujukan istilah yang berlaku.
- [data-model.md](data-model.md) — rancangan model data v1.
- [runtime.md](runtime.md) — alur eksekusi lokal.
- [scheduler.md](scheduler.md) — jadwal lokal.
- [storage.md](storage.md) — penyimpanan SQLite dan migrasi.
- [security.md](security.md) — batas keamanan.
- [size-budget.md](size-budget.md) — anggaran ukuran.

## Perencanaan

- [ROADMAP.md](ROADMAP.md) — arah pengembangan per versi.
- [release.md](release.md) — prosedur rilis.
- [PLUGINS.md](PLUGINS.md) — rencana Plugin, belum diimplementasikan.
- [PARITY.md](PARITY.md) — aturan mengambil inspirasi tanpa menyalin.

## Catatan kerja

Folder [ideas/](ideas/) berisi catatan gagasan dan progres lama. Isinya tidak mengikat, dan sebagian ditulis sebelum remodel model data v1.

Istilah lama yang tidak dipakai lagi: `Sasaran` menjadi Goal, `Pekerjaan` menjadi Proyek, dan `Run` menjadi Eksekusi. Rujuk [DOMAIN.md](DOMAIN.md) dan [data-model.md](data-model.md) bila ada perbedaan.
