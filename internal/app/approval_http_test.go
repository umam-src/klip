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

func TestApprovalLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-approval", Name: "Approval"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-approval", RuangID: "ruang-approval", Title: "Pekerjaan", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}

	h := New(config.Default(t.TempDir()), nil, repo).Handler()
	post := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		return res
	}
	if res := post("/api/v1/pekerjaan/pekerjaan-approval/approval", `{"id":"approval-1","reason":"tindakan berisiko"}`); res.Code != http.StatusCreated {
		t.Fatalf("create approval: %d %s", res.Code, res.Body.String())
	}
	res := post("/api/v1/approval/approval-1/approve", `{}`)
	if res.Code != http.StatusOK {
		t.Fatalf("approve: %d %s", res.Code, res.Body.String())
	}
	var approval domain.Approval
	if err := json.NewDecoder(res.Body).Decode(&approval); err != nil {
		t.Fatal(err)
	}
	if approval.Status != domain.ApprovalApproved || approval.DecidedAt == nil {
		t.Fatalf("approval = %#v", approval)
	}
	if err := repo.DecideApproval(ctx, "approval-1", domain.ApprovalRejected, "late"); err == nil {
		t.Fatal("expected second decision to fail")
	}
}

func TestApprovalGate(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-gate", Name: "Gate"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-gate", RuangID: "ruang-gate", Title: "Pekerjaan", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	status, exists, err := repo.ApprovalGate(ctx, "pekerjaan-gate", nil)
	if err != nil || exists || status != "" {
		t.Fatalf("empty gate = %q %v %v", status, exists, err)
	}
	if err := repo.CreateApproval(ctx, domain.Approval{ID: "approval-gate", PekerjaanID: "pekerjaan-gate", Status: domain.ApprovalPending}); err != nil {
		t.Fatal(err)
	}
	status, exists, err = repo.ApprovalGate(ctx, "pekerjaan-gate", nil)
	if err != nil || !exists || status != domain.ApprovalPending {
		t.Fatalf("pending gate = %q %v %v", status, exists, err)
	}
}
