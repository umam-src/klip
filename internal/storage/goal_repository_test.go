package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

func TestGoalRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	root := domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Akar", Status: domain.GoalStatusActive}
	if err := repo.CreateGoal(ctx, root); err != nil {
		t.Fatal(err)
	}
	child := domain.Goal{ID: "goal-2", RuangID: "ruang-1", ParentGoalID: &root.ID, Title: "Turunan", Status: domain.GoalStatusActive}
	if err := repo.CreateGoal(ctx, child); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetGoal(ctx, child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ParentGoalID == nil || *got.ParentGoalID != root.ID || got.Title != child.Title {
		t.Fatalf("goal tidak sesuai: %+v", got)
	}

	items, err := repo.ListGoalByRuang(ctx, root.RuangID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("jumlah goal tidak sesuai: %d", len(items))
	}

	child.Status = domain.GoalStatusCompleted
	if err := repo.UpdateGoal(ctx, child); err != nil {
		t.Fatal(err)
	}
	got, err = repo.GetGoal(ctx, child.ID)
	if err != nil || got.Status != domain.GoalStatusCompleted {
		t.Fatalf("status goal tidak berubah: %+v, %v", got, err)
	}
}

func TestGoalRepositoryRejectsCrossWorkspaceParent(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, "file::memory:?cache=shared")
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
	parent := domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Induk", Status: domain.GoalStatusActive}
	if err := repo.CreateGoal(ctx, parent); err != nil {
		t.Fatal(err)
	}
	child := domain.Goal{ID: "goal-2", RuangID: "ruang-2", ParentGoalID: &parent.ID, Title: "Anak", Status: domain.GoalStatusActive}
	if err := repo.CreateGoal(ctx, child); err == nil || !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestGoalRepositoryRejectsSelfParent(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	goal := domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Goal", Status: domain.GoalStatusActive}
	goal.ParentGoalID = &goal.ID
	if err := repo.CreateGoal(ctx, goal); err == nil || !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestGoalRepositoryRejectsParentCycleOnUpdate(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	root := domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Akar", Status: domain.GoalStatusActive}
	child := domain.Goal{ID: "goal-2", RuangID: "ruang-1", ParentGoalID: &root.ID, Title: "Anak", Status: domain.GoalStatusActive}
	if err := repo.CreateGoal(ctx, root); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, child); err != nil {
		t.Fatal(err)
	}

	root.ParentGoalID = &child.ID
	if err := repo.UpdateGoal(ctx, root); err == nil || !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestGoalRepositoryRejectsInvalidTransition(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, "file::memory:?cache=shared")
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
	goal.Status = domain.GoalStatusActive
	if err := repo.UpdateGoal(ctx, goal); err != nil {
		t.Fatal(err)
	}
	goal.Status = domain.GoalStatusCompleted
	if err := repo.UpdateGoal(ctx, goal); err != nil {
		t.Fatal(err)
	}
	goal.Status = domain.GoalStatusActive
	if err := repo.UpdateGoal(ctx, goal); err == nil || !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
