package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// waitFor polls fn until it returns true or the timeout elapses, for
// asserting on the background refresh goroutine's effects without a
// fixed sleep.
func waitFor(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("condition not met within %v", timeout)
}

func TestWeatherCacheFetchesAndCaches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"current": map[string]any{"temperature_2m": 21.5, "weather_code": 3},
		})
	}))
	defer srv.Close()

	orig := weatherBaseURL
	weatherBaseURL = srv.URL
	defer func() { weatherBaseURL = orig }()

	c := NewWeatherCache()
	if got := c.Current(30, -97); got != nil {
		t.Fatalf("first call (cold cache) = %+v, want nil (fetch is async)", got)
	}

	waitFor(t, time.Second, func() bool { return c.Current(30, -97) != nil })
	got := c.Current(30, -97)
	if got.TempC != 21.5 || got.Code != 3 {
		t.Errorf("cached weather = %+v, want temp=21.5 code=3", got)
	}
}

func TestWeatherCacheOmittedOnFetchFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	orig := weatherBaseURL
	weatherBaseURL = srv.URL
	defer func() { weatherBaseURL = orig }()

	c := NewWeatherCache()
	c.Current(30, -97) // kicks off the (failing) background fetch

	// Give the failing fetch a moment to complete, then confirm the cache
	// stays empty rather than surfacing an error — "no weather block
	// shown" is the required graceful-failure behavior, not an error field.
	time.Sleep(100 * time.Millisecond)
	if got := c.Current(30, -97); got != nil {
		t.Errorf("weather after failed fetch = %+v, want nil", got)
	}
}

func TestDisplayDataOmitsWeatherWhenDisabled(t *testing.T) {
	deps := newTestDeps(t)
	mux := NewRouter(deps)

	data := getDisplayData(t, mux)
	if _, ok := data["weather"]; ok {
		t.Errorf("weather key present in display-data with weather disabled (default): %v", data["weather"])
	}
}
