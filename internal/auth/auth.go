// Package auth handles the single shared admin passphrase, session
// cookies, and login rate limiting for Waqti's /admin panel.
package auth

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/heyubaidullah/waqti/internal/db"
)

const (
	CookieName    = "waqti_session"
	sessionTTL    = 18 * time.Hour // within the spec's 12-24h range
	passphraseLen = 12
	// Unambiguous alphabet: no 0/O, 1/I/L, to reduce transcription errors
	// when a volunteer reads the bootstrap passphrase off a terminal.
	passphraseAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
)

var (
	ErrInvalidCredentials = errors.New("auth: invalid passphrase")
	ErrLocked             = errors.New("auth: too many failed attempts, try again later")
	ErrNoSession          = errors.New("auth: no valid session")
)

// Manager owns in-memory sessions and mediates login/logout against the DB.
type Manager struct {
	db *sql.DB

	mu       sync.RWMutex
	sessions map[string]Session
}

func NewManager(database *sql.DB) *Manager {
	return &Manager{
		db:       database,
		sessions: make(map[string]Session),
	}
}

// Bootstrap ensures an admin_credentials row exists. On first run it
// generates a random passphrase, hashes it, stores it, and prints it to
// stdout — the only place it is ever written in plaintext, and never to
// disk. There is no hardcoded default password.
func (m *Manager) Bootstrap() error {
	var count int
	if err := m.db.QueryRow(`SELECT COUNT(*) FROM admin_credentials`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	passphrase, err := generatePassphrase()
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(passphrase), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := m.db.Exec(`INSERT INTO admin_credentials (id, passphrase_hash) VALUES (1, ?)`, string(hash)); err != nil {
		return err
	}

	banner := strings.Repeat("=", 60)
	log.Printf("\n%s\n INITIAL ADMIN PASSPHRASE: %s\n Change this from the /admin settings page after first login.\n This passphrase will not be shown again.\n%s", banner, passphrase, banner)
	return nil
}

func generatePassphrase() (string, error) {
	b := make([]byte, passphraseLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, passphraseLen)
	for i, v := range b {
		out[i] = passphraseAlphabet[int(v)%len(passphraseAlphabet)]
	}
	// Group as XXXX-XXXX-XXXX for readability.
	return fmt.Sprintf("%s-%s-%s", out[0:4], out[4:8], out[8:12]), nil
}

// Login validates passphrase against the stored hash, subject to rate
// limiting keyed on ipOrSession. Every attempt (success or failure) is
// recorded to the login_attempts audit table.
func (m *Manager) Login(passphrase, ipOrSession string) (string, error) {
	locked, err := IsLocked(m.db, ipOrSession)
	if err != nil {
		return "", err
	}
	if locked {
		return "", ErrLocked
	}

	var hash string
	err = m.db.QueryRow(`SELECT passphrase_hash FROM admin_credentials WHERE id = 1`).Scan(&hash)
	if err != nil {
		return "", err
	}

	compareErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passphrase))
	success := compareErr == nil

	if recordErr := db.RecordLoginAttempt(m.db, ipOrSession, success); recordErr != nil {
		return "", recordErr
	}
	if !success {
		return "", ErrInvalidCredentials
	}

	token, err := newToken()
	if err != nil {
		return "", err
	}
	now := time.Now()
	m.mu.Lock()
	m.sessions[token] = Session{Token: token, CreatedAt: now, ExpiresAt: now.Add(sessionTTL)}
	m.mu.Unlock()

	return token, nil
}

// Logout invalidates a session token.
func (m *Manager) Logout(token string) {
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
}

// Valid reports whether token refers to a live, unexpired session.
func (m *Manager) Valid(token string) bool {
	if token == "" {
		return false
	}
	m.mu.RLock()
	s, ok := m.sessions[token]
	m.mu.RUnlock()
	if !ok {
		return false
	}
	if s.expired(time.Now()) {
		m.mu.Lock()
		delete(m.sessions, token)
		m.mu.Unlock()
		return false
	}
	return true
}

// SetSessionCookie writes the session cookie for a newly-created session.
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		// Secure is intentionally false: this is a LAN-only kiosk
		// deployment served over plain HTTP, per the spec's
		// "local-network-scoped" session requirement. Accepted risk.
		Secure: false,
		MaxAge: int(sessionTTL.Seconds()),
	})
}

// ClearSessionCookie expires the session cookie on logout.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// Middleware rejects any request without a valid session cookie with 401.
// Every state-changing admin endpoint must be wrapped in this.
func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil || !m.Valid(cookie.Value) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ClientIP extracts a best-effort client identifier for rate limiting on a
// LAN-only deployment (no trusted reverse proxy, so RemoteAddr is used
// as-is rather than trusting X-Forwarded-For).
func ClientIP(r *http.Request) string {
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}
