package ai

import (
	"testing"

	"github.com/umam-src/klip/internal/config"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantID   string
	}{
		{name: "ollama", provider: "ollama", wantID: "ollama"},
		{name: "openai compatible", provider: "openai-compatible", wantID: "openai-compatible"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(config.AIConfig{
				Provider: tt.provider,
				BaseURL:  "http://127.0.0.1:1234",
				Model:    "local-model",
			})
			if err != nil {
				t.Fatalf("NewProvider() error = %v", err)
			}
			if provider.ID() != tt.wantID {
				t.Fatalf("ID() = %q, want %q", provider.ID(), tt.wantID)
			}
		})
	}
}

func TestNewProviderRejectsMissingModel(t *testing.T) {
	_, err := NewProvider(config.AIConfig{
		Provider: "ollama",
		BaseURL:  "http://127.0.0.1:11434",
	})
	if err == nil {
		t.Fatal("NewProvider() seharusnya menolak model kosong")
	}
}

func TestNewProviderRejectsUnknownProvider(t *testing.T) {
	_, err := NewProvider(config.AIConfig{
		Provider: "unknown",
		BaseURL:  "http://127.0.0.1:1234",
		Model:    "local-model",
	})
	if err == nil {
		t.Fatal("NewProvider() seharusnya menolak provider tidak dikenal")
	}
}
