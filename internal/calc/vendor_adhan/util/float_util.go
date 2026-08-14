// Adapted from github.com/mnadev/adhango (MIT License), commit 61c85f04c17a00e63d833605756d0b4743c4e5aa.
// See LICENSE in internal/calc/vendor_adhan/. Vendored (not a live dependency) per
// Waqti's offline-first / no-supply-chain-risk requirement.
package util

import "math"

func NormalizeWithBound(value float64, max float64) float64 {
	return value - (max * math.Floor(value/max))
}

func UnwindAngle(value float64) float64 {
	return NormalizeWithBound(value, 360.0)
}

func ClosestAngle(angle float64) float64 {
	if angle >= -180 && angle <= 180 {
		return angle

	}

	return angle - (360.0 * math.Round(angle/360.0))
}
