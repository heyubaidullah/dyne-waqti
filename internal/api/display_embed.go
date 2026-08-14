package api

import (
	"io/fs"
	"net/http"

	"github.com/heyubaidullah/waqti/internal/api/displayui"
)

// displayHandler serves the embedded, built vanilla-JS display kiosk view
// at /display/. fs.Sub can only fail if "dist" doesn't exist in the
// embedded tree, which go:embed guarantees at compile time — so a
// failure here is a build-time invariant violation, not a runtime
// condition to recover from.
func displayHandler() http.Handler {
	sub, err := fs.Sub(displayui.FS, "dist")
	if err != nil {
		panic("displayui: embedded dist/ missing: " + err.Error())
	}
	return http.StripPrefix("/display/", http.FileServer(http.FS(sub)))
}
