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

type prayerTimesRequest struct {
	Date            string `json:"date,omitempty"` // YYYY-MM-DD, defaults to today
	FajrIqamah      string `json:"fajr_iqamah"`
	DhuhrIqamah     string `json:"dhuhr_iqamah"`
	AsrIqamah       string `json:"asr_iqamah"`
	MaghribIqamah   string `json:"maghrib_iqamah"`
	IshaIqamah      string `json:"isha_iqamah"`
	JumuahIqamah    string `json:"jumuah_iqamah"`
	HijriAdjustDays *int   `json:"hijri_adjust_days,omitempty"`
}

func (d *Deps) handleUpdatePrayerTimes(w http.ResponseWriter, r *http.Request) {
	var req prayerTimesRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	for _, v := range []string{req.FajrIqamah, req.DhuhrIqamah, req.AsrIqamah, req.MaghribIqamah, req.IshaIqamah, req.JumuahIqamah} {
		if !isHHMM(v) {
			respondError(w, http.StatusBadRequest, "iqamah times must be HH:MM (24h)")
			return
		}
	}

	date := req.Date
	if date == "" {
		settings, err := loadDisplaySettings(d.DB)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to load settings")
			return
		}
		loc, err := time.LoadLocation(settings.Timezone)
		if err != nil {
			loc = time.UTC
		}
		date = time.Now().In(loc).Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", date); err != nil {
		respondError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}

	if err := db.UpsertPrayerSchedule(d.DB, db.PrayerSchedule{
		Date: date, FajrIqamah: req.FajrIqamah, DhuhrIqamah: req.DhuhrIqamah, AsrIqamah: req.AsrIqamah,
		MaghribIqamah: req.MaghribIqamah, IshaIqamah: req.IshaIqamah, JumuahIqamah: req.JumuahIqamah,
	}); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save prayer schedule")
		return
	}

	if req.HijriAdjustDays != nil {
		if err := db.SetSetting(d.DB, SettingHijriAdjustDays, strconv.Itoa(*req.HijriAdjustDays)); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save hijri adjustment")
			return
		}
	}

	d.afterAdminWrite("prayer-times")
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok", "date": date})
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

	id, err := db.InsertSlide(d.DB, db.Slide{
		Title: title, Type: slideType, ContentURLOrText: contentURLOrText,
		ArabicText: sql.NullString{String: r.FormValue("arabic_text"), Valid: r.FormValue("arabic_text") != ""},
		IsActive: true, ExpirationDate: expiration, DisplayDurationSec: duration,
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
