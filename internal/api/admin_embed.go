package api

import (
	"io/fs"
	"net/http"

	"github.com/heyubaidullah/waqti/internal/api/adminui"
)

// adminHandler serves the embedded, built React admin app at /admin/.
// fs.Sub can only fail if "dist" doesn't exist in the embedded tree, which
// go:embed guarantees at compile time — so a failure here is a build-time
// invariant violation, not a runtime condition to recover from.
func adminHandler() http.Handler {
	sub, err := fs.Sub(adminui.FS, "dist")
	if err != nil {
		panic("adminui: embedded dist/ missing: " + err.Error())
	}
	return http.StripPrefix("/admin/", http.FileServer(http.FS(sub)))
}
