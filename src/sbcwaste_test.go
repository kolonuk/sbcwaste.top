package main

import (
	"testing"
)

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