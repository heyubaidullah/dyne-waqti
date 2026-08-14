// Package calc computes prayer times and the Hijri date for Waqti.
// It wraps vendored, offline, dependency-free astronomical (vendor_adhan)
// and Umm al-Qura tabular (vendor_hijri) calculation code — see the
// attribution headers and LICENSE files inside those directories.
package calc

import (
	"fmt"
	"time"

	vcalc "github.com/heyubaidullah/waqti/internal/calc/vendor_adhan/calc"
	vdata "github.com/heyubaidullah/waqti/internal/calc/vendor_adhan/data"
	vutil "github.com/heyubaidullah/waqti/internal/calc/vendor_adhan/util"
)

// Method is a supported prayer-time calculation method.
type Method string

const (
	MethodMWL       Method = "MWL"
	MethodISNA      Method = "ISNA"
	MethodEgyptian  Method = "EGYPTIAN"
	MethodKarachi   Method = "KARACHI"
	MethodUmmAlQura Method = "UMM_AL_QURA"
)

var methodMap = map[Method]vcalc.CalculationMethod{
	MethodMWL:       vcalc.MUSLIM_WORLD_LEAGUE,
	MethodISNA:      vcalc.NORTH_AMERICA,
	MethodEgyptian:  vcalc.EGYPTIAN,
	MethodKarachi:   vcalc.KARACHI,
	MethodUmmAlQura: vcalc.UMM_AL_QURA,
}

// AsrMethod is the juristic school used for Asr timing.
type AsrMethod string

const (
	AsrStandard AsrMethod = "SHAFI" // Shafi/Hanbali/Maliki — shadow length 1x
	AsrHanafi   AsrMethod = "HANAFI" // Hanafi — shadow length 2x, later Asr
)

var asrMap = map[AsrMethod]vcalc.AsrJuristicMethod{
	AsrStandard: vcalc.SHAFI_HANBALI_MALIKI,
	AsrHanafi:   vcalc.HANAFI,
}

// Times holds the five daily prayer times plus sunrise, all in the
// caller-supplied IANA timezone.
type Times struct {
	Fajr    time.Time
	Sunrise time.Time
	Dhuhr   time.Time
	Asr     time.Time
	Maghrib time.Time
	Isha    time.Time
}

// Calculate computes prayer times for the local calendar date (year/month/day
// of `date` as observed in `tz`) at the given coordinates. All times are
// returned already converted into `tz`, so the caller never has to reason
// about UTC or DST offsets directly.
func Calculate(date time.Time, lat, lon float64, tz *time.Location, method Method, asr AsrMethod) (Times, error) {
	vMethod, ok := methodMap[method]
	if !ok {
		return Times{}, fmt.Errorf("calc: unknown method %q", method)
	}
	vAsr, ok := asrMap[asr]
	if !ok {
		return Times{}, fmt.Errorf("calc: unknown asr method %q", asr)
	}

	coords, err := vutil.NewCoordinates(lat, lon)
	if err != nil {
		return Times{}, fmt.Errorf("calc: %w", err)
	}

	localDate := date.In(tz)
	dateComponents := &vdata.DateComponents{
		Year:  localDate.Year(),
		Month: int(localDate.Month()),
		Day:   localDate.Day(),
	}

	params := vcalc.GetMethodParameters(vMethod)
	params.Madhab = vAsr

	pt, err := vcalc.NewPrayerTimes(coords, dateComponents, params)
	if err != nil {
		return Times{}, fmt.Errorf("calc: %w", err)
	}

	return Times{
		Fajr:    pt.Fajr.In(tz),
		Sunrise: pt.Sunrise.In(tz),
		Dhuhr:   pt.Dhuhr.In(tz),
		Asr:     pt.Asr.In(tz),
		Maghrib: pt.Maghrib.In(tz),
		Isha:    pt.Isha.In(tz),
	}, nil
}

// IqamahOffsets holds the number of minutes after each Adhan (computed
// prayer time) that Iqamah is held.
type IqamahOffsets struct {
	FajrMin    int
	DhuhrMin   int
	AsrMin     int
	MaghribMin int
	IshaMin    int
}

// ApplyIqamahOffsets returns the Iqamah times derived from Adhan times t.
func ApplyIqamahOffsets(t Times, o IqamahOffsets) Times {
	return Times{
		Fajr:    t.Fajr.Add(time.Duration(o.FajrMin) * time.Minute),
		Sunrise: t.Sunrise,
		Dhuhr:   t.Dhuhr.Add(time.Duration(o.DhuhrMin) * time.Minute),
		Asr:     t.Asr.Add(time.Duration(o.AsrMin) * time.Minute),
		Maghrib: t.Maghrib.Add(time.Duration(o.MaghribMin) * time.Minute),
		Isha:    t.Isha.Add(time.Duration(o.IshaMin) * time.Minute),
	}
}
