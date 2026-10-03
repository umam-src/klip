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

func TestHandlerTugasAgen(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("storage.Open() error = %v", err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)

	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang 1"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen 1", Role: "Pelaksana"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan 1", Status: domain.StatusDraft}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas 1", Status: domain.StatusDraft}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}

	handler := New(config.Default(t.TempDir()), nil, repo).Handler()

	put := httptest.NewRequest(http.MethodPut, "/api/v1/tugas/tugas-1/agen", strings.NewReader(`{"agen_id":"agen-1"}`))
	put.Header.Set("Content-Type", "application/json")
	putRes := httptest.NewRecorder()
	handler.ServeHTTP(putRes, put)
	if putRes.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, body = %s", putRes.Code, putRes.Body.String())
	}
	if !strings.Contains(putRes.Body.String(), `"agen_id":"agen-1"`) {
		t.Fatalf("PUT body = %s", putRes.Body.String())
	}

	get := httptest.NewRequest(http.MethodGet, "/api/v1/tugas/tugas-1/agen", nil)
	getRes := httptest.NewRecorder()
	handler.ServeHTTP(getRes, get)
	if getRes.Code != http.StatusOK {
		t.Fatalf("GET status = %d, body = %s", getRes.Code, getRes.Body.String())
	}
	if !strings.Contains(getRes.Body.String(), `"agen_id":"agen-1"`) {
		t.Fatalf("GET body = %s", getRes.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/tugas/tugas-1/agen", nil)
	deleteRes := httptest.NewRecorder()
	handler.ServeHTTP(deleteRes, deleteReq)
	if deleteRes.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, body = %s", deleteRes.Code, deleteRes.Body.String())
	}

	getEmpty := httptest.NewRequest(http.MethodGet, "/api/v1/tugas/tugas-1/agen", nil)
	getEmptyRes := httptest.NewRecorder()
	handler.ServeHTTP(getEmptyRes, getEmpty)
	if getEmptyRes.Code != http.StatusOK || strings.TrimSpace(getEmptyRes.Body.String()) != "null" {
		t.Fatalf("empty GET status = %d, body = %q", getEmptyRes.Code, getEmptyRes.Body.String())
	}
}

func TestHandlerTugasAgenRejectsInvalidAssignment(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("storage.Open() error = %v", err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)

	for _, ruang := range []domain.Ruang{
		{ID: "ruang-1", Name: "Ruang 1"},
		{ID: "ruang-2", Name: "Ruang 2"},
	} {
		if err := repo.CreateRuang(ctx, ruang); err != nil {
			t.Fatalf("CreateRuang(%s) error = %v", ruang.ID, err)
		}
	}
	for _, agent := range []domain.Agen{
		{ID: "agen-1", RuangID: "ruang-1", Name: "Agen 1", Role: "Pelaksana"},
		{ID: "agen-2", RuangID: "ruang-2", Name: "Agen 2", Role: "Pelaksana"},
	} {
		if err := repo.CreateAgen(ctx, agent); err != nil {
			t.Fatalf("CreateAgen(%s) error = %v", agent.ID, err)
		}
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan 1", Status: domain.StatusDraft}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas 1", Status: domain.StatusDraft}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}

	handler := New(config.Default(t.TempDir()), nil, repo).Handler()

	cases := []struct {
		name string
		body string
		want int
	}{
		{name: "cross room", body: `{"agen_id":"agen-2"}`, want: http.StatusBadRequest},
		{name: "missing agent", body: `{"agen_id":"agen-missing"}`, want: http.StatusNotFound},
		{name: "empty agent", body: `{"agen_id":""}`, want: http.StatusBadRequest},
		{name: "invalid json", body: `{`, want: http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			put := httptest.NewRequest(http.MethodPut, "/api/v1/tugas/tugas-1/agen", strings.NewReader(tc.body))
			put.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, put)
			if res.Code != tc.want {
				t.Fatalf("PUT status = %d, want %d, body = %s", res.Code, tc.want, res.Body.String())
			}
		})
	}

	missingTask := httptest.NewRequest(http.MethodGet, "/api/v1/tugas/tugas-missing/agen", nil)
	missingTaskRes := httptest.NewRecorder()
	handler.ServeHTTP(missingTaskRes, missingTask)
	if missingTaskRes.Code != http.StatusNotFound {
		t.Fatalf("missing task GET status = %d, want %d, body = %s", missingTaskRes.Code, http.StatusNotFound, missingTaskRes.Body.String())
	}
}
