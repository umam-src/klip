package ai

import (
	"errors"
	"net/http"
	"strings"

	"github.com/umam-src/klip/internal/config"
)

func NewProvider(cfg config.AIConfig) (AIProvider, error) {
	if strings.TrimSpace(cfg.Provider) == "" {
		return nil, errors.New("penyedia AI belum ditentukan")
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, errors.New("alamat penyedia AI belum ditentukan")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, errors.New("model AI belum ditentukan")
	}

	client := &http.Client{}
	switch strings.ToLower(cfg.Provider) {
	case "ollama":
		return &Ollama{BaseURL: cfg.BaseURL, Client: client}, nil
	case "openai-compatible", "openai_compatible":
		return &OpenAICompatible{BaseURL: cfg.BaseURL, APIKey: cfg.APIKey, Client: client}, nil
	default:
		return nil, errors.New("penyedia AI tidak dikenal: " + cfg.Provider)
	}
}
