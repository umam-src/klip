package storage

import (
	"context"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestRunRepositoryRoundTrip(t *testing.T) {
	db, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-run", Name: "Ruang Run", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-run", RuangID: "ruang-run", Name: "Agen Run", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-run", RuangID: "ruang-run", Title: "Pekerjaan Run", Status: domain.StatusReady, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}

	run := domain.Run{
		ID:          "run-1",
		PekerjaanID: "pekerjaan-run",
		AgenID:      "agen-run",
		Status:      domain.StatusCompleted,
		Program:     "printf",
		Arguments:   []string{"hello"},
		Stdout:      "hello",
		StartedAt:   now,
	}
	if err := repo.CreateRun(ctx, run); err != nil {
		t.Fatalf("CreateRun() error = %v", err)
	}

	got, err := repo.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got.Program != run.Program || got.Status != run.Status || got.Stdout != run.Stdout {
		t.Fatalf("GetRun() = %+v", got)
	}
	if len(got.Arguments) != 1 || got.Arguments[0] != "hello" {
		t.Fatalf("arguments = %#v", got.Arguments)
	}
	if !got.StartedAt.Equal(now) {
		t.Fatalf("started_at = %v, want %v", got.StartedAt, now)
	}

	list, err := repo.ListRunsByPekerjaan(ctx, run.PekerjaanID)
	if err != nil {
		t.Fatalf("ListRunsByPekerjaan() error = %v", err)
	}
	if len(list) != 1 || list[0].ID != run.ID {
		t.Fatalf("runs = %+v", list)
	}
}

func TestRunRepositoryRequiresRelations(t *testing.T) {
	db, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	err = repo.CreateRun(context.Background(), domain.Run{ID: "run-invalid", Program: "printf"})
	if err == nil {
		t.Fatal("CreateRun() error = nil, want error")
	}
}
