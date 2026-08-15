package api

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/heyubaidullah/waqti/internal/calc"
	"github.com/heyubaidullah/waqti/internal/db"
)

// Setting keys stored in the generic `settings` key/value table.
const (
	SettingTimezone           = "timezone"
	SettingLatitude           = "latitude"
	SettingLongitude          = "longitude"
	SettingCalcMethod         = "calc_method"
	SettingAsrMethod          = "asr_method"
	SettingHijriAdjustDays    = "hijri_adjust_days"
	SettingBlackout           = "blackout"
	SettingIqamahFajrMin      = "iqamah_fajr_min"
	SettingIqamahDhuhrMin     = "iqamah_dhuhr_min"
	SettingIqamahAsrMin       = "iqamah_asr_min"
	SettingIqamahMaghribMin   = "iqamah_maghrib_min"
	SettingIqamahIshaMin      = "iqamah_isha_min"
	SettingLogoURL            = "logo_url"
	SettingTimingsDurationSec = "timings_duration_sec"
)

// defaultSettings seed a usable-out-of-the-box configuration (UTC, ISNA, 0/0
// coordinates) so /api/v1/display-data works before an admin has configured
// anything. An admin must still set real coordinates/timezone for accurate
// prayer times; this only prevents Phase A from requiring UI-driven setup
// before the backend is testable end-to-end.
var defaultSettings = map[string]string{
	// San Antonio, TX (Central Time) — a friendlier out-of-the-box default
	// than UTC/0,0 for this deployment. Only seeded on a fresh data/
	// directory; an admin can change it from /admin at any time.
	SettingTimezone:           "America/Chicago",
	SettingLatitude:           "29.4241",
	SettingLongitude:          "-98.4936",
	SettingCalcMethod:         string(calc.MethodISNA),
	SettingAsrMethod:          string(calc.AsrStandard),
	SettingHijriAdjustDays:    "0",
	SettingBlackout:           "0",
	SettingIqamahFajrMin:      "20",
	SettingIqamahDhuhrMin:     "10",
	SettingIqamahAsrMin:       "10",
	SettingIqamahMaghribMin:   "5",
	SettingIqamahIshaMin:      "10",
	SettingLogoURL:            "",
	SettingTimingsDurationSec: "15",
}

// SeedDefaultSettings inserts any missing setting keys with defaults. It
// never overwrites a key an admin has already set.
func SeedDefaultSettings(database *sql.DB) error {
	for key, value := range defaultSettings {
		_, err := db.GetSetting(database, key)
		if errors.Is(err, db.ErrNotFound) {
			if err := db.SetSetting(database, key, value); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// DisplaySettings is the resolved, typed view of the settings table used to
// compute prayer times and Hijri date.
type DisplaySettings struct {
	Timezone           string
	Latitude           float64
	Longitude          float64
	CalcMethod         calc.Method
	AsrMethod          calc.AsrMethod
	HijriAdjustDays    int
	Blackout           bool
	LogoURL            string
	TimingsDurationSec int
	IqamahOffsets      calc.IqamahOffsets
}

func loadDisplaySettings(database *sql.DB) (DisplaySettings, error) {
	get := func(key, fallback string) string {
		v, err := db.GetSetting(database, key)
		if err != nil {
			return fallback
		}
		return v
	}
	atoiOr := func(s string, fallback int) int {
		n, err := strconv.Atoi(s)
		if err != nil {
			return fallback
		}
		return n
	}
	atofOr := func(s string, fallback float64) float64 {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fallback
		}
		return f
	}

	return DisplaySettings{
		Timezone:           get(SettingTimezone, "UTC"),
		Latitude:           atofOr(get(SettingLatitude, "0"), 0),
		Longitude:          atofOr(get(SettingLongitude, "0"), 0),
		CalcMethod:         calc.Method(get(SettingCalcMethod, string(calc.MethodISNA))),
		AsrMethod:          calc.AsrMethod(get(SettingAsrMethod, string(calc.AsrStandard))),
		HijriAdjustDays:    atoiOr(get(SettingHijriAdjustDays, "0"), 0),
		Blackout:           get(SettingBlackout, "0") == "1",
		LogoURL:            get(SettingLogoURL, ""),
		TimingsDurationSec: atoiOr(get(SettingTimingsDurationSec, "15"), 15),
		IqamahOffsets: calc.IqamahOffsets{
			FajrMin:    atoiOr(get(SettingIqamahFajrMin, "20"), 20),
			DhuhrMin:   atoiOr(get(SettingIqamahDhuhrMin, "10"), 10),
			AsrMin:     atoiOr(get(SettingIqamahAsrMin, "10"), 10),
			MaghribMin: atoiOr(get(SettingIqamahMaghribMin, "5"), 5),
			IshaMin:    atoiOr(get(SettingIqamahIshaMin, "10"), 10),
		},
	}, nil
}
