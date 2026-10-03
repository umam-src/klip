package storage

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrateRejectsLegacyDatabase(t *testing.T) {
	cases := []struct {
		name  string
		setup []string
		want  string
	}{
		{
			name: "versi skema lebih baru",
			setup: []string{
				`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
				`INSERT INTO schema_migrations(version) VALUES (9)`,
			},
			want: "lebih baru",
		},
		{
			name:  "tabel legacy tanpa versi skema",
			setup: []string{`CREATE TABLE pekerjaan (id TEXT PRIMARY KEY)`},
			want:  "database legacy terdeteksi",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatalf("sql.Open() error = %v", err)
			}
			defer db.Close()
			db.SetMaxOpenConns(1)

			for _, statement := range tc.setup {
				if _, err := db.ExecContext(ctx, statement); err != nil {
					t.Fatalf("setup statement %q: %v", statement, err)
				}
			}
			err = migrate(ctx, db)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("migrate() error = %v, want mengandung %q", err, tc.want)
			}
		})
	}
}
