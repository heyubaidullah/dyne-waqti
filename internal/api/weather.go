package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// weatherRefreshInterval caps how often a new fetch is attempted, successful
// or not — Open-Meteo's current conditions don't change fast enough to
// justify polling more often, and this also throttles retries while the
// kiosk is offline.
const weatherRefreshInterval = 15 * time.Minute

// weatherStaleAfter is how long a last-known-good reading is still shown
// while newer fetches keep failing (e.g. the kiosk lost internet) before
// the display just stops showing weather at all, per the "fail gracefully,
// no weather block" requirement.
const weatherStaleAfter = 2 * time.Hour

// weatherBaseURL is a var (not a const) so tests can point it at an
// httptest.Server instead of the real network.
var weatherBaseURL = "https://api.open-meteo.com/v1/forecast"

var weatherHTTPClient = &http.Client{Timeout: 4 * time.Second}

type weatherView struct {
	Temp float64 `json:"temp"`
	Unit string  `json:"unit"` // "C" or "F" — whatever the masjid configured
	Code int     `json:"code"`
}

// WeatherCache holds the last successfully fetched reading plus enough
// bookkeeping to refresh it lazily, in the background, without ever
// blocking a display-data request on a slow or unreachable network call.
type WeatherCache struct {
	mu          sync.Mutex
	lastGood    *weatherView
	lastGoodAt  time.Time
	lastAttempt time.Time
	fetching    bool
}

func NewWeatherCache() *WeatherCache {
	return &WeatherCache{}
}

// Current returns the best available weather reading for (lat, lon) in the
// requested unit, kicking off a background refresh if the cache is stale
// or was fetched in a different unit (e.g. the admin just switched
// Fahrenheit/Celsius). It never blocks on the network — a cold or expired
// cache simply returns nil this call and (if a refresh wasn't already in
// flight) populates itself in time for a later one.
func (c *WeatherCache) Current(lat, lon float64, unit string) *weatherView {
	c.mu.Lock()
	unitChanged := c.lastGood != nil && c.lastGood.Unit != unit
	needsRefresh := (time.Since(c.lastAttempt) >= weatherRefreshInterval || unitChanged) && !c.fetching
	if needsRefresh {
		c.fetching = true
		c.lastAttempt = time.Now()
	}
	var result *weatherView
	// A cached reading in the wrong unit is worse than none — never show
	// e.g. a Celsius value mislabeled as Fahrenheit while the refetch is
	// still in flight.
	if c.lastGood != nil && c.lastGood.Unit == unit && time.Since(c.lastGoodAt) < weatherStaleAfter {
		result = c.lastGood
	}
	c.mu.Unlock()

	if needsRefresh {
		go c.refresh(lat, lon, unit)
	}
	return result
}

type openMeteoResponse struct {
	Current struct {
		Temperature float64 `json:"temperature_2m"`
		WeatherCode int     `json:"weather_code"`
	} `json:"current"`
}

func (c *WeatherCache) refresh(lat, lon float64, unit string) {
	defer func() {
		c.mu.Lock()
		c.fetching = false
		c.mu.Unlock()
	}()

	req, err := http.NewRequest(http.MethodGet, weatherBaseURL, nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	q.Set("latitude", formatCoord(lat))
	q.Set("longitude", formatCoord(lon))
	q.Set("current", "temperature_2m,weather_code")
	if unit == "F" {
		q.Set("temperature_unit", "fahrenheit")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := weatherHTTPClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}

	var parsed openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return
	}

	c.mu.Lock()
	c.lastGood = &weatherView{Temp: parsed.Current.Temperature, Unit: unit, Code: parsed.Current.WeatherCode}
	c.lastGoodAt = time.Now()
	c.mu.Unlock()
}

func formatCoord(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
