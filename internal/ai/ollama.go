package ai

import (
	"context"
	"net/http"
)

type Ollama struct {
	BaseURL string
	Client  *http.Client
}

func (p *Ollama) ID() string { return "ollama" }

func (p *Ollama) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	compatible := &OpenAICompatible{BaseURL: p.BaseURL, Client: p.Client}
	return compatible.Chat(ctx, req)
}
