package auth

import (
	"database/sql"
	"time"

	"github.com/heyubaidullah/waqti/internal/db"
)

const (
	maxFailuresBeforeLockout = 5
	lockoutWindow            = 15 * time.Minute
)

// IsLocked reports whether ipOrSession has had maxFailuresBeforeLockout or
// more failed login attempts within the last lockoutWindow. Because the
// check always looks back lockoutWindow from "now", this is a sliding
// window: continued failed attempts keep extending the effective lockout.
func IsLocked(database *sql.DB, ipOrSession string) (bool, error) {
	count, err := db.CountRecentFailures(database, ipOrSession, time.Now().Add(-lockoutWindow))
	if err != nil {
		return false, err
	}
	return count >= maxFailuresBeforeLockout, nil
}
