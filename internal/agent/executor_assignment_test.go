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
	db, repo := newNativeExecutionTestRepository(t, ctx)
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-2", RuangID: "ruang-1", Name: "Agen 2"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-1", RuangID: "ruang-1", ProyekID: "proyek-1", Title: "Tugas", Status: domain.StatusReady}); err != nil {
		t.Fatalf("CreateTugasNative() error = %v", err)
	}
	if err := repo.CreatePenugasan(ctx, domain.Penugasan{ID: "penugasan-1", TugasID: "tugas-1", AgenID: "agen-1", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("CreatePenugasan() error = %v", err)
	}

	program, args := testCommand("should-not-run")
	_, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{ProyekID: "proyek-1", TugasID: ptrID("tugas-1"), AgenID: "agen-2", Program: program, Arguments: args})
	if !errors.Is(err, storage.ErrInvalid) {
		t.Fatalf("Execute() error = %v, want storage.ErrInvalid", err)
	}
	if rows, err := repo.ListEksekusiByProyek(ctx, "proyek-1"); err != nil || len(rows) != 0 {
		t.Fatalf("eksekusi = %v, error = %v", rows, err)
	}
	_ = db
}

func TestExecutorRunsAssignedTaskWithMatchingAgent(t *testing.T) {
	ctx := context.Background()
	_, repo := newNativeExecutionTestRepository(t, ctx)
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-1", RuangID: "ruang-1", ProyekID: "proyek-1", Title: "Tugas", Status: domain.StatusReady}); err != nil {
		t.Fatalf("CreateTugasNative() error = %v", err)
	}
	if err := repo.CreatePenugasan(ctx, domain.Penugasan{ID: "penugasan-1", TugasID: "tugas-1", AgenID: "agen-1", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("CreatePenugasan() error = %v", err)
	}

	program, args := testCommand("hello")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{ProyekID: "proyek-1", TugasID: ptrID("tugas-1"), AgenID: "agen-1", Program: program, Arguments: args})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Eksekusi.Status != domain.StatusCompleted || result.Eksekusi.AgenID != "agen-1" {
		t.Fatalf("eksekusi = %+v", result.Eksekusi)
	}
}

func ptrID(value string) *domain.ID {
	id := domain.ID(value)
	return &id
}
