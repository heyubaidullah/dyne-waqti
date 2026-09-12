// Package db owns the SQLite connection, idempotent schema migration,
// typed query helpers, and the backup job for Waqti.
package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO required
)

//go:embed schema.sql
var schemaFS embed.FS

// Open opens (creating if necessary) the SQLite database at path and
// configures pragmas suitable for a small, single-process local app.
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite allows only one writer at a time; a single connection avoids
	// SQLITE_BUSY errors from this process's own concurrent goroutines.
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// Migrate applies the embedded schema. Every statement is idempotent
// (CREATE TABLE/INDEX IF NOT EXISTS), so this is safe to run on every
// startup and never overwrites existing data. It then applies any pending
// ALTER-statement steps needed to bring an existing database (created by an
// older schema.sql) up to date with columns added since — CREATE TABLE IF
// NOT EXISTS alone can't add a column to a table that already exists.
func Migrate(db *sql.DB) error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(schema)); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := addColumnIfMissing(tx, "slides", "display_mode", "TEXT NOT NULL DEFAULT 'full'"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return tx.Commit()
}

// addColumnIfMissing runs an ALTER TABLE ... ADD COLUMN only if the column
// doesn't already exist — a fresh database gets the column straight from
// CREATE TABLE (schema.sql), so this only ever fires for a database created
// before that column was added, and only once.
func addColumnIfMissing(tx *sql.Tx, table, column, definition string) error {
	rows, err := tx.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == column {
			return rows.Err()
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, definition))
	return err
}
