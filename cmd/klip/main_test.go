package main

import (
	"os"
	"testing"
)

func BenchmarkProcessEnvironment(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = os.UserConfigDir()
	}
}
