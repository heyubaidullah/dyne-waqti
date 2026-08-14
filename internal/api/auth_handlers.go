package api

import (
	"net/http"

	"github.com/heyubaidullah/waqti/internal/auth"
)

type loginRequest struct {
	Passphrase string `json:"passphrase"`
}

func (d *Deps) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	token, err := d.Auth.Login(req.Passphrase, auth.ClientIP(r))
	switch err {
	case nil:
		auth.SetSessionCookie(w, token)
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case auth.ErrLocked:
		w.Header().Set("Retry-After", "900")
		respondError(w, http.StatusTooManyRequests, err.Error())
	case auth.ErrInvalidCredentials:
		respondError(w, http.StatusUnauthorized, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "login failed")
	}
}

func (d *Deps) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.CookieName); err == nil {
		d.Auth.Logout(cookie.Value)
	}
	auth.ClearSessionCookie(w)
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleSessionCheck lets the frontend distinguish "logged in" from "not"
// on load. d.Auth.Middleware already supplies the 401 when the cookie is
// missing or expired, so a reachable 200 here is the entire contract.
func (d *Deps) handleSessionCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
