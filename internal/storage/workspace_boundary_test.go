package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

func TestProyekRejectsCrossRuangGoal(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	for _, ruang := range []domain.Ruang{
		{ID: "ruang-1", Name: "Ruang 1"},
		{ID: "ruang-2", Name: "Ruang 2"},
	} {
		if err := repo.CreateRuang(ctx, ruang); err != nil {
			t.Fatal(err)
		}
	}

	goal := domain.Goal{ID: "goal-2", RuangID: "ruang-2", Title: "Goal ruang 2", Status: domain.GoalStatusActive}
	if err := repo.CreateGoal(ctx, goal); err != nil {
		t.Fatal(err)
	}

	err = repo.CreateProyek(ctx, domain.Proyek{
		ID:      "proyek-1",
		RuangID: "ruang-1",
		GoalID:  goal.ID,
		Title:   "Proyek lintas ruang",
		Status:  domain.StatusDraft,
	})
	if err == nil || !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected cross-ruang goal to be rejected, got %v", err)
	}
}
