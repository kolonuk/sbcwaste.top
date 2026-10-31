package main

import "fmt"

// version and buildNumber are injected at build time via -ldflags (see the
// Dockerfile and .github/workflows/google-cloudrun-docker.yml). version comes
// from the repo-root VERSION file (MAJOR.MINOR, bumped by the version skill);
// buildNumber is the total git commit count at build time, so it advances by
// exactly 1 for every commit regardless of which branch/environment builds
// it. Unset (e.g. plain `go run ./src` in local dev) falls back to "dev".
var (
	version     = "dev"
	buildNumber = "0"
	commitSHA   = "unknown"
)

// fullVersion returns the MAJOR.MINOR.BUILD string, e.g. "1.0.427".
func fullVersion() string {
	return fmt.Sprintf("%s.%s", version, buildNumber)
}
