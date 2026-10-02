package ai

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryChatRetriesTransientProviderError(t *testing.T) {
	attempts := 0
	response, err := retryChat(context.Background(), RetryConfig{Attempts: 3, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond}, func(context.Context) (ChatResponse, error) {
		attempts++
		if attempts < 3 {
			return ChatResponse{}, providerHTTPError{status: 503}
		}
		return ChatResponse{Content: "ok"}, nil
	})
	if err != nil || response.Content != "ok" || attempts != 3 {
		t.Fatalf("response=%+v err=%v attempts=%d", response, err, attempts)
	}
}

func TestRetryChatDoesNotRetryClientError(t *testing.T) {
	attempts := 0
	want := providerHTTPError{status: 400}
	_, err := retryChat(context.Background(), RetryConfig{Attempts: 3, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond}, func(context.Context) (ChatResponse, error) {
		attempts++
		return ChatResponse{}, want
	})
	if !errors.Is(err, want) || attempts != 1 {
		t.Fatalf("err=%v attempts=%d", err, attempts)
	}
}
