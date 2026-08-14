package db

import (
	"path/filepath"
	"testing"
)

// TestBackupIsRestorable proves a backup file is a valid, complete,
// independently-openable SQLite database: known rows written before the
// backup must round-trip when the backup file itself is opened directly.
func TestBackupIsRestorable(t *testing.T) {
	database := openTestDB(t)

	if err := SetSetting(database, "timezone", "America/Chicago"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if err := UpsertPrayerSchedule(database, PrayerSchedule{
		Date: "2026-08-13", FajrIqamah: "05:30", DhuhrIqamah: "13:30",
		AsrIqamah: "17:15", MaghribIqamah: "19:45", IshaIqamah: "21:00", JumuahIqamah: "13:30",
	}); err != nil {
		t.Fatalf("UpsertPrayerSchedule: %v", err)
	}

	backupsDir := t.TempDir()
	backupPath, err := Backup(database, backupsDir)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	restored, err := Open(backupPath)
	if err != nil {
		t.Fatalf("Open(backupPath): %v", err)
	}
	defer restored.Close()

	tz, err := GetSetting(restored, "timezone")
	if err != nil {
		t.Fatalf("GetSetting on restored db: %v", err)
	}
	if tz != "America/Chicago" {
		t.Errorf("restored timezone = %q, want %q", tz, "America/Chicago")
	}

	sched, err := GetPrayerSchedule(restored, "2026-08-13")
	if err != nil {
		t.Fatalf("GetPrayerSchedule on restored db: %v", err)
	}
	if sched.FajrIqamah != "05:30" {
		t.Errorf("restored FajrIqamah = %q, want %q", sched.FajrIqamah, "05:30")
	}
}

func TestBackupRetentionKeepsOnlyLastN(t *testing.T) {
	database := openTestDB(t)
	backupsDir := t.TempDir()

	var lastPath string
	for i := 0; i < backupRetention+3; i++ {
		path, err := Backup(database, backupsDir)
		if err != nil {
			t.Fatalf("Backup #%d: %v", i, err)
		}
		lastPath = path
	}

	matches, err := filepath.Glob(filepath.Join(backupsDir, "waqti-*.db"))
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(matches) != backupRetention {
		t.Errorf("backup count = %d, want %d", len(matches), backupRetention)
	}
	found := false
	for _, m := range matches {
		if m == lastPath {
			found = true
		}
	}
	if !found {
		t.Errorf("most recent backup %q was pruned", lastPath)
	}
}
