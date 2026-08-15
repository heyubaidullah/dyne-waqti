# Waqti

Offline-first mosque digital signage: a single Go binary serving a 24/7
`/display` kiosk view (prayer times, Hijri date, flyers, Iqamah countdown,
Janazah alerts) and a password-protected `/admin` panel for managing it.

**Status: v0.1.0-poc** — backend, the React `/admin` panel, and the
vanilla-JS `/display` kiosk view are all built, tested, and merged. Not
yet done: CI/release automation, a live NSSM/Windows-service run on real
hardware, and a genuine multi-day soak test of `/display`.

## Requirements

There's no `requirements.txt` — this is a Go project (dependencies are
declared in `go.mod`/`go.sum` and fetched automatically by `go build`),
with two small frontends built separately via npm
(`package-lock.json`/`npm ci` in `web/admin/` and `web/display/`, also
automatic). Nothing to install by hand beyond the tools themselves:

- **Go 1.22+** — to build. The compiled binary has no runtime dependencies
  (pure Go SQLite driver — no CGO, no C toolchain needed on the host).
- **Node.js 18+ / npm** — only needed at build time, to compile the React
  admin panel and the Tailwind CSS for the display view. Both get
  embedded into the single Go binary (`go:embed`) — the finished
  `waqti`/`waqti.exe` needs no Node/npm at runtime, including on the
  kiosk host itself.
- **`make`** — the build is driven by the `Makefile`. Present by default
  on macOS/Linux. On Windows, either install it (e.g. `choco install
  make`) or run the underlying commands directly — see
  [Windows setup from scratch](#windows-setup-from-scratch) below.

## Build & run

```sh
make build      # -> ./waqti (builds both frontends first, then the Go binary)
make run        # build + run
make test       # go test ./...
make build-windows   # cross-compile a Windows .exe from any host — no Node/Go needed on the target machine
make build-backend-only   # Go-only rebuild, skips the npm builds (uses whatever's already in internal/api/*ui/dist)
```

On first run, the server creates `data/` next to the binary (or at
`WAQTI_DATA_DIR` if set), generates a random admin passphrase, and prints
it once to the log:

```
============================================================
 INITIAL ADMIN PASSPHRASE: X64S-8C32-7N6Z
 Change this from the /admin settings page after first login.
 This passphrase will not be shown again.
============================================================
```

Note it down — it is never written to disk in plaintext and is not shown
again. Log in at `/admin`. (There's currently no in-UI way to change the
passphrase after the fact — a known gap, not yet built.)

## Configuration

| Env var | Default | Purpose |
|---|---|---|
| `WAQTI_DATA_DIR` | `<dir of executable>/data` | Where `waqti.db`, `uploads/`, and `backups/` live. Never resolved relative to the working directory, so `git pull` can't orphan it. |
| `WAQTI_ADDR` | `:3000` | HTTP listen address. |

Mosque-specific settings (timezone, coordinates, calculation method, Asr
juristic method, Iqamah offsets, Hijri adjustment) live in the `settings`
table and are seeded with safe defaults (UTC, 0/0, ISNA) on first run —
set real values from the `/admin` panel's settings section (or directly
via `POST /api/v1/admin/settings`).

## Data & backups

`data/` is gitignored and excluded from version control. On startup the
schema is created idempotently (`CREATE TABLE IF NOT EXISTS`) and existing
data is never touched. Every admin write triggers an atomic `VACUUM INTO`
snapshot to `data/backups/`, plus a safety-net snapshot every 6 hours; the
most recent 7 backups are retained.

## Windows setup from scratch

Two ways to get a working `waqti.exe` on a Windows machine — pick based on
whether you want the kiosk PC itself to have a dev toolchain on it.

**Option A — cross-compile elsewhere, copy just the binary (recommended
for the actual kiosk host).** Per the single-binary design, the kiosk PC
needs nothing but the `.exe` and `data/` next to it — no Go, no Node.
From any machine that already has this repo built (macOS/Linux/WSL):
```sh
make build-windows   # -> ./waqti.exe
```
Copy `waqti.exe` (plus `scripts/` if you want the NSSM service installer)
to the Windows machine and skip straight to "Run the binary" below.

**Option B — clone and build directly on Windows** (what you'd do to test
the full dev loop, or if there's no other machine handy):
1. Install [Go 1.22+](https://go.dev/dl/), [Node.js LTS](https://nodejs.org/),
   and [Git](https://git-scm.com/download/win) if not already present.
2. Install `make` — easiest via [Chocolatey](https://chocolatey.org/):
   `choco install make`. (Alternative: skip `make` entirely and run the
   commands below by hand in PowerShell.)
3. `git clone https://github.com/heyubaidullah/dyne-waqti.git && cd dyne-waqti`
4. `make build` — or, without `make`:
   ```powershell
   cd web\admin;   npm ci; npm run build; cd ..\..
   cd web\display; npm ci; npm run build; cd ..\..
   New-Item -ItemType Directory -Force internal\api\displayui\dist | Out-Null
   Copy-Item web\display\index.html internal\api\displayui\dist\
   Copy-Item -Recurse -Force web\display\js internal\api\displayui\dist\
   Copy-Item -Recurse -Force web\display\fonts internal\api\displayui\dist\
   go build -o waqti.exe .\cmd\server
   ```

**Run the binary:**
```powershell
.\waqti.exe
```
It prints the one-time admin passphrase to the console and listens on
`:3000` by default — open `http://localhost:3000/display` and
`http://localhost:3000/admin` in a browser to confirm it works before
setting up the kiosk service.

**Install as a background service + kiosk launch** (once you've confirmed
it runs):
1. Run `scripts\install-service.bat` as Administrator to register the
   `WaqtiService` background service via [NSSM](https://nssm.cc/) (NSSM
   itself isn't bundled — download it separately and put `nssm.exe` on
   `PATH` first).
2. Add a shortcut to the Startup folder that launches Chrome in kiosk mode:
   ```bat
   @echo off
   timeout /t 5
   start "" "C:\Program Files\Google\Chrome\Application\chrome.exe" --kiosk http://localhost:3000/display --noerrdialogs --disable-infobars --kiosk-printing
   ```
3. To update later, run `scripts\Update-Software.bat` — it `git pull`s,
   rebuilds, and restarts the service without touching `data/`.

## REST/SSE API

See `mosque-display-agent-prompt.md` for the full endpoint table. All
`/api/v1/admin/*` endpoints and `POST /api/v1/auth/logout` require a valid
session cookie (issued by `POST /api/v1/auth/login`); `GET
/api/v1/display-data` and `GET /api/v1/sse` are public and read-only.

## License

MIT — see `LICENSE`. Vendored calculation code under `internal/calc/` is
adapted from third-party MIT-licensed projects; see the `LICENSE` files
inside `internal/calc/vendor_adhan/` and `internal/calc/vendor_hijri/`.

---

Developed by Ubaidullah Khan at [Dyne Labs](https://www.dynelabs.org) © 2026
