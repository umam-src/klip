package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Ollama struct {
	BaseURL string
	Client  *http.Client
}

type ollamaModelListResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func (p *Ollama) ID() string { return "ollama" }

func (p *Ollama) Check(ctx context.Context) error {
	if strings.TrimSpace(p.BaseURL) == "" {
		return fmt.Errorf("base URL provider AI kosong")
	}
	url := strings.TrimRight(p.BaseURL, "/") + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("buat pemeriksaan provider AI: %w", err)
	}
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("hubungi provider AI: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return providerHTTPError{status: resp.StatusCode}
	}
	return nil
}

func (p *Ollama) ListModels(ctx context.Context) ([]Model, error) {
	if strings.TrimSpace(p.BaseURL) == "" {
		return nil, fmt.Errorf("base URL provider AI kosong")
	}
	url := strings.TrimRight(p.BaseURL, "/") + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("buat permintaan daftar model: %w", err)
	}
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hubungi provider AI: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, providerHTTPError{status: resp.StatusCode}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("baca daftar model: %w", err)
	}
	var decoded ollamaModelListResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("decode daftar model: %w", err)
	}
	models := make([]Model, 0, len(decoded.Models))
	for _, item := range decoded.Models {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		models = append(models, Model{ID: item.Name})
	}
	return models, nil
}

func (p *Ollama) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	compatible := &OpenAICompatible{BaseURL: p.BaseURL, Client: p.Client}
	return compatible.Chat(ctx, req)
}
