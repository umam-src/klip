package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func TestExecutorPersistsSessionAndRun(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Jalankan tugas", Status: domain.StatusReady}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}

	program, args := testCommand("hello")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{
		PekerjaanID: "pekerjaan-1",
		AgenID:      "agen-1",
		Program:     program,
		Arguments:   args,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Sesi.Status != domain.StatusCompleted || result.Run.Status != domain.StatusCompleted {
		t.Fatalf("statuses = sesi:%q run:%q", result.Sesi.Status, result.Run.Status)
	}
	if result.Run.ExitCode == nil || *result.Run.ExitCode != 0 {
		t.Fatalf("exit code = %v", result.Run.ExitCode)
	}
	if result.Run.Stdout == "" {
		t.Fatal("stdout kosong")
	}

	storedRun, err := repo.GetRun(ctx, result.Run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if storedRun.Status != domain.StatusCompleted || storedRun.PekerjaanID != "pekerjaan-1" {
		t.Fatalf("stored run = %+v", storedRun)
	}
	storedSesi, err := repo.GetSesi(ctx, result.Sesi.ID)
	if err != nil {
		t.Fatalf("GetSesi() error = %v", err)
	}
	if storedSesi.Status != domain.StatusCompleted || storedSesi.PekerjaanID != "pekerjaan-1" {
		t.Fatalf("stored sesi = %+v", storedSesi)
	}
}

func TestExecutorPersistsFailure(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	_ = repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek"})
	_ = repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"})
	_ = repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Gagal", Status: domain.StatusReady})

	program, args := testCommand("echo-error")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{
		PekerjaanID: "pekerjaan-1",
		AgenID:      "agen-1",
		Program:     program,
		Arguments:   args,
	})
	if !errors.Is(err, ErrProcessFailed) {
		t.Fatalf("Execute() error = %v, want ErrProcessFailed", err)
	}
	if result.Run.Status != domain.StatusFailed || result.Sesi.Status != domain.StatusFailed {
		t.Fatalf("statuses = sesi:%q run:%q", result.Sesi.Status, result.Run.Status)
	}
}

func TestExecutorRejectsInvalidProcessInputBeforeStart(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	_ = repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek"})
	_ = repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"})
	_ = repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Input invalid", Status: domain.StatusReady})

	_, err = (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{
		PekerjaanID: "pekerjaan-1",
		AgenID:      "agen-1",
		Program:     "printf\x00",
	})
	if !errors.Is(err, storage.ErrInvalid) {
		t.Fatalf("Execute() error = %v, want storage.ErrInvalid", err)
	}

	runs, err := repo.ListRunsByPekerjaan(ctx, "pekerjaan-1")
	if err != nil {
		t.Fatalf("ListRunsByPekerjaan() error = %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("runs = %d, want 0", len(runs))
	}
}
