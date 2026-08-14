package auth

import (
	"path/filepath"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/heyubaidullah/waqti/internal/db"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}
	return NewManager(database)
}

// bootstrapWithKnownPassphrase inserts credentials directly so the test
// knows the plaintext passphrase (Bootstrap() only logs a random one).
func bootstrapWithKnownPassphrase(t *testing.T, m *Manager, passphrase string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(passphrase), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}
	if _, err := m.db.Exec(`INSERT INTO admin_credentials (id, passphrase_hash) VALUES (1, ?)`, string(hash)); err != nil {
		t.Fatalf("insert admin_credentials: %v", err)
	}
}

func TestBootstrapGeneratesCredentialsOnlyOnce(t *testing.T) {
	m := newTestManager(t)

	if err := m.Bootstrap(); err != nil {
		t.Fatalf("first Bootstrap: %v", err)
	}
	var hashAfterFirst string
	if err := m.db.QueryRow(`SELECT passphrase_hash FROM admin_credentials WHERE id = 1`).Scan(&hashAfterFirst); err != nil {
		t.Fatalf("query hash: %v", err)
	}
	if hashAfterFirst == "" {
		t.Fatal("expected a non-empty passphrase hash after bootstrap")
	}

	if err := m.Bootstrap(); err != nil {
		t.Fatalf("second Bootstrap: %v", err)
	}
	var hashAfterSecond string
	if err := m.db.QueryRow(`SELECT passphrase_hash FROM admin_credentials WHERE id = 1`).Scan(&hashAfterSecond); err != nil {
		t.Fatalf("query hash: %v", err)
	}
	if hashAfterFirst != hashAfterSecond {
		t.Error("Bootstrap regenerated credentials on second run; it must be a no-op once a row exists")
	}
}

func TestLoginSuccessAndFailure(t *testing.T) {
	m := newTestManager(t)
	bootstrapWithKnownPassphrase(t, m, "correct-horse-battery-staple")

	if _, err := m.Login("wrong-passphrase", "10.0.0.5"); err != ErrInvalidCredentials {
		t.Errorf("Login(wrong) err = %v, want ErrInvalidCredentials", err)
	}

	token, err := m.Login("correct-horse-battery-staple", "10.0.0.5")
	if err != nil {
		t.Fatalf("Login(correct): %v", err)
	}
	if token == "" {
		t.Fatal("Login returned empty token")
	}
	if !m.Valid(token) {
		t.Error("newly issued token should be valid")
	}

	m.Logout(token)
	if m.Valid(token) {
		t.Error("token should be invalid after Logout")
	}
}

func TestLoginLockoutAfterFiveFailures(t *testing.T) {
	m := newTestManager(t)
	bootstrapWithKnownPassphrase(t, m, "correct-horse-battery-staple")
	ip := "10.0.0.9"

	for i := 0; i < 5; i++ {
		if _, err := m.Login("wrong", ip); err != ErrInvalidCredentials {
			t.Fatalf("attempt %d: err = %v, want ErrInvalidCredentials", i, err)
		}
	}

	// 6th attempt, even with the CORRECT passphrase, must be locked out.
	if _, err := m.Login("correct-horse-battery-staple", ip); err != ErrLocked {
		t.Errorf("6th attempt err = %v, want ErrLocked", err)
	}

	// A different client is unaffected by this IP's lockout.
	if _, err := m.Login("correct-horse-battery-staple", "10.0.0.10"); err != nil {
		t.Errorf("different IP should not be locked out: %v", err)
	}
}
