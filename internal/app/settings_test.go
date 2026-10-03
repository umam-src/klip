package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/umam-src/klip/internal/ai"
	"github.com/umam-src/klip/internal/config"
)

func TestHandlerSettingsUpdate(t *testing.T) {
	dataDir := t.TempDir()
	cfg := config.Default(dataDir)
	cfg.AI.Provider = "ollama"
	cfg.AI.BaseURL = "http://old.example"
	cfg.AI.Model = "model-lama"
	app := New(cfg, fakeProvider{})

	body := `{"provider":"ollama","base_url":"http://127.0.0.1:11434","model":"model-baru"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", stringsReader(body))
	res := httptest.NewRecorder()
	app.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusOK, res.Body.String())
	}

	data, err := os.ReadFile(filepath.Join(dataDir, "config.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if saved.AI.BaseURL != "http://127.0.0.1:11434" || saved.AI.Model != "model-baru" {
		t.Fatalf("saved AI config = %+v", saved.AI)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	getRes := httptest.NewRecorder()
	app.Handler().ServeHTTP(getRes, getReq)
	if getRes.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", getRes.Code, http.StatusOK)
	}
	if got := getRes.Body.String(); !containsAll(got, "model-baru", "127.0.0.1:11434") {
		t.Fatalf("GET body = %q", got)
	}
}

func TestHandlerSettingsUpdateRejectsIncompleteRequest(t *testing.T) {
	app := New(config.Default(t.TempDir()), fakeProvider{})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", stringsReader(`{"provider":"ollama"}`))
	res := httptest.NewRecorder()
	app.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandlerSettingsUpdatePreservesAPIKey(t *testing.T) {
	cfg := config.Default(t.TempDir())
	cfg.AI.Provider = "openai-compatible"
	cfg.AI.APIKey = "secret-not-logged"
	app := New(cfg, fakeProvider{})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", stringsReader(`{"provider":"openai-compatible","base_url":"http://127.0.0.1:8080","model":"model-baru"}`))
	res := httptest.NewRecorder()
	app.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if app.config.AI.APIKey != "secret-not-logged" {
		t.Fatalf("API key was not preserved")
	}
}

func stringsReader(value string) *stringReaderValue {
	return &stringReaderValue{value: value}
}

type stringReaderValue struct {
	value string
	offset int
}

func (r *stringReaderValue) Read(p []byte) (int, error) {
	if r.offset >= len(r.value) {
		return 0, context.Canceled
	}
	n := copy(p, r.value[r.offset:])
	r.offset += n
	return n, nil
}

func containsAll(value string, needles ...string) bool {
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			return false
		}
	}
	return true
}

var _ ai.AIProvider = fakeProvider{}
