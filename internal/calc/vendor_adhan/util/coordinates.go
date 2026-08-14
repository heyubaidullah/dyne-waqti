// Adapted from github.com/mnadev/adhango (MIT License), commit 61c85f04c17a00e63d833605756d0b4743c4e5aa.
// See LICENSE in internal/calc/vendor_adhan/. Vendored (not a live dependency) per
// Waqti's offline-first / no-supply-chain-risk requirement.
package util

import (
	"fmt"
)

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

func NewCoordinates(latitude float64, longitude float64) (*Coordinates, error) {
	if latitude > 90 || latitude < -90 {
		return nil, fmt.Errorf("latitude must be a number between -90 and 90 inclusive")
	}

	if longitude > 180 || longitude < -180 {
		return nil, fmt.Errorf("longitude must be a number between -180 and 180 inclusive")
	}

	return &Coordinates{
		Latitude:  latitude,
		Longitude: longitude,
	}, nil
}
