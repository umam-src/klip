package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

func TestProyekRepositoryRequiresGoal(t *testing.T) {
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

	proyek := domain.Proyek{
		ID:      "proyek-1",
		RuangID: "ruang-1",
		GoalID:  "goal-tidak-ada",
		Title:   "Proyek",
		Status:  domain.StatusDraft,
	}
	if err := repo.CreateProyek(ctx, proyek); err == nil || !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestProyekRepositoryRejectsCrossWorkspaceGoal(t *testing.T) {
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
	goal := domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Goal", Status: domain.GoalStatusActive}
	if err := repo.CreateGoal(ctx, goal); err != nil {
		t.Fatal(err)
	}

	proyek := domain.Proyek{
		ID:      "proyek-1",
		RuangID: "ruang-2",
		GoalID:  goal.ID,
		Title:   "Proyek",
		Status:  domain.StatusDraft,
	}
	if err := repo.CreateProyek(ctx, proyek); err == nil || !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestProyekRepositoryKeepsWorkspaceAndGoalContext(t *testing.T) {
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
	proyek := domain.Proyek{
		ID:      "proyek-1",
		RuangID: "ruang-1",
		GoalID:  goal.ID,
		Title:   "Proyek",
		Status:  domain.StatusDraft,
	}
	if err := repo.CreateProyek(ctx, proyek); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetProyek(ctx, proyek.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.RuangID != goal.RuangID || got.GoalID != goal.ID {
		t.Fatalf("konteks proyek tidak sesuai: %+v", got)
	}
}
