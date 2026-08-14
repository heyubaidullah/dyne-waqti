// Adapted from github.com/mnadev/adhango (MIT License), commit 61c85f04c17a00e63d833605756d0b4743c4e5aa.
// See LICENSE in internal/calc/vendor_adhan/. Vendored (not a live dependency) per
// Waqti's offline-first / no-supply-chain-risk requirement.
package data

import "time"

type DateComponents struct {
	Year  int
	Month int
	Day   int
}

// NewDateComponents creates a new DateComponents using the year, month and day in `t`.
func NewDateComponents(t time.Time) *DateComponents {
	return &DateComponents{Year: t.Year(), Month: int(t.Month()), Day: t.Day()}
}
