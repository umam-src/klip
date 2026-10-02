package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/config"
)

func TestHandlerRejectsNonJSONWrite(t *testing.T) {
	handler := New(config.Default(t.TempDir()), fakeProvider{}).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ruang", strings.NewReader(`{"id":"ruang-1","name":"Uji"}`))
	req.Header.Set("Content-Type", "text/plain")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnsupportedMediaType)
	}
}
