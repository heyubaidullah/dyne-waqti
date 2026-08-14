# Waqti

Offline-first mosque digital signage: a single Go binary serving a 24/7
`/display` kiosk view (prayer times, Hijri date, flyers, Iqamah countdown,
Janazah alerts) and a password-protected `/admin` panel for managing it.

**Status: v0.1.0-poc, Phase A** — the backend (auth, database, prayer-time
and Hijri calculation, REST/SSE API) is complete and tested. The React
`/admin` UI and vanilla-JS `/display` UI are not yet built; `GET /admin` and
`GET /display` currently serve placeholder pages. The full REST API is live
and can be exercised directly (see below) or from a future frontend.

## Requirements

- Go 1.22+ to build. The compiled binary has no runtime dependencies (pure
  Go SQLite driver — no CGO, no C toolchain needed on the host).

## Build & run

```sh
make build      # -> ./waqti
make run        # build + run
make test       # go test ./...
make build-windows   # cross-compile a Windows .exe from any host
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
again. Log in via `POST /api/v1/auth/login` and change it from `/admin`
once that UI ships.

## Configuration

| Env var | Default | Purpose |
|---|---|---|
| `WAQTI_DATA_DIR` | `<dir of executable>/data` | Where `waqti.db`, `uploads/`, and `backups/` live. Never resolved relative to the working directory, so `git pull` can't orphan it. |
| `WAQTI_ADDR` | `:3000` | HTTP listen address. |

Mosque-specific settings (timezone, coordinates, calculation method, Asr
juristic method, Iqamah offsets, Hijri adjustment) live in the `settings`
table and are seeded with safe defaults (UTC, 0/0, ISNA) on first run —
set real values via `POST /api/v1/admin/prayer-times` or the future admin
UI.

## Data & backups

`data/` is gitignored and excluded from version control. On startup the
schema is created idempotently (`CREATE TABLE IF NOT EXISTS`) and existing
data is never touched. Every admin write triggers an atomic `VACUUM INTO`
snapshot to `data/backups/`, plus a safety-net snapshot every 6 hours; the
most recent 7 backups are retained.

## Windows deployment (kiosk host)

1. `make build-windows` (or build directly on the Windows host).
2. Run `scripts\install-service.bat` as Administrator to register the
   `WaqtiService` background service via [NSSM](https://nssm.cc/).
3. Add a shortcut to the Startup folder that launches Chrome in kiosk mode:
   ```bat
   @echo off
   timeout /t 5
   start "" "C:\Program Files\Google\Chrome\Application\chrome.exe" --kiosk http://localhost:3000/display --noerrdialogs --disable-infobars --kiosk-printing
   ```
4. To update, run `scripts\Update-Software.bat` — it `git pull`s, rebuilds,
   and restarts the service without touching `data/`.

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

Developed by Ubaidullah Khan at Dyne Labs (c) 2026
