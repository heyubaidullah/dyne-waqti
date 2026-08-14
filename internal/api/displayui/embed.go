// Package displayui embeds the built /display kiosk view (web/display,
// compiled via Tailwind CLI into this package's dist/ directory — the
// plain ES6 JS/HTML need no bundling) so it ships inside the single Go
// binary with no runtime Node/npm dependency on the kiosk host.
//
// dist/ is gitignored except a committed placeholder index.html, so a
// fresh clone still builds/tests with zero Node dependency; run `make
// build-display` to replace it with the real app.
package displayui

import "embed"

//go:embed all:dist
var FS embed.FS
