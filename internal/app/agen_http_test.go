package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func TestHandlerAgentDetailAndUpdate(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-http", Name: "Ruang HTTP"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-root", RuangID: "ruang-http", Name: "Root", Role: "Pemimpin"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-child", RuangID: "ruang-http", Name: "Child", Role: "Pelaksana"}); err != nil {
		t.Fatal(err)
	}

	handler := New(config.Default(t.TempDir()), fakeProvider{}, repo).Handler()

	get := httptest.NewRequest(http.MethodGet, "/api/v1/ruang/ruang-http/agen?agent_id=agen-child", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, get)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"role":"Pelaksana"`) {
		t.Fatalf("get agent status = %d; body = %s", res.Code, res.Body.String())
	}

	update := httptest.NewRequest(http.MethodPut, "/api/v1/ruang/ruang-http/agen?agent_id=agen-child", strings.NewReader(`{"name":"Child Baru","role":"Operator","description":"Deskripsi","parent_id":"agen-root","provider_id":"openai-compatible","model_id":"model-1"}`))
	update.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, update)
	if res.Code != http.StatusOK {
		t.Fatalf("update agent status = %d; body = %s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	for _, want := range []string{`"name":"Child Baru"`, `"role":"Operator"`, `"parent_id":"agen-root"`, `"model_id":"model-1"`, `"status":"active"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("update response = %s, want %s", body, want)
		}
	}

	badParent := httptest.NewRequest(http.MethodPut, "/api/v1/ruang/ruang-http/agen?agent_id=agen-child", strings.NewReader(`{"name":"Child","role":"Operator","parent_id":"missing"}`))
	badParent.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, badParent)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("bad parent status = %d; body = %s", res.Code, res.Body.String())
	}

	missing := httptest.NewRequest(http.MethodGet, "/api/v1/ruang/ruang-http/agen?agent_id=missing", nil)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, missing)
	if res.Code != http.StatusNotFound {
		t.Fatalf("missing agent status = %d; body = %s", res.Code, res.Body.String())
	}
}
