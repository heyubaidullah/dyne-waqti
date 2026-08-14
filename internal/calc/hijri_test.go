package calc

import (
	"testing"
	"time"
)

// TestHijriKnownDates spot-checks known Gregorian <-> Umm al-Qura
// correspondences (widely published reference dates).
func TestHijriKnownDates(t *testing.T) {
	cases := []struct {
		name      string
		gregorian time.Time
		wantYear  int
		wantMonth int
		wantDay   int
	}{
		// 1 Muharram 1447H corresponds to 2025-06-26 (Umm al-Qura).
		{"1447 New Year", time.Date(2025, 6, 26, 12, 0, 0, 0, time.UTC), 1447, 1, 1},
		// 1 Ramadan 1447H corresponds to 2026-02-18 (Umm al-Qura).
		{"1447 Ramadan start", time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC), 1447, 9, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := GregorianToHijri(c.gregorian, 0)
			if err != nil {
				t.Fatalf("GregorianToHijri: %v", err)
			}
			if got.Year != c.wantYear || got.Month != c.wantMonth || got.Day != c.wantDay {
				t.Errorf("GregorianToHijri(%v) = %+v, want {%d %d %d}",
					c.gregorian, got, c.wantYear, c.wantMonth, c.wantDay)
			}
		})
	}
}

// TestHijriAdjustmentRollover verifies the admin's manual +/-day adjustment
// correctly rolls over a Hijri month (and, at the year boundary, a year).
func TestHijriAdjustmentRollover(t *testing.T) {
	newYear := time.Date(2025, 6, 26, 12, 0, 0, 0, time.UTC) // 1 Muharram 1447H

	base, err := GregorianToHijri(newYear, 0)
	if err != nil {
		t.Fatalf("GregorianToHijri base: %v", err)
	}
	if base.Year != 1447 || base.Month != 1 || base.Day != 1 {
		t.Fatalf("base = %+v, want {1447 1 1}", base)
	}

	// -1 day must roll back across both the month AND year boundary, to
	// the last day of the prior Hijri year's final month (29 Dhu al-Hijjah 1446).
	minusOne, err := GregorianToHijri(newYear, -1)
	if err != nil {
		t.Fatalf("GregorianToHijri -1: %v", err)
	}
	if minusOne.Year != 1446 || minusOne.Month != 12 || minusOne.Day != 29 {
		t.Errorf("-1 day adjustment = %+v, want {1446 12 29}", minusOne)
	}

	// +2 days should advance into month 1, day 3, same year.
	plusTwo, err := GregorianToHijri(newYear, 2)
	if err != nil {
		t.Fatalf("GregorianToHijri +2: %v", err)
	}
	if plusTwo.Year != 1447 || plusTwo.Month != 1 || plusTwo.Day != 3 {
		t.Errorf("+2 day adjustment = %+v, want {1447 1 3}", plusTwo)
	}
}

func TestHijriOutOfRangeReturnsError(t *testing.T) {
	_, err := GregorianToHijri(time.Date(1800, 1, 1, 12, 0, 0, 0, time.UTC), 0)
	if err == nil {
		t.Error("expected error for out-of-range date, got nil")
	}
}
