package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestRepositoryCoreEntities(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewRepository(db)
	created := time.Date(2026, 10, 2, 5, 0, 0, 0, time.UTC)
	updated := created.Add(time.Minute)

	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang", CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen", Role: "worker", Status: domain.AgenStatusInactive, CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Goal", Description: "Tujuan", Status: domain.GoalStatusActive, CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateProyek(ctx, domain.Proyek{ID: "proyek-1", RuangID: "ruang-1", GoalID: "goal-1", Title: "Proyek", Status: domain.StatusDraft, CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-1", RuangID: "ruang-1", ProyekID: "proyek-1", Title: "Tugas", Status: domain.StatusReady, CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatal(err)
	}
	parent := domain.ID("tugas-1")
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-2", RuangID: "ruang-1", ProyekID: "proyek-1", ParentID: &parent, Title: "Subtugas", Status: domain.StatusDraft, CreatedAt: created.Add(time.Second), UpdatedAt: updated.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateHasilKerja(ctx, domain.HasilKerja{ID: "hasil-1", RuangID: "ruang-1", ProyekID: "proyek-1", TugasID: &parent, Kind: "file", Name: "hasil.txt", Path: "hasil/hasil.txt", CreatedAt: updated}); err != nil {
		t.Fatal(err)
	}

	ruang, err := repo.GetRuang(ctx, "ruang-1")
	if err != nil || ruang.Name != "Ruang" || !ruang.CreatedAt.Equal(created) || !ruang.UpdatedAt.Equal(updated) {
		t.Fatalf("GetRuang() = %+v, error = %v", ruang, err)
	}
	agent, err := repo.GetAgen(ctx, "agen-1")
	if err != nil || agent.RuangID != "ruang-1" || agent.Role != "worker" || agent.Status != domain.AgenStatusInactive {
		t.Fatalf("GetAgen() = %+v, error = %v", agent, err)
	}
	goal, err := repo.GetGoal(ctx, "goal-1")
	if err != nil || goal.RuangID != "ruang-1" || goal.Status != domain.GoalStatusActive {
		t.Fatalf("GetGoal() = %+v, error = %v", goal, err)
	}
	proyek, err := repo.GetProyek(ctx, "proyek-1")
	if err != nil || proyek.GoalID != "goal-1" || proyek.Status != domain.StatusDraft {
		t.Fatalf("GetProyek() = %+v, error = %v", proyek, err)
	}
	tugas, err := repo.GetTugasNative(ctx, "tugas-2")
	if err != nil || tugas.ParentID == nil || *tugas.ParentID != parent || tugas.ProyekID != "proyek-1" {
		t.Fatalf("GetTugasNative() = %+v, error = %v", tugas, err)
	}
	hasil, err := repo.GetHasilKerja(ctx, "hasil-1")
	if err != nil || hasil.TugasID == nil || *hasil.TugasID != parent || hasil.Path != "hasil/hasil.txt" {
		t.Fatalf("GetHasilKerja() = %+v, error = %v", hasil, err)
	}
}

func TestRepositoryNotFoundAndValidation(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewRepository(db)
	if _, err := repo.GetRuang(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetRuang() error = %v", err)
	}
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "", Name: "tanpa id"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Goal", Status: domain.GoalStatusActive}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateProyek(ctx, domain.Proyek{ID: "proyek-1", RuangID: "ruang-1", GoalID: "goal-1", Title: "Proyek", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-1", RuangID: "ruang-1", ProyekID: "proyek-1", Title: "Tugas", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "", RuangID: "ruang-1", ProyekID: "proyek-1", Title: "tanpa id"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateTugasNative() error = %v", err)
	}
	if err := repo.CreateProyek(ctx, domain.Proyek{ID: "proyek-2", RuangID: "ruang-1", GoalID: "missing", Title: "Proyek"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateProyek(missing goal) error = %v", err)
	}
}

func TestRepositoryWorkspaceIntegrity(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewRepository(db)
	for _, ruang := range []domain.Ruang{{ID: "ruang-1", Name: "Satu"}, {ID: "ruang-2", Name: "Dua"}} {
		if err := repo.CreateRuang(ctx, ruang); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Goal 1", Status: domain.GoalStatusActive}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-2", RuangID: "ruang-2", Title: "Goal 2", Status: domain.GoalStatusActive}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateProyek(ctx, domain.Proyek{ID: "proyek-1", RuangID: "ruang-1", GoalID: "goal-1", Title: "Proyek 1", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen 1"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-2", RuangID: "ruang-2", Name: "Agen 2"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-1", RuangID: "ruang-1", ProyekID: "proyek-1", Title: "Tugas", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateEksekusi(ctx, domain.Eksekusi{ID: "eksekusi-1", RuangID: "ruang-1", ProyekID: "proyek-1", TugasID: idPtrRepo("tugas-1"), AgenID: "agen-1", Status: domain.StatusRunning, Program: "echo", StartedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateEksekusi(ctx, domain.Eksekusi{ID: "eksekusi-salah", RuangID: "ruang-1", ProyekID: "proyek-1", AgenID: "agen-2", Status: domain.StatusRunning, Program: "echo", StartedAt: time.Now().UTC()}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cross-workspace execution error = %v", err)
	}
}

func idPtrRepo(value string) *domain.ID {
	id := domain.ID(value)
	return &id
}
