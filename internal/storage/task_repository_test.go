package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestUpdateTugasStatusEnforcesLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}

	when := time.Date(2026, 10, 2, 5, 0, 0, 0, time.UTC)
	if err := repo.UpdateTugasStatus(ctx, "tugas-1", domain.StatusRunning, when); err != nil {
		t.Fatalf("ready -> running: %v", err)
	}
	if err := repo.UpdateTugasStatus(ctx, "tugas-1", domain.StatusCompleted, when.Add(time.Minute)); err != nil {
		t.Fatalf("running -> completed: %v", err)
	}
	if err := repo.UpdateTugasStatus(ctx, "tugas-1", domain.StatusRunning, when.Add(2*time.Minute)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("completed -> running error = %v, want ErrInvalid", err)
	}

	tugas, err := repo.GetTugas(ctx, "tugas-1")
	if err != nil {
		t.Fatal(err)
	}
	if tugas.Status != domain.StatusCompleted {
		t.Fatalf("status = %q, want completed", tugas.Status)
	}

	if err := repo.UpdateTugasStatus(ctx, "missing", domain.StatusReady, when); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing task error = %v, want ErrNotFound", err)
	}
	if err := repo.UpdateTugasStatus(ctx, "tugas-1", domain.Status("unknown"), when); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown status error = %v, want ErrInvalid", err)
	}
}
