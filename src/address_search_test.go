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

func TestSearchAddressHandler_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(`my_callback({"data":[["10001234567", "123", "Test Street, Swindon"]]})`))
	}))
	defer server.Close()

	// This is a bit of a hack, but we need to override the function that fetches the address data
	// so that it points to our test server.
	originalFetchAddressData := fetchAddressData
	fetchAddressData = func(url string) (*AddressResponse, error) {
		// We can ignore the URL passed in and just use our test server's URL
		return originalFetchAddressData(server.URL)
	}
	defer func() { fetchAddressData = originalFetchAddressData }()

	req := httptest.NewRequest(http.MethodGet, "/search-address?q=test", nil)
	rr := httptest.NewRecorder()

	SearchAddressHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rr.Code)
	}

	expectedBody := `[{"address":"Test Street, Swindon","uprn":"10001234567"}]` + "\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected body %q; got %q", expectedBody, rr.Body.String())
	}
}
