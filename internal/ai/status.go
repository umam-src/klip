package ai

import "context"

type HealthChecker interface {
	Check(ctx context.Context) error
}
