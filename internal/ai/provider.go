package ai

import "context"

type Message struct {
	Role    string
	Content string
}

type ChatRequest struct {
	Model       string
	Messages    []Message
	Temperature *float64
	MaxTokens   int
	Stream      bool
}

type ChatResponse struct {
	Model   string
	Content string
}

type AIProvider interface {
	ID() string
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
