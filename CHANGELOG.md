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

## [1.6.0] - 2026-07-29 (builds 218–219)

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

## Version history (reconstructed, builds 1–217; live-tracked from 1.6.0 / build 218 onward)

Presented oldest-first, as a chronological narrative of the repo. Reconstructed from 220 commits on the `main`/`dev` history via `git log` subjects, dates, authors, and `Co-Authored-By` trailers (an earlier draft of this file cross-checked against `git log --all`, which double-counts commits across refs — 215 "unique" was that inflated figure; the real linear count on `main` is 220). Author "kolonuk" / "John Wood" (john@kolon.co.uk, kolonuk@gmail.com) is one person, the repo owner/maintainer, referred to as "John Wood (kolonuk)" below.

### v1.0.1 — 2024-07-15 — Initial commit
Repo created (John Wood). A single seed commit.

### v1.0.2 — 2024-07-21 — Initial scaffold
John Wood (kolonuk) adds the initial project scaffold (the "Initial" commit). Human-authored.

### v1.0.3–v1.0.22 — 2024-07-23 to 2024-07-25 — First working CI/CD deploy
A concentrated burst of 20 commits by John Wood (kolonuk), almost all titled things like "github actions debug output test", "github actions echo test", "github action update iam service account" — trial-and-error getting the first Google Cloud Run deploy pipeline working, plus the initial `google-cloudrun-docker.yml` workflow. Culminates in "github actions complete - part 1" (build 22, 2024-07-25). All human-authored; no AI involvement at this stage. The repo then goes dormant for roughly 14 months.

