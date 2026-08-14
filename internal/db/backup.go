package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const backupRetention = 7

// Backup takes a consistent snapshot of the database into backupsDir using
// SQLite's VACUUM INTO, which is atomic and safe to run against a live,
// open database (unlike a raw file copy). It then prunes old backups,
// keeping only the most recent backupRetention files.
func Backup(db *sql.DB, backupsDir string) (string, error) {
	now := time.Now().UTC()
	// Nanosecond suffix keeps filenames unique (and still chronologically
	// sortable) even when backups are triggered faster than once/second,
	// e.g. by tests or rapid consecutive admin saves.
	name := fmt.Sprintf("waqti-%s-%09d.db", now.Format("20060102-150405"), now.Nanosecond())
	dest := filepath.Join(backupsDir, name)

	// VACUUM INTO requires a path with forward slashes to be unambiguous
	// inside the SQL string even on Windows.
	sqlPath := filepath.ToSlash(dest)
	if _, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s'", sqlPath)); err != nil {
		return "", fmt.Errorf("backup: %w", err)
	}

	if err := pruneBackups(backupsDir, backupRetention); err != nil {
		return dest, fmt.Errorf("backup succeeded but prune failed: %w", err)
	}
	return dest, nil
}

func pruneBackups(backupsDir string, keep int) error {
	matches, err := filepath.Glob(filepath.Join(backupsDir, "waqti-*.db"))
	if err != nil {
		return err
	}
	// Filenames are timestamp-prefixed (YYYYMMDD-HHMMSS), so lexicographic
	// sort is chronological.
	sort.Strings(matches)

	if len(matches) <= keep {
		return nil
	}
	for _, old := range matches[:len(matches)-keep] {
		if err := os.Remove(old); err != nil {
			return err
		}
	}
	return nil
}
