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
}

type emergencyView struct {
	Title        string `json:"title"`
	DeceasedName string `json:"deceased_name"`
	PrayerTime   string `json:"prayer_time"`
	Location     string `json:"location"`
}

type displayDataResponse struct {
	Now                string          `json:"now"`
	Timezone           string          `json:"timezone"`
	Hijri              hijriView       `json:"hijri"`
	AdhanTimes         prayerTimesView `json:"adhan_times"`
	IqamahTimes        iqamahView      `json:"iqamah_times"`
	Slides             []slideView     `json:"slides"`
	Emergency          *emergencyView  `json:"emergency"`
	Blackout           bool            `json:"blackout"`
	LogoURL            string          `json:"logo_url,omitempty"`
	TimingsDurationSec int             `json:"timings_duration_sec"`
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

	adhan, err := calc.Calculate(now, settings.Latitude, settings.Longitude, loc, settings.CalcMethod, settings.AsrMethod)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "prayer time calculation failed: "+err.Error())
		return
	}
	computedIqamah := calc.ApplyIqamahOffsets(adhan, settings.IqamahOffsets)

	iqamah := iqamahView{
		Fajr:    computedIqamah.Fajr.Format("15:04"),
		Dhuhr:   computedIqamah.Dhuhr.Format("15:04"),
		Asr:     computedIqamah.Asr.Format("15:04"),
		Maghrib: computedIqamah.Maghrib.Format("15:04"),
		Isha:    computedIqamah.Isha.Format("15:04"),
		Jumuah:  computedIqamah.Dhuhr.Format("15:04"),
	}
	if stored, err := db.GetPrayerSchedule(d.DB, dateStr); err == nil {
		iqamah = iqamahView{
			Fajr: stored.FajrIqamah, Dhuhr: stored.DhuhrIqamah, Asr: stored.AsrIqamah,
			Maghrib: stored.MaghribIqamah, Isha: stored.IshaIqamah, Jumuah: stored.JumuahIqamah,
		}
	} else if !errors.Is(err, db.ErrNotFound) {
		respondError(w, http.StatusInternalServerError, "failed to load prayer schedule")
		return
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

	respondJSON(w, http.StatusOK, displayDataResponse{
		Now:      now.Format(time.RFC3339),
		Timezone: settings.Timezone,
		Hijri:    hijriView{Year: hijri.Year, Month: hijri.Month, Day: hijri.Day},
		AdhanTimes: prayerTimesView{
			Fajr: adhan.Fajr.Format("15:04"), Sunrise: adhan.Sunrise.Format("15:04"),
			Dhuhr: adhan.Dhuhr.Format("15:04"), Asr: adhan.Asr.Format("15:04"),
			Maghrib: adhan.Maghrib.Format("15:04"), Isha: adhan.Isha.Format("15:04"),
		},
		IqamahTimes:        iqamah,
		Slides:             slideViews,
		Emergency:          emergency,
		Blackout:           settings.Blackout,
		LogoURL:            settings.LogoURL,
		TimingsDurationSec: settings.TimingsDurationSec,
	})
}
