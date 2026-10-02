package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func TestKomentarPekerjaanAndTugasThread(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-komentar", Name: "Komentar"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-komentar", RuangID: "ruang-komentar", Title: "Pekerjaan", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-komentar", PekerjaanID: "pekerjaan-komentar", Title: "Tugas", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}

	h := New(config.Default(t.TempDir()), nil, repo).Handler()
	post := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		return res
	}
	if res := post("/api/v1/pekerjaan/pekerjaan-komentar/komentar", `{"id":"comment-1","body":"konteks"}`); res.Code != http.StatusCreated {
		t.Fatalf("create pekerjaan comment: %d %s", res.Code, res.Body.String())
	}
	if res := post("/api/v1/pekerjaan/pekerjaan-komentar/komentar", `{"id":"comment-2","parent_id":"comment-1","body":"balasan"}`); res.Code != http.StatusCreated {
		t.Fatalf("create reply: %d %s", res.Code, res.Body.String())
	}
	if res := post("/api/v1/tugas/tugas-komentar/komentar", `{"id":"comment-3","body":"catatan tugas"}`); res.Code != http.StatusCreated {
		t.Fatalf("create task comment: %d %s", res.Code, res.Body.String())
	}

	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/pekerjaan/pekerjaan-komentar/komentar", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("list pekerjaan comments: %d", res.Code)
	}
	var pekerjaanComments []domain.Komentar
	if err := json.NewDecoder(res.Body).Decode(&pekerjaanComments); err != nil {
		t.Fatal(err)
	}
	if len(pekerjaanComments) != 2 || pekerjaanComments[1].ParentID == nil || *pekerjaanComments[1].ParentID != "comment-1" {
		t.Fatalf("comments = %#v", pekerjaanComments)
	}

	res = httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/tugas/tugas-komentar/komentar", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("list task comments: %d", res.Code)
	}
	var taskComments []domain.Komentar
	if err := json.NewDecoder(res.Body).Decode(&taskComments); err != nil {
		t.Fatal(err)
	}
	if len(taskComments) != 1 || taskComments[0].TugasID == nil || *taskComments[0].TugasID != "tugas-komentar" {
		t.Fatalf("task comments = %#v", taskComments)
	}
}

func TestKomentarRejectsCrossContextParent(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-cross", Name: "Cross"}); err != nil {
		t.Fatal(err)
	}
	for _, p := range []domain.Pekerjaan{{ID: "p1", RuangID: "ruang-cross", Title: "P1"}, {ID: "p2", RuangID: "ruang-cross", Title: "P2"}} {
		if err := repo.CreatePekerjaan(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.CreateKomentar(ctx, domain.Komentar{ID: "parent", PekerjaanID: "p1", Body: "parent"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateKomentar(ctx, domain.Komentar{ID: "child", PekerjaanID: "p2", ParentID: ptrID("parent"), Body: "invalid"}); err == nil {
		t.Fatal("expected cross-context parent rejection")
	}
}

func ptrID(value string) *domain.ID {
	id := domain.ID(value)
	return &id
}
