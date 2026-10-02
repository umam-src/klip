package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

func TestAgenSkillRepository(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}

	relations := []domain.AgenSkill{
		{AgenID: "agen-1", SkillName: "ringkas"},
		{AgenID: "agen-1", SkillName: "tulis"},
	}
	for _, relation := range relations {
		if err := repo.AssignSkill(ctx, relation); err != nil {
			t.Fatalf("AssignSkill(%+v) error = %v", relation, err)
		}
	}

	got, err := repo.ListSkillsByAgen(ctx, "agen-1")
	if err != nil {
		t.Fatalf("ListSkillsByAgen() error = %v", err)
	}
	if len(got) != 2 || got[0].SkillName != "ringkas" || got[1].SkillName != "tulis" {
		t.Fatalf("ListSkillsByAgen() = %+v, want sorted relations", got)
	}

	if err := repo.RemoveSkill(ctx, relations[0]); err != nil {
		t.Fatalf("RemoveSkill() error = %v", err)
	}
	got, err = repo.ListSkillsByAgen(ctx, "agen-1")
	if err != nil {
		t.Fatalf("ListSkillsByAgen() after remove error = %v", err)
	}
	if len(got) != 1 || got[0].SkillName != "tulis" {
		t.Fatalf("ListSkillsByAgen() after remove = %+v", got)
	}
}

func TestAgenSkillRepositoryRejectsInvalidAndDuplicate(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}

	invalid := []domain.AgenSkill{
		{AgenID: "", SkillName: "ringkas"},
		{AgenID: "agen-1", SkillName: "Ringkas"},
		{AgenID: "agen-1", SkillName: "../ringkas"},
	}
	for _, relation := range invalid {
		if err := repo.AssignSkill(ctx, relation); !errors.Is(err, ErrInvalid) {
			t.Fatalf("AssignSkill(%+v) error = %v, want ErrInvalid", relation, err)
		}
	}

	relation := domain.AgenSkill{AgenID: "agen-1", SkillName: "ringkas"}
	if err := repo.AssignSkill(ctx, relation); err != nil {
		t.Fatalf("AssignSkill() error = %v", err)
	}
	if err := repo.AssignSkill(ctx, relation); err == nil {
		t.Fatal("AssignSkill() duplicate error = nil")
	}
	if err := repo.RemoveSkill(ctx, domain.AgenSkill{AgenID: "agen-1", SkillName: "tidak-ada"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("RemoveSkill() missing error = %v, want ErrNotFound", err)
	}
}

func TestAgenSkillCascadeWithAgen(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	if err := repo.AssignSkill(ctx, domain.AgenSkill{AgenID: "agen-1", SkillName: "ringkas"}); err != nil {
		t.Fatalf("AssignSkill() error = %v", err)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM agen WHERE id = ?`, "agen-1"); err != nil {
		t.Fatalf("delete agen error = %v", err)
	}
	rows, err := repo.ListSkillsByAgen(ctx, "agen-1")
	if err != nil {
		t.Fatalf("ListSkillsByAgen() after cascade error = %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("ListSkillsByAgen() after cascade = %+v, want empty", rows)
	}
}