### v1.0.23–v1.0.28 — 2025-10-01 to 2025-10-07 — Scraper fixes, CodeQL, dev/prod environments (PR #2, #3, #4)
Work resumes: CodeQL scanning is added (builds 23–25, John Wood, human). "Fix scraper and add features" (PR #2, build 26, John Wood) is co-authored by `google-labs-jules[bot]` per its commit trailer — Google's Jules AI agent assisted even though John is the recorded author. "Add development and production environments" (PR #3, build 27) is authored directly by Jules. "Update codeql.yml" (PR #4, build 28, John Wood, human). This marks the start of heavy Jules involvement in the project.

### v1.0.29–v1.0.33 — 2025-10-08 — CI/CD hardening and auth fixes (PR #5–#12)
A run of Jules-authored PRs fixing OIDC/WIF authentication, deployment permissions, manual workflow triggers, and Artifact Registry naming (including a revert by John Wood of a bad rename, PR #10, and its redo, PR #12, build 33). All AI-authored (Jules), reviewed/merged by John Wood.

### v1.1.34 — 2025-10-08 — Address lookup + ICS calendar export (PR #15) — MINOR bump
Jules adds address lookup and .ics calendar link generation for bin collections — a core user-facing feature, and the first MINOR-worthy addition. AI-authored (Jules).

### v1.2.35 — 2025-10-08 — Caching layer: SQLite + Firestore (PR #16/#18) — MINOR bump
Jules adds the first caching layer backed by SQLite and Firestore (PR #16 was misnamed and redone as PR #18, build 35). AI-authored (Jules).

### v1.2.36–v1.2.41 — 2025-10-09 — Refactors, tests, XML/YAML fixes (PR #19–#27)
Jules-authored: backend refactor and unit tests (#19, #20), project cleanup and new features (#22), XML/YAML output fixes (#27). AI-authored (Jules), merged by John Wood.

### v1.3.42 — 2025-10-10 — Running-costs/billing automation via BigQuery (PR #28) — MINOR bump
Jules automates the running-costs page using BigQuery billing data. AI-authored (Jules).

### v1.3.43–v1.3.54 — 2025-10-10 — More fixes ahead of rate limiting (PR #29–#38ish)
Jules-authored fixes including a syntax-error correction (#29), frontend UI improvements, and SSRF mitigation. AI-authored (Jules).

### v1.4.55 — 2025-10-10 — Per-IP rate limiting (PR #39/#40) — MINOR bump
Jules implements and then adjusts per-IP rate limiting. AI-authored (Jules).

### v1.4.56–v1.4.68 — 2025-10-11 to 2025-10-14 — Container/runtime stability (PR #41–#53)
Jules fixes Chrome/chromedp stability in the container (headless scraping dependency), deployment timeouts, static file serving, plus adds SEO/security-policy files, clipboard-copy and direct calendar links. All AI-authored (Jules).

### v1.4.69–v1.4.102 — 2025-10-16 to 2025-10-27 — Vulnerability-scan CI hardening (PR #54–#85)
A long iterative series mostly authored by Jules, hardening the CI vulnerability-scanning pipeline (Artifact Analysis/Trivy integration, polling robustness, GCP service enablement) plus scraper updates for site layout changes, TLS/cert fixes, cache observability headers, environment-specific caching, and conditional icon scraping. Two commits ("Refactor CI workflow...", "Refactor vulnerability scan checks...") are direct human edits by John Wood. Predominantly AI-authored (Jules).

### v1.4.103–v1.4.120 — 2025-10-29 to 2025-11-18 — Test suite overhaul, CI fixes, environment-based caching (PR #87–#108)
Jules overhauls the test suite to use live data (#88), fixes CI timeouts and build failures, and adds environment-based caching with cache statistics (#108) and graceful BigQuery error handling (#107). John Wood makes a small manual "Minor updates and cleaning" pass (build 117). AI-authored (Jules), human-reviewed.

### v1.4.121–v1.4.135 — 2026-01-28 to 2026-02-06 — Multi-environment CI/CD, Go version and cache refactor
After a ~2-month gap, Jules updates the CI/CD workflow for multi-environment deployment (PR #111, build 121). John Wood (kolonuk) then does a manual hardening pass: workflow cleanup, CVE version bumps, and a refactor genericizing the cache to store raw bytes with hashed keys plus adding Trivy scanning to CI. Mixed: one Jules PR, rest human.

### v1.4.136–v1.4.160 — 2026-02-19 to 2026-02-21 — Security hardening flurry, Go 1.24 migration (PR #114–#122)
A dense human-driven push by John Wood (kolonuk): several "security updates" commits (starting build 136), a path-traversal fix accepted from GitHub's Copilot Autofix suggestion (commit 679f88e, flagged as AI-suggested/human-applied), and an extended trial-and-error fixing Go version compatibility (1.23 to stable 1.24) across Dockerfile and CI, plus static file/CSS serving fixes. All human-authored except the one Copilot Autofix commit.

### v1.4.161–v1.4.162 — 2026-03-02 — Upcoming collections grid
Jules adds an "upcoming collections" grid to address search results and reorders page sections; interestingly the commit trailer shows kolonuk as co-author on Jules's own commit. Primarily AI-authored (Jules), human-involved.

### v1.4.163–v1.4.169 — 2026-03-19 to 2026-06-23 — Dependency bumps
Routine automated dependency updates via `dependabot[bot]` (grpc, trivy-action, go.opentelemetry.io/otel). Automated, not AI-authored in the coding sense.

### v1.4.170–v1.4.195 — 2026-06-23 to 2026-06-24 — Claude-led security/efficiency review, distroless base image (PR #123–#131)
The project's first explicit Claude Code involvement: branches named `claude/security-efficiency-review` and `claude/repo-security-efficiency-review-*` produce a series of commits authored directly as `Claude <noreply@anthropic.com>` (starting build 170) — fixing govulncheck CI failures (Go 1.25 compatibility, SSA-analysis panics), x/net CVEs, gosec findings, and a broader "security, efficiency, and repo cleanup review." The container base image is switched to `distroless/static` to eliminate OS-level CVEs, and `package-lock.json` is added to pin ESLint dependencies (after a corrupted-hash version is removed). John Wood merges these PRs alongside a dependabot Go-modules bump (PR #127) and the earlier upcoming-collections-grid PR (#123). AI-authored (Claude), human-reviewed/merged.

### v1.4.196 — 2026-07-01 — Rate-limiter and billing-route hardening
kolonuk (now committing from kolonuk@gmail.com) fixes rate-limiter IP spoofing, a dead billing route, and CI/CSP gaps. The commit trailer shows `Claude Sonnet 5` as co-author — the maintainer is now pair-programming directly with Claude Code rather than routing everything through separate agent PRs.

### v1.4.197–v1.4.214 — 2026-07-28 — Container hardening, billing CSV/cost-tree UI, govulncheck fixes
kolonuk, with `Claude Sonnet 5` co-author trailers on nearly every commit: govulncheck reachability-scan fix, container hardening and CI supply-chain pinning, README corrections, a fixed Firestore cache-miss bug, a fixed post-deploy health check, and making the billing CSV the source of truth with a year/month cost tree UI. Human-directed, AI-assisted (Claude Code).

### v1.5.215–v1.5.217 — 2026-07-28 to 2026-07-29 — Docker Hub image mirror + local dev workflow (PR #134, #135) — MINOR bump
kolonuk adds a local Docker dev workflow (forced SQLite) and a Docker Hub image mirror job (build 215), with a same-day follow-up fix for a GitHub Actions `secrets` context restriction (build 216), then adds a lint skill (actionlint plus the rest of the local lint/build/test suite) as a required pre-commit/pre-push step (build 217). AI-assisted (Claude Code) throughout.

---

*Judgement calls: PR numbers above are taken from commit subjects, not cross-checked against GitHub's PR API. Dates are commit dates (`%ad`), not merge/PR-open dates, and may not exactly match when a PR was opened. Where a commit's primary author differs from its `Co-Authored-By` trailer (e.g. PR #2, PR #112, and the 2026-03-02 grid commit), both are noted rather than picking one. Build numbers are exact (`git rev-list --count <hash>`, verified 2026-07-29); the six MINOR-bump points are a judgement call about what counts as "notable" — a different maintainer might draw that line differently.*
