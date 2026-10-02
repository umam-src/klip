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
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	created := time.Date(2026, 10, 2, 5, 0, 0, 0, time.UTC)
	updated := created.Add(time.Minute)

	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek", CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen Lokal", ProviderID: "ollama", ModelID: "qwen", CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Bangun inti", Status: domain.StatusReady, CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Langkah pertama", Status: domain.StatusReady, Position: 1, CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}
	parent := domain.ID("tugas-1")
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-2", PekerjaanID: "pekerjaan-1", ParentID: &parent, Title: "Langkah kedua", Status: domain.StatusDraft, Position: 2, CreatedAt: created.Add(time.Second), UpdatedAt: updated.Add(time.Second)}); err != nil {
		t.Fatalf("CreateTugas(parent) error = %v", err)
	}
	if err := repo.CreateHasil(ctx, domain.Hasil{ID: "hasil-1", PekerjaanID: "pekerjaan-1", TugasID: &parent, Kind: "file", Name: "hasil.txt", Path: "hasil/hasil.txt", CreatedAt: updated}); err != nil {
		t.Fatalf("CreateHasil() error = %v", err)
	}

	ruang, err := repo.GetRuang(ctx, "ruang-1")
	if err != nil || ruang.Name != "Proyek" || !ruang.CreatedAt.Equal(created) || !ruang.UpdatedAt.Equal(updated) {
		t.Fatalf("GetRuang() = %+v, error = %v", ruang, err)
	}

	agent, err := repo.GetAgen(ctx, "agen-1")
	if err != nil || agent.RuangID != "ruang-1" || agent.ProviderID != "ollama" {
		t.Fatalf("GetAgen() = %+v, error = %v", agent, err)
	}

	pekerjaan, err := repo.GetPekerjaan(ctx, "pekerjaan-1")
	if err != nil || pekerjaan.Status != domain.StatusReady || pekerjaan.RuangID != "ruang-1" {
		t.Fatalf("GetPekerjaan() = %+v, error = %v", pekerjaan, err)
	}

	tugas, err := repo.GetTugas(ctx, "tugas-2")
	if err != nil || tugas.ParentID == nil || *tugas.ParentID != parent || tugas.Position != 2 {
		t.Fatalf("GetTugas() = %+v, error = %v", tugas, err)
	}

	hasil, err := repo.GetHasil(ctx, "hasil-1")
	if err != nil || hasil.TugasID == nil || *hasil.TugasID != parent || hasil.Path != "hasil/hasil.txt" {
		t.Fatalf("GetHasil() = %+v, error = %v", hasil, err)
	}
}

func TestRepositoryNotFoundAndValidation(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if _, err := repo.GetRuang(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetRuang() error = %v, want ErrNotFound", err)
	}
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "", Name: "tanpa id"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateRuang() error = %v, want ErrInvalid", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", Name: "tanpa ruang"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateAgen() error = %v, want ErrInvalid", err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "negatif", Position: -1}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateTugas() error = %v, want ErrInvalid", err)
	}
	if err := repo.CreateHasil(ctx, domain.Hasil{ID: "hasil-1", PekerjaanID: "pekerjaan-1", Kind: "file", Name: "hasil"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateHasil() error = %v, want ErrInvalid", err)
	}
}

func TestRepositoryForeignKeys(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	err = repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "missing", Name: "Agen"})
	if err == nil {
		t.Fatal("CreateAgen() error = nil, want foreign-key error")
	}

	if err = repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "missing", Title: "Tugas"}); err == nil {
		t.Fatal("CreateTugas() error = nil, want foreign-key error")
	}
}

func TestRepositoryResultIntegrity(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	for _, job := range []domain.Pekerjaan{
		{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Satu"},
		{ID: "pekerjaan-2", RuangID: "ruang-1", Title: "Dua"},
	} {
		if err := repo.CreatePekerjaan(ctx, job); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-2", PekerjaanID: "pekerjaan-2", Title: "Tugas"}); err != nil {
		t.Fatal(err)
	}

	valid := domain.Hasil{ID: "hasil-valid", PekerjaanID: "pekerjaan-1", TugasID: domainID("tugas-1"), Kind: "file", Name: "hasil.txt", Path: "hasil/hasil.txt"}
	if err := repo.CreateHasil(ctx, valid); err != nil {
		t.Fatalf("valid result rejected: %v", err)
	}

	for _, result := range []domain.Hasil{
		{ID: "hasil-abs", PekerjaanID: "pekerjaan-1", Kind: "file", Name: "hasil.txt", Path: "/tmp/hasil.txt"},
		{ID: "hasil-dot", PekerjaanID: "pekerjaan-1", Kind: "file", Name: "hasil.txt", Path: "./hasil.txt"},
		{ID: "hasil-parent", PekerjaanID: "pekerjaan-1", Kind: "file", Name: "hasil.txt", Path: "../hasil.txt"},
		{ID: "hasil-name", PekerjaanID: "pekerjaan-1", Kind: "file", Name: "sub/hasil.txt", Path: "hasil.txt"},
	} {
		if err := repo.CreateHasil(ctx, result); !errors.Is(err, ErrInvalid) {
			t.Errorf("CreateHasil(%q) error = %v, want ErrInvalid", result.Path, err)
		}
	}

	crossJob := valid
	crossJob.ID = "hasil-cross-job"
	crossJob.PekerjaanID = "pekerjaan-2"
	if err := repo.CreateHasil(ctx, crossJob); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cross-job result error = %v, want ErrInvalid", err)
	}

	missingTask := valid
	missingTask.ID = "hasil-missing-task"
	missingTask.TugasID = domainID("missing")
	if err := repo.CreateHasil(ctx, missingTask); !errors.Is(err, ErrInvalid) {
		t.Fatalf("missing task result error = %v, want ErrInvalid", err)
	}
}

func domainID(value string) *domain.ID {
	id := domain.ID(value)
	return &id
}
