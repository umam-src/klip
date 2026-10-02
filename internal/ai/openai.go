package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAICompatible struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
	Retry   RetryConfig
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature *float64  `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type chatStreamResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (p *OpenAICompatible) ID() string { return "openai-compatible" }

func (p *OpenAICompatible) Check(ctx context.Context) error {
	if strings.TrimSpace(p.BaseURL) == "" {
		return fmt.Errorf("base URL provider AI kosong")
	}
	url := strings.TrimRight(p.BaseURL, "/") + "/v1/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("buat pemeriksaan provider AI: %w", err)
	}
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
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

func (p *OpenAICompatible) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if strings.TrimSpace(p.BaseURL) == "" {
		return ChatResponse{}, fmt.Errorf("base URL provider AI kosong")
	}
	if strings.TrimSpace(req.Model) == "" {
		return ChatResponse{}, fmt.Errorf("model AI kosong")
	}
	return retryChat(ctx, p.Retry, func(ctx context.Context) (ChatResponse, error) {
		return p.chatOnce(ctx, req)
	})
}

func (p *OpenAICompatible) Stream(ctx context.Context, req ChatRequest, emit func(ChatResponse) error) error {
	if strings.TrimSpace(p.BaseURL) == "" {
		return fmt.Errorf("base URL provider AI kosong")
	}
	if strings.TrimSpace(req.Model) == "" {
		return fmt.Errorf("model AI kosong")
	}
	if emit == nil {
		return fmt.Errorf("callback streaming AI kosong")
	}
	return p.streamOnce(ctx, req, emit)
}

func (p *OpenAICompatible) chatOnce(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	payload, err := json.Marshal(chatRequest{
		Model: req.Model, Messages: req.Messages, Temperature: req.Temperature,
		MaxTokens: req.MaxTokens, Stream: false,
	})
	if err != nil {
		return ChatResponse{}, fmt.Errorf("encode request AI: %w", err)
	}

	url := strings.TrimRight(p.BaseURL, "/") + "/v1/chat/completions"
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("buat request AI: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		reqHTTP.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("hubungi provider AI: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("baca respons AI: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ChatResponse{}, providerHTTPError{status: resp.StatusCode}
	}

	var decoded chatResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return ChatResponse{}, fmt.Errorf("decode respons AI: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("respons AI tidak memiliki pilihan")
	}

	return ChatResponse{Model: decoded.Model, Content: decoded.Choices[0].Message.Content}, nil
}

func (p *OpenAICompatible) streamOnce(ctx context.Context, req ChatRequest, emit func(ChatResponse) error) error {
	payload, err := json.Marshal(chatRequest{
		Model: req.Model, Messages: req.Messages, Temperature: req.Temperature,
		MaxTokens: req.MaxTokens, Stream: true,
	})
	if err != nil {
		return fmt.Errorf("encode request AI: %w", err)
	}

	url := strings.TrimRight(p.BaseURL, "/") + "/v1/chat/completions"
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("buat request streaming AI: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		reqHTTP.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return fmt.Errorf("hubungi provider AI: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return providerHTTPError{status: resp.StatusCode}
	}

	scanner := bufio.NewScanner(io.LimitReader(resp.Body, 8<<20))
	const maxLine = 1 << 20
	scanner.Buffer(make([]byte, 4<<10), maxLine)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			return nil
		}
		var decoded chatStreamResponse
		if err := json.Unmarshal([]byte(data), &decoded); err != nil {
			return fmt.Errorf("decode potongan streaming AI: %w", err)
		}
		if len(decoded.Choices) == 0 || decoded.Choices[0].Delta.Content == "" {
			continue
		}
		if err := emit(ChatResponse{Model: decoded.Model, Content: decoded.Choices[0].Delta.Content}); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("baca streaming AI: %w", err)
	}
	return nil
}
