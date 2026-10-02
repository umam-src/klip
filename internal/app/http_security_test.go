package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/storage"
)

func TestHandlerRejectsNonJSONWrite(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	handler := New(config.Default(t.TempDir()), fakeProvider{}, storage.NewRepository(db)).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ruang", strings.NewReader(`{"id":"ruang-1","name":"Uji"}`))
	req.Header.Set("Content-Type", "text/plain")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnsupportedMediaType)
	}
}
