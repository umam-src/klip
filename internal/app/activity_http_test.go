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
	base := time.Date(2026, 10, 2, 10, 0, 0, 0, time.UTC)
	for _, event := range []domain.Event{
		{ID: "event-old", PekerjaanID: "pekerjaan-activity", Type: domain.EventExecutionStarted, Message: "prompt rahasia" + " credential-token", CreatedAt: base},
		{ID: "event-new", PekerjaanID: "pekerjaan-activity", Type: domain.EventExecutionCompleted, Message: "credential-token", CreatedAt: base.Add(time.Minute)},
		{ID: "event-other", PekerjaanID: "pekerjaan-other", Type: domain.EventExecutionFailed, Message: "jangan tampil", CreatedAt: base.Add(2 * time.Minute)},
	} {
		if err := repo.AppendEvent(ctx, event); err != nil {
			t.Fatalf("AppendEvent(%s) error = %v", event.ID, err)
		}
	}

	handler := New(config.Default(t.TempDir()), nil, repo).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pekerjaan/pekerjaan-activity/aktivitas?limit=1", nil)
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
		req := httptest.NewRequest(http.MethodGet, "/api/v1/pekerjaan/pekerjaan-activity/aktivitas?limit="+value, nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("limit %q: status = %d, want %d", value, res.Code, http.StatusBadRequest)
		}
	}
}
