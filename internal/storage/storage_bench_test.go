package storage

import (
	"context"
	"testing"
)

func BenchmarkOpenMemory(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db, err := Open(ctx, ":memory:")
		if err != nil {
			b.Fatalf("Open() error = %v", err)
		}
		if err := db.Close(); err != nil {
			b.Fatalf("Close() error = %v", err)
		}
	}
}

func BenchmarkSQLiteWriteRead(b *testing.B) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		b.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := db.ExecContext(ctx, "INSERT INTO ruang (id, name, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
			string(rune(i))+"-bench", "benchmark"); err != nil {
			b.Fatalf("insert error = %v", err)
		}
		var name string
		if err := db.QueryRowContext(ctx, "SELECT name FROM ruang WHERE id = ?", string(rune(i))+"-bench").Scan(&name); err != nil {
			b.Fatalf("select error = %v", err)
		}
	}
}
