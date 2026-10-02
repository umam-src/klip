package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/umam-src/klip/internal/i18n"
)

type Config struct {
	DataDir string   `json:"data_dir,omitempty"`
	Listen  string   `json:"listen,omitempty"`
	Locale  string   `json:"locale,omitempty"`
	AI      AIConfig `json:"ai"`
}

type AIConfig struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key,omitempty"`
	Model    string `json:"model"`
}

func Default(dataDir string) Config {
	return Config{
		DataDir: dataDir,
		Listen:  "127.0.0.1:8787",
		Locale:  i18n.DefaultLocale,
		AI: AIConfig{
			Provider: "ollama",
			BaseURL:  "http://127.0.0.1:11434",
		},
	}
}

func Load(path string, fallback Config) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return fallback, nil
	}
	if err != nil {
		return Config{}, err
	}

	cfg := fallback
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.DataDir == "" {
		cfg.DataDir = fallback.DataDir
	}
	if cfg.Listen == "" {
		cfg.Listen = fallback.Listen
	}
	cfg.Locale = i18n.Normalize(cfg.Locale)
	if cfg.AI.Provider == "" {
		cfg.AI.Provider = fallback.AI.Provider
	}
	if cfg.AI.BaseURL == "" {
		cfg.AI.BaseURL = fallback.AI.BaseURL
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	cfg.Locale = i18n.Normalize(cfg.Locale)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}
