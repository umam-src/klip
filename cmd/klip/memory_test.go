package main

import (
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/umam-src/klip/internal/app"
	"github.com/umam-src/klip/internal/config"
)

func BenchmarkAppHandlerMemory(b *testing.B) {
	cfg := config.Default(b.TempDir())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler := app.New(cfg, nil).Handler()
		req := httptest.NewRequest("GET", "/health", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkRuntimeMemorySnapshot(b *testing.B) {
	var stats runtime.MemStats
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.ReadMemStats(&stats)
	}
}
