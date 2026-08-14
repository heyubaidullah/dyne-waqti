package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDisplayRedirectsToTrailingSlash(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/display", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want 301", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/display/" {
		t.Errorf("Location = %q, want %q", loc, "/display/")
	}
}

func TestDisplayServesEmbeddedIndex(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/display/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if rec.Body.Len() == 0 {
		t.Error("expected non-empty body")
	}
}
