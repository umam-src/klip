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

func seedKomentarProyek(t *testing.T, repo *storage.Repository, ruangID domain.ID, proyekIDs ...domain.ID) {
	t.Helper()
	ctx := context.Background()
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: ruangID, Name: "Komentar"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-komentar", RuangID: ruangID, Title: "Goal", Status: domain.GoalStatusActive}); err != nil {
		t.Fatal(err)
	}
	for _, id := range proyekIDs {
		if err := repo.CreateProyek(ctx, domain.Proyek{ID: id, RuangID: ruangID, GoalID: "goal-komentar", Title: string(id), Status: domain.StatusDraft}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestKomentarProyekAndTugasThread(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)
	seedKomentarProyek(t, repo, "ruang-komentar", "proyek-komentar")
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-komentar", RuangID: "ruang-komentar", ProyekID: "proyek-komentar", Title: "Tugas", Status: domain.StatusDraft}); err != nil {
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
	if res := post("/api/v1/proyek/proyek-komentar/komentar", `{"id":"comment-1","body":"konteks"}`); res.Code != http.StatusCreated {
		t.Fatalf("create proyek comment: %d %s", res.Code, res.Body.String())
	}
	if res := post("/api/v1/proyek/proyek-komentar/komentar", `{"id":"comment-2","parent_id":"comment-1","body":"balasan"}`); res.Code != http.StatusCreated {
		t.Fatalf("create reply: %d %s", res.Code, res.Body.String())
	}
	if res := post("/api/v1/tugas/tugas-komentar/komentar", `{"id":"comment-3","body":"catatan tugas"}`); res.Code != http.StatusCreated {
		t.Fatalf("create task comment: %d %s", res.Code, res.Body.String())
	}

	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/proyek/proyek-komentar/komentar", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("list proyek comments: %d", res.Code)
	}
	var proyekComments []domain.Komentar
	if err := json.NewDecoder(res.Body).Decode(&proyekComments); err != nil {
		t.Fatal(err)
	}
	if len(proyekComments) != 2 || proyekComments[1].ParentID == nil || *proyekComments[1].ParentID != "comment-1" {
		t.Fatalf("comments = %#v", proyekComments)
	}
	if proyekComments[0].RuangID != "ruang-komentar" || proyekComments[0].ProyekID != "proyek-komentar" {
		t.Fatalf("konteks komentar = %#v", proyekComments[0])
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
	seedKomentarProyek(t, repo, "ruang-cross", "p1", "p2")
	if err := repo.CreateKomentar(ctx, domain.Komentar{ID: "parent", RuangID: "ruang-cross", ProyekID: "p1", Body: "parent"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateKomentar(ctx, domain.Komentar{ID: "child", RuangID: "ruang-cross", ProyekID: "p2", ParentID: ptrID("parent"), Body: "invalid"}); err == nil {
		t.Fatal("expected cross-context parent rejection")
	}
}

func ptrID(value string) *domain.ID {
	id := domain.ID(value)
	return &id
}
