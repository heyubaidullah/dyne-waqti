package db

import (
	"database/sql"
	"errors"
	"time"
)

// ErrNotFound is returned by single-row lookups that find no matching row.
var ErrNotFound = errors.New("db: not found")

// --- settings (generic key/value: location, timezone, calc method, ...) ---

func GetSetting(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return value, err
}

func SetSetting(db *sql.DB, key, value string) error {
	_, err := db.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

// --- prayer_schedules ---

type PrayerSchedule struct {
	Date          string // YYYY-MM-DD
	FajrIqamah    string
	DhuhrIqamah   string
	AsrIqamah     string
	MaghribIqamah string
	IshaIqamah    string
	JumuahIqamah  string
}

func GetPrayerSchedule(db *sql.DB, date string) (*PrayerSchedule, error) {
	var s PrayerSchedule
	err := db.QueryRow(`
		SELECT date, fajr_iqamah, dhuhr_iqamah, asr_iqamah, maghrib_iqamah, isha_iqamah, jumuah_iqamah
		FROM prayer_schedules WHERE date = ?`, date).
		Scan(&s.Date, &s.FajrIqamah, &s.DhuhrIqamah, &s.AsrIqamah, &s.MaghribIqamah, &s.IshaIqamah, &s.JumuahIqamah)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func UpsertPrayerSchedule(db *sql.DB, s PrayerSchedule) error {
	_, err := db.Exec(`
		INSERT INTO prayer_schedules (date, fajr_iqamah, dhuhr_iqamah, asr_iqamah, maghrib_iqamah, isha_iqamah, jumuah_iqamah)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE SET
			fajr_iqamah = excluded.fajr_iqamah,
			dhuhr_iqamah = excluded.dhuhr_iqamah,
			asr_iqamah = excluded.asr_iqamah,
			maghrib_iqamah = excluded.maghrib_iqamah,
			isha_iqamah = excluded.isha_iqamah,
			jumuah_iqamah = excluded.jumuah_iqamah`,
		s.Date, s.FajrIqamah, s.DhuhrIqamah, s.AsrIqamah, s.MaghribIqamah, s.IshaIqamah, s.JumuahIqamah)
	return err
}

// --- slides ---

type Slide struct {
	ID                 int64
	Title              string
	Type               string // "image" | "text_verse"
	ContentURLOrText   string
	ArabicText         sql.NullString
	IsActive           bool
	ExpirationDate     sql.NullString
	DisplayDurationSec int
	CreatedAt          time.Time
}

func ListSlides(db *sql.DB, activeOnly bool) ([]Slide, error) {
	q := `SELECT id, title, type, content_url_or_text, arabic_text, is_active, expiration_date, display_duration_sec, created_at FROM slides`
	if activeOnly {
		q += ` WHERE is_active = 1`
	}
	q += ` ORDER BY created_at DESC`

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Slide
	for rows.Next() {
		var s Slide
		var isActive int
		if err := rows.Scan(&s.ID, &s.Title, &s.Type, &s.ContentURLOrText, &s.ArabicText, &isActive, &s.ExpirationDate, &s.DisplayDurationSec, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.IsActive = isActive != 0
		out = append(out, s)
	}
	return out, rows.Err()
}

func GetSlide(db *sql.DB, id int64) (*Slide, error) {
	var s Slide
	var isActive int
	err := db.QueryRow(`
		SELECT id, title, type, content_url_or_text, arabic_text, is_active, expiration_date, display_duration_sec, created_at
		FROM slides WHERE id = ?`, id).
		Scan(&s.ID, &s.Title, &s.Type, &s.ContentURLOrText, &s.ArabicText, &isActive, &s.ExpirationDate, &s.DisplayDurationSec, &s.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	s.IsActive = isActive != 0
	return &s, nil
}

func InsertSlide(db *sql.DB, s Slide) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO slides (title, type, content_url_or_text, arabic_text, is_active, expiration_date, display_duration_sec)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		s.Title, s.Type, s.ContentURLOrText, s.ArabicText, boolToInt(s.IsActive), s.ExpirationDate, s.DisplayDurationSec)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateSlide(db *sql.DB, id int64, isActive bool, expirationDate sql.NullString) error {
	res, err := db.Exec(`UPDATE slides SET is_active = ?, expiration_date = ? WHERE id = ?`,
		boolToInt(isActive), expirationDate, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func DeleteSlide(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM slides WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// --- emergency_notices ---

type EmergencyNotice struct {
	ID           int64
	Title        string
	DeceasedName string
	PrayerTime   string
	Location     string
	IsActive     bool
	CreatedAt    time.Time
}

// GetActiveEmergency returns the currently active Janazah notice, if any.
func GetActiveEmergency(db *sql.DB) (*EmergencyNotice, error) {
	var n EmergencyNotice
	var isActive int
	err := db.QueryRow(`
		SELECT id, title, deceased_name, prayer_time, location, is_active, created_at
		FROM emergency_notices WHERE is_active = 1 ORDER BY created_at DESC LIMIT 1`).
		Scan(&n.ID, &n.Title, &n.DeceasedName, &n.PrayerTime, &n.Location, &isActive, &n.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	n.IsActive = isActive != 0
	return &n, nil
}

// PublishEmergency deactivates any existing notice and inserts a new active one.
func PublishEmergency(db *sql.DB, n EmergencyNotice) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE emergency_notices SET is_active = 0 WHERE is_active = 1`); err != nil {
		return 0, err
	}
	res, err := tx.Exec(`
		INSERT INTO emergency_notices (title, deceased_name, prayer_time, location, is_active)
		VALUES (?, ?, ?, ?, 1)`, n.Title, n.DeceasedName, n.PrayerTime, n.Location)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// DismissActiveEmergency deactivates whatever notice is currently active.
func DismissActiveEmergency(db *sql.DB) error {
	_, err := db.Exec(`UPDATE emergency_notices SET is_active = 0 WHERE is_active = 1`)
	return err
}

// --- login_attempts (rate limiting audit log) ---

func RecordLoginAttempt(db *sql.DB, ipOrSession string, success bool) error {
	_, err := db.Exec(`INSERT INTO login_attempts (ip_or_session, success) VALUES (?, ?)`,
		ipOrSession, boolToInt(success))
	return err
}

// CountRecentFailures returns the number of failed attempts for ipOrSession since `since`.
func CountRecentFailures(db *sql.DB, ipOrSession string, since time.Time) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM login_attempts
		WHERE ip_or_session = ? AND success = 0 AND attempted_at > ?`,
		ipOrSession, since.UTC().Format("2006-01-02 15:04:05")).Scan(&count)
	return count, err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
