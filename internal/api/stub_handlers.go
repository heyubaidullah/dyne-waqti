package api

import (
	"fmt"
	"net/http"

	"github.com/heyubaidullah/waqti/internal/config"
)

// Placeholder page for Phase A/B, before the real vanilla-JS /display
// frontend (Phase C) is embedded via go:embed. /admin is now the real
// embedded React app — see admin_embed.go.

const stubPageTemplate = `<!doctype html>
<html><head><title>%s</title></head>
<body style="font-family:sans-serif;background:#0B0F17;color:#F8FAFC;padding:2rem">
<h1>%s</h1>
<p>%s</p>
</body></html>`

func handleDisplayStub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, stubPageTemplate, config.DisplayName+" — Display",
		config.DisplayName, "The display UI is not built yet — see GET /api/v1/display-data for live data.")
}
