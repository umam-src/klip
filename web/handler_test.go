package web

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHandlerServesLocalUI(t *testing.T) {
    handler := Handler()
    request := httptest.NewRequest(http.MethodGet, "/", nil)
    response := httptest.NewRecorder()

    handler.ServeHTTP(response, request)

    if response.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
    }
    if response.Header().Get("Content-Type") != "text/html; charset=utf-8" {
        t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
    }
}

func TestHandlerServesAssets(t *testing.T) {
    handler := Handler()
    request := httptest.NewRequest(http.MethodGet, "/assets/app.css", nil)
    response := httptest.NewRecorder()

    handler.ServeHTTP(response, request)

    if response.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
    }
}

func TestHandlerRejectsUnknownPath(t *testing.T) {
    handler := Handler()
    request := httptest.NewRequest(http.MethodGet, "/unknown", nil)
    response := httptest.NewRecorder()

    handler.ServeHTTP(response, request)

    if response.Code != http.StatusNotFound {
        t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
    }
}
