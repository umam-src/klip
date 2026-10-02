package ai

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

const (
	defaultRetryAttempts  = 3
	defaultRetryBaseDelay = 250 * time.Millisecond
	defaultRetryMaxDelay  = 2 * time.Second
)

type RetryConfig struct {
	Attempts  int
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

func (c RetryConfig) normalized() RetryConfig {
	if c.Attempts <= 0 {
		c.Attempts = defaultRetryAttempts
	}
	if c.BaseDelay <= 0 {
		c.BaseDelay = defaultRetryBaseDelay
	}
	if c.MaxDelay <= 0 {
		c.MaxDelay = defaultRetryMaxDelay
	}
	if c.MaxDelay < c.BaseDelay {
		c.MaxDelay = c.BaseDelay
	}
	return c
}

func retryChat(ctx context.Context, cfg RetryConfig, fn func(context.Context) (ChatResponse, error)) (ChatResponse, error) {
	cfg = cfg.normalized()
	var lastErr error
	for attempt := 0; attempt < cfg.Attempts; attempt++ {
		response, err := fn(ctx)
		if err == nil {
			return response, nil
		}
		lastErr = err
		if !retryableError(err) || attempt == cfg.Attempts-1 {
			break
		}
		delay := retryDelay(cfg, attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ChatResponse{}, ctx.Err()
		case <-timer.C:
		}
	}
	return ChatResponse{}, lastErr
}

func retryDelay(cfg RetryConfig, attempt int) time.Duration {
	factor := math.Pow(2, float64(attempt))
	delay := time.Duration(float64(cfg.BaseDelay) * factor)
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}
	jitter := time.Duration(rand.Int63n(int64(delay/4) + 1))
	return delay + jitter
}

func retryableError(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var statusErr interface{ HTTPStatusCode() int }
	if errors.As(err, &statusErr) {
		code := statusErr.HTTPStatusCode()
		return code == http.StatusRequestTimeout || code == http.StatusTooManyRequests || code >= 500
	}
	return false
}

type providerHTTPError struct {
	status int
}

func (e providerHTTPError) Error() string {
	return fmt.Sprintf("provider AI mengembalikan HTTP %d", e.status)
}

func (e providerHTTPError) HTTPStatusCode() int { return e.status }
