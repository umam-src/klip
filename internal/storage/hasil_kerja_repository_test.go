package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func siapkanKonteksHasilKerja(t *testing.T) (context.Context, *Repository) {
	t.Helper()
	ctx, repo := buatKonteksPeristiwa(t)
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateEksekusi(ctx, domain.Eksekusi{ID: "eksekusi-1", RuangID: "ruang-1", ProyekID: "proyek-1", AgenID: "agen-1", Status: domain.StatusRunning, Program: "echo", StartedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	return ctx, repo
}

func TestHasilKerjaDapatDitelusuriKeGoal(t *testing.T) {
	ctx, repo := siapkanKonteksHasilKerja(t)
	hasil := domain.HasilKerja{ID: "hasil-1", RuangID: "ruang-1", ExecutionID: idPtr("eksekusi-1"), ProyekID: "proyek-1", Kind: "file", Name: "laporan.txt", Path: "hasil/laporan.txt"}
	if err := repo.CreateHasilKerja(ctx, hasil); err != nil {
		t.Fatalf("CreateHasilKerja() error = %v", err)
	}

	got, err := repo.GetHasilKerja(ctx, hasil.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ExecutionID == nil {
		t.Fatal("hasil kerja kehilangan referensi eksekusi")
	}
	eksekusi, err := repo.GetEksekusi(ctx, *got.ExecutionID)
	if err != nil {
		t.Fatal(err)
	}
	proyek, err := repo.GetProyek(ctx, eksekusi.ProyekID)
	if err != nil {
		t.Fatal(err)
	}
	goal, err := repo.GetGoal(ctx, proyek.GoalID)
	if err != nil {
		t.Fatal(err)
	}
	if got.RuangID != "ruang-1" || eksekusi.RuangID != got.RuangID || proyek.RuangID != got.RuangID || goal.RuangID != got.RuangID {
		t.Fatalf("rantai hasil kerja melintasi ruang: hasil=%s eksekusi=%s proyek=%s goal=%s", got.RuangID, eksekusi.RuangID, proyek.RuangID, goal.RuangID)
	}
	if goal.ID != "goal-1" {
		t.Fatalf("goal = %s, want goal-1", goal.ID)
	}

	daftar, err := repo.ListHasilKerjaByEksekusi(ctx, "eksekusi-1")
	if err != nil || len(daftar) != 1 || daftar[0].ID != hasil.ID {
		t.Fatalf("ListHasilKerjaByEksekusi() = %+v, %v", daftar, err)
	}
}

func TestHasilKerjaMenolakEksekusiProyekLain(t *testing.T) {
	ctx, repo := siapkanKonteksHasilKerja(t)
	if err := repo.CreateProyek(ctx, domain.Proyek{ID: "proyek-2", RuangID: "ruang-1", GoalID: "goal-1", Title: "Proyek lain", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	err := repo.CreateHasilKerja(ctx, domain.HasilKerja{ID: "hasil-1", RuangID: "ruang-1", ExecutionID: idPtr("eksekusi-1"), ProyekID: "proyek-2", Kind: "file", Name: "a.txt", Path: "hasil/a.txt"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateHasilKerja() error = %v, want ErrInvalid", err)
	}
}

func TestHasilKerjaMenolakProyekRuangLain(t *testing.T) {
	ctx, repo := siapkanKonteksHasilKerja(t)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-2", Name: "Ruang 2"}); err != nil {
		t.Fatal(err)
	}
	err := repo.CreateHasilKerja(ctx, domain.HasilKerja{ID: "hasil-1", RuangID: "ruang-2", ExecutionID: idPtr("eksekusi-1"), ProyekID: "proyek-1", Kind: "file", Name: "a.txt", Path: "hasil/a.txt"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateHasilKerja() error = %v, want ErrInvalid", err)
	}
}

func TestHasilKerjaMenolakEksekusiTidakAda(t *testing.T) {
	ctx, repo := siapkanKonteksHasilKerja(t)
	err := repo.CreateHasilKerja(ctx, domain.HasilKerja{ID: "hasil-1", RuangID: "ruang-1", ExecutionID: idPtr("eksekusi-hilang"), ProyekID: "proyek-1", Kind: "file", Name: "a.txt", Path: "hasil/a.txt"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateHasilKerja() error = %v, want ErrInvalid", err)
	}
}
