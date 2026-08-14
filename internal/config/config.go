// Package config resolves runtime paths and settings for the Waqti server.
// Nothing here depends on the current working directory or the git
// checkout location, so an Update-Software.bat "git pull" can never orphan
// or overwrite live data.
package config

import (
	"os"
	"path/filepath"
)

const (
	// AppName is the internal/config identifier (log lines, service name, etc).
	// DisplayName is the human-facing name shown in the UI and README.
	// Both are defined once here so a future rebrand is a single-line change.
	AppName     = "waqti"
	DisplayName = "Waqti"
)

// Config holds resolved filesystem paths and the HTTP listen address.
type Config struct {
	DataDir    string
	DBPath     string
	UploadsDir string
	BackupsDir string
	Addr       string
}

// Load resolves the data directory and HTTP address, creating any missing
// directories. It never deletes or truncates existing files.
//
// DataDir resolution order:
//  1. WAQTI_DATA_DIR environment variable, if set.
//  2. "data" next to the running executable.
// It is never resolved relative to the current working directory, since a
// git working tree (which cwd may point into) is not a safe place to store
// data across an "Update-Software.bat" git pull.
func Load() (*Config, error) {
	dataDir := os.Getenv("WAQTI_DATA_DIR")
	if dataDir == "" {
		exe, err := os.Executable()
		if err != nil {
			return nil, err
		}
		exe, err = filepath.EvalSymlinks(exe)
		if err != nil {
			return nil, err
		}
		dataDir = filepath.Join(filepath.Dir(exe), "data")
	}

	uploadsDir := filepath.Join(dataDir, "uploads")
	backupsDir := filepath.Join(dataDir, "backups")

	for _, dir := range []string{dataDir, uploadsDir, backupsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	addr := os.Getenv("WAQTI_ADDR")
	if addr == "" {
		addr = ":3000"
	}

	return &Config{
		DataDir:    dataDir,
		DBPath:     filepath.Join(dataDir, "waqti.db"),
		UploadsDir: uploadsDir,
		BackupsDir: backupsDir,
		Addr:       addr,
	}, nil
}
