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

	agents, err := repo.ListAgenByRuang(ctx, "ruang-1")
	if err != nil || len(agents) != 1 || agents[0].ID != "agen-1" {
		t.Fatalf("ListAgenByRuang() = %+v, error = %v", agents, err)
	}

	jobs, err := repo.ListPekerjaanByRuang(ctx, "ruang-1")
	if err != nil || len(jobs) != 1 || jobs[0].ID != "pekerjaan-1" {
		t.Fatalf("ListPekerjaanByRuang() = %+v, error = %v", jobs, err)
	}

	tasks, err := repo.ListTugasByPekerjaan(ctx, "pekerjaan-1")
	if err != nil || len(tasks) != 2 || tasks[0].ID != "tugas-1" || tasks[1].ID != "tugas-2" {
		t.Fatalf("ListTugasByPekerjaan() = %+v, error = %v", tasks, err)
	}

	results, err := repo.ListHasilByPekerjaan(ctx, "pekerjaan-1")
	if err != nil || len(results) != 1 || results[0].ID != "hasil-1" {
		t.Fatalf("ListHasilByPekerjaan() = %+v, error = %v", results, err)
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
