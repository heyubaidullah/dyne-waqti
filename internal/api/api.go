// Package api implements the REST/SSE HTTP layer for Waqti:
// public /display endpoints, authenticated /admin mutating endpoints, and
// the SSE broadcaster that pushes admin changes to connected displays.
package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/heyubaidullah/waqti/internal/auth"
	"github.com/heyubaidullah/waqti/internal/config"
	"github.com/heyubaidullah/waqti/internal/db"
)

const maxJSONBodyBytes = 1 << 20 // 1MB, generous for the small JSON bodies used here

// Deps holds shared dependencies injected into every handler.
type Deps struct {
	DB          *sql.DB
	Auth        *auth.Manager
	Cfg         *config.Config
	Broadcaster *Broadcaster
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("respondJSON encode error: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return false
	}
	return true
}

// afterAdminWrite runs the shared side effects every mutating admin
// endpoint performs on success: a backup snapshot and an SSE broadcast so
// connected /display clients pick up the change immediately.
func (d *Deps) afterAdminWrite(event string) {
	if _, err := db.Backup(d.DB, d.Cfg.BackupsDir); err != nil {
		log.Printf("backup after admin write failed: %v", err)
	}
	d.Broadcaster.Publish(event)
}
