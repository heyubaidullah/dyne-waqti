# Contributing / branch flow

Three branches matter here, each with a different job:

- **feature branches** (`feature/...`, `fix/...`, `chore/...`) — where all
  work actually happens. Branch off `main`.
- **`main`** — the integration branch. Always buildable, always tested.
  Every feature branch merges here first, via PR.
- **`release`** — only ever moves forward by merging `main` into it.
  Never receives a feature branch directly. Tagging a commit on
  `release` (`vX.Y.Z`) is what triggers a real GitHub Release (see
  [Releases](#releases) below).

## Day-to-day workflow

1. Branch off `main`: `git checkout -b feature/your-thing main`
2. Commit your work — one logical change per commit, short messages, no
   AI-attribution trailers.
3. Push and open a PR into `main`.
4. Once it's tested and looks good, merge it.

That's it for regular work. `release` is not part of this loop — you
never target it directly, and you don't need to think about it until
it's actually time to ship a build to end users.

## `release` is protected

`release` has branch protection turned on:

- No direct pushes — every change arrives via PR.
- The PR's source branch must be exactly `main` — enforced by the
  `check-source` required status check
  ([`.github/workflows/enforce-release-source.yml`](.github/workflows/enforce-release-source.yml)),
  which fails (and blocks merging) for any other head branch. A PR from
  `feature/whatever` straight into `release` will never go green.

This means the only way code reaches `release` is: land it on `main`
first, then open `main` → `release`.

## Releases

Once `main` has everything you want in the next build:

1. Open a PR from `main` into `release` and merge it (fast-forward,
   since `release` never diverges from `main` on its own).
2. Tag the resulting commit on `release`:
   ```sh
   git checkout release && git pull
   git tag v1.2.3
   git push origin v1.2.3
   ```
3. Pushing a `v*` tag triggers
   [`.github/workflows/release.yml`](.github/workflows/release.yml),
   which cross-compiles `waqti.exe` (`make build-windows` — pure Go
   SQLite driver, no Windows runner or CGO toolchain needed) and
   publishes it as a GitHub Release asset automatically. Nothing to
   upload by hand.
4. Update [`CHANGELOG.md`](CHANGELOG.md) with what shipped, in the same
   PR that bumps `main` before step 1, so it's already on `release` by
   the time the tag goes out.

Version numbers follow [semver](https://semver.org/): breaking changes
bump the major version, new features bump minor, fixes bump patch.
