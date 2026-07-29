# Changelog

| Version | Date | Change | Author |
|---|---|---|---|
| 1.0.1 | 2024-07-15 | Repository created | John Wood |
| 1.0.2 | 2024-07-21 | Initial project scaffold | John Wood |
| 1.0.22 | 2024-07-25 | Google Cloud Run CI/CD pipeline: build, auth, and deploy workflow | John Wood |
| 1.0.25 | 2025-10-01 | CodeQL security scanning added | John Wood |
| 1.0.26 | 2025-10-01 | Scraper fixed and features added | John Wood, Jules |
| 1.0.27 | 2025-10-07 | Separate development and production environments added | Jules |
| 1.0.28 | 2025-10-07 | CodeQL workflow updated | John Wood |
| 1.0.29 | 2025-10-08 | CI/CD manual trigger added; verification enhanced | Jules |
| 1.0.30 | 2025-10-08 | OIDC authentication permissions fixed | Jules |
| 1.0.31 | 2025-10-08 | Chrome instance reused across requests (startup performance) | Jules |
| 1.0.32 | 2025-10-08 | Deployment authentication/permission errors fixed | Jules |
| 1.0.33 | 2025-10-08 | Artifact Registry repository naming fixed, after a bad rename and revert | Jules, John Wood |
| 1.1.34 | 2025-10-08 | **MINOR bump** — address lookup + ICS calendar export added | Jules |
| 1.2.35 | 2025-10-08 | **MINOR bump** — SQLite/Firestore caching layer added | Jules |
| 1.2.36 | 2025-10-09 | Go application refactored; deployment workflow improved | Jules |
| 1.2.37 | 2025-10-09 | Backend refactored; unit tests added | Jules |
| 1.2.38 | 2025-10-10 | Deployment health check fixed | Jules |
| 1.2.39 | 2025-10-10 | Project structure cleaned up; new features added | Jules |
| 1.2.40 | 2025-10-10 | Council research notes added (Bristol, Manchester) | Jules |
| 1.2.41 | 2025-10-10 | XML/YAML output format bugs fixed | Jules |
| 1.3.42 | 2025-10-10 | **MINOR bump** — running-costs/billing automation via BigQuery added | Jules |
| 1.3.43 | 2025-10-10 | Test syntax error fixed | Jules |
| 1.3.44 | 2025-10-10 | Security hardened; CI/CD refactored; src structure adapted | Jules |
| 1.3.47 | 2025-10-10 | Application security, robustness, and efficiency improved | Jules |
| 1.3.48 | 2025-10-10 | Build/deploy workflow split into separate jobs | Jules |
| 1.3.49 | 2025-10-10 | Tests added; docs improved | Jules |
| 1.3.50 | 2025-10-10 | Frontend UI improved | Jules |
| 1.3.51 | 2025-10-10 | SSRF vulnerability (G107) fixed | Jules |
| 1.3.52 | 2025-10-10 | Image cleanup CI job added | Jules |
| 1.3.53 | 2025-10-10 | Historical and live billing data combined | Jules |
| 1.4.54 | 2025-10-10 | **MINOR bump** — per-IP rate limiting added | Jules |
| 1.4.55 | 2025-10-11 | Rate limiting adjusted | Jules |
| 1.4.56 | 2025-10-11 | Copy-to-clipboard and direct calendar links added | Jules |
| 1.4.57 | 2025-10-11 | SEO/dev-info/security-policy files added | Jules |
| 1.4.58 | 2025-10-11 | Routing and Chromium path handling refactored | Jules |
| 1.4.59 | 2025-10-11 | Tests run from repo root in deploy workflow | Jules |
| 1.4.60 | 2025-10-11 | Dynamic icon caching added | Jules |
| 1.4.68 | 2025-10-14 | Container startup and deployment-timeout issues fixed, after several iterations | Jules |
| 1.4.69 | 2025-10-16 | Static file serving and routing fixed | Jules |
| 1.4.72 | 2025-10-16 | CI vulnerability scanning added (Artifact Analysis, GCP scan checks) | Jules |
| 1.4.92 | 2025-10-27 | Address search, billing logic, and support text improved | Jules |
| 1.4.93 | 2025-10-27 | Base image packages updated to fix CVE-2023-45853 | Jules |
| 1.4.94 | 2025-10-27 | CI validation jobs and status badges added | Jules |
| 1.4.97 | 2025-10-27 | TLS certificate issue in address search fixed | Jules |
| 1.4.99 | 2025-10-28 | Scraper updated for council website layout change | Jules |
| 1.4.101 | 2025-10-29 | Conditional icon scraping added | Jules |
| 1.4.102 | 2025-10-29 | Cache observability headers and environment-specific caching added | Jules |
| 1.4.104 | 2025-10-30 | Test suite overhauled to use live data | Jules |
| 1.4.114 | 2025-11-13 | Trivy-flagged vulnerabilities resolved | Jules |
| 1.4.115 | 2025-11-18 | Insecure TLS-bypass parameter added for local dev | Jules |
| 1.4.119 | 2025-11-18 | BigQuery errors handled gracefully | Jules |
| 1.4.120 | 2025-11-18 | Environment-based caching and cache statistics added | Jules |
| 1.4.121 | 2026-01-28 | CI/CD workflow updated for multi-environment deployment | Jules |
| 1.4.127 | 2026-02-04 | CI/CD workflow cleaned up; CVE version bumps applied | John Wood |
| 1.4.128 | 2026-02-06 | Cache refactored to store raw bytes with hashed keys; Trivy scanning added to CI | John Wood |
| 1.4.129 | 2026-02-06 | Go version updated to 1.24.13 | John Wood |
| 1.4.134 | 2026-02-19 | Path-traversal vulnerability fixed, from a GitHub Copilot Autofix suggestion | John Wood, GitHub Copilot Autofix |
| 1.4.136 | 2026-02-19 | Security updates applied | John Wood |
| 1.4.148 | 2026-02-20 | CI workflow triggers restricted; gosec action pinned | John Wood |
| 1.4.153 | 2026-02-20 | Go version compatibility fixed (1.23 to stable 1.24), after several iterations | John Wood |
| 1.4.160 | 2026-02-20 | Static file and CSS/JS serving fixed | John Wood |
| 1.4.162 | 2026-03-02 | Upcoming collections grid added | Jules |
| 1.4.170 | 2026-03-19 | Dependency bumps (grpc, trivy-action) | dependabot |
| 1.4.178 | 2026-06-23 | Dependency bump (opentelemetry/otel) | dependabot |
| 1.4.179 | 2026-06-23 | Security, efficiency, and repo cleanup review fixes | Claude |
| 1.4.182 | 2026-06-23 | x/net CVEs fixed; govulncheck Go 1.25 compatibility fixed; gosec G706 fixed | Claude |
| 1.4.188 | 2026-06-24 | JS bugs in upcoming-collections feature fixed | John Wood |
| 1.4.191 | 2026-06-24 | Container base image switched to distroless/static to eliminate OS CVEs | Claude, John Wood |
| 1.4.193 | 2026-06-24 | cleanup-images CI job fixed to normalize digests and loop deletes | Claude |
| 1.4.195 | 2026-06-24 | package-lock.json added to pin ESLint dependencies, after a corrupted version | John Wood |
| 1.4.196 | 2026-07-01 | Rate-limiter IP spoofing, dead billing route, and CI/CSP gaps fixed | kolonuk, Claude |
| 1.4.198 | 2026-07-28 | Container hardened; CI supply chain pinned; reachable CVEs cleared | kolonuk, Claude |
| 1.4.199 | 2026-07-28 | govulncheck CI failure fixed (reachability scan, not module scan) | kolonuk, Claude |
| 1.4.206 | 2026-07-28 | README corrected: undocumented endpoints, stale claims, icons bug | kolonuk |
| 1.4.207 | 2026-07-29 | UI numbering-gap bug fixed | kolonuk |
| 1.4.209 | 2026-07-29 | Post-deploy health check fixed to bypass Cloudflare, not the app | kolonuk, Claude |
| 1.4.210 | 2026-07-29 | Firestore cache-miss handling fixed | kolonuk, Claude |
| 1.4.214 | 2026-07-29 | Billing CSV made the source of truth; year/month cost-tree UI added | kolonuk, Claude |
| 1.5.217 | 2026-07-29 | **MINOR bump** — Docker Hub image mirror + local dev workflow added; lint skill added | kolonuk, Claude |
| 1.6.225 | 2026-07-29 | **MINOR bump** — semantic versioning system added: VERSION file, `/api/version` endpoint, version footer, this changelog, `version` skill | kolonuk, Claude |
