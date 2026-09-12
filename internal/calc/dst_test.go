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

// TestDSTSouthernHemisphere covers Australia/Sydney, where DST runs
// opposite the Northern Hemisphere's calendar (starts ~October, ends
// ~April) — proves the DST-transition handling isn't accidentally only
// correct for the Northern Hemisphere zones exercised above.
func TestDSTSouthernHemisphere(t *testing.T) {
	const (
		sydneyLat = -33.8688
		sydneyLon = 151.2093
	)

	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	// 2026-04-05: Sydney's DST ends (clocks go back), opposite month from
	// the Northern Hemisphere fall-back case above.
	before := time.Date(2026, 4, 4, 12, 0, 0, 0, loc)
	transition := time.Date(2026, 4, 5, 12, 0, 0, 0, loc)

	beforeTimes, err := Calculate(before, sydneyLat, sydneyLon, loc, MethodMWL, AsrStandard)
	if err != nil {
		t.Fatalf("Calculate(before): %v", err)
	}
	assertOrdered(t, beforeTimes)

	transitionTimes, err := Calculate(transition, sydneyLat, sydneyLon, loc, MethodMWL, AsrStandard)
	if err != nil {
		t.Fatalf("Calculate(transition): %v", err)
	}
	assertOrdered(t, transitionTimes)

	// After DST ends, Sydney is AEST (UTC+10), not AEDT (UTC+11).
	if _, offset := transitionTimes.Dhuhr.Zone(); offset != 10*3600 {
		t.Errorf("2026-04-05 Dhuhr UTC offset = %d, want 36000 (AEST)", offset)
	}

	// 2026-10-04: Sydney's DST begins (clocks go forward).
	springForward := time.Date(2026, 10, 4, 12, 0, 0, 0, loc)
	springTimes, err := Calculate(springForward, sydneyLat, sydneyLon, loc, MethodMWL, AsrStandard)
	if err != nil {
		t.Fatalf("Calculate(springForward): %v", err)
	}
	assertOrdered(t, springTimes)
	if _, offset := springTimes.Dhuhr.Zone(); offset != 11*3600 {
		t.Errorf("2026-10-04 Dhuhr UTC offset = %d, want 39600 (AEDT)", offset)
	}
}

// TestDSTWithOffsetsApplied proves saved Azaan/Iqamah offsets survive a DST
// transition intact — the actual root cause of the pilot's "timings changed
// overnight" report was a data-model bug (per-date storage silently
// reverting the next day), not a DST calculation bug, but this closes the
// loop by confirming offset application itself is DST-transition-safe.
func TestDSTWithOffsetsApplied(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}
	azaan := Offsets{FajrMin: 3, DhuhrMin: 0, AsrMin: 0, MaghribMin: 2, IshaMin: 0}
	iqamah := Offsets{FajrMin: 20, DhuhrMin: 10, AsrMin: 10, MaghribMin: 5, IshaMin: 10}

	for _, date := range []time.Time{
		time.Date(2026, 3, 8, 12, 0, 0, 0, loc),  // spring-forward
		time.Date(2026, 11, 1, 12, 0, 0, 0, loc), // fall-back
	} {
		calculated, err := Calculate(date, nycLat, nycLon, loc, MethodISNA, AsrStandard)
		if err != nil {
			t.Fatalf("Calculate: %v", err)
		}
		azaanTimes := ApplyOffsets(calculated, azaan)
		iqamahTimes := ApplyOffsets(azaanTimes, iqamah)
		assertOrdered(t, iqamahTimes)

		wantFajrAzaan := calculated.Fajr.Add(3 * time.Minute)
		if !azaanTimes.Fajr.Equal(wantFajrAzaan) {
			t.Errorf("%v: Fajr azaan = %v, want %v", date, azaanTimes.Fajr, wantFajrAzaan)
		}
		wantFajrIqamah := wantFajrAzaan.Add(20 * time.Minute)
		if !iqamahTimes.Fajr.Equal(wantFajrIqamah) {
			t.Errorf("%v: Fajr iqamah = %v, want %v (azaan + 20min, chained not from raw calculated)", date, iqamahTimes.Fajr, wantFajrIqamah)
		}
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
