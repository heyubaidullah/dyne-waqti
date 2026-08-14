// Adapted from github.com/mnadev/adhango (MIT License), commit 61c85f04c17a00e63d833605756d0b4743c4e5aa.
// See LICENSE in internal/calc/vendor_adhan/. Vendored (not a live dependency) per
// Waqti's offline-first / no-supply-chain-risk requirement.
package calc

type HighLatitudeRule int64

const (
	NO_HIGH_LATITUDE_RULE HighLatitudeRule = iota

	// Fajr will never be earlier than the middle of the night, and Isha will never be later than
	// the middle of the night.
	MIDDLE_OF_THE_NIGHT

	// Fajr will never be earlier than the beginning of the last seventh of the night, and Isha will
	// never be later than the end of the first seventh of the night.
	SEVENTH_OF_THE_NIGHT

	// Similar to SEVENTH_OF_THE_NIGHT, but instead of 1/7th, the fraction of the night used
	// is fajrAngle / 60 and ishaAngle/60.
	TWILIGHT_ANGLE
)
