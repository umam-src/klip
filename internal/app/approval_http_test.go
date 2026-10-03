package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func siapkanApproval(t *testing.T) (context.Context, *storage.Repository, http.Handler) {
	t.Helper()
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-approval", Name: "Approval"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-approval", RuangID: "ruang-approval", Title: "Goal", Status: domain.GoalStatusActive}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateProyek(ctx, domain.Proyek{ID: "proyek-approval", RuangID: "ruang-approval", GoalID: "goal-approval", Title: "Proyek", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-approval", RuangID: "ruang-approval", ProyekID: "proyek-approval", Title: "Tugas", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}
	return ctx, repo, New(config.Default(t.TempDir()), nil, repo).Handler()
}

func TestApprovalLifecycle(t *testing.T) {
	ctx, repo, handler := siapkanApproval(t)
	post := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		return res
	}
	if res := post("/api/v1/tugas/tugas-approval/approval", `{"id":"approval-1","reason":"tindakan berisiko"}`); res.Code != http.StatusCreated {
		t.Fatalf("create approval: %d %s", res.Code, res.Body.String())
	}
	if events, err := repo.ListEventsByProyek(ctx, "proyek-approval", 10); err != nil || len(events) != 0 {
		t.Fatalf("approval create events = %#v, err = %v; approval bukan peristiwa eksekusi", events, err)
	}
	res := post("/api/v1/approval/approval-1/approve", `{}`)
	if res.Code != http.StatusOK {
		t.Fatalf("approve: %d %s", res.Code, res.Body.String())
	}
	var approval domain.Approval
	if err := json.NewDecoder(res.Body).Decode(&approval); err != nil {
		t.Fatal(err)
	}
	if approval.Status != domain.ApprovalApproved || approval.DecidedAt == nil || approval.ProyekID != "proyek-approval" || approval.TugasID == nil {
		t.Fatalf("approval = %#v", approval)
	}
	if events, err := repo.ListEventsByProyek(ctx, "proyek-approval", 10); err != nil || len(events) != 0 {
		t.Fatalf("approval decision events = %#v, err = %v; approval bukan peristiwa eksekusi", events, err)
	}
	if err := repo.DecideApproval(ctx, "approval-1", domain.ApprovalRejected, "late"); err == nil {
		t.Fatal("expected second decision to fail")
	}
}

func TestApprovalGate(t *testing.T) {
	ctx, repo, _ := siapkanApproval(t)
	status, exists, err := repo.ApprovalGate(ctx, "proyek-approval", nil)
	if err != nil || exists || status != "" {
		t.Fatalf("empty gate = %q %v %v", status, exists, err)
	}
	if err := repo.CreateApproval(ctx, domain.Approval{ID: "approval-gate", RuangID: "ruang-approval", ProyekID: "proyek-approval", Status: domain.ApprovalPending}); err != nil {
		t.Fatal(err)
	}
	status, exists, err = repo.ApprovalGate(ctx, "proyek-approval", nil)
	if err != nil || !exists || status != domain.ApprovalPending {
		t.Fatalf("pending gate = %q %v %v", status, exists, err)
	}
}
