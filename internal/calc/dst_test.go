package calc

import (
	"testing"
	"time"
)

// New York City coordinates — used for all DST test cases below.
const (
	nycLat = 40.7128
	nycLon = -74.0060
)

// assertOrdered checks Fajr < Sunrise < Dhuhr < Asr < Maghrib < Isha, which
// must hold regardless of which side of a DST transition the date falls on.
func assertOrdered(t *testing.T, times Times) {
	t.Helper()
	ordered := []struct {
		name string
		val  time.Time
	}{
		{"Fajr", times.Fajr},
		{"Sunrise", times.Sunrise},
		{"Dhuhr", times.Dhuhr},
		{"Asr", times.Asr},
		{"Maghrib", times.Maghrib},
		{"Isha", times.Isha},
	}
	for i := 1; i < len(ordered); i++ {
		if !ordered[i].val.After(ordered[i-1].val) {
			t.Errorf("%s (%v) is not after %s (%v)", ordered[i].name, ordered[i].val, ordered[i-1].name, ordered[i-1].val)
		}
	}
}

// TestDSTSpringForward covers America/New_York, 2026-03-08 — the wall clock
// jumps from 2:00am to 3:00am. A wrong Iqamah time here is the kind of
// failure a congregation notices immediately.
func TestDSTSpringForward(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	before := time.Date(2026, 3, 7, 12, 0, 0, 0, loc)
	transition := time.Date(2026, 3, 8, 12, 0, 0, 0, loc)

	beforeTimes, err := Calculate(before, nycLat, nycLon, loc, MethodISNA, AsrStandard)
	if err != nil {
		t.Fatalf("Calculate(before): %v", err)
	}
	assertOrdered(t, beforeTimes)

	transitionTimes, err := Calculate(transition, nycLat, nycLon, loc, MethodISNA, AsrStandard)
	if err != nil {
		t.Fatalf("Calculate(transition): %v", err)
	}
	assertOrdered(t, transitionTimes)

	// The transition date's offset must be EDT (UTC-4), not EST (UTC-5).
	if _, offset := transitionTimes.Dhuhr.Zone(); offset != -4*3600 {
		t.Errorf("2026-03-08 Dhuhr UTC offset = %d, want -14400 (EDT)", offset)
	}
	// Fajr and Dhuhr must land on the requested calendar date (no
	// accidental wraparound across the missing 2am-3am hour).
	if transitionTimes.Fajr.Day() != 8 {
		t.Errorf("Fajr date = %v, want day 8", transitionTimes.Fajr)
	}
	if transitionTimes.Dhuhr.Day() != 8 {
		t.Errorf("Dhuhr date = %v, want day 8", transitionTimes.Dhuhr)
	}
}

// TestDSTFallBack covers America/New_York, 2026-11-01 — the ambiguous hour
// where 1:00am-2:00am wall-clock time occurs twice.
func TestDSTFallBack(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	transition := time.Date(2026, 11, 1, 12, 0, 0, 0, loc)
	after := time.Date(2026, 11, 2, 12, 0, 0, 0, loc)

	transitionTimes, err := Calculate(transition, nycLat, nycLon, loc, MethodISNA, AsrStandard)
	if err != nil {
		t.Fatalf("Calculate(transition): %v", err)
	}
	assertOrdered(t, transitionTimes)

	afterTimes, err := Calculate(after, nycLat, nycLon, loc, MethodISNA, AsrStandard)
	if err != nil {
		t.Fatalf("Calculate(after): %v", err)
	}
	assertOrdered(t, afterTimes)

	// The transition date's offset must already be EST (UTC-5): DST ends
	// at 2am, and Dhuhr (noon-ish) falls well after the switch.
	if _, offset := transitionTimes.Dhuhr.Zone(); offset != -5*3600 {
		t.Errorf("2026-11-01 Dhuhr UTC offset = %d, want -18000 (EST)", offset)
	}

	// Consecutive days' Maghrib must not differ by ~25 hours (a sign the
	// ambiguous hour was double-counted) or ~23 hours (skipped).
	diff := afterTimes.Maghrib.Sub(transitionTimes.Maghrib)
	if diff < 23*time.Hour || diff > 25*time.Hour {
		t.Errorf("Maghrib gap across fall-back = %v, want roughly 24h", diff)
	}
}

// TestDSTAllCalculationMethods sanity-checks every supported method stays
// internally consistent across the spring-forward date.
func TestDSTAllCalculationMethods(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}
	date := time.Date(2026, 3, 8, 12, 0, 0, 0, loc)

	for _, method := range []Method{MethodMWL, MethodISNA, MethodEgyptian, MethodKarachi, MethodUmmAlQura} {
		times, err := Calculate(date, nycLat, nycLon, loc, method, AsrStandard)
		if err != nil {
			t.Fatalf("Calculate(%s): %v", method, err)
		}
		assertOrdered(t, times)
	}
}
