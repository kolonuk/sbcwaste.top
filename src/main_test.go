package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, expectedContentType)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("handler returned invalid JSON: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("handler returned wrong status: got %v want %v",
			response["status"], "ok")
	}
}

func TestRootHandler(t *testing.T) {
	// Mock file server handler that returns a specific header so we know it was called
	mockFileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Mock-Server", "called")
		w.WriteHeader(http.StatusOK)
	})

	handler := rootHandler(mockFileServer)

	// Test Case 1: Root path "/" should be served by file server
	reqRoot, _ := http.NewRequest("GET", "/", nil)
	rrRoot := httptest.NewRecorder()
	handler.ServeHTTP(rrRoot, reqRoot)

	if rrRoot.Header().Get("X-Mock-Server") != "called" {
		t.Error("rootHandler did not delegate to fileServer for '/' path")
	}

	// Test Case 2: Non-existent file path should be handled by WasteCollection
	// We assume "non-existent-file" does not exist in the "static" directory
	reqWaste, _ := http.NewRequest("GET", "/non-existent-file", nil)
	rrWaste := httptest.NewRecorder()
	handler.ServeHTTP(rrWaste, reqWaste)

	// WasteCollection (mocked via its behavior in rootHandler)
	// Since WasteCollection writes output directly, we check for its signature.
	// But wait, rootHandler calls WasteCollection directly.
	// WasteCollection checks for UPRN or shows help.
	// For "/non-existent-file", parseRequestParams treats it as UPRN "non-existent-file"?
	// parseRequestParams splits by /. [0] = "non-existent-file".
	// It validates UPRN regex. "non-existent-file" fails regex.
	// So WasteCollection normally returns 400 Bad Request: "invalid UPRN format".
	// Or "UPRN not provided" if empty.

	if rrWaste.Code == http.StatusOK && rrWaste.Header().Get("X-Mock-Server") == "called" {
		t.Error("rootHandler incorrectly delegated to fileServer for non-existent file")
	}

	// We expect WasteCollection to be called.
	// "non-existent-file" is not a valid UPRN, so WasteCollection should return 400.
	// Or the help page if UPRN not provided (but here it is parsed as UPRN).
	// Let's check if the body contains "invalid UPRN format" (from WasteCollection -> parseRequestParams)
	// OR "UPRN not provided"
	// Actually, if I pass "/1234567890", it's a valid UPRN and should call WasteCollection logic.
	// But I want a path that doesn't exist AND is not a UPRN?
	// If I use a valid UPRN as path: "/1234567890"
	// stat("static/1234567890") -> err
	// rootHandler -> WasteCollection
	// WasteCollection -> parseRequestParams -> UPRN="1234567890".
	// Fetch/Cache -> returns data.

	// Let's rely on the fact that mockFileServer is NOT called.
	if rrWaste.Header().Get("X-Mock-Server") == "called" {
		t.Error("rootHandler should not call fileServer for missing files")
	}
}
