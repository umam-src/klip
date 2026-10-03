package storage

import (
	"context"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestTugasAgenAssignment(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	repo := NewRepository(db)

	for _, ruang := range []domain.Ruang{{ID: "ruang-1", Name: "Ruang 1"}, {ID: "ruang-2", Name: "Ruang 2"}} {
		if err := repo.CreateRuang(ctx, ruang); err != nil {
			t.Fatalf("CreateRuang() error = %v", err)
		}
	}
	for _, agen := range []domain.Agen{
		{ID: "agen-1", RuangID: "ruang-1", Name: "Agen 1", Role: "Pelaksana"},
		{ID: "agen-2", RuangID: "ruang-2", Name: "Agen 2", Role: "Pelaksana"},
		{ID: "agen-3", RuangID: "ruang-1", Name: "Agen 3", Role: "Pelaksana"},
	} {
		if err := repo.CreateAgen(ctx, agen); err != nil {
			t.Fatalf("CreateAgen() error = %v", err)
		}
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan 1", Status: domain.StatusDraft}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas 1", Status: domain.StatusDraft}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}

	assignedAt := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	if err := repo.AssignTugasToAgen(ctx, "tugas-1", "agen-1", assignedAt); err != nil {
		t.Fatalf("assign tugas: %v", err)
	}

	assignment, err := repo.GetTugasAssignment(ctx, "tugas-1")
	if err != nil {
		t.Fatalf("get assignment: %v", err)
	}
	if assignment.AgenID != "agen-1" || !assignment.AssignedAt.Equal(assignedAt) {
		t.Fatalf("assignment = %#v", assignment)
	}

	reassignedAt := assignedAt.Add(time.Minute)
	if err := repo.AssignTugasToAgen(ctx, "tugas-1", "agen-3", reassignedAt); err != nil {
		t.Fatalf("reassign tugas: %v", err)
	}
	assignment, err = repo.GetTugasAssignment(ctx, "tugas-1")
	if err != nil {
		t.Fatalf("get reassigned task: %v", err)
	}
	if assignment.AgenID != "agen-3" || !assignment.AssignedAt.Equal(reassignedAt) {
		t.Fatalf("reassigned assignment = %#v", assignment)
	}

	tasks, err := repo.ListTugasByAgen(ctx, "agen-1")
	if err != nil {
		t.Fatalf("list old agent tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("old agent tasks = %#v", tasks)
	}
	tasks, err = repo.ListTugasByAgen(ctx, "agen-3")
	if err != nil {
		t.Fatalf("list new agent tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "tugas-1" {
		t.Fatalf("new agent tasks = %#v", tasks)
	}

	if err := repo.AssignTugasToAgen(ctx, "tugas-1", "agen-2", assignedAt); err == nil {
		t.Fatal("expected cross-room assignment to fail")
	}

	if err := repo.UnassignTugas(ctx, "tugas-1"); err != nil {
		t.Fatalf("unassign task: %v", err)
	}
	if _, err := repo.GetTugasAssignment(ctx, "tugas-1"); err != ErrNotFound {
		t.Fatalf("expected missing assignment, got %v", err)
	}
}
