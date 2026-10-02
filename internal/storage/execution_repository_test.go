package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestFinalizeExecutionRollsBackAsOneUnit(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	fixture := time.Now().UTC()
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas", Status: domain.StatusRunning}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateSesi(ctx, domain.Sesi{ID: "sesi-1", PekerjaanID: "pekerjaan-1", AgenID: "agen-1", Status: domain.StatusRunning, StartedAt: fixture}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateRun(ctx, domain.Run{ID: "run-1", PekerjaanID: "pekerjaan-1", TugasID: idPtr("tugas-1"), AgenID: "agen-1", Status: domain.StatusRunning, Program: "echo", StartedAt: fixture}); err != nil {
		t.Fatal(err)
	}

	err = repo.FinalizeExecution(ctx, "run-1", "sesi-1", idPtr("tugas-missing"), domain.StatusCompleted, intPtr(0), "ok", "", fixture)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("FinalizeExecution() error = %v, want ErrInvalid", err)
	}

	run, err := repo.GetRun(ctx, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != domain.StatusRunning || run.FinishedAt != nil {
		t.Fatalf("run berubah setelah rollback: %+v", run)
	}
	sesi, err := repo.GetSesi(ctx, "sesi-1")
	if err != nil {
		t.Fatal(err)
	}
	if sesi.Status != domain.StatusRunning || sesi.FinishedAt != nil {
		t.Fatalf("sesi berubah setelah rollback: %+v", sesi)
	}
	tugas, err := repo.GetTugas(ctx, "tugas-1")
	if err != nil {
		t.Fatal(err)
	}
	if tugas.Status != domain.StatusRunning {
		t.Fatalf("tugas berubah setelah rollback: %q", tugas.Status)
	}
}

func intPtr(value int) *int {
	return &value
}
