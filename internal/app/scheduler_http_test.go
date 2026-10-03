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

// siapkanJadwal membuat dua Ruang Kerja terpisah, masing-masing dengan Goal,
// Proyek, Tugas, dan Agen, untuk menguji batas ruang_id pada jadwal.
func siapkanJadwal(t *testing.T) http.Handler {
	t.Helper()
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := storage.NewRepository(db)

	for _, ruangID := range []domain.ID{"ruang-a", "ruang-b"} {
		if err := repo.CreateRuang(ctx, domain.Ruang{ID: ruangID, Name: string(ruangID)}); err != nil {
			t.Fatal(err)
		}
		goalID := domain.ID("goal-" + string(ruangID))
		proyekID := domain.ID("proyek-" + string(ruangID))
		if err := repo.CreateGoal(ctx, domain.Goal{ID: goalID, RuangID: ruangID, Title: "Goal", Status: domain.GoalStatusActive}); err != nil {
			t.Fatal(err)
		}
		if err := repo.CreateProyek(ctx, domain.Proyek{ID: proyekID, RuangID: ruangID, GoalID: goalID, Title: "Proyek", Status: domain.StatusDraft}); err != nil {
			t.Fatal(err)
		}
		if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: domain.ID("tugas-" + string(ruangID)), RuangID: ruangID, ProyekID: proyekID, Title: "Tugas", Status: domain.StatusReady}); err != nil {
			t.Fatal(err)
		}
		if err := repo.CreateAgen(ctx, domain.Agen{ID: domain.ID("agen-" + string(ruangID)), RuangID: ruangID, Name: "Agen"}); err != nil {
			t.Fatal(err)
		}
	}
	return New(config.Default(t.TempDir()), nil, repo).Handler()
}

func kirimJadwal(handler http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/scheduler", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func TestBuatJadwalMengambilRuangDariProyek(t *testing.T) {
	handler := siapkanJadwal(t)
	res := kirimJadwal(handler, `{"name":"Harian","proyek_id":"proyek-ruang-a","tugas_id":"tugas-ruang-a","agen_id":"agen-ruang-a","program":"echo","interval_seconds":60}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("buat jadwal: %d %s", res.Code, res.Body.String())
	}
	var jadwal map[string]any
	if err := json.NewDecoder(res.Body).Decode(&jadwal); err != nil {
		t.Fatal(err)
	}
	if jadwal["ruang_id"] != "ruang-a" {
		t.Fatalf("ruang_id = %v, ingin ruang-a", jadwal["ruang_id"])
	}
}

func TestBuatJadwalMenolakKonteksLintasRuang(t *testing.T) {
	handler := siapkanJadwal(t)
	kasus := map[string]string{
		"agen dari ruang lain":  `{"name":"x","proyek_id":"proyek-ruang-a","agen_id":"agen-ruang-b","program":"echo","interval_seconds":60}`,
		"tugas dari ruang lain": `{"name":"x","proyek_id":"proyek-ruang-a","tugas_id":"tugas-ruang-b","agen_id":"agen-ruang-a","program":"echo","interval_seconds":60}`,
		"proyek tidak ada":      `{"name":"x","proyek_id":"proyek-tidak-ada","agen_id":"agen-ruang-a","program":"echo","interval_seconds":60}`,
	}
	for nama, body := range kasus {
		t.Run(nama, func(t *testing.T) {
			if res := kirimJadwal(handler, body); res.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, ingin 400; body = %s", res.Code, res.Body.String())
			}
		})
	}
}
