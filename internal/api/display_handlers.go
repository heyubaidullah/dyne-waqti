package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/heyubaidullah/waqti/internal/calc"
	"github.com/heyubaidullah/waqti/internal/db"
)

type prayerTimesView struct {
	Fajr    string `json:"fajr"`
	Sunrise string `json:"sunrise"`
	Dhuhr   string `json:"dhuhr"`
	Asr     string `json:"asr"`
	Maghrib string `json:"maghrib"`
	Isha    string `json:"isha"`
}

type iqamahView struct {
	Fajr    string `json:"fajr"`
	Dhuhr   string `json:"dhuhr"`
	Asr     string `json:"asr"`
	Maghrib string `json:"maghrib"`
	Isha    string `json:"isha"`
	Jumuah  string `json:"jumuah"`
}

type jumuahSlotView struct {
	Label  string `json:"label"`
	Iqamah string `json:"iqamah"`
}

type hijriView struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type slideView struct {
	ID                 int64  `json:"id"`
	Title              string `json:"title"`
	Type               string `json:"type"`
	ContentURLOrText   string `json:"content_url_or_text"`
	ArabicText         string `json:"arabic_text,omitempty"`
	DisplayDurationSec int    `json:"display_duration_sec"`
	DisplayMode        string `json:"display_mode"`
}

type emergencyView struct {
	Title        string `json:"title"`
	DeceasedName string `json:"deceased_name"`
	PrayerTime   string `json:"prayer_time"`
	Location     string `json:"location"`
}

type displayDataResponse struct {
	Now                     string           `json:"now"`
	Timezone                string           `json:"timezone"`
	Hijri                   hijriView        `json:"hijri"`
	ShowGregorianDate       bool             `json:"show_gregorian_date"`
	AdhanTimes              prayerTimesView  `json:"adhan_times"`
	IqamahTimes             iqamahView       `json:"iqamah_times"`
	JumuahTimes             []jumuahSlotView `json:"jumuah_times"`
	Slides                  []slideView      `json:"slides"`
	Emergency               *emergencyView   `json:"emergency"`
	Blackout                bool             `json:"blackout"`
	LogoURL                 string           `json:"logo_url,omitempty"`
	LogoHeightPx            int              `json:"logo_height_px"`
	TimingsDurationSec      int              `json:"timings_duration_sec"`
	DisplayFontScale        string           `json:"display_font_scale"`
	SilenceDurationAfterMin int              `json:"silence_duration_after_min"`
	MasjidName              string           `json:"masjid_name,omitempty"`
	ShowMasjidName          bool             `json:"show_masjid_name"`
	ShowMasjidLogoBanner    bool             `json:"show_masjid_logo_banner"`
	Weather                 *weatherView     `json:"weather,omitempty"`
}

func (d *Deps) handleDisplayData(w http.ResponseWriter, r *http.Request) {
	settings, err := loadDisplaySettings(d.DB)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}

	loc, err := time.LoadLocation(settings.Timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	dateStr := now.Format("2006-01-02")

	// Sunrise has no admin override — it's informational only, always the
	// astronomically calculated value. Azaan/Iqamah are the admin's exact,
	// fixed clock times (settings.AzaanTimes/IqamahTimes below), not derived
	// from this calculation at all.
	calculated, err := currentCalculatedTimes(settings)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "prayer time calculation failed: "+err.Error())
		return
	}

	jumuah1 := settings.Jumuah1Iqamah
	if jumuah1 == "" {
		jumuah1 = settings.IqamahTimes.Dhuhr
	}
	jumuahTimes := []jumuahSlotView{{Label: "Jumu'ah", Iqamah: jumuah1}}
	if settings.JumuahCount == 2 {
		jumuah2 := settings.Jumuah2Iqamah
		if jumuah2 == "" {
			jumuah2 = settings.IqamahTimes.Dhuhr
		}
		jumuahTimes = append(jumuahTimes, jumuahSlotView{Label: "Jumu'ah 2", Iqamah: jumuah2})
	}

	iqamah := iqamahView{
		Fajr:    settings.IqamahTimes.Fajr,
		Dhuhr:   settings.IqamahTimes.Dhuhr,
		Asr:     settings.IqamahTimes.Asr,
		Maghrib: settings.IqamahTimes.Maghrib,
		Isha:    settings.IqamahTimes.Isha,
		Jumuah:  jumuah1,
	}

	hijri, err := calc.GregorianToHijri(now, settings.HijriAdjustDays)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "hijri calculation failed: "+err.Error())
		return
	}

	slides, err := db.ListSlides(d.DB, true)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load slides")
		return
	}
	slideViews := make([]slideView, 0, len(slides))
	for _, s := range slides {
		// Basic expiration check (per spec, polish beyond this is out of
		// scope for v0.1.0-poc): skip slides whose expiration date has passed.
		if s.ExpirationDate.Valid && s.ExpirationDate.String < dateStr {
			continue
		}
		slideViews = append(slideViews, slideView{
			ID: s.ID, Title: s.Title, Type: s.Type, ContentURLOrText: s.ContentURLOrText,
			ArabicText: s.ArabicText.String, DisplayDurationSec: s.DisplayDurationSec,
			DisplayMode: s.DisplayMode,
		})
	}

	var emergency *emergencyView
	if notice, err := db.GetActiveEmergency(d.DB); err == nil {
		emergency = &emergencyView{
			Title: notice.Title, DeceasedName: notice.DeceasedName,
			PrayerTime: notice.PrayerTime, Location: notice.Location,
		}
	} else if !errors.Is(err, db.ErrNotFound) {
		respondError(w, http.StatusInternalServerError, "failed to load emergency notice")
		return
	}

	var weather *weatherView
	if settings.WeatherEnabled && d.Weather != nil {
		weather = d.Weather.Current(settings.Latitude, settings.Longitude, settings.WeatherUnit)
	}

	respondJSON(w, http.StatusOK, displayDataResponse{
		Now:               now.Format(time.RFC3339),
		Timezone:          settings.Timezone,
		Hijri:             hijriView{Year: hijri.Year, Month: hijri.Month, Day: hijri.Day},
		ShowGregorianDate: settings.ShowGregorianDate,
		AdhanTimes: prayerTimesView{
			Fajr: settings.AzaanTimes.Fajr, Sunrise: calculated.Sunrise.Format("15:04"),
			Dhuhr: settings.AzaanTimes.Dhuhr, Asr: settings.AzaanTimes.Asr,
			Maghrib: settings.AzaanTimes.Maghrib, Isha: settings.AzaanTimes.Isha,
		},
		IqamahTimes:             iqamah,
		JumuahTimes:             jumuahTimes,
		Slides:                  slideViews,
		Emergency:               emergency,
		Blackout:                settings.Blackout,
		LogoURL:                 settings.LogoURL,
		LogoHeightPx:            settings.LogoHeightPx,
		TimingsDurationSec:      settings.TimingsDurationSec,
		DisplayFontScale:        settings.DisplayFontScale,
		SilenceDurationAfterMin: settings.SilenceDurationAfterMin,
		MasjidName:              settings.MasjidName,
		ShowMasjidName:          settings.ShowMasjidName,
		ShowMasjidLogoBanner:    settings.ShowMasjidLogoBanner,
		Weather:                 weather,
	})
}
