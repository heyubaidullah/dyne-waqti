package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/heyubaidullah/waqti/internal/calc"
	"github.com/heyubaidullah/waqti/internal/db"
)

var validCalcMethods = map[string]bool{
	string(calc.MethodMWL): true, string(calc.MethodISNA): true, string(calc.MethodEgyptian): true,
	string(calc.MethodKarachi): true, string(calc.MethodUmmAlQura): true,
}

var validAsrMethods = map[string]bool{
	string(calc.AsrStandard): true, string(calc.AsrHanafi): true,
}

type settingsPayload struct {
	Timezone           string `json:"timezone"`
	Latitude           string `json:"latitude"`
	Longitude          string `json:"longitude"`
	CalcMethod         string `json:"calc_method"`
	AsrMethod          string `json:"asr_method"`
	HijriAdjustDays    string `json:"hijri_adjust_days"`
	IqamahFajrMin      string `json:"iqamah_fajr_min"`
	IqamahDhuhrMin     string `json:"iqamah_dhuhr_min"`
	IqamahAsrMin       string `json:"iqamah_asr_min"`
	IqamahMaghribMin   string `json:"iqamah_maghrib_min"`
	IqamahIshaMin      string `json:"iqamah_isha_min"`
	TimingsDurationSec string `json:"timings_duration_sec"`
}

func (d *Deps) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	s, err := loadDisplaySettings(d.DB)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settingsPayload{
		Timezone:           s.Timezone,
		Latitude:           strconv.FormatFloat(s.Latitude, 'f', -1, 64),
		Longitude:          strconv.FormatFloat(s.Longitude, 'f', -1, 64),
		CalcMethod:         string(s.CalcMethod),
		AsrMethod:          string(s.AsrMethod),
		HijriAdjustDays:    strconv.Itoa(s.HijriAdjustDays),
		IqamahFajrMin:      strconv.Itoa(s.IqamahOffsets.FajrMin),
		IqamahDhuhrMin:     strconv.Itoa(s.IqamahOffsets.DhuhrMin),
		IqamahAsrMin:       strconv.Itoa(s.IqamahOffsets.AsrMin),
		IqamahMaghribMin:   strconv.Itoa(s.IqamahOffsets.MaghribMin),
		IqamahIshaMin:      strconv.Itoa(s.IqamahOffsets.IshaMin),
		TimingsDurationSec: strconv.Itoa(s.TimingsDurationSec),
	})
}

func (d *Deps) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req settingsPayload
	if !decodeJSON(w, r, &req) {
		return
	}

	if _, err := time.LoadLocation(req.Timezone); err != nil {
		respondError(w, http.StatusBadRequest, "invalid timezone: "+err.Error())
		return
	}
	if _, err := strconv.ParseFloat(req.Latitude, 64); err != nil {
		respondError(w, http.StatusBadRequest, "latitude must be a number")
		return
	}
	if _, err := strconv.ParseFloat(req.Longitude, 64); err != nil {
		respondError(w, http.StatusBadRequest, "longitude must be a number")
		return
	}
	if !validCalcMethods[req.CalcMethod] {
		respondError(w, http.StatusBadRequest, "invalid calc_method")
		return
	}
	if !validAsrMethods[req.AsrMethod] {
		respondError(w, http.StatusBadRequest, "invalid asr_method")
		return
	}
	intFields := map[string]string{
		"hijri_adjust_days": req.HijriAdjustDays, "iqamah_fajr_min": req.IqamahFajrMin,
		"iqamah_dhuhr_min": req.IqamahDhuhrMin, "iqamah_asr_min": req.IqamahAsrMin,
		"iqamah_maghrib_min": req.IqamahMaghribMin, "iqamah_isha_min": req.IqamahIshaMin,
		"timings_duration_sec": req.TimingsDurationSec,
	}
	for name, v := range intFields {
		if _, err := strconv.Atoi(v); err != nil {
			respondError(w, http.StatusBadRequest, name+" must be an integer")
			return
		}
	}

	values := map[string]string{
		SettingTimezone: req.Timezone, SettingLatitude: req.Latitude, SettingLongitude: req.Longitude,
		SettingCalcMethod: req.CalcMethod, SettingAsrMethod: req.AsrMethod,
		SettingHijriAdjustDays: req.HijriAdjustDays, SettingIqamahFajrMin: req.IqamahFajrMin,
		SettingIqamahDhuhrMin: req.IqamahDhuhrMin, SettingIqamahAsrMin: req.IqamahAsrMin,
		SettingIqamahMaghribMin: req.IqamahMaghribMin, SettingIqamahIshaMin: req.IqamahIshaMin,
		SettingTimingsDurationSec: req.TimingsDurationSec,
	}
	for key, value := range values {
		if err := db.SetSetting(d.DB, key, value); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save settings")
			return
		}
	}

	d.afterAdminWrite("settings")
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
