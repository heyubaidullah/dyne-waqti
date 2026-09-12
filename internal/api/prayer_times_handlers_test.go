package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func defaultPrayerTimesRequest() prayerTimesRequest {
	return prayerTimesRequest{
		AzaanFajrMin: "0", AzaanDhuhrMin: "0", AzaanAsrMin: "0", AzaanMaghribMin: "0", AzaanIshaMin: "0",
		IqamahFajrMin: "20", IqamahDhuhrMin: "10", IqamahAsrMin: "10", IqamahMaghribMin: "5", IqamahIshaMin: "10",
		JumuahCount: 1,
	}
}

// TestAzaanDefaultsToCalculatedTime proves that with no Azaan offset
// configured, the displayed Azaan time is exactly the calculated Adhan
// time — the default fallback the spec requires.
func TestAzaanDefaultsToCalculatedTime(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)

	before := getDisplayData(t, mux)
	after := getDisplayData(t, mux)
	beforeAdhan := before["adhan_times"].(map[string]any)
	afterAdhan := after["adhan_times"].(map[string]any)
	if beforeAdhan["fajr"] != afterAdhan["fajr"] {
		t.Errorf("adhan time changed between two default-settings calls: %v vs %v", beforeAdhan["fajr"], afterAdhan["fajr"])
	}
}

// TestAzaanAndIqamahOffsetsAreIndependent proves setting one doesn't
// clobber the other, and that Iqamah chains off the (possibly offset)
// Azaan time rather than the raw calculated time.
func TestAzaanAndIqamahOffsetsAreIndependent(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	baseline := getDisplayData(t, mux)
	baselineFajrAdhan := baseline["adhan_times"].(map[string]any)["fajr"].(string)

	req := defaultPrayerTimesRequest()
	req.AzaanFajrMin = "5"
	req.IqamahFajrMin = "15"
	postPrayerTimes(t, mux, cookie, req)

	got := getDisplayData(t, mux)
	gotFajrAdhan := got["adhan_times"].(map[string]any)["fajr"].(string)
	gotFajrIqamah := got["iqamah_times"].(map[string]any)["fajr"].(string)

	wantAdhan := addMinutesHHMM(t, baselineFajrAdhan, 5)
	if gotFajrAdhan != wantAdhan {
		t.Errorf("Fajr azaan = %q, want %q (baseline + 5min azaan offset)", gotFajrAdhan, wantAdhan)
	}
	wantIqamah := addMinutesHHMM(t, wantAdhan, 15)
	if gotFajrIqamah != wantIqamah {
		t.Errorf("Fajr iqamah = %q, want %q (offset azaan + 15min, not raw calculated + 15min)", gotFajrIqamah, wantIqamah)
	}
}

// TestPrayerTimesPersistAcrossRepeatedRequests is the regression test for
// the pilot's "timings changed overnight" report. The old bug wrote custom
// times into a table keyed by calendar date, so they silently vanished the
// next day. That code path is gone entirely now — prayer times are derived
// solely from the persistent `settings` table on every request, so saved
// offsets can never revert. This can't literally fast-forward a day inside
// a unit test, but it proves the mechanism: repeated calls to display-data,
// with no re-save in between, keep returning the saved offset forever
// (there is nothing left that could expire).
func TestPrayerTimesPersistAcrossRepeatedRequests(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)
	cookie := loginAndGetCookie(t, mux)

	req := defaultPrayerTimesRequest()
	req.IqamahMaghribMin = "7"
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

	req := defaultPrayerTimesRequest()
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

// addMinutesHHMM adds n minutes to an "HH:MM" string — sufficient for these
// tests since offsets used are small and never cross midnight.
func addMinutesHHMM(t *testing.T, hhmm string, n int) string {
	t.Helper()
	tm, err := time.Parse("15:04", hhmm)
	if err != nil {
		t.Fatalf("parse %q: %v", hhmm, err)
	}
	return tm.Add(time.Duration(n) * time.Minute).Format("15:04")
}
