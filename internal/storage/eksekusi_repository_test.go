package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestEksekusiRepositoryRejectsMissingProject(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen", Role: "Pelaksana"}); err != nil {
		t.Fatal(err)
	}

	err = repo.CreateEksekusi(ctx, domain.Eksekusi{
		ID: "eksekusi-1", RuangID: "ruang-1", ProyekID: "proyek-tidak-ada", AgenID: "agen-1",
		Status: domain.StatusReady, Program: "echo", StartedAt: time.Now().UTC(),
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("error = %v, want ErrInvalid", err)
	}
}

func TestEksekusiRepositoryRejectsCrossWorkspaceProject(t *testing.T) {
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
	goal := domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Goal", Status: domain.GoalStatusActive}
	if err := repo.CreateGoal(ctx, goal); err != nil {
		t.Fatal(err)
	}
	proyek := domain.Proyek{ID: "proyek-1", RuangID: "ruang-1", GoalID: goal.ID, Title: "Proyek", Status: domain.StatusDraft}
	if err := repo.CreateProyek(ctx, proyek); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-2", RuangID: "ruang-2", Name: "Agen", Role: "Pelaksana"}); err != nil {
		t.Fatal(err)
	}

	err = repo.CreateEksekusi(ctx, domain.Eksekusi{
		ID: "eksekusi-1", RuangID: "ruang-2", ProyekID: proyek.ID, AgenID: "agen-2",
		Status: domain.StatusReady, Program: "echo", StartedAt: time.Now().UTC(),
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("error = %v, want ErrInvalid", err)
	}
}

func TestEksekusiRepositoryPreservesGoalTraceability(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	goal := domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Goal", Status: domain.GoalStatusActive}
	if err := repo.CreateGoal(ctx, goal); err != nil {
		t.Fatal(err)
	}
	proyek := domain.Proyek{ID: "proyek-1", RuangID: "ruang-1", GoalID: goal.ID, Title: "Proyek", Status: domain.StatusDraft}
	if err := repo.CreateProyek(ctx, proyek); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen", Role: "Pelaksana"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateEksekusi(ctx, domain.Eksekusi{
		ID: "eksekusi-1", RuangID: "ruang-1", ProyekID: proyek.ID, AgenID: "agen-1",
		Status: domain.StatusCompleted, Program: "echo", Arguments: []string{"selesai"}, StartedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetEksekusi(ctx, "eksekusi-1")
	if err != nil {
		t.Fatal(err)
	}
	var goalID, ruangID string
	if err := db.QueryRowContext(ctx, `SELECT p.goal_id, p.ruang_id FROM proyek p WHERE p.id = ?`, got.ProyekID).Scan(&goalID, &ruangID); err != nil {
		t.Fatal(err)
	}
	if goalID != string(goal.ID) || ruangID != string(goal.RuangID) || got.RuangID != goal.RuangID {
		t.Fatalf("traceability tidak sesuai: eksekusi=%+v goal=%s ruang=%s", got, goalID, ruangID)
	}
}
