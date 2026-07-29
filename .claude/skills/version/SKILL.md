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
4. Add a new entry at the top of `CHANGELOG.md`, immediately below the
   `## Versioning scheme` section and above whatever entry is currently
   first (newest-first ordering for real, tracked releases — the
   `Pre-1.0 history` section below them stays as-is, don't touch it).
   Use this shape:

   ```markdown
   ## [MAJOR.MINOR.0] - YYYY-MM-DD

   One-paragraph description of what was added and why it's MINOR-worthy.

   Author: <name>. AI-assisted by <model/agent name>, if applicable.
   ```

   Use today's date. The `.0` placeholder for BUILD in the heading is
   intentional — the real build number isn't known until CI actually builds
   this commit; note that explicitly in the entry if it'd otherwise be
   confusing.
5. Stage `VERSION` and `CHANGELOG.md` together with the feature's own code
   changes (don't make the version bump a separate, unrelated commit) unless
   the user asks otherwise.

## Steps to bump MAJOR (only when explicitly requested)

Same as above, except: increment MAJOR by 1, reset MINOR to `0`
(e.g. `1.4` → `2.0`), and say plainly in the CHANGELOG entry why this is a
major version (what's breaking or what the re-launch represents).
