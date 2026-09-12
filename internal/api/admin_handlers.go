package api

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/heyubaidullah/waqti/internal/db"
)

func isHHMM(s string) bool {
	_, err := time.Parse("15:04", s)
	return err == nil
}

// --- POST /api/v1/admin/prayer-times ---
//
// Exact, fixed Azaan/Iqamah clock times plus Jumu'ah — an admin sets the
// actual time, period, rather than an offset from a calculated time. This
// is also what fixed the pilot's "timings changed overnight" bug: these
// are persistent settings (never date-scoped), so nothing can revert them
// the next calendar day the way the old per-date prayer_schedules table did.

type prayerTimesRequest struct {
	AzaanFajrTime     string `json:"azaan_fajr_time"`
	AzaanDhuhrTime    string `json:"azaan_dhuhr_time"`
	AzaanAsrTime      string `json:"azaan_asr_time"`
	AzaanMaghribTime  string `json:"azaan_maghrib_time"`
	AzaanIshaTime     string `json:"azaan_isha_time"`
	IqamahFajrTime    string `json:"iqamah_fajr_time"`
	IqamahDhuhrTime   string `json:"iqamah_dhuhr_time"`
	IqamahAsrTime     string `json:"iqamah_asr_time"`
	IqamahMaghribTime string `json:"iqamah_maghrib_time"`
	IqamahIshaTime    string `json:"iqamah_isha_time"`
	JumuahCount       int    `json:"jumuah_count"`
	Jumuah1Iqamah     string `json:"jumuah_1_iqamah"` // HH:MM, optional
	Jumuah2Iqamah     string `json:"jumuah_2_iqamah"` // HH:MM, optional
}

// prayerTimesResponse adds a read-only "what the calculator says for today"
// reference alongside the admin's fixed times — helps them pick sensible
// values without the app ever silently substituting the calculated time.
type prayerTimesResponse struct {
	prayerTimesRequest
	Calculated *prayerTimesView `json:"calculated,omitempty"`
}

func (d *Deps) handleGetPrayerTimes(w http.ResponseWriter, r *http.Request) {
	s, err := loadDisplaySettings(d.DB)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}

	resp := prayerTimesResponse{
		prayerTimesRequest: prayerTimesRequest{
			AzaanFajrTime: s.AzaanTimes.Fajr, AzaanDhuhrTime: s.AzaanTimes.Dhuhr, AzaanAsrTime: s.AzaanTimes.Asr,
			AzaanMaghribTime: s.AzaanTimes.Maghrib, AzaanIshaTime: s.AzaanTimes.Isha,
			IqamahFajrTime: s.IqamahTimes.Fajr, IqamahDhuhrTime: s.IqamahTimes.Dhuhr, IqamahAsrTime: s.IqamahTimes.Asr,
			IqamahMaghribTime: s.IqamahTimes.Maghrib, IqamahIshaTime: s.IqamahTimes.Isha,
			JumuahCount: s.JumuahCount, Jumuah1Iqamah: s.Jumuah1Iqamah, Jumuah2Iqamah: s.Jumuah2Iqamah,
		},
	}
	if calculated, err := currentCalculatedTimes(s); err == nil {
		resp.Calculated = &prayerTimesView{
			Fajr: calculated.Fajr.Format("15:04"), Sunrise: calculated.Sunrise.Format("15:04"),
			Dhuhr: calculated.Dhuhr.Format("15:04"), Asr: calculated.Asr.Format("15:04"),
			Maghrib: calculated.Maghrib.Format("15:04"), Isha: calculated.Isha.Format("15:04"),
		}
	}
	respondJSON(w, http.StatusOK, resp)
}

