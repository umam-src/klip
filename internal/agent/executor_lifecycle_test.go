package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func TestExecutorPersistsTaskExecutionContext(t *testing.T) {
	ctx := context.Background()
	_, repo := newNativeExecutionTestRepository(t, ctx)
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-1", RuangID: "ruang-1", ProyekID: "proyek-1", Title: "Jalankan perintah", Status: domain.StatusReady}); err != nil {
		t.Fatalf("CreateTugasNative() error = %v", err)
	}

	program, args := testCommand("hello")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{ProyekID: "proyek-1", TugasID: ptrID("tugas-1"), AgenID: "agen-1", Program: program, Arguments: args})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Eksekusi.TugasID == nil || *result.Eksekusi.TugasID != "tugas-1" {
		t.Fatalf("task = %v", result.Eksekusi.TugasID)
	}
	stored, err := repo.GetEksekusi(ctx, result.Eksekusi.ID)
	if err != nil {
		t.Fatalf("GetEksekusi() error = %v", err)
	}
	if stored.Status != domain.StatusCompleted || stored.TugasID == nil || *stored.TugasID != "tugas-1" {
		t.Fatalf("stored = %+v", stored)
	}
}

func TestExecutorMarksTimeoutAsFailed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, repo := newNativeExecutionTestRepository(t, context.Background())

	program, args := testCommand("sleep")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{ProyekID: "proyek-1", AgenID: "agen-1", Program: program, Arguments: args})
	if !errors.Is(err, ErrTimedOut) {
		t.Fatalf("Execute() error = %v, want ErrTimedOut", err)
	}
	if result.Eksekusi.Status != domain.StatusFailed {
		t.Fatalf("status = %q, want failed", result.Eksekusi.Status)
	}
}

func TestExecutorMarksCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, repo := newNativeExecutionTestRepository(t, context.Background())

	program, args := testCommand("sleep")
	done := make(chan ExecutionResult, 1)
	errs := make(chan error, 1)
	go func() {
		result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{ProyekID: "proyek-1", AgenID: "agen-1", Program: program, Arguments: args})
		done <- result
		errs <- err
	}()

	deadline := time.Now().Add(time.Second)
	for {
		rows, err := repo.ListEksekusiByProyek(context.Background(), "proyek-1")
		if err != nil {
			t.Fatalf("ListEksekusiByProyek() error = %v", err)
		}
		if len(rows) == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("eksekusi tidak dimulai dalam batas waktu pengujian")
		}
		time.Sleep(time.Millisecond)
	}

	cancel()
	result := <-done
	err := <-errs
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("Execute() error = %v, want ErrCancelled", err)
	}
	if result.Eksekusi.Status != domain.StatusCancelled {
		t.Fatalf("status = %q, want cancelled", result.Eksekusi.Status)
	}
}

func TestExecutorRejectsTaskFromAnotherProyek(t *testing.T) {
	ctx := context.Background()
	_, repo := newNativeExecutionTestRepository(t, ctx)
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-2", RuangID: "ruang-1", ProyekID: "proyek-2", Title: "Tugas lain", Status: domain.StatusReady}); err != nil {
		// Proyek kedua belum ada sehingga repository harus menolak lintas konteks.
		if !errors.Is(err, storage.ErrInvalid) {
			t.Fatalf("CreateTugasNative() error = %v, want storage.ErrInvalid", err)
		}
		return
	}

	program, args := testCommand("hello")
	_, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{ProyekID: "proyek-1", TugasID: ptrID("tugas-2"), AgenID: "agen-1", Program: program, Arguments: args})
	if !errors.Is(err, storage.ErrInvalid) {
		t.Fatalf("Execute() error = %v, want storage.ErrInvalid", err)
	}
}

func TestExecutorRejectsNonReadyTask(t *testing.T) {
	ctx := context.Background()
	_, repo := newNativeExecutionTestRepository(t, ctx)
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-done", RuangID: "ruang-1", ProyekID: "proyek-1", Title: "Sudah selesai", Status: domain.StatusCompleted}); err != nil {
		t.Fatalf("CreateTugasNative() error = %v", err)
	}

	program, args := testCommand("hello")
	_, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{ProyekID: "proyek-1", TugasID: ptrID("tugas-done"), AgenID: "agen-1", Program: program, Arguments: args})
	if !errors.Is(err, storage.ErrInvalid) {
		t.Fatalf("Execute() error = %v, want storage.ErrInvalid", err)
	}
}

func TestExecutorRejectsAgentFromAnotherRuang(t *testing.T) {
	ctx := context.Background()
	_, repo := newNativeExecutionTestRepository(t, ctx)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-2", Name: "Ruang lain"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-2", RuangID: "ruang-2", Name: "Agen lain"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}

	program, args := testCommand("hello")
	_, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{ProyekID: "proyek-1", AgenID: "agen-2", Program: program, Arguments: args})
	if !errors.Is(err, storage.ErrInvalid) {
		t.Fatalf("Execute() error = %v, want storage.ErrInvalid", err)
	}
}
