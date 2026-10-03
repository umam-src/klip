package agent

import (
	"context"
	"testing"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

// Regresi: AppendEvent pernah tidak dipanggil kode produksi sehingga riwayat
// aktivitas selalu kosong. Executor harus mencatat peristiwa mulai dan akhir.
func TestExecutorRecordsEventsForCompletedExecution(t *testing.T) {
	ctx := context.Background()
	_, repo := newNativeExecutionTestRepository(t, ctx)

	program, args := testCommand("hello")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{ProyekID: "proyek-1", AgenID: "agen-1", Program: program, Arguments: args})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertExecutionEvents(t, ctx, repo, result.Eksekusi, domain.EventExecutionStarted, domain.EventExecutionCompleted)
}

func TestExecutorRecordsEventsForFailedExecution(t *testing.T) {
	ctx := context.Background()
	_, repo := newNativeExecutionTestRepository(t, ctx)

	program, args := testCommand("echo-error")
	result, err := (Executor{Repo: repo}).Execute(ctx, ExecutionRequest{ProyekID: "proyek-1", AgenID: "agen-1", Program: program, Arguments: args})
	if err == nil {
		t.Fatal("Execute() error = nil, want kegagalan proses")
	}

	assertExecutionEvents(t, ctx, repo, result.Eksekusi, domain.EventExecutionStarted, domain.EventExecutionFailed)
}

func assertExecutionEvents(t *testing.T, ctx context.Context, repo *storage.Repository, eksekusi domain.Eksekusi, want ...string) {
	t.Helper()
	events, err := repo.ListEventsByProyek(ctx, eksekusi.ProyekID, 50)
	if err != nil {
		t.Fatalf("ListEventsByProyek() error = %v", err)
	}
	got := map[string]int{}
	for _, event := range events {
		if event.ExecutionID == nil || *event.ExecutionID != eksekusi.ID {
			continue
		}
		got[event.Type]++
		if event.RuangID != eksekusi.RuangID {
			t.Fatalf("ruang peristiwa = %q, want %q", event.RuangID, eksekusi.RuangID)
		}
		// Isi proses tidak boleh bocor ke jejak aktivitas.
		if event.Message != "" {
			t.Fatalf("peristiwa memuat isi proses: %q", event.Message)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("jenis peristiwa = %v, want %v", got, want)
	}
	for _, jenis := range want {
		if got[jenis] != 1 {
			t.Fatalf("peristiwa %q = %d, want 1 (semua: %v)", jenis, got[jenis], got)
		}
	}
}