func (d *Deps) handleUpdatePrayerTimes(w http.ResponseWriter, r *http.Request) {
	var req prayerTimesRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	timeFields := map[string]string{
		"azaan_fajr_time": req.AzaanFajrTime, "azaan_dhuhr_time": req.AzaanDhuhrTime, "azaan_asr_time": req.AzaanAsrTime,
		"azaan_maghrib_time": req.AzaanMaghribTime, "azaan_isha_time": req.AzaanIshaTime,
		"iqamah_fajr_time": req.IqamahFajrTime, "iqamah_dhuhr_time": req.IqamahDhuhrTime, "iqamah_asr_time": req.IqamahAsrTime,
		"iqamah_maghrib_time": req.IqamahMaghribTime, "iqamah_isha_time": req.IqamahIshaTime,
	}
	for name, v := range timeFields {
		if !isHHMM(v) {
			respondError(w, http.StatusBadRequest, name+" must be HH:MM (24h)")
			return
		}
	}
	if req.JumuahCount != 1 && req.JumuahCount != 2 {
		respondError(w, http.StatusBadRequest, "jumuah_count must be 1 or 2")
		return
	}
	if req.Jumuah1Iqamah != "" && !isHHMM(req.Jumuah1Iqamah) {
		respondError(w, http.StatusBadRequest, "jumuah_1_iqamah must be HH:MM (24h) or blank")
		return
	}
	if req.JumuahCount == 2 && req.Jumuah2Iqamah != "" && !isHHMM(req.Jumuah2Iqamah) {
		respondError(w, http.StatusBadRequest, "jumuah_2_iqamah must be HH:MM (24h) or blank")
		return
	}

	values := map[string]string{
		SettingAzaanFajrTime: req.AzaanFajrTime, SettingAzaanDhuhrTime: req.AzaanDhuhrTime, SettingAzaanAsrTime: req.AzaanAsrTime,
		SettingAzaanMaghribTime: req.AzaanMaghribTime, SettingAzaanIshaTime: req.AzaanIshaTime,
		SettingIqamahFajrTime: req.IqamahFajrTime, SettingIqamahDhuhrTime: req.IqamahDhuhrTime, SettingIqamahAsrTime: req.IqamahAsrTime,
		SettingIqamahMaghribTime: req.IqamahMaghribTime, SettingIqamahIshaTime: req.IqamahIshaTime,
		SettingJumuahCount: strconv.Itoa(req.JumuahCount), SettingJumuah1Iqamah: req.Jumuah1Iqamah,
		SettingJumuah2Iqamah: req.Jumuah2Iqamah,
	}
	for key, value := range values {
		if err := db.SetSetting(d.DB, key, value); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save prayer times")
			return
		}
	}

	d.afterAdminWrite("prayer-times")
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- POST /api/v1/admin/janazah ---

type janazahRequest struct {
	Action       string `json:"action"` // "publish" | "dismiss"
	Title        string `json:"title,omitempty"`
	DeceasedName string `json:"deceased_name,omitempty"`
	PrayerTime   string `json:"prayer_time,omitempty"`
	Location     string `json:"location,omitempty"`
}

func (d *Deps) handleJanazah(w http.ResponseWriter, r *http.Request) {
	var req janazahRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	switch req.Action {
	case "publish":
		if req.Title == "" || req.DeceasedName == "" || req.PrayerTime == "" || req.Location == "" {
			respondError(w, http.StatusBadRequest, "title, deceased_name, prayer_time, and location are required")
			return
		}
		if _, err := db.PublishEmergency(d.DB, db.EmergencyNotice{
			Title: req.Title, DeceasedName: req.DeceasedName, PrayerTime: req.PrayerTime, Location: req.Location,
		}); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to publish notice")
			return
		}
	case "dismiss":
		if err := db.DismissActiveEmergency(d.DB); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to dismiss notice")
			return
		}
	default:
		respondError(w, http.StatusBadRequest, `action must be "publish" or "dismiss"`)
		return
	}

	d.afterAdminWrite("janazah")
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- POST /api/v1/admin/blackout ---

type blackoutRequest struct {
	Active bool `json:"active"`
}

func (d *Deps) handleBlackout(w http.ResponseWriter, r *http.Request) {
	var req blackoutRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	value := "0"
	if req.Active {
		value = "1"
	}
	if err := db.SetSetting(d.DB, SettingBlackout, value); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save blackout state")
		return
	}

	d.afterAdminWrite("blackout")
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- GET /api/v1/admin/slides ---

type adminSlideView struct {
	ID                 int64  `json:"id"`
	Title              string `json:"title"`
	Type               string `json:"type"`
	ContentURLOrText   string `json:"content_url_or_text"`
	ArabicText         string `json:"arabic_text,omitempty"`
	IsActive           bool   `json:"is_active"`
	ExpirationDate     string `json:"expiration_date,omitempty"`
	DisplayDurationSec int    `json:"display_duration_sec"`
	DisplayMode        string `json:"display_mode"`
}

func (d *Deps) handleListAllSlides(w http.ResponseWriter, r *http.Request) {
	slides, err := db.ListSlides(d.DB, false)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load slides")
		return
	}
	views := make([]adminSlideView, 0, len(slides))
	for _, s := range slides {
		views = append(views, adminSlideView{
			ID: s.ID, Title: s.Title, Type: s.Type, ContentURLOrText: s.ContentURLOrText,
			ArabicText: s.ArabicText.String, IsActive: s.IsActive,
			ExpirationDate: s.ExpirationDate.String, DisplayDurationSec: s.DisplayDurationSec,
			DisplayMode: s.DisplayMode,
		})
	}
	respondJSON(w, http.StatusOK, views)
}

