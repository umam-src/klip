package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
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
	if response.Result().Header.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatalf("content type = %q", response.Result().Header.Get("Content-Type"))
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

func TestAssignmentUIContract(t *testing.T) {
	index, err := Files.ReadFile("index.html")
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	page := string(index)
	for _, want := range []string{`id="task-form"`, `name="agen_id"`, `id="task-agent"`} {
		if !strings.Contains(page, want) {
			t.Fatalf("index.html missing assignment UI contract %q", want)
		}
	}
	app, err := Files.ReadFile("assets/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	code := string(app)
	for _, want := range []string{`/api/v1/tugas/${encodeURIComponent(task.id)}/agen`, `}, 'PUT');`, `{ agen_id: data.agen_id }`, `Agen pelaksana`} {
		if !strings.Contains(code, want) {
			t.Fatalf("app.js missing assignment contract %q", want)
		}
	}
}
