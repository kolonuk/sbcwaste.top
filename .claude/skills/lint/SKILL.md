---
name: lint
description: Run this repo's full local lint/build/test suite (actionlint for GitHub Actions workflow files, go build/vet/test, eslint, npm audit). Use before every `git commit` or `git push` to any branch, not just main — and always when a `.github/workflows/*.yml` file changed. Also use when explicitly asked to lint, preflight, or validate the repo.
---

# Local preflight checks

Run these, in order, and fix any failure before committing or pushing. Stop
and report the failure rather than committing broken code — do not skip a
step because an earlier one looks unrelated.

## 1. Workflow file lint (actionlint) — run whenever `.github/workflows/*.yml` changed

```sh
actionlint .github/workflows/*.yml
```

If `actionlint` isn't installed:

```sh
go install github.com/rhysd/actionlint/cmd/actionlint@latest
$(go env GOPATH)/bin/actionlint .github/workflows/*.yml
```

This is not optional for workflow changes. A plain YAML parser (`yaml.safe_load`,
`@action-validator/cli`) will happily accept a workflow file that GitHub's
Actions runner rejects outright — e.g. referencing the `secrets` context in a
job-level `if:` (only `github`, `inputs`, `needs`, `vars` are valid there;
`secrets` is only valid in step-level `if:`). That exact mistake shipped to
`main` once already in this repo and produced a run that failed in under a
second with zero jobs queued, no useful error from the GitHub API — only
`actionlint` catches it before push.

## 2. Go build, vet, test

```sh
go build -o /tmp/sbcwaste-lint-build ./src && rm -f /tmp/sbcwaste-lint-build
go vet ./src/...
go test -count=1 ./src/...
```

(`-count=1` disables the test cache — without it a stale pass can hide a real
regression.)

## 3. JS lint + dependency audit (matches the `validate-js` CI job)

```sh
npx eslint static/script.js static/costs.js static/version.js
npm audit --audit-level=high
```

## 4. HTML/CSS validation (matches the `validate-static` CI job), if available

CI runs `html5validator` (via the `Cyb3r-Jak3/html5validator-action` GitHub
Action) against `./static`. There's no simple local equivalent bundled in
this repo, so this step is best-effort locally — if `html5validator` isn't
installed, skip it and rely on CI to catch markup issues, but say so
explicitly rather than silently treating it as passed.

## Notes

- `gosec`, `govulncheck`, and the two Trivy scans also run in CI but are
  slower / need extra setup (Gosec/govulncheck binaries, Trivy). They're not
  part of this fast local loop — CI is the backstop for those. Steps 1–3
  above are fast enough to run before every commit.
- This applies to every branch, not just `main` — a bad workflow file or a
  failing test is just as much a problem on `dev`/`test`/a feature branch.
