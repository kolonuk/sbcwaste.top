package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

// Mock Collections data for testing
var testCollections = &Collections{
	Collections: []Collection{
		{
			Type:            "Refuse",
			CollectionDates: []string{"2024-01-01", "2024-01-08"},
			IconURL:         "http://example.com/refuse.png",
			IconDataURI:     "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=",
		},
		{
			Type:            "Recycling",
			CollectionDates: []string{"2024-01-02", "2024-01-09"},
			IconURL:         "http://example.com/recycling.png",
			IconDataURI:     "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=",
		},
	},
	Address: "123 Test Street, Test Town, TE57 1NG",
}

func TestFormatAsJSON(t *testing.T) {
	w := httptest.NewRecorder()

	formatAsJSON(w, testCollections)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %v", contentType)
	}

	var collections Collections
	err := json.Unmarshal(body, &collections)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(collections.Collections) != 2 {
		t.Errorf("Expected 2 collections, got %d", len(collections.Collections))
	}
}

func TestFormatAsICS(t *testing.T) {
	req := httptest.NewRequest("GET", "/12345/ics", nil)
	w := httptest.NewRecorder()
	params, _ := parseRequestParams(req)

	formatAsICS(w, testCollections, params)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "text/calendar" {
		t.Errorf("Expected Content-Type text/calendar, got %v", contentType)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, "BEGIN:VCALENDAR") {
		t.Error("ICS output does not contain BEGIN:VCALENDAR")
	}

	if !strings.Contains(bodyStr, "SUMMARY:Refuse") {
		t.Error("ICS output does not contain Refuse summary")
	}

	if !strings.Contains(bodyStr, "SUMMARY:Recycling") {
		t.Error("ICS output does not contain Recycling summary")
	}
}

func TestFormatAsXML(t *testing.T) {
	w := httptest.NewRecorder()

	formatAsXML(w, testCollections)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "application/xml" {
		t.Errorf("Expected Content-Type application/xml, got %v", contentType)
	}

	var collections Collections
	err := xml.Unmarshal(body, &collections)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML: %v", err)
	}

	if len(collections.Collections) != 2 {
		t.Errorf("Expected 2 collections, got %d", len(collections.Collections))
	}
}

func TestFormatAsYAML(t *testing.T) {
	w := httptest.NewRecorder()

	formatAsYAML(w, testCollections)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "application/x-yaml" {
		t.Errorf("Expected Content-Type application/x-yaml, got %v", contentType)
	}

	var collections Collections
	err := yaml.Unmarshal(body, &collections)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	if len(collections.Collections) != 2 {
		t.Errorf("Expected 2 collections, got %d", len(collections.Collections))
	}
}

func TestDebugOutput(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	os.Setenv("APP_ENV", "development")
	defer os.Unsetenv("APP_ENV")

	req := httptest.NewRequest("GET", "/12345/json?debug=yes", nil)
	w := httptest.NewRecorder()

	// We need to mock the fetchCollectionsFromSBC function to control the output
	originalFetchCollectionsFromSBC := fetchCollectionsFromSBC
	fetchCollectionsFromSBC = func(ctx context.Context, params *requestParams) (*Collections, error) {
		return testCollections, nil
	}
	defer func() { fetchCollectionsFromSBC = originalFetchCollectionsFromSBC }()

	WasteCollection(w, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Cache miss for UPRN: 12345") {
		t.Errorf("Expected log to contain 'Cache miss for UPRN: 12345', but it didn't. Log: %s", logOutput)
	}
}

func TestIconsInJSON(t *testing.T) {
	w := httptest.NewRecorder()

	formatAsJSON(w, testCollections)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	var collections Collections
	err := json.Unmarshal(body, &collections)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	for _, c := range collections.Collections {
		if c.IconDataURI == "" {
			t.Errorf("Expected IconDataURI for %s, but it was empty", c.Type)
		}
	}
}
