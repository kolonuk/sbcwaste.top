# Changelog

All notable changes to sbcwaste.top are documented here.

## Versioning scheme

This project uses **MAJOR.MINOR.BUILD** (e.g. `1.0.427`):

- **MAJOR** — bumped manually, by maintainer decision, for significant re-launches or breaking changes.
- **MINOR** — bumped manually for notable feature additions (e.g. the SQLite/Firestore caching layer, per-IP rate limiting, or the Docker Hub mirror + local dev workflow).
- **BUILD** — auto-incremented, one per commit, at build time (it is literally the git commit count).

Semantic versioning was introduced on **2026-07-29**, starting the scheme at `1.0.0`. Entries below the `[1.0.0]` line are a **reconstructed historical narrative** covering 2024-07-15 through 2026-07-29 — no real version numbers existed for this period, so none are invented. The `[1.0.0]` entry and everything after it are real, tracked releases.

---

## [1.0.0] - 2026-07-29

Introduces versioning for the project:

- A `VERSION` file at the repo root holding `MAJOR.MINOR` (starts at `1.0`), bumped manually by the `version` skill.
- `src/version.go` — `version`/`buildNumber`/`commitSHA` vars set at build time via `-ldflags`, with `fullVersion()` returning `MAJOR.MINOR.BUILD`.
- `BUILD` is the total git commit count at build time (`git rev-list --count HEAD`), computed in CI (which has full history) and passed into `docker build` as `--build-arg`s, then baked into the binary — so it advances by exactly 1 per commit with no bot-commit loop required.
- A `GET /api/version` JSON endpoint (`{"version": "1.0.NNN", "commit": "abcdef1"}`).
- A `?version=yes` query parameter on the waste-collection endpoint (alongside the existing `?debug=yes`, `?icons=yes`, `?cachestats=yes`), adding a `version` field to the JSON/XML/YAML output.
- A version footer on both HTML pages (`index.html`, `costs.html`), populated client-side from `/api/version`.
- A new `.claude/skills/version` skill documenting how/when to bump MINOR vs. BUILD vs. MAJOR, and to add a CHANGELOG entry.

Author: kolonuk (John Wood). AI-assisted by Claude (Claude Code).

---

## Pre-1.0 history (2024-07-15 – 2026-07-29)

Presented oldest-first, as a chronological narrative — this reads more naturally as "the story of the repo" than newest-first would for a history section (the `[1.0.0]` release above is intentionally the exception, kept at the top per Keep a Changelog convention). Reconstructed from 215 unique commits (deduped from ~425 raw `git log --all` entries that included the same commits replicated across branches/refs) via `git log` subjects, dates, authors, and `Co-Authored-By` trailers. Author "kolonuk" / "John Wood" (john@kolon.co.uk, kolonuk@gmail.com) is one person, the repo owner/maintainer, and is referred to as "John Wood (kolonuk)" below.

### 2024-07-15 — Initial commit
Repo created (John Wood). A single seed commit.

### 2024-07-21 — Initial scaffold
John Wood (kolonuk) adds the initial project scaffold (the "Initial" commit). Human-authored.

### 2024-07-23 to 2024-07-25 — First working CI/CD deploy
A concentrated burst of ~16 commits by John Wood (kolonuk), almost all titled things like "github actions debug output test", "github actions echo test", "github action update iam service account" — trial-and-error getting the first Google Cloud Run deploy pipeline working, plus the initial `google-cloudrun-docker.yml` workflow. Culminates in "github actions complete - part 1" (2024-07-25). All human-authored; no AI involvement at this stage. The repo then goes dormant for roughly 14 months.

