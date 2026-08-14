BINARY := waqti
PKG := ./cmd/server

.PHONY: build test run clean

build:
	go build -o $(BINARY) $(PKG)

test:
	go test ./...

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY) $(BINARY).exe

# Windows cross-compile target used by the release workflow (Phase A ships
# Windows-first per spec; modernc.org/sqlite is pure Go so no CGO toolchain
# is required for this cross-compile).
build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY).exe $(PKG)
