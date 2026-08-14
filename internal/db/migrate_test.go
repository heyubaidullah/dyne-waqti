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
