package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func TestHandlerAktivitasReturnsSafeNewestLimitedHistory(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-activity", Name: "Aktivitas"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-activity", RuangID: "ruang-activity", Title: "Goal Aktivitas"}); err != nil {
		t.Fatalf("CreateGoal() error = %v", err)
	}
	for _, proyek := range []domain.Proyek{
		{ID: "proyek-activity", RuangID: "ruang-activity", GoalID: "goal-activity", Title: "Proyek Aktivitas"},
		{ID: "proyek-other", RuangID: "ruang-activity", GoalID: "goal-activity", Title: "Proyek Lain"},
	} {
		if err := repo.CreateProyek(ctx, proyek); err != nil {
			t.Fatalf("CreateProyek(%s) error = %v", proyek.ID, err)
		}
	}

	base := time.Date(2026, 10, 2, 10, 0, 0, 0, time.UTC)
	for _, event := range []domain.Event{
		{ID: "event-old", RuangID: "ruang-activity", ProyekID: "proyek-activity", Type: domain.EventExecutionStarted, Message: "prompt rahasia credential-token", CreatedAt: base},
		{ID: "event-new", RuangID: "ruang-activity", ProyekID: "proyek-activity", Type: domain.EventExecutionCompleted, Message: "credential-token", CreatedAt: base.Add(time.Minute)},
		{ID: "event-other", RuangID: "ruang-activity", ProyekID: "proyek-other", Type: domain.EventExecutionFailed, Message: "jangan tampil", CreatedAt: base.Add(2 * time.Minute)},
	} {
		if err := repo.AppendEvent(ctx, event); err != nil {
			t.Fatalf("AppendEvent(%s) error = %v", event.ID, err)
		}
	}

	handler := New(config.Default(t.TempDir()), nil, repo).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/proyek/proyek-activity/aktivitas?limit=1", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	body := res.Body.String()
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, body)
	}
	var got []activityResponse
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].ID != "event-new" {
		t.Fatalf("activity = %#v, want newest single event", got)
	}
	if got[0].Summary != "Eksekusi selesai" {
		t.Fatalf("summary = %q", got[0].Summary)
	}
	if strings.Contains(body, "credential-token") || strings.Contains(body, "prompt rahasia") {
		t.Fatalf("response leaked event message: %s", body)
	}
}

func TestHandlerAktivitasRejectsInvalidLimit(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	handler := New(config.Default(t.TempDir()), nil, repo).Handler()
	for _, value := range []string{"0", "101", "abc"} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/proyek/proyek-activity/aktivitas?limit="+value, nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("limit %q: status = %d, want %d", value, res.Code, http.StatusBadRequest)
		}
	}
}
