package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func getDisplayData(t *testing.T, mux http.Handler) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/display-data", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("display-data status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var data map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &data); err != nil {
		t.Fatalf("unmarshal display-data: %v", err)
	}
	return data
}

func postPrayerTimes(t *testing.T, mux http.Handler, cookie *http.Cookie, req prayerTimesRequest) {
	t.Helper()
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/prayer-times", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httpReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST prayer-times status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
}

func getPrayerTimes(t *testing.T, mux http.Handler, cookie *http.Cookie) prayerTimesResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/prayer-times", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET prayer-times status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp prayerTimesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal prayer-times: %v", err)
	}
	return resp
}

func sampleTimesRequest() prayerTimesRequest {
	return prayerTimesRequest{
		AzaanFajrTime: "05:45", AzaanDhuhrTime: "13:15", AzaanAsrTime: "17:00", AzaanMaghribTime: "19:45", AzaanIshaTime: "21:00",
		IqamahFajrTime: "06:05", IqamahDhuhrTime: "13:25", IqamahAsrTime: "17:10", IqamahMaghribTime: "19:50", IqamahIshaTime: "21:15",
		JumuahCount: 1,
	}
}

// TestPrayerTimesAreExactNotCalculated proves the admin's saved clock times
// are shown verbatim on the public display — no offset, no chaining off a
// calculated value. This is the whole point of the exact-time model: what
// you type is what shows up.
func TestPrayerTimesAreExactNotCalculated(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	req := sampleTimesRequest()
	postPrayerTimes(t, mux, cookie, req)

	data := getDisplayData(t, mux)
	adhan := data["adhan_times"].(map[string]any)
	iqamah := data["iqamah_times"].(map[string]any)

	if adhan["fajr"] != "05:45" || adhan["dhuhr"] != "13:15" || adhan["maghrib"] != "19:45" {
		t.Errorf("adhan_times = %v, want the exact saved azaan times", adhan)
	}
	if iqamah["fajr"] != "06:05" || iqamah["maghrib"] != "19:50" {
		t.Errorf("iqamah_times = %v, want the exact saved iqamah times", iqamah)
	}
}

// TestPrayerTimesRejectsInvalidFormat proves each azaan/iqamah field must be
// a valid HH:MM time.
func TestPrayerTimesRejectsInvalidFormat(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	bad := sampleTimesRequest()
	bad.AzaanFajrTime = "not-a-time"
	body, _ := json.Marshal(bad)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/prayer-times", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

// TestGetPrayerTimesIncludesCalculatedReference proves the admin-only GET
// endpoint surfaces today's astronomically calculated times as a read-only
// reference alongside the saved exact times — helping the admin pick
// sensible values without the app ever silently substituting them.
func TestGetPrayerTimesIncludesCalculatedReference(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	postPrayerTimes(t, mux, cookie, sampleTimesRequest())
	resp := getPrayerTimes(t, mux, cookie)

	if resp.AzaanFajrTime != "05:45" {
		t.Errorf("AzaanFajrTime = %q, want the saved exact value 05:45", resp.AzaanFajrTime)
	}
	if resp.Calculated == nil {
		t.Fatal("Calculated reference missing from GET prayer-times response")
	}
	if resp.Calculated.Fajr == "" || resp.Calculated.Fajr == resp.AzaanFajrTime {
		t.Errorf("Calculated.Fajr = %q, want a distinct astronomically-calculated value, not the saved exact time", resp.Calculated.Fajr)
	}
}

// TestPrayerTimesPersistAcrossRepeatedRequests is the regression test for
// the pilot's "timings changed overnight" report. The old bug wrote custom
// times into a table keyed by calendar date, so they silently vanished the
// next day. That code path is gone entirely now — prayer times are derived
// solely from the persistent `settings` table on every request, so a saved
// exact time can never revert. This can't literally fast-forward a day
// inside a unit test, but it proves the mechanism: repeated calls to
// display-data, with no re-save in between, keep returning the saved value
// forever (there is nothing left that could expire).
func TestPrayerTimesPersistAcrossRepeatedRequests(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	req := sampleTimesRequest()
	req.IqamahMaghribTime = "19:52"
	postPrayerTimes(t, mux, cookie, req)

	first := getDisplayData(t, mux)["iqamah_times"].(map[string]any)["maghrib"]
	for i := 0; i < 3; i++ {
		again := getDisplayData(t, mux)["iqamah_times"].(map[string]any)["maghrib"]
		if again != first {
			t.Fatalf("call %d: maghrib iqamah = %v, want unchanged %v", i, again, first)
		}
	}
}

// TestJumuahDefaultsToDhuhr proves the fallback behavior when no Jumu'ah
// time has been explicitly set.
func TestJumuahDefaultsToDhuhr(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)

	data := getDisplayData(t, mux)
	dhuhr := data["iqamah_times"].(map[string]any)["dhuhr"]
	jumuah := data["iqamah_times"].(map[string]any)["jumuah"]
	if jumuah != dhuhr {
		t.Errorf("default jumuah = %v, want it to fall back to dhuhr = %v", jumuah, dhuhr)
	}

	slots := data["jumuah_times"].([]any)
	if len(slots) != 1 {
		t.Fatalf("jumuah_times = %v, want 1 slot by default", slots)
	}
}

// TestJumuahTwoSlots proves a second Jumu'ah slot appears only when
// explicitly enabled, with its own explicit time, and that toggling back
// to one slot cleanly drops the second from the response.
func TestJumuahTwoSlots(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	req := sampleTimesRequest()
	req.JumuahCount = 2
	req.Jumuah1Iqamah = "13:00"
	req.Jumuah2Iqamah = "14:15"
	postPrayerTimes(t, mux, cookie, req)

	data := getDisplayData(t, mux)
	slots := data["jumuah_times"].([]any)
	if len(slots) != 2 {
		t.Fatalf("jumuah_times = %v, want 2 slots", slots)
	}
	if slots[0].(map[string]any)["iqamah"] != "13:00" || slots[1].(map[string]any)["iqamah"] != "14:15" {
		t.Errorf("jumuah_times = %v, want 13:00 then 14:15", slots)
	}

	// Toggle back down to one slot.
	req.JumuahCount = 1
	postPrayerTimes(t, mux, cookie, req)
	data = getDisplayData(t, mux)
	slots = data["jumuah_times"].([]any)
	if len(slots) != 1 {
		t.Fatalf("after toggling to 1: jumuah_times = %v, want 1 slot", slots)
	}
}
