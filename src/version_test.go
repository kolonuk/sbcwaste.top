package main

import "testing"

func TestFullVersion(t *testing.T) {
	originalVersion, originalBuildNumber := version, buildNumber
	defer func() { version, buildNumber = originalVersion, originalBuildNumber }()

	version = "1.0"
	buildNumber = "427"

	if got, want := fullVersion(), "1.0.427"; got != want {
		t.Errorf("fullVersion() = %q, want %q", got, want)
	}
}
