package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

func TestAgenHierarchy(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-2", Name: "Ruang lain"}); err != nil {
		t.Fatal(err)
	}

	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-root", RuangID: "ruang-1", Name: "Root", Role: "Pemimpin", Status: domain.AgenStatusActive}); err != nil {
		t.Fatalf("create root: %v", err)
	}
	parent := domain.ID("agen-root")
	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-child", RuangID: "ruang-1", ParentID: &parent, Name: "Child", Role: "Pelaksana", Status: domain.AgenStatusInactive}); err != nil {
		t.Fatalf("create child: %v", err)
	}

	child, err := repo.GetAgenWithParent(ctx, "agen-child")
	if err != nil {
		t.Fatalf("get child: %v", err)
	}
	if child.ParentID == nil || *child.ParentID != parent || child.Role != "Pelaksana" || child.Status != domain.AgenStatusInactive {
		t.Fatalf("child = %+v, want parent, role, and status", child)
	}

	items, err := repo.ListAgenByRuangWithParent(ctx, "ruang-1")
	if err != nil {
		t.Fatalf("list hierarchy: %v", err)
	}
	if len(items) != 2 || items[0].ParentID != nil || items[1].ParentID == nil {
		t.Fatalf("hierarchy = %+v, want root then child", items)
	}

	crossRoomParent := parent
	err = repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-cross", RuangID: "ruang-2", ParentID: &crossRoomParent, Name: "Cross", Role: "Pelaksana"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("cross-room parent error = %v, want ErrInvalid", err)
	}

	missingParent := domain.ID("missing")
	err = repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-missing", RuangID: "ruang-1", ParentID: &missingParent, Name: "Missing", Role: "Pelaksana"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("missing parent error = %v, want ErrInvalid", err)
	}

	self := domain.ID("agen-root")
	err = repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-root", RuangID: "ruang-1", ParentID: &self, Name: "Duplicate", Role: "Pemimpin"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("self parent error = %v, want ErrInvalid", err)
	}
}

func TestAgenHierarchySchema(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	var version int
	if err := db.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		t.Fatalf("schema version: %v", err)
	}
	if version != schemaVersion {
		t.Fatalf("schema version = %d, want %d", version, schemaVersion)
	}

	for _, column := range []string{"parent_id", "role", "status"} {
		var found int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pragma_table_info('agen') WHERE name = ?", column).Scan(&found); err != nil {
			t.Fatalf("%s column: %v", column, err)
		}
		if found != 1 {
			t.Fatalf("%s column count = %d, want 1", column, found)
		}
	}
}
