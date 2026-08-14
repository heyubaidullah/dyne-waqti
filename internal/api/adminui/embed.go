// Package adminui embeds the built React admin app (web/admin, built via
// Vite into this package's dist/ directory) so it ships inside the single
// Go binary with no runtime Node/npm dependency on the kiosk host.
//
// dist/ is gitignored except a committed placeholder index.html, so a
// fresh clone still builds/tests with zero Node dependency; run `make
// build-frontend` to replace it with the real app.
package adminui

import "embed"

//go:embed all:dist
var FS embed.FS
