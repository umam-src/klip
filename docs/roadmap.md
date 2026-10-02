# Roadmap Klip

Roadmap ini mengikuti TODO.md. Item yang belum selesai dikelompokkan sebagai issue pada Milestone #2.

## P0 — Fondasi
- Pemeriksaan ukuran Docker image

## P0 — Domain inti
- Relasi hierarki agen
- Thread dan komentar tugas
- Persetujuan
- Activity log

## P0 — AI provider
- Adapter llama.cpp/llama-server khusus bila kebutuhan native muncul
- Streaming

## P1 — UI
- Thread tugas
- Antrean persetujuan
- Riwayat aktivitas
- Halaman provider/model
- Pengaturan

## P1 — Skills dan hasil
- Format skill sederhana
- Skill lokal
- Relasi skill ke agen
- Penyimpanan hasil
- Metadata hasil
- Download/view hasil melalui UI

## P1 — Scheduler
- Jadwal lokal
- Heartbeat agen
- Job queue sederhana
- Pencegahan duplicate run
- Retry policy
- Riwayat scheduler

## P1 — Data dan keamanan
- Backup hasil
- Proteksi CSRF jika diperlukan oleh mode UI
- Audit aksi penting

## P2 — Optimasi
- Hapus dependency yang tidak diperlukan
- Optimasi static asset
- Uji Docker image <100 MiB
- Hard fail Docker image >100 MiB
- Benchmark memory idle

## P2 — CI/CD hemat menit
- Browser E2E tidak dijalankan pada setiap commit jika tidak diperlukan
- Cross-platform build pada release

## P3 — Integrasi tambahan
- Provider cloud opsional
- Git integration
- Connector tambahan berdasarkan kebutuhan nyata
- Import/export tambahan
- Integrasi chat
