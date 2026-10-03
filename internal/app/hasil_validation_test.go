package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func TestValidateHasilPathAndName(t *testing.T) {
	valid := domain.HasilKerja{ID: "hasil-1", RuangID: "ruang-1", ProyekID: "proyek-1", Kind: "file", Name: "hasil.txt", Path: "hasil/hasil.txt"}
	if err := validateHasil(valid); err != nil {
		t.Fatalf("valid hasil rejected: %v", err)
	}

	for _, path := range []string{"../hasil.txt", "hasil/../hasil.txt", "/tmp/hasil.txt", "./hasil.txt", "hasil\\hasil.txt", "hasil\x00.txt"} {
		value := valid
		value.Path = path
		if err := validateHasil(value); err == nil {
			t.Errorf("path %q accepted", path)
		}
	}
	for _, name := range []string{"", "hasil/hasil.txt", "hasil\\hasil.txt", "hasil\x00.txt"} {
		value := valid
		value.Name = name
		if err := validateHasil(value); err == nil {
			t.Errorf("name %q accepted", name)
		}
	}
}

func TestHandlerHasilRejectsCrossProyekTugas(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Goal", Status: domain.GoalStatusActive}); err != nil {
		t.Fatal(err)
	}
	for _, proyek := range []domain.Proyek{
		{ID: "proyek-1", RuangID: "ruang-1", GoalID: "goal-1", Title: "Satu", Status: domain.StatusReady},
		{ID: "proyek-2", RuangID: "ruang-1", GoalID: "goal-1", Title: "Dua", Status: domain.StatusReady},
	} {
		if err := repo.CreateProyek(ctx, proyek); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-1", RuangID: "ruang-1", ProyekID: "proyek-1", Title: "Tugas", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}

	app := New(config.Default(t.TempDir()), fakeProvider{}, repo)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/proyek/proyek-2/hasil", strings.NewReader(`{"id":"hasil-1","tugas_id":"tugas-1","kind":"file","name":"hasil.txt","path":"hasil.txt"}`))
	res := httptest.NewRecorder()
	app.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusBadRequest, res.Body.String())
	}
}
