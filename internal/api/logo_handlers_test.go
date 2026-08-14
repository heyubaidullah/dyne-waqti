package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadLogoValidAndReflectedInDisplayData(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	req := multipartSlideRequest(t, map[string]string{}, "file", "logo.png", tinyPNG)
	req.URL.Path = "/api/v1/admin/logo"
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var uploadResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &uploadResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	logoURL := uploadResp["logo_url"]
	if logoURL == "" || filepath.Base(logoURL) == "logo.png" {
		t.Fatalf("logo_url = %q, want a generated /uploads/<uuid>.png path", logoURL)
	}
	if _, err := os.Stat(filepath.Join(deps.Cfg.UploadsDir, filepath.Base(logoURL))); err != nil {
		t.Errorf("uploaded logo file not found: %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/display-data", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var displayData map[string]any
	json.Unmarshal(rec.Body.Bytes(), &displayData)
	if displayData["logo_url"] != logoURL {
		t.Errorf("display-data logo_url = %v, want %q", displayData["logo_url"], logoURL)
	}
}

func TestUploadLogoRejectsDisallowedType(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	req := multipartSlideRequest(t, map[string]string{}, "file", "logo.png", []byte("not an image"))
	req.URL.Path = "/api/v1/admin/logo"
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestDeleteLogoClearsSettingAndRemovesFile(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	req := multipartSlideRequest(t, map[string]string{}, "file", "logo.png", tinyPNG)
	req.URL.Path = "/api/v1/admin/logo"
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var uploadResp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &uploadResp)
	logoPath := filepath.Join(deps.Cfg.UploadsDir, filepath.Base(uploadResp["logo_url"]))

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/admin/logo", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200", rec.Code)
	}

	if _, err := os.Stat(logoPath); !os.IsNotExist(err) {
		t.Errorf("logo file still exists after delete: %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/display-data", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var displayData map[string]any
	json.Unmarshal(rec.Body.Bytes(), &displayData)
	if v, ok := displayData["logo_url"]; ok && v != "" {
		t.Errorf("display-data logo_url = %v, want empty/absent after delete", v)
	}
}

func TestLogoEndpointsRequireAuth(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/logo", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("DELETE unauth status = %d, want 401", rec.Code)
	}
}
