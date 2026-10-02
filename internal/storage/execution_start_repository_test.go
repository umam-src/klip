package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestStartExecutionRollsBackAsOneUnit(t *testing.T) {
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

	tugasID := domain.ID("tugas-missing")
	sesi := domain.Sesi{ID: "sesi-1", PekerjaanID: "pekerjaan-1", AgenID: "agen-1", Status: domain.StatusRunning, StartedAt: fixture}
	run := domain.Run{ID: "run-1", PekerjaanID: "pekerjaan-1", TugasID: &tugasID, AgenID: "agen-1", Status: domain.StatusRunning, Program: "echo", StartedAt: fixture}

	_, err = repo.StartExecution(ctx, sesi, run)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("StartExecution() error = %v, want ErrNotFound", err)
	}

	if _, err := repo.GetSesi(ctx, sesi.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("sesi tersimpan setelah rollback: %v", err)
	}
	if _, err := repo.GetRun(ctx, run.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("run tersimpan setelah rollback: %v", err)
	}
}
