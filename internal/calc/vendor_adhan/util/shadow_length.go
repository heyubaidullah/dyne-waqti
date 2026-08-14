// Adapted from github.com/mnadev/adhango (MIT License), commit 61c85f04c17a00e63d833605756d0b4743c4e5aa.
// See LICENSE in internal/calc/vendor_adhan/. Vendored (not a live dependency) per
// Waqti's offline-first / no-supply-chain-risk requirement.
package util

type ShadowLength int64

const (
	SINGLE ShadowLength = iota

	DOUBLE
)

var ShadowLengthToFloatMap = map[ShadowLength]float64{
	SINGLE: 1.0,
	DOUBLE: 2.0,
}
