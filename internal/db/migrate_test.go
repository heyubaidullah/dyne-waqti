package db

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return database
}

func TestMigrateIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	if err := Migrate(database); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := SetSetting(database, "timezone", "America/New_York"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	// Re-running migrate must not wipe or error on existing data.
	if err := Migrate(database); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	got, err := GetSetting(database, "timezone")
	if err != nil {
		t.Fatalf("GetSetting after re-migrate: %v", err)
	}
	if got != "America/New_York" {
		t.Errorf("timezone = %q, want %q", got, "America/New_York")
	}
}

// TestMigrateAddsDisplayModeColumnToExistingSlides simulates a database
// created before the `display_mode` column existed (a pre-v0.2 install) and
// proves Migrate adds it via ALTER TABLE, defaulting existing rows to
// "full", without touching any other column's data. Also proves running
// Migrate a second time (already-migrated database) is a safe no-op.
func TestMigrateAddsDisplayModeColumnToExistingSlides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	// Hand-create the pre-v0.2 slides table shape (no display_mode column)
	// to stand in for a database that predates this release.
	if _, err := database.Exec(`
		CREATE TABLE slides (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			type TEXT NOT NULL,
			content_url_or_text TEXT NOT NULL,
			arabic_text TEXT,
			is_active INTEGER DEFAULT 1,
			expiration_date TEXT,
			display_duration_sec INTEGER DEFAULT 10,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		t.Fatalf("create legacy slides table: %v", err)
	}
	if _, err := database.Exec(`
		INSERT INTO slides (title, type, content_url_or_text, display_duration_sec)
		VALUES ('Existing Flyer', 'image', '/uploads/old.png', 12)`); err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	if err := Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	slides, err := ListSlides(database, false)
	if err != nil {
		t.Fatalf("ListSlides: %v", err)
	}
	if len(slides) != 1 {
		t.Fatalf("len(slides) = %d, want 1", len(slides))
	}
	if slides[0].Title != "Existing Flyer" || slides[0].DisplayDurationSec != 12 {
		t.Errorf("existing row data corrupted: %+v", slides[0])
	}
	if slides[0].DisplayMode != "full" {
		t.Errorf("DisplayMode = %q, want default %q", slides[0].DisplayMode, "full")
	}

	// Re-running Migrate must not error (column already exists) or touch data.
	if err := Migrate(database); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	slides, err = ListSlides(database, false)
	if err != nil {
		t.Fatalf("ListSlides after second migrate: %v", err)
	}
	if len(slides) != 1 || slides[0].DisplayMode != "full" {
		t.Errorf("slides after second migrate = %+v, want unchanged", slides)
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	database := openTestDB(t)

	if _, err := GetSetting(database, "missing"); err != ErrNotFound {
		t.Fatalf("GetSetting(missing) err = %v, want ErrNotFound", err)
	}

	if err := SetSetting(database, "calc_method", "ISNA"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if err := SetSetting(database, "calc_method", "MWL"); err != nil {
		t.Fatalf("SetSetting overwrite: %v", err)
	}
	got, err := GetSetting(database, "calc_method")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if got != "MWL" {
		t.Errorf("calc_method = %q, want %q", got, "MWL")
	}
}
