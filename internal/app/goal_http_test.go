package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func TestGoalAPIUsesNativeContract(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	app := New(config.Config{}, nil, repo)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/ruang/ruang-1/goal", strings.NewReader(`{"id":"goal-1","title":"Rilis","description":"Tujuan"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status tidak sesuai: %d, body=%s", response.Code, response.Body.String())
	}

	var goal domain.Goal
	if err := json.NewDecoder(response.Body).Decode(&goal); err != nil {
		t.Fatal(err)
	}
	if goal.ID != "goal-1" || goal.RuangID != "ruang-1" || goal.Status != domain.GoalStatusActive {
		t.Fatalf("goal tidak sesuai: %+v", goal)
	}

	legacy := httptest.NewRequest(http.MethodGet, "/api/v1/ruang/ruang-1/sasaran", nil)
	legacyResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(legacyResponse, legacy)
	if legacyResponse.Code != http.StatusNotFound {
		t.Fatalf("endpoint legacy masih tersedia: %d", legacyResponse.Code)
	}
}
