// Adapted from github.com/mnadev/adhango (MIT License), commit 61c85f04c17a00e63d833605756d0b4743c4e5aa.
// See LICENSE in internal/calc/vendor_adhan/. Vendored (not a live dependency) per
// Waqti's offline-first / no-supply-chain-risk requirement.
package calc

type NightPortions struct {
	Fajr float64
	Isha float64
}

func NewNightPortions(fajr float64, isha float64) (*NightPortions, error) {
	return &NightPortions{fajr, isha}, nil
}
