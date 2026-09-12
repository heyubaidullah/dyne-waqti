package api

import (
	"net/http"
	"strconv"

	"github.com/heyubaidullah/waqti/internal/db"
)

var validFontScales = map[string]bool{"small": true, "medium": true, "large": true}
var validWeatherUnits = map[string]bool{"C": true, "F": true}

const (
	minSilenceDurationAfterMin = 1
	maxSilenceDurationAfterMin = 15
)

type displaySettingsPayload struct {
	DisplayFontScale        string `json:"display_font_scale"`
	ShowGregorianDate       bool   `json:"show_gregorian_date"`
	SilenceDurationAfterMin string `json:"silence_duration_after_min"`
	MasjidName              string `json:"masjid_name"`
	ShowMasjidName          bool   `json:"show_masjid_name"`
	ShowMasjidLogoBanner    bool   `json:"show_masjid_logo_banner"`
	WeatherEnabled          bool   `json:"weather_enabled"`
	WeatherUnit             string `json:"weather_unit"`
}

func (d *Deps) handleGetDisplaySettings(w http.ResponseWriter, r *http.Request) {
	s, err := loadDisplaySettings(d.DB)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, displaySettingsPayload{
		DisplayFontScale:        s.DisplayFontScale,
		ShowGregorianDate:       s.ShowGregorianDate,
		SilenceDurationAfterMin: strconv.Itoa(s.SilenceDurationAfterMin),
		MasjidName:              s.MasjidName,
		ShowMasjidName:          s.ShowMasjidName,
		ShowMasjidLogoBanner:    s.ShowMasjidLogoBanner,
		WeatherEnabled:          s.WeatherEnabled,
		WeatherUnit:             s.WeatherUnit,
	})
}

func (d *Deps) handleUpdateDisplaySettings(w http.ResponseWriter, r *http.Request) {
	var req displaySettingsPayload
	if !decodeJSON(w, r, &req) {
		return
	}

	if !validFontScales[req.DisplayFontScale] {
		respondError(w, http.StatusBadRequest, "display_font_scale must be small, medium, or large")
		return
	}
	silenceMin, err := strconv.Atoi(req.SilenceDurationAfterMin)
	if err != nil || silenceMin < minSilenceDurationAfterMin || silenceMin > maxSilenceDurationAfterMin {
		respondError(w, http.StatusBadRequest, "silence_duration_after_min must be an integer between 1 and 15")
		return
	}
	if !validWeatherUnits[req.WeatherUnit] {
		respondError(w, http.StatusBadRequest, "weather_unit must be C or F")
		return
	}

	values := map[string]string{
		SettingDisplayFontScale:        req.DisplayFontScale,
		SettingShowGregorianDate:       boolToSetting(req.ShowGregorianDate),
		SettingSilenceDurationAfterMin: req.SilenceDurationAfterMin,
		SettingMasjidName:              req.MasjidName,
		SettingShowMasjidName:          boolToSetting(req.ShowMasjidName),
		SettingShowMasjidLogoBanner:    boolToSetting(req.ShowMasjidLogoBanner),
		SettingWeatherEnabled:          boolToSetting(req.WeatherEnabled),
		SettingWeatherUnit:             req.WeatherUnit,
	}
	for key, value := range values {
		if err := db.SetSetting(d.DB, key, value); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save display settings")
			return
		}
	}

	d.afterAdminWrite("display-settings")
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func boolToSetting(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
