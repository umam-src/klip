package storage

import (
	"context"
	"testing"
)

func TestAISettingsRoundTrip(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if got, found, err := repo.GetAISettings(ctx); err != nil {
		t.Fatalf("GetAISettings() error = %v", err)
	} else if found {
		t.Fatalf("GetAISettings() found settings before save: %#v", got)
	}

	want := AISettings{
		Provider: "llama.cpp",
		BaseURL:  "http://127.0.0.1:8080",
		Model:    "qwen2.5-0.5b-instruct-q4_k_m",
	}
	if err := repo.SaveAISettings(ctx, want); err != nil {
		t.Fatalf("SaveAISettings() error = %v", err)
	}

	got, found, err := repo.GetAISettings(ctx)
	if err != nil {
		t.Fatalf("GetAISettings() error = %v", err)
	}
	if !found {
		t.Fatal("GetAISettings() found = false after save")
	}
	if got != want {
		t.Fatalf("GetAISettings() = %#v, want %#v", got, want)
	}

	updated := AISettings{
		Provider: "ollama",
		BaseURL:  "http://127.0.0.1:11434",
		Model:    "qwen3:8b",
	}
	if err := repo.SaveAISettings(ctx, updated); err != nil {
		t.Fatalf("SaveAISettings() update error = %v", err)
	}

	got, found, err = repo.GetAISettings(ctx)
	if err != nil {
		t.Fatalf("GetAISettings() after update error = %v", err)
	}
	if !found {
		t.Fatal("GetAISettings() found = false after update")
	}
	if got != updated {
		t.Fatalf("GetAISettings() after update = %#v, want %#v", got, updated)
	}
}
