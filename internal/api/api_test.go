package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/heyubaidullah/waqti/internal/auth"
	"github.com/heyubaidullah/waqti/internal/config"
	"github.com/heyubaidullah/waqti/internal/db"
)

const testPassphrase = "correct-horse-battery-staple"

func newTestDeps(t *testing.T) *Deps {
	t.Helper()
	dir := t.TempDir()

	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}
	if err := SeedDefaultSettings(database); err != nil {
		t.Fatalf("SeedDefaultSettings: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(testPassphrase), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	if _, err := database.Exec(`INSERT INTO admin_credentials (id, passphrase_hash) VALUES (1, ?)`, string(hash)); err != nil {
		t.Fatalf("insert admin_credentials: %v", err)
	}

	uploadsDir := filepath.Join(dir, "uploads")
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		t.Fatalf("mkdir uploads: %v", err)
	}
	backupsDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupsDir, 0o755); err != nil {
		t.Fatalf("mkdir backups: %v", err)
	}

	return &Deps{
		DB:          database,
		Auth:        auth.NewManager(database),
		Cfg:         &config.Config{DataDir: dir, UploadsDir: uploadsDir, BackupsDir: backupsDir},
		Broadcaster: NewBroadcaster(),
	}
}

func loginAndGetCookie(t *testing.T, mux http.Handler) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(loginRequest{Passphrase: testPassphrase})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.CookieName {
			return c
		}
	}
	t.Fatal("no session cookie set on successful login")
	return nil
}

func TestUnauthenticatedRequestToAdminEndpointIs401(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)

	body, _ := json.Marshal(blackoutRequest{Active: true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/blackout", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}

	got, err := db.GetSetting(deps.DB, SettingBlackout)
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if got != "0" {
		t.Errorf("blackout setting = %q, want unchanged %q", got, "0")
	}
}

func TestLoginThenAuthenticatedRequestSucceeds(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)

	cookie := loginAndGetCookie(t, mux)

	body, _ := json.Marshal(blackoutRequest{Active: true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/blackout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	got, err := db.GetSetting(deps.DB, SettingBlackout)
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if got != "1" {
		t.Errorf("blackout setting = %q, want %q", got, "1")
	}
}

func TestLoginRateLimitReturns429After5Failures(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)

	doLogin := func(passphrase string) int {
		body, _ := json.Marshal(loginRequest{Passphrase: passphrase})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.0.2.1:5555"
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec.Code
	}

	for i := 0; i < 5; i++ {
		if code := doLogin("wrong"); code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want 401", i, code)
		}
	}

	if code := doLogin(testPassphrase); code != http.StatusTooManyRequests {
		t.Errorf("6th attempt (correct passphrase) status = %d, want 429", code)
	}
}

func multipartSlideRequest(t *testing.T, fields map[string]string, fileFieldName, fileName string, fileContent []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("WriteField: %v", err)
		}
	}
	if fileFieldName != "" {
		fw, err := w.CreateFormFile(fileFieldName, fileName)
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := fw.Write(fileContent); err != nil {
			t.Fatalf("write file content: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/slides", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

// tinyPNG is a minimal valid 1x1 PNG (enough for http.DetectContentType to
// recognize it as image/png).
var tinyPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, 0x00, 0x00, 0x00,
	0x0C, 0x49, 0x44, 0x41, 0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
	0x00, 0x03, 0x01, 0x01, 0x00, 0x18, 0xDD, 0x8D, 0xB0, 0x00, 0x00, 0x00,
	0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

func TestUploadValidPNGIsSavedWithGeneratedFilename(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	req := multipartSlideRequest(t, map[string]string{"title": "Flyer", "type": "image"},
		"file", "../../../etc/passwd.png", tinyPNG)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	contentURL, _ := resp["content_url_or_text"].(string)
	if contentURL == "" {
		t.Fatal("expected content_url_or_text in response")
	}
	savedName := filepath.Base(contentURL)
	if savedName == "passwd.png" || savedName == "../../../etc/passwd.png" {
		t.Errorf("original filename leaked into saved path: %q", contentURL)
	}

	// The file must exist strictly inside uploadsDir (no path traversal).
	if _, err := os.Stat(filepath.Join(deps.Cfg.UploadsDir, savedName)); err != nil {
		t.Errorf("uploaded file not found at expected generated-name path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(deps.Cfg.UploadsDir, "..", "etc", "passwd.png")); err == nil {
		t.Error("path traversal: file was written outside uploadsDir")
	}
}

func TestUploadRejectsDisallowedFileType(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	// A script disguised with a .png extension: content sniffing must
	// catch this since it never trusts the extension.
	maliciousContent := []byte("#!/bin/sh\necho pwned\n")
	req := multipartSlideRequest(t, map[string]string{"title": "Flyer", "type": "image"},
		"file", "evil.png", maliciousContent)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestUploadRejectsOversizedFile(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	oversized := make([]byte, maxUploadBytes+1024)
	copy(oversized, tinyPNG)

	req := multipartSlideRequest(t, map[string]string{"title": "Flyer", "type": "image"},
		"file", "big.png", oversized)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}
