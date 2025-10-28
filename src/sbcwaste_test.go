package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// loadTestHTMLFile loads a specific HTML file for testing.
func loadTestHTMLFile(t *testing.T, filename string) *goquery.Document {
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open test data file %s: %v", filename, err)
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		t.Fatalf("Failed to parse test HTML from %s: %v", filename, err)
	}
	return doc
}

func TestParseCollectionsFromLiveFixture(t *testing.T) {
	doc := loadTestHTMLFile(t, "../testdata/sbc_response_live.html")
	collections, err := parseCollections(doc)
	if err != nil {
		t.Fatalf("parseCollections failed: %v", err)
	}

	if len(collections.Collections) != 2 {
		t.Fatalf("Expected 2 collections, got %d", len(collections.Collections))
	}

	// Test Recycling and food waste collection
	recycling := collections.Collections[0]
	if recycling.Type != "Recycling and food waste" {
		t.Errorf("Expected first collection type to be 'Recycling and food waste', got '%s'", recycling.Type)
	}
	expectedRecyclingDates := []string{"2025-11-07", "2025-11-21", "2025-12-05", "2025-12-19"}
	if !equalSlices(recycling.CollectionDates, expectedRecyclingDates) {
		t.Errorf("Expected recycling dates %v, got %v", expectedRecyclingDates, recycling.CollectionDates)
	}

	// Test Rubbish bin and food waste collection
	rubbish := collections.Collections[1]
	if rubbish.Type != "Rubbish bin and food waste" {
		t.Errorf("Expected second collection type to be 'Rubbish bin and food waste', got '%s'", rubbish.Type)
	}
	expectedRubbishDates := []string{"2025-10-31", "2025-11-14", "2025-11-28", "2025-12-12"}
	if !equalSlices(rubbish.CollectionDates, expectedRubbishDates) {
		t.Errorf("Expected rubbish dates %v, got %v", expectedRubbishDates, rubbish.CollectionDates)
	}
}

func TestParseRequestParams(t *testing.T) {
	testCases := []struct {
		name          string
		url           string
		expectedUPRN  string
		expectedError string
	}{
		{
			name:          "Valid UPRN in path",
			url:           "/1234567890",
			expectedUPRN:  "1234567890",
			expectedError: "",
		},
		{
			name:          "Valid UPRN in query",
			url:           "/?uprn=1234567890",
			expectedUPRN:  "1234567890",
			expectedError: "",
		},
		{
			name:          "Invalid UPRN format",
			url:           "/not-a-uprn",
			expectedUPRN:  "",
			expectedError: "invalid UPRN format",
		},
		{
			name:          "Missing UPRN",
			url:           "/",
			expectedUPRN:  "",
			expectedError: "UPRN not provided",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			params, err := parseRequestParams(req)

			if tc.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.expectedError)
				}
				if err.Error() != tc.expectedError {
					t.Fatalf("expected error %q, got %q", tc.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if params.uprn != tc.expectedUPRN {
				t.Errorf("expected uprn %q, got %q", tc.expectedUPRN, params.uprn)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid date",
			input:   "Tuesday, 7 May 2024",
			want:    "2024-05-07",
			wantErr: false,
		},
		{
			name:    "Another valid date",
			input:   "Friday, 31 January 2025",
			want:    "2025-01-31",
			wantErr: false,
		},
		{
			name:    "Invalid date format",
			input:   "2024-05-07",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Malformed date string",
			input:   "Not a date",
			want:    "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseDate(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseDate() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if got != tc.want {
				t.Errorf("parseDate() = %v, want %v", got, tc.want)
			}
		})
	}
}
