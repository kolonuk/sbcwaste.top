# Changelog

Versioning: `MAJOR.MINOR.BUILD`. MAJOR is manual (only for a major re-launch or breaking change), MINOR is manual for notable features (marked "MINOR bump" below), BUILD auto-increments once per commit (`git rev-list --count HEAD` at build time).

One row per notable change, oldest first. "kolonuk" and "John Wood" are the same person (repo owner/maintainer).

| Version | Date | Change | Author |
|---|---|---|---|
| 1.0.1 | 2024-07-15 | Repository created | John Wood |
| 1.0.2 | 2024-07-21 | Initial project scaffold | John Wood |
| 1.0.22 | 2024-07-25 | Google Cloud Run CI/CD pipeline: build, auth, and deploy workflow | John Wood |
| 1.0.28 | 2025-10-07 | CodeQL scanning added; scraper fixed; separate dev/prod environments added | John Wood, Jules |
| 1.0.33 | 2025-10-08 | CI/CD authentication and Artifact Registry naming fixed | Jules |
| 1.1.34 | 2025-10-08 | **MINOR bump** — address lookup + ICS calendar export added | Jules |
| 1.2.35 | 2025-10-08 | **MINOR bump** — SQLite/Firestore caching layer added | Jules |
| 1.2.41 | 2025-10-09 | Backend refactor, unit tests, XML/YAML output fixes | Jules |
| 1.3.42 | 2025-10-10 | **MINOR bump** — running-costs/billing automation via BigQuery added | Jules |
| 1.3.54 | 2025-10-10 | Bug fixes, UI improvements, SSRF mitigation | Jules |
| 1.4.55 | 2025-10-10 | **MINOR bump** — per-IP rate limiting added | Jules |
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
| 1.5.217 | 2026-07-29 | **MINOR bump** — Docker Hub image mirror + local dev workflow added; lint skill added | kolonuk, Claude |
| 1.6.224 | 2026-07-29 | **MINOR bump** — semantic versioning system added: VERSION file, `/api/version` endpoint, version footer, this changelog, `version` skill | kolonuk, Claude |
