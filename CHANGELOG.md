# Changelog

All notable changes to this project are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning
follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [1.1.0] - 2026-09-12

Pilot-feedback release, built from the first live masjid deployment.

### Fixed

- Prayer/Iqamah times set from the admin panel could silently revert to
  uncustomized defaults the next calendar day. The old "Today's prayer
  times" form saved to a table keyed by the current date, so a masjid's
  custom times only ever applied to that single day — this looked like
  "timings changed overnight without anyone touching them." Prayer times
  are now derived solely from persistent settings on every request, so
  nothing date-scoped remains to revert.
- Broadened timezone verification beyond a single US zone: named zones
  across a wide spread (half-hour offsets, no-DST zones, Southern
  Hemisphere DST, historically-tricky European zones) are now covered by
  tests, plus a Southern Hemisphere DST-transition regression test.
  (The underlying fix — embedding the IANA tzdata database — already
  shipped in 1.0.1; this closes out the investigation with broader proof.)
- The admin panel could crash to a blank white page on the Location &
  calculation settings section, caused by a removed MUI API
  (`InputProps` on `TextField`/`Autocomplete`, replaced by `slotProps`).
- The Idle screen's "Next prayer in ..." countdown showed raw minutes past
  60 (e.g. "101:10") instead of rolling over to hours ("1:41:10").

### Added

- Exact Azaan and Iqamah clock times, set independently per prayer —
  replaces the old single Iqamah-offset-from-calculated-time model.
  Non-technical staff type the actual time (e.g. "5:47 AM"), same as
  reading it off a printed prayer schedule; today's astronomically
  calculated time is shown alongside as a reference only, never applied
  automatically. Whatever is saved shows up exactly as entered, every
  day, until changed again.
- Jumu'ah 1 / Jumu'ah 2 support — masjids that hold a second Friday
  prayer can add a second time slot; the display renders one or two rows
  accordingly.
- **Display Settings** panel: prayer-timings font size (small/medium/
  large), a toggle for showing the Gregorian date alongside the Hijri
  date, a configurable silence-screen duration after prayer time
  (default 7 minutes, 1-15 minute range), and an opt-in weather display
  (via Open-Meteo, no API key, fails silently offline, admin-selectable
  Fahrenheit or Celsius).
- Always-visible "Powered by Waqti" attribution banner on `/display`,
  redesigned per masjid feedback: masjid name/logo on the left (both
  independently toggleable, no longer shown separately in the corner),
  "Powered by Waqti" plus the Waqti mark and "Free and Open Source" on
  the right (not configurable).
- Full Screen / In Screen display mode for image flyers — In Screen
  shows the flyer above the persistent prayer-times ribbon instead of
  filling the whole screen.
- In-admin Help/FAQ page, plus an inline tooltip on the timezone field.
- `--reset-passphrase` CLI flag (plus `scripts\Reset-Passphrase.bat` on
  Windows) to recover a lost admin passphrase without touching any other
  data, and a "Forgot the passphrase?" hint on the login screen pointing
  to it.

## [1.0.1] - 2026-08-16

### Fixed

- `time.LoadLocation` failed with `unknown time zone America/Chicago` (or
  any other zone) when running the released `waqti.exe` on Windows,
  even though the same binary worked fine when built and run locally on
  macOS/Linux. Windows has no system IANA timezone database the way
  `/usr/share/zoneinfo` provides on macOS/Linux, so the binary now
  embeds its own copy via a blank `time/tzdata` import (~400KB), which
  Go uses as an automatic fallback wherever the OS doesn't supply one.

## [1.0.0] - 2026-08-15

First tagged release. A single Go binary serving an offline-first
`/display` kiosk view and a password-protected `/admin` panel — built,
tested, and ready to hand to a mosque as a working `.exe`.

### Added

**Backend**
- Go server with a pure-Go SQLite driver (no CGO, no C toolchain needed
  to build or run) and idempotent schema migrations.
- Vendored Adhan prayer-time and Umm al-Qura Hijri date calculation code
  (offline-first — no third-party API dependency).
- Passphrase-based admin auth: random passphrase generated on first run,
  bcrypt session cookies, login rate limiting.
- REST + Server-Sent Events API: prayer/Iqamah times, Hijri date, slide
  (flyer/verse) management, emergency (Janazah) notices, screen
  blackout, masjid logo upload, settings.
- Automatic `VACUUM INTO` backups on every admin write, plus a 6-hourly
  safety-net snapshot; most recent 7 retained.
- Graceful shutdown on SIGINT/SIGTERM.

**Admin panel** (`/admin`, React + MUI, embedded into the binary)
- Login gate, dashboard shell.
- Location & calculation settings: timezone (with a friendly picker for
  common zones), latitude/longitude, calculation method, Asr juristic
  method, Hijri day adjustment.
- Per-prayer Iqamah time entry (12-hour AM/PM inputs) with default
  Adhan-relative offsets as a fallback.
- Masjid logo upload/replace/remove, with an admin-configurable display
  height (no fixed size — every mosque's logo differs).
- Flyer/announcement (slide) manager: image flyers or text_verse
  (Quran/Hadith) slides, per-slide display duration, optional expiration
  date, Arabic text field, guidance on recommended flyer dimensions.
- Configurable full-screen timings-page duration.
- Emergency controls: instant screen blackout, Janazah notice
  publishing.
- Dashboard sections ordered by how often staff touch them: prayer times
  → location/calculation settings → flyers → emergency controls (last).

**Display kiosk** (`/display`, vanilla JS + Tailwind, embedded into the
binary)
- Client-side state machine (Idle / Countdown / Silence / Blackout /
  Emergency) driven by a server-time-anchored clock, never the kiosk
  machine's own system timezone.
- Live SSE updates plus a 60s safety-net poll (self-heals from a missed
  event or a midnight day-rollover).
- Idle content cycle: image flyers go full-screen for their configured
  duration, then a full-screen prayer-times interlude, and repeat; text
  slides show their content in the upper portion of the screen with a
  persistent prayer-times banner in the lower portion, simultaneously,
  with no interlude needed.
- 12-hour prayer time labels, the upcoming prayer boxed/highlighted,
  Hijri date shown with the month written out.
- Masjid logo (admin-configurable height) and the Waqti attribution
  logo shown together during Idle/Countdown.

**Branding & docs**
- Waqti brand assets (icons, wordmarks, lockups) committed to the repo.
- Copyright symbol and a hyperlinked Dyne Labs credit in both UIs and
  the README.
- README split into an end-user path (download the prebuilt `.exe`, no
  toolchain needed) and a developer path (clone + build from source).

**Release infrastructure**
- `.github/workflows/release.yml` — pushing a `v*` tag on `release`
  cross-compiles `waqti.exe` and publishes it as a GitHub Release asset
  automatically.
- `release` branch protection: PR-only, and a required status check
  (`.github/workflows/enforce-release-source.yml`) that only allows
  `main` as the merge source — see [CONTRIBUTING.md](CONTRIBUTING.md).

### Known limitations

- No in-UI way to change the admin passphrase after first login.
- Not yet soak-tested running continuously on real kiosk hardware over
  multiple days.
