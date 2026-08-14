BINARY := waqti
PKG := ./cmd/server

.PHONY: build build-backend-only build-frontend build-windows test run clean

# Requires Node/npm on PATH. Builds the React admin app into
# internal/api/adminui/dist, then the Go binary (which embeds it).
build: build-frontend
	go build -o $(BINARY) $(PKG)

# Go-only build against whatever is currently in internal/api/adminui/dist
# (the committed placeholder, or a real app from a previous build-frontend
# run) — use this to iterate on the backend without Node installed.
build-backend-only:
	go build -o $(BINARY) $(PKG)

build-frontend:
	cd web/admin && npm ci && npm run build

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
