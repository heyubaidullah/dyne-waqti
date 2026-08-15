BINARY := waqti
PKG := ./cmd/server

.PHONY: build build-backend-only build-frontend build-admin build-display build-windows test run clean

# Requires Node/npm on PATH. Builds the admin + display frontends into
# their respective internal/api/*ui/dist directories, then the Go binary
# (which embeds both).
build: build-frontend
	go build -o $(BINARY) $(PKG)

# Go-only build against whatever is currently in internal/api/adminui/dist
# and internal/api/displayui/dist (the committed placeholders, or real
# builds from a previous build-frontend run) — use this to iterate on the
# backend without Node installed.
build-backend-only:
	go build -o $(BINARY) $(PKG)

build-frontend: build-admin build-display

build-admin:
	cd web/admin && npm ci && npm run build

# Tailwind CLI compiles CSS straight into internal/api/displayui/dist;
# the plain ES6 JS/HTML need no bundling, just copying alongside it.
build-display:
	cd web/display && npm ci && npm run build
	mkdir -p internal/api/displayui/dist
	cp web/display/index.html internal/api/displayui/dist/
	cp -r web/display/js internal/api/displayui/dist/
	cp -r web/display/fonts internal/api/displayui/dist/
	cp -r web/display/branding internal/api/displayui/dist/

test:
	go test ./...

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY) $(BINARY).exe

# Windows cross-compile target used by the release workflow (Phase A ships
# Windows-first per spec; modernc.org/sqlite is pure Go so no CGO toolchain
# is required for this cross-compile).
build-windows: build-frontend
	GOOS=windows GOARCH=amd64 go build -o $(BINARY).exe $(PKG)
