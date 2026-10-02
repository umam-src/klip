package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/storage"
)

func TestHandlerAgentHierarchy(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, ruangForHierarchyTest("ruang-http")); err != nil {
		t.Fatal(err)
	}

	handler := New(config.Default(t.TempDir()), fakeProvider{}, repo).Handler()
	create := func(body string) string {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ruang/ruang-http/agen", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusCreated {
			t.Fatalf("create agent status = %d; body = %s", res.Code, res.Body.String())
		}
		return res.Body.String()
	}

	rootBody := create(`{"id":"agen-root","name":"Root"}`)
	if !strings.Contains(rootBody, `"parent_id"`) {
		// parent_id is omitted for root agents by design.
		if strings.Contains(rootBody, `"parent_id":`) {
			t.Fatalf("unexpected root parent: %s", rootBody)
		}
	}
	childBody := create(`{"id":"agen-child","parent_id":"agen-root","name":"Child"}`)
	if !strings.Contains(childBody, `"parent_id":"agen-root"`) {
		t.Fatalf("child response = %s", childBody)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ruang/ruang-http/agen", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("list agents status = %d; body = %s", res.Code, res.Body.String())
	}
	if body := res.Body.String(); !strings.Contains(body, `"id":"agen-root"`) || !strings.Contains(body, `"id":"agen-child"`) || !strings.Contains(body, `"parent_id":"agen-root"`) {
		t.Fatalf("hierarchy response = %s", body)
	}
}

func ruangForHierarchyTest(id string) (ruang storageTestRuang) {
	return storageTestRuang{ID: id, Name: "Ruang HTTP"}
}

type storageTestRuang struct {
	ID   string
	Name string
}

func (r storageTestRuang) toDomain() (result struct {
	ID   string
	Name string
}) {
	return struct {
		ID   string
		Name string
	}{ID: r.ID, Name: r.Name}
}
