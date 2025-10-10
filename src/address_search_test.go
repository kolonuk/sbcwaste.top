package main

import (
	"net/http"
	"net/http/httptest"
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