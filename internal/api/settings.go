package api

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/heyubaidullah/waqti/internal/calc"
	"github.com/heyubaidullah/waqti/internal/db"
)

// Setting keys stored in the generic `settings` key/value table.
const (
	SettingTimezone        = "timezone"
	SettingLatitude        = "latitude"
	SettingLongitude       = "longitude"
	SettingCalcMethod      = "calc_method"
	SettingAsrMethod       = "asr_method"
	SettingHijriAdjustDays = "hijri_adjust_days"
	SettingBlackout        = "blackout"
	SettingAzaanFajrTime   = "azaan_fajr_time"
	SettingAzaanDhuhrTime  = "azaan_dhuhr_time"
	SettingAzaanAsrTime    = "azaan_asr_time"
	SettingAzaanIshaTime   = "azaan_isha_time"
	SettingIqamahFajrTime  = "iqamah_fajr_time"
	SettingIqamahDhuhrTime = "iqamah_dhuhr_time"
	SettingIqamahAsrTime   = "iqamah_asr_time"
	SettingIqamahIshaTime  = "iqamah_isha_time"
	// Maghrib is the one prayer where Azaan can't sensibly be a fixed clock
	// time — sunset itself moves by minutes per day. Azaan always tracks
	// the calculated sunset time automatically; only the Iqamah delay after
	// it is admin-configurable.
	SettingIqamahMaghribOffsetMin = "iqamah_maghrib_offset_min"
	SettingJumuahCount            = "jumuah_count"
	SettingJumuah1Iqamah          = "jumuah_1_iqamah"
	SettingJumuah2Iqamah          = "jumuah_2_iqamah"
	SettingLogoURL                = "logo_url"
	SettingLogoHeightPx           = "logo_height_px"
	SettingTimingsDurationSec     = "timings_duration_sec"

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
	SettingTimezone:        "America/Chicago",
	SettingLatitude:        "29.4241",
	SettingLongitude:       "-98.4936",
	SettingCalcMethod:      string(calc.MethodISNA),
	SettingAsrMethod:       string(calc.AsrStandard),
	SettingHijriAdjustDays: "0",
	SettingBlackout:        "0",
	// Placeholder clock times for a fresh install — bear no relationship to
	// the seeded San Antonio coordinates above; an admin is expected to set
	// these to their own masjid's actual posted prayer schedule immediately,
	// the same way the coordinates need setting for accurate calculation.
	SettingAzaanFajrTime:           "06:00",
	SettingAzaanDhuhrTime:          "13:15",
	SettingAzaanAsrTime:            "17:00",
	SettingAzaanIshaTime:           "21:00",
	SettingIqamahFajrTime:          "06:20",
	SettingIqamahDhuhrTime:         "13:25",
	SettingIqamahAsrTime:           "17:10",
	SettingIqamahIshaTime:          "21:15",
	SettingIqamahMaghribOffsetMin:  "5",
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

// PrayerTimeSet holds one HH:MM clock time per prayer — an admin's exact,
// fixed schedule (for either Azaan or Iqamah), not a formula. It stays
// exactly what was entered until an admin changes it again; nothing
// recalculates it in the background. Maghrib is deliberately excluded:
// sunset moves too much day to day for a fixed clock time to stay
// accurate, so its Azaan always tracks the calculated sunset instead — see
// IqamahMaghribOffsetMin below.
type PrayerTimeSet struct {
	Fajr  string
	Dhuhr string
	Asr   string
	Isha  string
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
	AzaanTimes              PrayerTimeSet
	IqamahTimes             PrayerTimeSet
	IqamahMaghribOffsetMin  int
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
		AzaanTimes: PrayerTimeSet{
			Fajr:  get(SettingAzaanFajrTime, "06:00"),
			Dhuhr: get(SettingAzaanDhuhrTime, "13:15"),
			Asr:   get(SettingAzaanAsrTime, "17:00"),
			Isha:  get(SettingAzaanIshaTime, "21:00"),
		},
		IqamahTimes: PrayerTimeSet{
			Fajr:  get(SettingIqamahFajrTime, "06:20"),
			Dhuhr: get(SettingIqamahDhuhrTime, "13:25"),
			Asr:   get(SettingIqamahAsrTime, "17:10"),
			Isha:  get(SettingIqamahIshaTime, "21:15"),
		},
		IqamahMaghribOffsetMin:  atoiOr(get(SettingIqamahMaghribOffsetMin, "5"), 5),
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

// currentCalculatedTimes returns today's astronomically-calculated prayer
// times for s's configured location — used only as a read-only reference
// shown to the admin alongside the exact Azaan/Iqamah times they set (which
// are fixed, not derived from this). Never surfaced on the public
// /api/v1/display-data endpoint.
func currentCalculatedTimes(s DisplaySettings) (calc.Times, error) {
	loc, err := time.LoadLocation(s.Timezone)
	if err != nil {
		loc = time.UTC
	}
	return calc.Calculate(time.Now().In(loc), s.Latitude, s.Longitude, loc, s.CalcMethod, s.AsrMethod)
}
