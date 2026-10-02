package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestFinalizeExecutionRejectsMismatchedTaskWithoutMutation(t *testing.T) {
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
	for _, tugas := range []domain.Tugas{
		{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas 1", Status: domain.StatusRunning, Position: 0},
		{ID: "tugas-2", PekerjaanID: "pekerjaan-1", Title: "Tugas 2", Status: domain.StatusRunning, Position: 1},
	} {
		if err := repo.CreateTugas(ctx, tugas); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.CreateSesi(ctx, domain.Sesi{ID: "sesi-1", PekerjaanID: "pekerjaan-1", AgenID: "agen-1", Status: domain.StatusRunning, StartedAt: fixture}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateRun(ctx, domain.Run{ID: "run-1", PekerjaanID: "pekerjaan-1", TugasID: idPtr("tugas-1"), AgenID: "agen-1", Status: domain.StatusRunning, Program: "echo", StartedAt: fixture}); err != nil {
		t.Fatal(err)
	}

	err = repo.FinalizeExecution(ctx, "run-1", "sesi-1", idPtr("tugas-2"), domain.StatusCompleted, intPtr(0), "ok", "", fixture)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("FinalizeExecution() error = %v, want ErrInvalid", err)
	}

	for _, id := range []domain.ID{"tugas-1", "tugas-2"} {
		tugas, err := repo.GetTugas(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if tugas.Status != domain.StatusRunning {
			t.Fatalf("tugas %s status = %q, want running", id, tugas.Status)
		}
	}
	run, err := repo.GetRun(ctx, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != domain.StatusRunning || run.FinishedAt != nil {
		t.Fatalf("run berubah setelah input tugas tidak konsisten: %+v", run)
	}
}

func TestFinalizeExecutionDerivesTaskFromRun(t *testing.T) {
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

	if err := repo.FinalizeExecution(ctx, "run-1", "sesi-1", nil, domain.StatusCompleted, intPtr(0), "ok", "", fixture); err != nil {
		t.Fatalf("FinalizeExecution() error = %v", err)
	}

	tugas, err := repo.GetTugas(ctx, "tugas-1")
	if err != nil {
		t.Fatal(err)
	}
	if tugas.Status != domain.StatusCompleted {
		t.Fatalf("tugas status = %q, want completed", tugas.Status)
	}
}
