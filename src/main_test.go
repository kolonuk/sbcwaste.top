package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

import (
	"os"
)

func TestMain(m *testing.M) {
	// Initialize a real browser for the tests
	initializeBrowser()
	// Wait for the browser to be ready before running tests
	<-browserReady

	// Run the tests
	code := m.Run()

	// Clean up
	if browserCancel != nil {
		browserCancel()
	}
	os.Exit(code)
}

func TestWasteCollection(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(WasteCollection)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := "<h1>sbcwaste - Swindon Borough Council Waste Collection API</h1>"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want to contain %v",
			rr.Body.String(), expected)
	}
}