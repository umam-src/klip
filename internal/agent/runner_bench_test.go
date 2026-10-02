package agent

import (
	"context"
	"testing"
)

func BenchmarkRunnerRun(b *testing.B) {
	program, args := testCommand("hello")
	runner := NewRunner(1, 1<<20)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := runner.Run(ctx, Command{Program: program, Args: args}); err != nil {
			b.Fatalf("Run() error = %v", err)
		}
	}
}

func BenchmarkRunnerRunConcurrent(b *testing.B) {
	program, args := testCommand("hello")
	runner := NewRunner(4, 1<<20)
	ctx := context.Background()

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := runner.Run(ctx, Command{Program: program, Args: args}); err != nil {
				b.Errorf("Run() error = %v", err)
			}
		}
	})
}
