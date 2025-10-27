package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSearchAddressHandler_MissingQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/search-address", nil)
	rr := httptest.NewRecorder()

	SearchAddressHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d; got %d", http.StatusBadRequest, rr.Code)
	}

	expectedBody := "Query parameter 'q' is required\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected body %q; got %q", expectedBody, rr.Body.String())
	}
}

func TestSearchAddressHandler_Success_Integration(t *testing.T) {
	testCases := []struct {
		name  string
		query string
	}{
		{"Postcode", "SN2 2DY"},
		{"StreetName", "Kemble Drive"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This is an integration test that hits the actual Swindon council API.
			// It's intended to be run manually to diagnose issues.
			t.Skip("Skipping integration test to avoid external network calls in CI")

			escapedQuery := url.QueryEscape(tc.query)
			req := httptest.NewRequest(http.MethodGet, "/search-address?q="+escapedQuery, nil)
			rr := httptest.NewRecorder()

			SearchAddressHandler(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status %d; got %d", http.StatusOK, rr.Code)
			}

			// The response should not contain any HTML tags.
			if strings.Contains(rr.Body.String(), "<b>") {
				t.Errorf("expected body to not contain HTML tags; got %q", rr.Body.String())
			}
		})
	}
}
