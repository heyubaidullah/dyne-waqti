package calc

import (
	"fmt"
	"time"

	vhijri "github.com/heyubaidullah/waqti/internal/calc/vendor_hijri/hijri"
)

// HijriDate is a Y/M/D Islamic calendar date under the Umm al-Qura system.
type HijriDate struct {
	Year  int
	Month int
	Day   int
}

// GregorianToHijri converts a Gregorian date to its Umm al-Qura Hijri
// equivalent, deterministically (no dependency on browser/OS Intl support).
//
// adjustDays layers the admin's manual moon-sighting adjustment (typically
// -2..+2) on top of the tabular result. It is applied to the Gregorian
// input before conversion rather than to the Hijri output fields directly:
// since the Umm al-Qura mapping is a monotonic day-for-day correspondence
// (via Julian Day Number), shifting the input by N days is exactly
// equivalent to shifting the output by N days, and this way month/year
// rollovers (Hijri months are 29 or 30 days) are handled correctly by the
// same tabular lookup instead of needing separate rollover arithmetic.
func GregorianToHijri(t time.Time, adjustDays int) (HijriDate, error) {
	adjusted := t.AddDate(0, 0, adjustDays)
	uq, err := vhijri.CreateUmmAlQuraDate(adjusted)
	if err != nil {
		return HijriDate{}, fmt.Errorf("hijri: %w", err)
	}
	return HijriDate{
		Year:  int(uq.Year),
		Month: int(uq.Month),
		Day:   int(uq.Day),
	}, nil
}
