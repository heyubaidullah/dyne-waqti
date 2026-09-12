package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestSessionCheckRequiresAuth(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/session", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want 401", rec.Code)
	}

	cookie := loginAndGetCookie(t, mux)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/session", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("authenticated status = %d, want 200", rec.Code)
	}
}

func TestGetSettingsReturnsDefaults(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var got settingsPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Timezone != "America/Chicago" || got.CalcMethod != "ISNA" {
		t.Errorf("defaults = %+v, want timezone=America/Chicago calc_method=ISNA", got)
	}
	if got.TimingsDurationSec != "15" {
		t.Errorf("TimingsDurationSec default = %q, want %q", got.TimingsDurationSec, "15")
	}
}

func TestUpdateSettingsValidatesAndPersists(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	valid := settingsPayload{
		Timezone: "America/New_York", Latitude: "40.7128", Longitude: "-74.0060",
		CalcMethod: "MWL", AsrMethod: "HANAFI", HijriAdjustDays: "1",
		LogoHeightPx: "150", TimingsDurationSec: "15",
	}

	// Invalid timezone is rejected.
	bad := valid
	bad.Timezone = "Not/A_Zone"
	body, _ := json.Marshal(bad)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid timezone status = %d, want 400", rec.Code)
	}

	// Invalid calc_method is rejected.
	bad = valid
	bad.CalcMethod = "MADE_UP"
	body, _ = json.Marshal(bad)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid calc_method status = %d, want 400", rec.Code)
	}

	// Valid payload is accepted and persisted.
	body, _ = json.Marshal(valid)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid payload status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var got settingsPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Timezone != valid.Timezone || got.CalcMethod != valid.CalcMethod || got.AsrMethod != valid.AsrMethod {
		t.Errorf("persisted settings = %+v, want %+v", got, valid)
	}
	if got.TimingsDurationSec != valid.TimingsDurationSec {
		t.Errorf("TimingsDurationSec = %q, want %q", got.TimingsDurationSec, valid.TimingsDurationSec)
	}
}

// TestUpdateSettingsAcceptsWideTimezoneSpread proves tzdata coverage isn't
// US-centric: a half-hour-offset zone, a zone with no DST at all, a
// Southern-Hemisphere zone (DST flips relative to the Northern Hemisphere),
// and a historically-tricky European zone all save and reload correctly.
// This is the broadened verification for the pilot's "timezone rejected"
// report — cmd/server/main.go already embeds the full IANA database
// (`_ "time/tzdata"`, shipped in v1.0.1), so this proves that fix generalizes
// well beyond the single US zone the older test covered.
func TestUpdateSettingsAcceptsWideTimezoneSpread(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	zones := []string{
		"Asia/Kolkata",     // UTC+5:30 — also covers Hyderabad, India (no distinct IANA zone)
		"Asia/Riyadh",      // no DST at all
		"Australia/Sydney", // Southern Hemisphere — DST months opposite the Northern Hemisphere
		"Europe/Istanbul",  // historically changed DST rules more than once
	}

	for _, zone := range zones {
		t.Run(zone, func(t *testing.T) {
			payload := settingsPayload{
				Timezone: zone, Latitude: "0", Longitude: "0",
				CalcMethod: "MWL", AsrMethod: "SHAFI", HijriAdjustDays: "0",
				LogoHeightPx: "150", TimingsDurationSec: "15",
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("save %s status = %d, want 200; body=%s", zone, rec.Code, rec.Body.String())
			}

			req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
			req.AddCookie(cookie)
			rec = httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			var got settingsPayload
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got.Timezone != zone {
				t.Errorf("reloaded timezone = %q, want %q", got.Timezone, zone)
			}

			// display-data must also load successfully with this zone —
			// proves time.LoadLocation succeeds end-to-end, not just at
			// the settings-save validation step.
			req = httptest.NewRequest(http.MethodGet, "/api/v1/display-data", nil)
			rec = httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("display-data with %s status = %d, want 200; body=%s", zone, rec.Code, rec.Body.String())
			}
			var displayData map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &displayData); err != nil {
				t.Fatalf("unmarshal display-data: %v", err)
			}
			if displayData["timezone"] != zone {
				t.Errorf("display-data timezone = %v, want %q", displayData["timezone"], zone)
			}
		})
	}
}

func TestUpdateSettingsRejectsNonIntegerTimingsDuration(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	bad := settingsPayload{
		Timezone: "America/New_York", Latitude: "40.7128", Longitude: "-74.0060",
		CalcMethod: "MWL", AsrMethod: "HANAFI", HijriAdjustDays: "1",
		LogoHeightPx: "150", TimingsDurationSec: "not-a-number",
	}
	body, _ := json.Marshal(bad)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestListAllSlidesIncludesInactive(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	// Create an image slide, then deactivate it.
	req := multipartSlideRequest(t, map[string]string{"title": "Flyer", "type": "image"}, "file", "flyer.png", tinyPNG)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create slide status = %d, want 201", rec.Code)
	}
	var created map[string]any
	json.Unmarshal(rec.Body.Bytes(), &created)
	id := int64(created["id"].(float64))

	patchBody, _ := json.Marshal(map[string]any{"is_active": false})
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/slides/"+strconv.FormatInt(id, 10), bytes.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch status = %d, want 200", rec.Code)
	}

	// Public display-data must NOT include the inactive slide.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/display-data", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var displayData map[string]any
	json.Unmarshal(rec.Body.Bytes(), &displayData)
	if slides, _ := displayData["slides"].([]any); len(slides) != 0 {
		t.Errorf("public display-data included %d slides, want 0 (inactive)", len(slides))
	}

	// Admin's all-slides listing MUST include it.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/slides", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var all []adminSlideView
	if err := json.Unmarshal(rec.Body.Bytes(), &all); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(all) != 1 || all[0].IsActive {
		t.Errorf("admin slides = %+v, want 1 inactive slide", all)
	}
}