// --- POST /api/v1/admin/slides (multipart) ---

func (d *Deps) handleCreateSlide(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+1<<20)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		respondError(w, http.StatusBadRequest, "invalid multipart form: "+err.Error())
		return
	}

	title := r.FormValue("title")
	slideType := r.FormValue("type")
	if title == "" {
		respondError(w, http.StatusBadRequest, "title is required")
		return
	}

	var contentURLOrText string
	switch slideType {
	case "image":
		file, header, err := r.FormFile("file")
		if err != nil {
			respondError(w, http.StatusBadRequest, "file is required for image slides")
			return
		}
		defer file.Close()

		filename, err := saveUpload(d.Cfg.UploadsDir, file, header)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		contentURLOrText = "/uploads/" + filename
	case "text_verse":
		content := r.FormValue("content")
		if content == "" {
			respondError(w, http.StatusBadRequest, "content is required for text_verse slides")
			return
		}
		contentURLOrText = content
	default:
		respondError(w, http.StatusBadRequest, `type must be "image" or "text_verse"`)
		return
	}

	duration := 10
	if v := r.FormValue("display_duration_sec"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			duration = n
		}
	}

	var expiration sql.NullString
	if v := r.FormValue("expiration_date"); v != "" {
		if _, err := time.Parse("2006-01-02", v); err != nil {
			respondError(w, http.StatusBadRequest, "expiration_date must be YYYY-MM-DD")
			return
		}
		expiration = sql.NullString{String: v, Valid: true}
	}

	displayMode := r.FormValue("display_mode")
	if displayMode == "" {
		displayMode = "full"
	}
	if displayMode != "full" && displayMode != "in_screen" {
		respondError(w, http.StatusBadRequest, `display_mode must be "full" or "in_screen"`)
		return
	}

	id, err := db.InsertSlide(d.DB, db.Slide{
		Title: title, Type: slideType, ContentURLOrText: contentURLOrText,
		ArabicText: sql.NullString{String: r.FormValue("arabic_text"), Valid: r.FormValue("arabic_text") != ""},
		IsActive:   true, ExpirationDate: expiration, DisplayDurationSec: duration, DisplayMode: displayMode,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save slide")
		return
	}

	d.afterAdminWrite("slides")
	respondJSON(w, http.StatusCreated, map[string]any{"id": id, "content_url_or_text": contentURLOrText})
}

// --- PATCH /api/v1/admin/slides/{id} ---

type updateSlideRequest struct {
	IsActive       *bool   `json:"is_active,omitempty"`
	ExpirationDate *string `json:"expiration_date,omitempty"`
}

func (d *Deps) handleUpdateSlide(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid slide id")
		return
	}

	var req updateSlideRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	existing, err := db.GetSlide(d.DB, id)
	if err == db.ErrNotFound {
		respondError(w, http.StatusNotFound, "slide not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load slide")
		return
	}

	isActive := existing.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	expiration := existing.ExpirationDate
	if req.ExpirationDate != nil {
		if *req.ExpirationDate == "" {
			expiration = sql.NullString{}
		} else {
			if _, err := time.Parse("2006-01-02", *req.ExpirationDate); err != nil {
				respondError(w, http.StatusBadRequest, "expiration_date must be YYYY-MM-DD")
				return
			}
			expiration = sql.NullString{String: *req.ExpirationDate, Valid: true}
		}
	}

	if err := db.UpdateSlide(d.DB, id, isActive, expiration); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update slide")
		return
	}

	d.afterAdminWrite("slides")
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- DELETE /api/v1/admin/slides/{id} ---

func (d *Deps) handleDeleteSlide(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid slide id")
		return
	}

	existing, err := db.GetSlide(d.DB, id)
	if err == db.ErrNotFound {
		respondError(w, http.StatusNotFound, "slide not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load slide")
		return
	}

	if err := db.DeleteSlide(d.DB, id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete slide")
		return
	}

	if existing.Type == "image" {
		// Best-effort cleanup; a failure here doesn't affect the DB state
		// the client cares about.
		_ = os.Remove(filepath.Join(d.Cfg.UploadsDir, filepath.Base(existing.ContentURLOrText)))
	}

	d.afterAdminWrite("slides")
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
