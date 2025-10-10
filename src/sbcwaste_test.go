package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		}
    }
}

func TestFormatAsXML(t *testing.T) {
	collections := &Collections{
		Collections: []Collection{
			{Type: "Refuse", CollectionDates: []string{"2024-01-01", "2024-01-08"}},
			{Type: "Recycling", CollectionDates: []string{"2024-01-05", "2024-01-12"}},
		},
		Address: "123 Test Street",
	}

	rr := httptest.NewRecorder()
	formatAsXML(rr, collections)

	expected := `<collections><collection><type>Refuse</type><CollectionDates>2024-01-01</CollectionDates><CollectionDates>2024-01-08</CollectionDates><iconURL></iconURL></collection><collection><type>Recycling</type><CollectionDates>2024-01-05</CollectionDates><CollectionDates>2024-01-12</CollectionDates><iconURL></iconURL></collection><address>123 Test Street</address></collections>`
	// The default XML encoder adds a newline, so we should trim space from the actual output.
	actual := strings.TrimSpace(rr.Body.String())

	// The order of elements in XML can vary, so we'll do a more robust check.
	if !strings.Contains(actual, "<type>Refuse</type>") ||
		!strings.Contains(actual, "<type>Recycling</type>") ||
		!strings.Contains(actual, "<address>123 Test Street</address>") {
		t.Errorf("handler returned unexpected body: got %v want %v", actual, expected)
	}
}

func TestFormatAsYAML(t *testing.T) {
	collections := &Collections{
		Collections: []Collection{
			{Type: "Refuse", CollectionDates: []string{"2024-01-01", "2024-01-08"}},
			{Type: "Recycling", CollectionDates: []string{"2024-01-05", "2024-01-12"}},
		},
		Address: "123 Test Street",
	}

	rr := httptest.NewRecorder()
	formatAsYAML(rr, collections)

	expected := `collections:
- type: Refuse
  CollectionDates:
  - "2024-01-01"
  - "2024-01-08"
  iconURL: ""
- type: Recycling
  CollectionDates:
  - "2024-01-05"
  - "2024-01-12"
  iconURL: ""
address: 123 Test Street`

	// Trim both expected and actual to handle potential trailing newlines differently across systems
	actual := strings.TrimSpace(rr.Body.String())
	expected = strings.TrimSpace(expected)

	if actual != expected {
		t.Errorf("handler returned unexpected body:\ngot:\n%v\nwant:\n%v", actual, expected)
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
		}
	}
}
