package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func TestExecutorUpdatesTaskLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	createExecutionFixture(t, ctx, repo)
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Jalankan perintah", Status: domain.StatusReady, Position: 0}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}

	program, args := testCommand("hello")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{PekerjaanID: "pekerjaan-1", TugasID: idPtr("tugas-1"), AgenID: "agen-1", Program: program, Arguments: args})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Run.TugasID == nil || *result.Run.TugasID != "tugas-1" {
		t.Fatalf("run task = %v", result.Run.TugasID)
	}

	tugas, err := repo.GetTugas(ctx, "tugas-1")
	if err != nil {
		t.Fatalf("GetTugas() error = %v", err)
	}
	if tugas.Status != domain.StatusCompleted {
		t.Fatalf("task status = %q, want %q", tugas.Status, domain.StatusCompleted)
	}
	if tugas.UpdatedAt.IsZero() {
		t.Fatal("task updated_at kosong")
	}
}

func TestExecutorMarksTaskFailed(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	createExecutionFixture(t, ctx, repo)
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Perintah gagal", Status: domain.StatusReady, Position: 0}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}

	program, args := testCommand("echo-error")
	_, err = (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{PekerjaanID: "pekerjaan-1", TugasID: idPtr("tugas-1"), AgenID: "agen-1", Program: program, Arguments: args})
	if !errors.Is(err, ErrProcessFailed) {
		t.Fatalf("Execute() error = %v, want ErrProcessFailed", err)
	}

	tugas, err := repo.GetTugas(ctx, "tugas-1")
	if err != nil {
		t.Fatalf("GetTugas() error = %v", err)
	}
	if tugas.Status != domain.StatusFailed {
		t.Fatalf("task status = %q, want %q", tugas.Status, domain.StatusFailed)
	}
}

func TestExecutorMarksTaskFailedOnTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	db, err := storage.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	createExecutionFixture(t, context.Background(), repo)
	if err := repo.CreateTugas(context.Background(), domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Perintah lambat", Status: domain.StatusReady, Position: 0}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}

	program, args := testCommand("sleep")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{PekerjaanID: "pekerjaan-1", TugasID: idPtr("tugas-1"), AgenID: "agen-1", Program: program, Arguments: args})
	if !errors.Is(err, ErrTimedOut) {
		t.Fatalf("Execute() error = %v, want ErrTimedOut", err)
	}
	if result.Run.Status != domain.StatusFailed || result.Sesi.Status != domain.StatusFailed {
		t.Fatalf("statuses = sesi:%q run:%q", result.Sesi.Status, result.Run.Status)
	}

	tugas, err := repo.GetTugas(context.Background(), "tugas-1")
	if err != nil {
		t.Fatalf("GetTugas() error = %v", err)
	}
	if tugas.Status != domain.StatusFailed {
		t.Fatalf("task status = %q, want %q", tugas.Status, domain.StatusFailed)
	}
}

func TestExecutorMarksTaskCancelledOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	db, err := storage.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	createExecutionFixture(t, context.Background(), repo)
	if err := repo.CreateTugas(context.Background(), domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Perintah dibatalkan", Status: domain.StatusReady, Position: 0}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}

	program, args := testCommand("hello")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{PekerjaanID: "pekerjaan-1", TugasID: idPtr("tugas-1"), AgenID: "agen-1", Program: program, Arguments: args})
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("Execute() error = %v, want ErrCancelled", err)
	}
	if result.Run.Status != domain.StatusCancelled || result.Sesi.Status != domain.StatusCancelled {
		t.Fatalf("statuses = sesi:%q run:%q", result.Sesi.Status, result.Run.Status)
	}

	tugas, err := repo.GetTugas(context.Background(), "tugas-1")
	if err != nil {
		t.Fatalf("GetTugas() error = %v", err)
	}
	if tugas.Status != domain.StatusCancelled {
		t.Fatalf("task status = %q, want %q", tugas.Status, domain.StatusCancelled)
	}
}

func TestExecutorRejectsTaskFromAnotherJob(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	createExecutionFixture(t, ctx, repo)
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-2", RuangID: "ruang-1", Title: "Pekerjaan lain", Status: domain.StatusReady}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-2", PekerjaanID: "pekerjaan-2", Title: "Tugas lain", Status: domain.StatusReady, Position: 0}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}

	program, args := testCommand("hello")
	_, err = (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{PekerjaanID: "pekerjaan-1", TugasID: idPtr("tugas-2"), AgenID: "agen-1", Program: program, Arguments: args})
	if !errors.Is(err, storage.ErrInvalid) {
		t.Fatalf("Execute() error = %v, want storage.ErrInvalid", err)
	}
}

func TestExecutorRejectsNonReadyTaskBeforeCreatingSession(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	createExecutionFixture(t, ctx, repo)
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-done", PekerjaanID: "pekerjaan-1", Title: "Sudah selesai", Status: domain.StatusCompleted, Position: 0}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}

	program, args := testCommand("hello")
	_, err = (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{PekerjaanID: "pekerjaan-1", TugasID: idPtr("tugas-done"), AgenID: "agen-1", Program: program, Arguments: args})
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

func TestExecutorRejectsAgentFromAnotherSpace(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	createExecutionFixture(t, ctx, repo)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-2", Name: "Ruang lain"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-2", RuangID: "ruang-2", Name: "Agen lain"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}

	program, args := testCommand("hello")
	_, err = (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{PekerjaanID: "pekerjaan-1", AgenID: "agen-2", Program: program, Arguments: args})
	if !errors.Is(err, storage.ErrInvalid) {
		t.Fatalf("Execute() error = %v, want storage.ErrInvalid", err)
	}
}

func createExecutionFixture(t *testing.T, ctx context.Context, repo *storage.Repository) {
	t.Helper()
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan", Status: domain.StatusReady}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}
}

func idPtr(id domain.ID) *domain.ID {
	return &id
}
