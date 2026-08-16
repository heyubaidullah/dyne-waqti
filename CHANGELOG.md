# Changelog

All notable changes to this project are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning
follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

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
