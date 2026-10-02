package ai

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Ollama struct {
	BaseURL string
	Client  *http.Client
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

func (p *Ollama) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	compatible := &OpenAICompatible{BaseURL: p.BaseURL, Client: p.Client}
	return compatible.Chat(ctx, req)
}
