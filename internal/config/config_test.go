package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/umam-src/klip/internal/i18n"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.json")
	want := Default(filepath.Join(dir, "data"))
	want.AI.Model = "local-model"

	if err := Save(path, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("permission config = %o, want 600", info.Mode().Perm())
	}

	got, err := Load(path, Default("fallback"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.DataDir != want.DataDir || got.Listen != want.Listen || got.Locale != want.Locale || got.AI != want.AI {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
}

func TestLoadUsesFallbackForMissingFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"ai":{"model":"local-model"}}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	got, err := Load(path, Default("fallback"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.DataDir != "fallback" || got.Listen != "127.0.0.1:8787" || got.Locale != i18n.DefaultLocale || got.AI.Provider != "ollama" {
		t.Fatalf("fallback tidak diterapkan: %+v", got)
	}
	if got.AI.Model != "local-model" {
		t.Fatalf("model = %q", got.AI.Model)
	}
}

func TestLoadNormalizesUnsupportedLocale(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"locale":"en-US"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	got, err := Load(path, Default("fallback"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Locale != i18n.DefaultLocale {
		t.Fatalf("locale = %q, want %q", got.Locale, i18n.DefaultLocale)
	}
}
