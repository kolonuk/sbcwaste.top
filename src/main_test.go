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

	// Test Case 2: UPRN path should be handled by WasteCollection, not fileServer
	reqWaste, _ := http.NewRequest("GET", "/1234567890", nil)
	rrWaste := httptest.NewRecorder()
	handler.ServeHTTP(rrWaste, reqWaste)

	if rrWaste.Header().Get("X-Mock-Server") == "called" {
		t.Error("rootHandler should not call fileServer for UPRN paths")
	}

	// Test Case 3: Static file path should be served by fileServer
	reqStatic, _ := http.NewRequest("GET", "/style.css", nil)
	rrStatic := httptest.NewRecorder()
	handler.ServeHTTP(rrStatic, reqStatic)

	if rrStatic.Header().Get("X-Mock-Server") != "called" {
		t.Error("rootHandler did not delegate to fileServer for static file path")
	}
}
