---
name: version
description: Bump the app's MAJOR.MINOR version and add a CHANGELOG.md entry. Use when a notable feature has just been added (e.g. caching, rate limiting, a new API endpoint, the Docker Hub mirror) and needs a MINOR bump, or when the user explicitly asks to bump the MAJOR version. Not needed for routine fixes/refactors/dependency bumps — those just ride the automatic BUILD number.
---

# Bumping the app version

sbcwaste.top uses **MAJOR.MINOR.BUILD** (e.g. `1.0.427`). See the top of
`CHANGELOG.md` for the full explanation. Summary:

- **BUILD** — fully automatic, never touch it. It's the git commit count at
  build time (`git rev-list --count HEAD`), computed in CI and baked into the
  binary via `-ldflags` (see `Dockerfile` and
  `.github/workflows/google-cloudrun-docker.yml`). Every commit advances it
  by exactly 1 with no action needed here.
- **MINOR** — bump for a notable, user-visible feature addition. This skill
  handles that case.
- **MAJOR** — bump **only** when the user explicitly asks for a major version
  change. Never infer a major bump yourself, even from a large refactor or a
  breaking change — ask if genuinely unsure whether the user wants one.

## What counts as MINOR-worthy

Judgement call, but examples of things that warranted (or would have
warranted) a MINOR bump in this project's history: the SQLite/Firestore
caching layer, per-IP rate limiting, the address-lookup/ICS calendar export
feature, the running-costs/billing page, and the Docker Hub image mirror +
local Docker dev workflow. Routine bug fixes, dependency bumps, CI tweaks,
refactors, and test additions do **not** warrant a bump — those just ride
the automatic BUILD number increment on their own commits.

If unsure whether a change clears the bar, ask the user rather than guessing.

## Steps to bump MINOR

1. Read the current `VERSION` file (repo root, format `MAJOR.MINOR`).
2. Increment MINOR by 1 (e.g. `1.0` → `1.1`). Leave MAJOR untouched.
3. Write the new value back to `VERSION` (single line, no trailing content
   beyond the newline).
4. Add a new row to the **bottom** of the table in `CHANGELOG.md` (the table
   is oldest-first, so the newest release goes last):

   ```markdown
   | MAJOR.MINOR.BUILD | YYYY-MM-DD | **MINOR bump** — brief description of what was added | Author(s) |
   ```

   Use today's date. **BUILD must be the real commit count, never a
   placeholder like `.0`** — every other row in the table uses the actual
   `git rev-list --count` of its commit, and a `.0` would be the only row
   that doesn't, which is inconsistent and confusing (this has been gotten
   wrong before). To get the real number: run `git rev-list --count HEAD`
   immediately before committing, then add 1 for the commit you're about to
   create (assuming nothing else lands in between — re-check right before
   `git commit` if there's any chance of that). Don't guess or reuse an old
   count.
5. Stage `VERSION` and `CHANGELOG.md` together with the feature's own code
   changes (don't make the version bump a separate, unrelated commit) unless
   the user asks otherwise.

## Steps to bump MAJOR (only when explicitly requested)

Same as above, except: increment MAJOR by 1, reset MINOR to `0`
(e.g. `1.4` → `2.0`), and say plainly in the CHANGELOG row why this is a
major version (what's breaking or what the re-launch represents). BUILD
still uses the real commit count, same rule as above — MAJOR/MINOR reset,
BUILD never does.
