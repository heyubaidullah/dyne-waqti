package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

// Session is an authenticated admin session, held only in memory. This is a
// single-admin, single-process, local-network app, so sessions do not need
// to survive a process restart — requiring re-login after a restart is an
// acceptable (arguably safer) tradeoff against the complexity of persisting
// session state.
type Session struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (s Session) expired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

// newToken generates a cryptographically random, URL-safe session token.
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
