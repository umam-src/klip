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
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas", Status: domain.StatusReady, Position: 0}); err != nil {
		t.Fatal(err)
	}

	tugasID := domain.ID("tugas-1")
	sesi := domain.Sesi{ID: "sesi-1", PekerjaanID: "pekerjaan-1", AgenID: "agen-1", Status: domain.StatusRunning, StartedAt: fixture}
	run := domain.Run{ID: "run-1", PekerjaanID: "pekerjaan-1", TugasID: &tugasID, AgenID: "agen-1", Status: domain.StatusRunning, Program: "echo", StartedAt: fixture}
	if err := repo.CreateRun(ctx, run); err != nil {
		t.Fatal(err)
	}

	_, err = repo.StartExecution(ctx, sesi, run)
	if err == nil {
		t.Fatal("StartExecution() error = nil, want duplicate run failure")
	}

	if _, err := repo.GetSesi(ctx, sesi.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("sesi tersimpan setelah rollback: %v", err)
	}
	storedRun, err := repo.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if storedRun.Status != domain.StatusRunning {
		t.Fatalf("run status = %q, want running", storedRun.Status)
	}
	storedTask, err := repo.GetTugas(ctx, tugasID)
	if err != nil {
		t.Fatalf("GetTugas() error = %v", err)
	}
	if storedTask.Status != domain.StatusReady {
		t.Fatalf("tugas status = %q, want ready", storedTask.Status)
	}
}

func TestStartExecutionCreatesConsistentRunningState(t *testing.T) {
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
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas", Status: domain.StatusReady, Position: 0}); err != nil {
		t.Fatal(err)
	}

	tugasID := domain.ID("tugas-1")
	sesi := domain.Sesi{ID: "sesi-1", PekerjaanID: "pekerjaan-1", AgenID: "agen-1", Status: domain.StatusRunning, StartedAt: fixture}
	run := domain.Run{ID: "run-1", PekerjaanID: "pekerjaan-1", TugasID: &tugasID, AgenID: "agen-1", Status: domain.StatusRunning, Program: "echo", StartedAt: fixture}

	started, err := repo.StartExecution(ctx, sesi, run)
	if err != nil {
		t.Fatalf("StartExecution() error = %v", err)
	}
	if !started.Sesi.StartedAt.Equal(started.Run.StartedAt) {
		t.Fatalf("started timestamps differ: sesi=%v run=%v", started.Sesi.StartedAt, started.Run.StartedAt)
	}

	storedTask, err := repo.GetTugas(ctx, tugasID)
	if err != nil {
		t.Fatal(err)
	}
	if storedTask.Status != domain.StatusRunning {
		t.Fatalf("tugas status = %q, want running", storedTask.Status)
	}
	storedSession, err := repo.GetSesi(ctx, sesi.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedSession.Status != domain.StatusRunning {
		t.Fatalf("sesi status = %q, want running", storedSession.Status)
	}
	storedRun, err := repo.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedRun.Status != domain.StatusRunning {
		t.Fatalf("run status = %q, want running", storedRun.Status)
	}
}
