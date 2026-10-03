package agent

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func TestExecutorPersistsEksekusi(t *testing.T) {
	ctx := context.Background()
	db, repo := newNativeExecutionTestRepository(t, ctx)
	program, args := testCommand("hello")

	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{
		ProyekID:  "proyek-1",
		AgenID:    "agen-1",
		Program:   program,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Eksekusi.Status != domain.StatusCompleted {
		t.Fatalf("status = %q", result.Eksekusi.Status)
	}
	if result.Eksekusi.ExitCode == nil || *result.Eksekusi.ExitCode != 0 {
		t.Fatalf("exit code = %v", result.Eksekusi.ExitCode)
	}
	if result.Eksekusi.Stdout == "" {
		t.Fatal("stdout kosong")
	}

	stored, err := repo.GetEksekusi(ctx, result.Eksekusi.ID)
	if err != nil {
		t.Fatalf("GetEksekusi() error = %v", err)
	}
	if stored.Status != domain.StatusCompleted || stored.ProyekID != "proyek-1" {
		t.Fatalf("stored eksekusi = %+v", stored)
	}
}

func TestExecutorPersistsFailure(t *testing.T) {
	ctx := context.Background()
	_, repo := newNativeExecutionTestRepository(t, ctx)
	program, args := testCommand("echo-error")

	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{
		ProyekID:  "proyek-1",
		AgenID:    "agen-1",
		Program:   program,
		Arguments: args,
	})
	if !errors.Is(err, ErrProcessFailed) {
		t.Fatalf("Execute() error = %v, want ErrProcessFailed", err)
	}
	if result.Eksekusi.Status != domain.StatusFailed {
		t.Fatalf("status = %q, want failed", result.Eksekusi.Status)
	}
}

func TestExecutorRejectsInvalidProcessInputBeforeStart(t *testing.T) {
	ctx := context.Background()
	_, repo := newNativeExecutionTestRepository(t, ctx)

	_, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{
		ProyekID: "proyek-1",
		AgenID:   "agen-1",
		Program:  "printf\x00",
	})
	if !errors.Is(err, storage.ErrInvalid) {
		t.Fatalf("Execute() error = %v, want storage.ErrInvalid", err)
	}

	eksekusi, err := repo.ListEksekusiByProyek(ctx, "proyek-1")
	if err != nil {
		t.Fatalf("ListEksekusiByProyek() error = %v", err)
	}
	if len(eksekusi) != 0 {
		t.Fatalf("eksekusi = %d, want 0", len(eksekusi))
	}
}

func TestExecutorRejectsRelativeWorkingDirectoryBeforeStart(t *testing.T) {
	ctx := context.Background()
	_, repo := newNativeExecutionTestRepository(t, ctx)

	program, args := testCommand("relative-dir")
	_, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{
		ProyekID: "proyek-1",
		AgenID:   "agen-1",
		Program:  program,
		Arguments: args,
		Dir:      filepath.Join("workspace", "task"),
	})
	if !errors.Is(err, storage.ErrInvalid) {
		t.Fatalf("Execute() error = %v, want storage.ErrInvalid", err)
	}

	eksekusi, err := repo.ListEksekusiByProyek(ctx, "proyek-1")
	if err != nil {
		t.Fatalf("ListEksekusiByProyek() error = %v", err)
	}
	if len(eksekusi) != 0 {
		t.Fatalf("eksekusi = %d, want 0", len(eksekusi))
	}
}

func newNativeExecutionTestRepository(t *testing.T, ctx context.Context) (*sql.DB, *storage.Repository) {
	t.Helper()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(ctx, `INSERT INTO goal (id, ruang_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, "goal-1", "ruang-1", "Goal", "", "active", now, now); err != nil {
		t.Fatalf("insert goal: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO proyek (id, ruang_id, goal_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "proyek-1", "ruang-1", "goal-1", "Proyek", "", domain.StatusReady, now, now); err != nil {
		t.Fatalf("insert proyek: %v", err)
	}
	return db, repo
}