### 2025-10-01 to 2025-10-07 — Scraper fixes, CodeQL, dev/prod environments (PR #2, #3, #4)
Work resumes. "Fix scraper and add features" (PR #2, John Wood) is co-authored by `google-labs-jules[bot]` per its commit trailer — Google's Jules AI agent assisted even though John is the recorded author. CodeQL scanning is added and updated (PR #4, John Wood, human). "Add development and production environments" (PR #3) is authored directly by Jules. This marks the start of heavy Jules involvement in the project.

### 2025-10-08 — CI/CD hardening and auth fixes (PR #5–#12)
A run of Jules-authored PRs fixing OIDC/WIF authentication, deployment permissions, manual workflow triggers, and Artifact Registry naming (including a revert by John Wood of a bad rename, PR #10). All AI-authored (Jules), reviewed/merged by John Wood.

### 2025-10-08 — Address lookup + ICS calendar export (PR #15)
Jules adds address lookup and .ics calendar link generation for bin collections — a core user-facing feature. AI-authored (Jules).

### 2025-10-08 — Caching layer: SQLite + Firestore (PR #16/#18)
Jules adds the first caching layer backed by SQLite and Firestore (PR #16 was misnamed and redone as PR #18). AI-authored (Jules). This is a minor-version-worthy feature addition in the new scheme's terms.

### 2025-10-09 to 2025-10-10 — Refactors, tests, billing automation, rate limiting (PR #19–#40)
A dense stretch of ~20 Jules-authored PRs: backend refactor and unit tests (#19, #20), project cleanup and new features (#22), XML/YAML output fixes (#27), automated running-costs page via BigQuery (#28), per-IP rate limiting (#39, #40), frontend UI improvements, SSRF mitigation, and an image-cleanup CI job. All AI-authored (Jules), merged by John Wood. Rate limiting is a minor-version-worthy feature addition in the new scheme's terms.

### 2025-10-11 to 2025-10-14 — Container/runtime stability (PR #41–#53)
Jules fixes Chrome/chromedp stability in the container (headless scraping dependency), deployment timeouts, static file serving, plus adds SEO/security-policy files, clipboard-copy and direct calendar links. All AI-authored (Jules).

### 2025-10-16 to 2025-10-27 — Vulnerability-scan CI hardening (PR #54–#85)
A long iterative series (~30 PRs) mostly authored by Jules, hardening the CI vulnerability-scanning pipeline (Artifact Analysis/Trivy integration, polling robustness, GCP service enablement) plus scraper updates for site layout changes, TLS/cert fixes, cache observability headers, environment-specific caching, and conditional icon scraping. Two commits ("Refactor CI workflow...", "Refactor vulnerability scan checks...") are direct human edits by John Wood. Predominantly AI-authored (Jules).

### 2025-10-29 to 2025-11-18 — Test suite overhaul, CI fixes, environment-based caching (PR #87–#108)
Jules overhauls the test suite to use live data (#88), fixes CI timeouts and build failures, and adds environment-based caching with cache statistics (#108) and graceful BigQuery error handling (#107). John Wood makes a small manual "Minor updates and cleaning" pass. AI-authored (Jules), human-reviewed.

### 2026-01-28 to 2026-02-06 — Multi-environment CI/CD, Go version and cache refactor
After a ~2-month gap, Jules updates the CI/CD workflow for multi-environment deployment (PR #111). John Wood (kolonuk) then does a manual hardening pass: workflow cleanup, CVE version bumps, and a refactor genericizing the cache to store raw bytes with hashed keys plus adding Trivy scanning to CI. Mixed: one Jules PR, rest human.

### 2026-02-19 to 2026-02-21 — Security hardening flurry, Go 1.24 migration (PR #114–#122)
A dense human-driven push by John Wood (kolonuk): several "security updates" commits, a path-traversal fix accepted from GitHub's Copilot Autofix suggestion (commit 679f88e, flagged as AI-suggested/human-applied), and an extended trial-and-error fixing Go version compatibility (1.23 to stable 1.24) across Dockerfile and CI, plus static file/CSS serving fixes. All human-authored except the one Copilot Autofix commit.

### 2026-03-02 — Upcoming collections grid
Jules adds an "upcoming collections" grid to address search results and reorders page sections; interestingly the commit trailer shows kolonuk as co-author on Jules's own commit. Primarily AI-authored (Jules), human-involved.

### 2026-03-19 to 2026-06-23 — Dependency bumps
Routine automated dependency updates via `dependabot[bot]` (grpc, trivy-action, go.opentelemetry.io/otel). Automated, not AI-authored in the coding sense.

### 2026-06-23 to 2026-06-24 — Claude-led security/efficiency review, distroless base image (PR #123–#131)
The project's first explicit Claude Code involvement: branches named `claude/security-efficiency-review` and `claude/repo-security-efficiency-review-*` produce a series of commits authored directly as `Claude <noreply@anthropic.com>` — fixing govulncheck CI failures (Go 1.25 compatibility, SSA-analysis panics), x/net CVEs, gosec findings, and a broader "security, efficiency, and repo cleanup review." The container base image is switched to `distroless/static` to eliminate OS-level CVEs, and `package-lock.json` is added to pin ESLint dependencies (after a corrupted-hash version is removed). John Wood merges these PRs alongside a dependabot Go-modules bump (PR #127) and the earlier upcoming-collections-grid PR (#123). AI-authored (Claude), human-reviewed/merged.

### 2026-07-01 — Rate-limiter and billing-route hardening (PR #131)
kolonuk (now committing from kolonuk@gmail.com) fixes rate-limiter IP spoofing, a dead billing route, and CI/CSP gaps. The commit trailer shows `Claude Sonnet 5` as co-author — the maintainer is now pair-programming directly with Claude Code rather than routing everything through separate agent PRs.

### 2026-07-28 to 2026-07-29 — Container hardening, billing CSV/cost tree, Docker Hub mirror, lint skill (PR #134, #135)
The final pre-1.0 stretch, all by kolonuk with `Claude Sonnet 5` co-author trailers on nearly every commit: govulncheck reachability-scan fix, container hardening and CI supply-chain pinning, README corrections, a fixed Firestore cache-miss bug, a fixed post-deploy health check, making the billing CSV the source of truth with a year/month cost tree UI, and — the standout feature addition of this stretch — a local Docker dev workflow (forced SQLite) plus a Docker Hub image mirror job (with a same-day follow-up fix for a GitHub Actions `secrets` context restriction). This mirror/local-dev addition is a minor-version-worthy feature in the new scheme's terms. The stretch closes with the addition of a lint skill (actionlint plus the rest of the local lint/build/test suite) as a required pre-commit/pre-push step. Human-directed, AI-assisted (Claude Code) throughout.

---

*Judgement calls: PR numbers above are taken from commit subjects, not cross-checked against GitHub's PR API. Dates are commit dates (`%ad`), not merge/PR-open dates, and may not exactly match when a PR was opened. Where a commit's primary author differs from its `Co-Authored-By` trailer (e.g. PR #2, PR #112, and the 2026-03-02 grid commit), both are noted rather than picking one.*
