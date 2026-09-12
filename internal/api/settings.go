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
	SettingAzaanFajrMin       = "azaan_fajr_min"
	SettingAzaanDhuhrMin      = "azaan_dhuhr_min"
	SettingAzaanAsrMin        = "azaan_asr_min"
	SettingAzaanMaghribMin    = "azaan_maghrib_min"
	SettingAzaanIshaMin       = "azaan_isha_min"
	SettingJumuahCount        = "jumuah_count"
	SettingJumuah1Iqamah      = "jumuah_1_iqamah"
	SettingJumuah2Iqamah      = "jumuah_2_iqamah"
	SettingLogoURL            = "logo_url"
	SettingLogoHeightPx       = "logo_height_px"
	SettingTimingsDurationSec = "timings_duration_sec"

	// Display Settings — visual/kiosk preferences, independent of the
	// location/calculation settings above.
	SettingDisplayFontScale        = "display_font_scale"
	SettingShowGregorianDate       = "show_gregorian_date"
	SettingSilenceDurationAfterMin = "silence_duration_after_min"
	SettingMasjidName              = "masjid_name"
	SettingShowMasjidName          = "show_masjid_name"
	SettingShowMasjidLogoBanner    = "show_masjid_logo_banner"
	SettingWeatherEnabled          = "weather_enabled"
	SettingWeatherUnit             = "weather_unit" // "C" or "F"
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
	SettingTimezone:                "America/Chicago",
	SettingLatitude:                "29.4241",
	SettingLongitude:               "-98.4936",
	SettingCalcMethod:              string(calc.MethodISNA),
	SettingAsrMethod:               string(calc.AsrStandard),
	SettingHijriAdjustDays:         "0",
	SettingBlackout:                "0",
	SettingIqamahFajrMin:           "20",
	SettingIqamahDhuhrMin:          "10",
	SettingIqamahAsrMin:            "10",
	SettingIqamahMaghribMin:        "5",
	SettingIqamahIshaMin:           "10",
	SettingAzaanFajrMin:            "0",
	SettingAzaanDhuhrMin:           "0",
	SettingAzaanAsrMin:             "0",
	SettingAzaanMaghribMin:         "0",
	SettingAzaanIshaMin:            "0",
	SettingJumuahCount:             "1",
	SettingJumuah1Iqamah:           "",
	SettingJumuah2Iqamah:           "",
	SettingLogoURL:                 "",
	SettingLogoHeightPx:            "150",
	SettingTimingsDurationSec:      "15",
	SettingDisplayFontScale:        "medium",
	SettingShowGregorianDate:       "1",
	SettingSilenceDurationAfterMin: "7",
	SettingMasjidName:              "",
	SettingShowMasjidName:          "0",
	SettingShowMasjidLogoBanner:    "0",
	SettingWeatherEnabled:          "0",
	SettingWeatherUnit:             "F",
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

// JumuahSlot is one Jumu'ah Iqamah entry — either of up to two the display
// renders, in list form rather than two hardcoded always-present fields.
type JumuahSlot struct {
	Label  string
	Iqamah string // HH:MM
}

// DisplaySettings is the resolved, typed view of the settings table used to
// compute prayer times and Hijri date.
type DisplaySettings struct {
	Timezone                string
	Latitude                float64
	Longitude               float64
	CalcMethod              calc.Method
	AsrMethod               calc.AsrMethod
	HijriAdjustDays         int
	Blackout                bool
	LogoURL                 string
	LogoHeightPx            int
	TimingsDurationSec      int
	AzaanOffsets            calc.Offsets
	IqamahOffsets           calc.Offsets
	JumuahCount             int
	Jumuah1Iqamah           string
	Jumuah2Iqamah           string
	DisplayFontScale        string
	ShowGregorianDate       bool
	SilenceDurationAfterMin int
	MasjidName              string
	ShowMasjidName          bool
	ShowMasjidLogoBanner    bool
	WeatherEnabled          bool
	WeatherUnit             string
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

	jumuahCount := atoiOr(get(SettingJumuahCount, "1"), 1)
	if jumuahCount != 2 {
		jumuahCount = 1
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
		LogoHeightPx:       atoiOr(get(SettingLogoHeightPx, "150"), 150),
		TimingsDurationSec: atoiOr(get(SettingTimingsDurationSec, "15"), 15),
		AzaanOffsets: calc.Offsets{
			FajrMin:    atoiOr(get(SettingAzaanFajrMin, "0"), 0),
			DhuhrMin:   atoiOr(get(SettingAzaanDhuhrMin, "0"), 0),
			AsrMin:     atoiOr(get(SettingAzaanAsrMin, "0"), 0),
			MaghribMin: atoiOr(get(SettingAzaanMaghribMin, "0"), 0),
			IshaMin:    atoiOr(get(SettingAzaanIshaMin, "0"), 0),
		},
		IqamahOffsets: calc.Offsets{
			FajrMin:    atoiOr(get(SettingIqamahFajrMin, "20"), 20),
			DhuhrMin:   atoiOr(get(SettingIqamahDhuhrMin, "10"), 10),
			AsrMin:     atoiOr(get(SettingIqamahAsrMin, "10"), 10),
			MaghribMin: atoiOr(get(SettingIqamahMaghribMin, "5"), 5),
			IshaMin:    atoiOr(get(SettingIqamahIshaMin, "10"), 10),
		},
		JumuahCount:             jumuahCount,
		Jumuah1Iqamah:           get(SettingJumuah1Iqamah, ""),
		Jumuah2Iqamah:           get(SettingJumuah2Iqamah, ""),
		DisplayFontScale:        get(SettingDisplayFontScale, "medium"),
		ShowGregorianDate:       get(SettingShowGregorianDate, "1") == "1",
		SilenceDurationAfterMin: atoiOr(get(SettingSilenceDurationAfterMin, "7"), 7),
		MasjidName:              get(SettingMasjidName, ""),
		ShowMasjidName:          get(SettingShowMasjidName, "0") == "1",
		ShowMasjidLogoBanner:    get(SettingShowMasjidLogoBanner, "0") == "1",
		WeatherEnabled:          get(SettingWeatherEnabled, "0") == "1",
		WeatherUnit:             get(SettingWeatherUnit, "F"),
	}, nil
}
