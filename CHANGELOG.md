# Changelog

All notable changes to sbcwaste.top are documented here.

## Versioning scheme

This project uses **MAJOR.MINOR.BUILD** (e.g. `1.6.220`):

- **MAJOR** — fixed at `1` by maintainer decision when versioning was introduced (2026-07-29); only bumped manually going forward, for a significant re-launch or breaking change.
- **MINOR** — bumped for notable feature additions.
- **BUILD** — auto-incremented, one per commit, at build time (it is literally the git commit count, `git rev-list --count HEAD`).

The version history below is **reconstructed from the actual commit graph**, not invented: every version boundary corresponds to a real commit, and its BUILD number is that commit's real `git rev-list --count` value (computed 2026-07-29, on the `main`/`dev` branch history, 220 commits at time of writing). MAJOR is `1` throughout, including retroactively — this project chose to start real tracking at `1.y.xxx` rather than `0.y.xxx`, and applied that same rule backwards across its whole history rather than inventing a separate pre-1.0 scheme. MINOR increments at each of the six feature additions judged notable enough to warrant it (see table below); every other commit just advances BUILD under whatever MINOR was current at the time.

| Version bump | Build | Date | Feature |
|---|---|---|---|
| 1.0 → 1.1 | 34 | 2025-10-08 | Address lookup + ICS calendar export |
| 1.1 → 1.2 | 35 | 2025-10-08 | SQLite/Firestore caching layer |
| 1.2 → 1.3 | 42 | 2025-10-10 | Running-costs/billing automation (BigQuery) |
| 1.3 → 1.4 | 55 | 2025-10-10 | Per-IP rate limiting |
| 1.4 → 1.5 | 215 | 2026-07-28 | Docker Hub image mirror + local Docker dev workflow |
| 1.5 → 1.6 | 218 | 2026-07-29 | This versioning system itself (`VERSION`, `/api/version`, footer, this file, the `version` skill) |

---

## [1.6.0] - 2026-07-29

Introduces versioning for the project:

- A `VERSION` file at the repo root holding `MAJOR.MINOR` (now `1.6`), bumped manually by the `version` skill.
- `src/version.go` — `version`/`buildNumber`/`commitSHA` vars set at build time via `-ldflags`, with `fullVersion()` returning `MAJOR.MINOR.BUILD`.
- `BUILD` is the total git commit count at build time (`git rev-list --count HEAD`), computed in CI (which has full history) and passed into `docker build` as `--build-arg`s, then baked into the binary — so it advances by exactly 1 per commit with no bot-commit loop required.
- A `GET /api/version` JSON endpoint (`{"version": "1.6.NNN", "commit": "abcdef1"}`).
- A `?version=yes` query parameter on the waste-collection endpoint (alongside the existing `?debug=yes`, `?icons=yes`, `?cachestats=yes`), adding a `version` field to the JSON/XML/YAML output.
- A version footer on both HTML pages (`index.html`, `costs.html`), populated client-side from `/api/version`.
- A new `.claude/skills/version` skill documenting how/when to bump MINOR vs. BUILD vs. MAJOR, and to add a CHANGELOG entry.
- This CHANGELOG.md, including the reconstructed version history below.

Author: kolonuk (John Wood). AI-assisted by Claude (Claude Code).

---

## Version history (reconstructed prior to 1.6.0; live-tracked from 1.6.0 onward)

One row per notable change, oldest first. "kolonuk" and "John Wood" are the same person (repo owner/maintainer).

| Version | Date | Change | Author |
|---|---|---|---|
| 1.0.1 | 2024-07-15 | Repository created | John Wood |
| 1.0.2 | 2024-07-21 | Initial project scaffold | John Wood |
| 1.0.22 | 2024-07-25 | Google Cloud Run CI/CD pipeline: build, auth, and deploy workflow | John Wood |
| 1.0.28 | 2025-10-07 | CodeQL scanning added; scraper fixed; separate dev/prod environments added | John Wood, Jules |
| 1.0.33 | 2025-10-08 | CI/CD authentication and Artifact Registry naming fixed | Jules |
| 1.1.34 | 2025-10-08 | Address lookup + ICS calendar export added | Jules |
| 1.2.35 | 2025-10-08 | SQLite/Firestore caching layer added | Jules |
| 1.2.41 | 2025-10-09 | Backend refactor, unit tests, XML/YAML output fixes | Jules |
| 1.3.42 | 2025-10-10 | Running-costs/billing automation via BigQuery added | Jules |
| 1.3.54 | 2025-10-10 | Bug fixes, UI improvements, SSRF mitigation | Jules |
| 1.4.55 | 2025-10-10 | Per-IP rate limiting added | Jules |
| 1.4.68 | 2025-10-14 | Container/runtime stability fixes; SEO/security-policy files added | Jules |
| 1.4.102 | 2025-10-27 | CI vulnerability-scanning hardened; scraper/TLS fixes; caching improvements | Jules, John Wood |
| 1.4.120 | 2025-11-18 | Test suite overhaul; environment-based caching with stats | Jules, John Wood |
| 1.4.135 | 2026-02-06 | Multi-environment CI/CD; cache refactor; Trivy scanning added | Jules, John Wood |
| 1.4.160 | 2026-02-21 | Security hardening; Go 1.24 migration; path-traversal fix | John Wood, GitHub Copilot Autofix |
| 1.4.162 | 2026-03-02 | Upcoming collections grid added | Jules |
| 1.4.169 | 2026-06-23 | Dependency bumps | dependabot |
| 1.4.195 | 2026-06-24 | Security/efficiency review; distroless base image; govulncheck/gosec fixes | Claude |
| 1.4.196 | 2026-07-01 | Rate-limiter and billing-route hardening | kolonuk, Claude |
| 1.4.214 | 2026-07-28 | Container hardening; billing CSV/cost-tree UI added | kolonuk, Claude |
| 1.5.217 | 2026-07-29 | Docker Hub image mirror + local dev workflow; lint skill added | kolonuk, Claude |

---

*Judgement calls: dates are commit dates, which may not exactly match when a PR was opened or merged. The six MINOR-bump points are a judgement call about what counts as "notable" — a different maintainer might draw that line differently.*
