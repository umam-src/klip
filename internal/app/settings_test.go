package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/storage"
)

func TestHandlerSettingsUpdate(t *testing.T) {
	db, err := storage.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	cfg := config.Default(t.TempDir())
	cfg.AI.Provider = "ollama"
	cfg.AI.BaseURL = "http://old.example"
	cfg.AI.Model = "model-lama"
	repo := storage.NewRepository(db)
	app := New(cfg, fakeProvider{}, repo)

	body := `{"provider":"ollama","base_url":"http://127.0.0.1:11434","model":"model-baru"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", strings.NewReader(body))
	res := httptest.NewRecorder()
	app.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusOK, res.Body.String())
	}

	settings, found, err := repo.GetAISettings(context.Background())
	if err != nil {
		t.Fatalf("GetAISettings() error = %v", err)
	}
	if !found {
		t.Fatal("saved settings not found in database")
	}
	if settings.Provider != "ollama" || settings.BaseURL != "http://127.0.0.1:11434" || settings.Model != "model-baru" {
		t.Fatalf("saved AI settings = %+v", settings)
	}

	var response settingsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Model != "model-baru" || response.BaseURL != "http://127.0.0.1:11434" {
		t.Fatalf("response = %+v", response)
	}
}

func TestHandlerSettingsUpdateRejectsIncompleteRequest(t *testing.T) {
	db, err := storage.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	app := New(config.Default(t.TempDir()), fakeProvider{}, storage.NewRepository(db))
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", strings.NewReader(`{"provider":"ollama"}`))
	res := httptest.NewRecorder()
	app.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandlerSettingsUpdatePreservesAPIKey(t *testing.T) {
	db, err := storage.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	cfg := config.Default(t.TempDir())
	cfg.AI.Provider = "openai-compatible"
	cfg.AI.APIKey = "secret-not-logged"
	app := New(cfg, fakeProvider{}, storage.NewRepository(db))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", strings.NewReader(`{"provider":"openai-compatible","base_url":"http://127.0.0.1:8080","model":"model-baru"}`))
	res := httptest.NewRecorder()
	app.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if app.config.AI.APIKey != "secret-not-logged" {
		t.Fatalf("API key was not preserved")
	}
}

func TestHandlerSettingsUpdateRequiresDatabase(t *testing.T) {
	app := New(config.Default(t.TempDir()), fakeProvider{})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", strings.NewReader(`{"provider":"ollama","base_url":"http://127.0.0.1:11434","model":"model"}`))
	res := httptest.NewRecorder()
	app.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusInternalServerError)
	}
}
