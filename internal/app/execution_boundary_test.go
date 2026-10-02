package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/config"
)

func TestHandlerRejectsPublicProcessExecution(t *testing.T) {
	handler := New(config.Default(t.TempDir()), fakeProvider{}).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/execute", strings.NewReader(`{"program":"echo","arguments":["unsafe"]}`))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
}
