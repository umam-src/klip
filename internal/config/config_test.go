package config

import (
	"os"
	"path/filepath"
	"testing"
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
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permission config = %o, want 600", info.Mode().Perm())
	}

	got, err := Load(path, Default("fallback"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.DataDir != want.DataDir || got.Listen != want.Listen || got.AI != want.AI {
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
	if got.DataDir != "fallback" || got.Listen != "127.0.0.1:8787" || got.AI.Provider != "ollama" {
		t.Fatalf("fallback tidak diterapkan: %+v", got)
	}
	if got.AI.Model != "local-model" {
		t.Fatalf("model = %q", got.AI.Model)
	}
}
