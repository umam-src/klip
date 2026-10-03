package storage

import (
	"context"
	"testing"
)

func TestAISettings(t *testing.T) {
	db, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	settings, found, err := repo.GetAISettings(context.Background())
	if err != nil {
		t.Fatalf("GetAISettings() error = %v", err)
	}
	if found {
		t.Fatal("GetAISettings() found settings before save")
	}
	if err := repo.SaveAISettings(context.Background(), AISettings{
		Provider: "llama cpp",
		BaseURL:  "http://127.0.0.1:8080",
		Model:    "qwen2.5-0.5b-instruct-q4_k_m",
	}); err != nil {
		t.Fatalf("SaveAISettings() error = %v", err)
	}

	settings, found, err = repo.GetAISettings(context.Background())
	if err != nil {
		t.Fatalf("GetAISettings() error = %v", err)
	}
	if !found {
		t.Fatal("GetAISettings() did not find saved settings")
	}
	if settings.Provider != "llama cpp" || settings.BaseURL != "http://127.0.0.1:8080" || settings.Model != "qwen2.5-0.5b-instruct-q4_k_m" {
		t.Fatalf("settings = %+v", settings)
	}
}

func TestAISettingsRejectIncomplete(t *testing.T) {
	db, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.SaveAISettings(context.Background(), AISettings{Provider: "ollama", BaseURL: "http://127.0.0.1:11434"}); err == nil {
		t.Fatal("SaveAISettings() error = nil, want validation error")
	}
}
