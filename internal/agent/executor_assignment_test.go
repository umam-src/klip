package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func TestExecutorRequiresAssignedAgentForAssignedTask(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil { t.Fatalf("Open() error = %v", err) }
	defer db.Close()
	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek"}); err != nil { t.Fatalf("CreateRuang() error = %v", err) }
	for _, agen := range []domain.Agen{{ID: "agen-1", RuangID: "ruang-1", Name: "Agen 1"}, {ID: "agen-2", RuangID: "ruang-1", Name: "Agen 2"}} {
		if err := repo.CreateAgen(ctx, agen); err != nil { t.Fatalf("CreateAgen() error = %v", err) }
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan", Status: domain.StatusReady}); err != nil { t.Fatalf("CreatePekerjaan() error = %v", err) }
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas", Status: domain.StatusReady}); err != nil { t.Fatalf("CreateTugas() error = %v", err) }
	if err := repo.AssignTugasToAgen(ctx, "tugas-1", "agen-1", time.Now().UTC()); err != nil { t.Fatalf("AssignTugasToAgen() error = %v", err) }

	program, args := testCommand("should-not-run")
	_, err = (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{PekerjaanID: "pekerjaan-1", TugasID: ptrID("tugas-1"), AgenID: "agen-2", Program: program, Arguments: args})
	if !errors.Is(err, storage.ErrInvalid) { t.Fatalf("Execute() error = %v, want storage.ErrInvalid", err) }
	runs, err := repo.ListRunsByPekerjaan(ctx, "pekerjaan-1")
	if err != nil { t.Fatalf("ListRunsByPekerjaan() error = %v", err) }
	if len(runs) != 0 { t.Fatalf("runs = %d, want 0", len(runs)) }
}

func TestExecutorRunsAssignedTaskWithMatchingAgent(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil { t.Fatalf("Open() error = %v", err) }
	defer db.Close()
	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek"}); err != nil { t.Fatalf("CreateRuang() error = %v", err) }
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen 1"}); err != nil { t.Fatalf("CreateAgen() error = %v", err) }
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan", Status: domain.StatusReady}); err != nil { t.Fatalf("CreatePekerjaan() error = %v", err) }
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas", Status: domain.StatusReady}); err != nil { t.Fatalf("CreateTugas() error = %v", err) }
	if err := repo.AssignTugasToAgen(ctx, "tugas-1", "agen-1", time.Now().UTC()); err != nil { t.Fatalf("AssignTugasToAgen() error = %v", err) }

	// "hello" is the successful cross-platform test command.
	program, args := testCommand("hello")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{PekerjaanID: "pekerjaan-1", TugasID: ptrID("tugas-1"), AgenID: "agen-1", Program: program, Arguments: args})
	if err != nil { t.Fatalf("Execute() error = %v", err) }
	if result.Run.Status != domain.StatusCompleted || result.Run.AgenID != "agen-1" { t.Fatalf("run = %+v", result.Run) }
}

func ptrID(value string) *domain.ID { id := domain.ID(value); return &id }
