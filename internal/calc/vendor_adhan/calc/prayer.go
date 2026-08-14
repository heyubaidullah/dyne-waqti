// Adapted from github.com/mnadev/adhango (MIT License), commit 61c85f04c17a00e63d833605756d0b4743c4e5aa.
// See LICENSE in internal/calc/vendor_adhan/. Vendored (not a live dependency) per
// Waqti's offline-first / no-supply-chain-risk requirement.
package calc

type Prayer int64

const (
	NO_PRAYER Prayer = iota

	FAJR

	SUNRISE

	DHUHR

	ASR

	MAGHRIB

	ISHA
)
