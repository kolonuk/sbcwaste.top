package main

import (
	"testing"
)

func TestFoldLine(t *testing.T) {
	// Test case 1: line shorter than maxLen
	shortLine := "This is a short line."
	expected := "This is a short line."
	if got := foldLine(shortLine); got != expected {
		t.Errorf("foldLine() = %q, want %q", got, expected)
	}

	// Test case 2: line longer than maxLen
	longLine := "This is a very long line that should definitely be folded into multiple lines because it exceeds the maximum length of 74 characters."
	// The function splits at 74 chars and adds a CRLF, then a space on the next line.
	expected_long := "This is a very long line that should definitely be folded into multiple li\r\n nes because it exceeds the maximum length of 74 characters."
	if got := foldLine(longLine); got != expected_long {
		t.Errorf("foldLine() = %q, want %q", got, expected_long)
	}

	// Test case 3: line exactly maxLen. Should not fold.
	exactLine := "This line is exactly 74 characters long, which is a perfect test case!!123"
	if len(exactLine) != 74 {
		t.Fatalf("Test setup error: exactLine should be 74 chars long, but is %d", len(exactLine))
	}
	if got := foldLine(exactLine); got != exactLine {
		t.Errorf("foldLine() = %q, want %q", got, exactLine)
	}
}

func TestGenerateUID(t *testing.T) {
	title := "Refuse Collection"
	date := "2025-01-01"
	location := "123 Test Street, Swindon"

	// Since the hash is deterministic, the output should always be the same.
	expectedUID := "ebf4c108fbcf5abafbbf161b309026ed6152b8bdecfa7c3fb924f481460cc087@sbcwaste.com"
	if got := generateUID(title, date, location); got != expectedUID {
		t.Errorf("generateUID() = %q, want %q", got, expectedUID)
	}

	// Test with different data to ensure it produces a different UID
	differentUID := generateUID("Recycling Collection", date, location)
	if differentUID == expectedUID {
		t.Errorf("generateUID() produced the same UID for different data")
	}
}