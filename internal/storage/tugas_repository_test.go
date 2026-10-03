package storage

import (
	"context"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

func TestTugasNativeRepositoryRoundTrip(t *testing.T) {
	db := newTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	ruangID := domain.ID("ruang-1")
	goalID := domain.ID("goal-1")
	proyekID := domain.ID("proyek-1")
	tugasID := domain.ID("tugas-1")

	mustExec(t, db, `INSERT INTO ruang (id, name, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, ruangID, "Ruang", "active", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z")
	mustExec(t, db, `INSERT INTO goal (id, ruang_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, goalID, ruangID, "Goal", "", "active", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z")
	mustExec(t, db, `INSERT INTO proyek (id, ruang_id, goal_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, proyekID, ruangID, goalID, "Proyek", "", "draft", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z")

	tugas := domain.Tugas{ID: tugasID, RuangID: ruangID, ProyekID: proyekID, Title: "Tugas", Status: domain.StatusDraft}
	if err := repo.CreateTugasNative(ctx, tugas); err != nil {
		t.Fatalf("CreateTugasNative() error = %v", err)
	}

	got, err := repo.GetTugasNative(ctx, tugasID)
	if err != nil {
		t.Fatalf("GetTugasNative() error = %v", err)
	}
	if got.ID != tugasID || got.RuangID != ruangID || got.ProyekID != proyekID || got.Title != tugas.Title {
		t.Fatalf("GetTugasNative() = %+v, want %+v", got, tugas)
	}

	list, err := repo.ListTugasByProyek(ctx, proyekID)
	if err != nil {
		t.Fatalf("ListTugasByProyek() error = %v", err)
	}
	if len(list) != 1 || list[0].ID != tugasID {
		t.Fatalf("ListTugasByProyek() = %+v, want one task %q", list, tugasID)
	}
}

func TestTugasNativeRepositoryRejectsCrossProjectParent(t *testing.T) {
	db := newTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	mustExec(t, db, `INSERT INTO ruang (id, name, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, "ruang-1", "Ruang", "active", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z")
	mustExec(t, db, `INSERT INTO goal (id, ruang_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, "goal-1", "ruang-1", "Goal", "", "active", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z")
	mustExec(t, db, `INSERT INTO proyek (id, ruang_id, goal_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "proyek-1", "ruang-1", "goal-1", "Proyek 1", "", "draft", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z")
	mustExec(t, db, `INSERT INTO proyek (id, ruang_id, goal_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "proyek-2", "ruang-1", "goal-1", "Proyek 2", "", "draft", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z")
	mustExec(t, db, `INSERT INTO tugas (id, ruang_id, proyek_id, title, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, "parent", "ruang-1", "proyek-2", "Parent", "draft", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z")

	parent := domain.ID("parent")
	err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "child", RuangID: "ruang-1", ProyekID: "proyek-1", ParentID: &parent, Title: "Child", Status: domain.StatusDraft})
	if err == nil {
		t.Fatal("CreateTugasNative() error = nil, want cross-project parent rejection")
	}
}
