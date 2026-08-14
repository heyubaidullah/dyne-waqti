package api

import (
	"fmt"
	"net/http"

	"github.com/heyubaidullah/waqti/internal/config"
)

// Placeholder pages for Phase A, before the real React /admin and vanilla-JS
// /display frontends (Phase B/C) are embedded via go:embed.

const stubPageTemplate = `<!doctype html>
<html><head><title>%s</title></head>
<body style="font-family:sans-serif;background:#0B0F17;color:#F8FAFC;padding:2rem">
<h1>%s</h1>
<p>%s</p>
</body></html>`

func handleDisplayStub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, stubPageTemplate, config.DisplayName+" — Display",
		config.DisplayName, "Phase A backend online. The display UI is not built yet — see GET /api/v1/display-data for live data.")
}

func handleAdminStub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, stubPageTemplate, config.DisplayName+" — Admin",
		config.DisplayName+" Admin", "Phase A backend online. The admin UI is not built yet.")
}
