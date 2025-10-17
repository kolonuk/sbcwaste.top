package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/yaml.v2"
)

// loadTestHTML loads the content of the test HTML file.
func loadTestHTML(t *testing.T) *goquery.Document {
	file, err := os.Open("../testdata/sbc_response.html")
	if err != nil {
		t.Fatalf("Failed to open test data file: %v", err)
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		t.Fatalf("Failed to parse test HTML: %v", err)
	}
	return doc
}

// TestParseCollectionsFromFixture tests the parsing of collections from the HTML fixture.
func TestParseCollectionsFromFixture(t *testing.T) {
	doc := loadTestHTML(t)
	collections, err := parseCollections(doc)
	if err != nil {
		t.Fatalf("parseCollections failed: %v", err)
	}

	if len(collections.Collections) != 2 {
		t.Fatalf("Expected 2 collections, got %d", len(collections.Collections))
	}

	// Test Refuse collection
	refuse := collections.Collections[0]
	if refuse.Type != "Refuse" {
		t.Errorf("Expected first collection type to be 'Refuse', got '%s'", refuse.Type)
	}
	expectedRefuseDates := []string{"2025-10-20", "2025-11-03", "2025-11-10"}
	if !equalSlices(refuse.CollectionDates, expectedRefuseDates) {
		t.Errorf("Expected refuse dates %v, got %v", expectedRefuseDates, refuse.CollectionDates)
	}

	// Test Recycling collection
	recycling := collections.Collections[1]
	if recycling.Type != "Recycling" {
		t.Errorf("Expected second collection type to be 'Recycling', got '%s'", recycling.Type)
	}
	expectedRecyclingDates := []string{"2025-10-27", "2025-11-17", "2025-11-24"}
	if !equalSlices(recycling.CollectionDates, expectedRecyclingDates) {
		t.Errorf("Expected recycling dates %v, got %v", expectedRecyclingDates, recycling.CollectionDates)
	}
}

// equalSlices checks if two string slices are equal.
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestOutputFormatsFromFixture tests all output formats using the HTML fixture.
func TestOutputFormatsFromFixture(t *testing.T) {
	doc := loadTestHTML(t)
	collections, err := parseCollections(doc)
	if err != nil {
		t.Fatalf("parseCollections failed: %v", err)
	}
	collections.Address = "Test Address"

	// Mock request params
	params := &requestParams{
		uprn:      "12345",
		output:    "", // will be set in subtests
		debugging: false,
		showIcons: false,
	}

	t.Run("JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		formatAsJSON(w, collections)
		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK, got %v", resp.Status)
		}
		if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %v", contentType)
		}

		var out Collections
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		if len(out.Collections) != 2 {
			t.Errorf("Expected 2 collections, got %d", len(out.Collections))
		}
	})

	t.Run("XML", func(t *testing.T) {
		w := httptest.NewRecorder()
		formatAsXML(w, collections)
		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK, got %v", resp.Status)
		}
		if contentType := resp.Header.Get("Content-Type"); contentType != "application/xml" {
			t.Errorf("Expected Content-Type application/xml, got %v", contentType)
		}

		var out Collections
		if err := xml.Unmarshal(body, &out); err != nil {
			t.Fatalf("Failed to unmarshal XML: %v", err)
		}
		if len(out.Collections) != 2 {
			t.Errorf("Expected 2 collections, got %d", len(out.Collections))
		}
	})

	t.Run("YAML", func(t *testing.T) {
		w := httptest.NewRecorder()
		formatAsYAML(w, collections)
		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK, got %v", resp.Status)
		}
		if contentType := resp.Header.Get("Content-Type"); contentType != "application/x-yaml" {
			t.Errorf("Expected Content-Type application/x-yaml, got %v", contentType)
		}

		var out Collections
		if err := yaml.Unmarshal(body, &out); err != nil {
			t.Fatalf("Failed to unmarshal YAML: %v", err)
		}
		if len(out.Collections) != 2 {
			t.Errorf("Expected 2 collections, got %d", len(out.Collections))
		}
	})

	t.Run("ICS", func(t *testing.T) {
		w := httptest.NewRecorder()
		formatAsICS(w, collections, params)
		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK, got %v", resp.Status)
		}
		if contentType := resp.Header.Get("Content-Type"); contentType != "text/calendar" {
			t.Errorf("Expected Content-Type text/calendar, got %v", contentType)
		}

		bodyStr := string(body)
		if !strings.Contains(bodyStr, "SUMMARY:Refuse") {
			t.Error("ICS output does not contain Refuse summary")
		}
		if !strings.Contains(bodyStr, "SUMMARY:Recycling") {
			t.Error("ICS output does not contain Recycling summary")
		}
		if !strings.Contains(bodyStr, "DTSTART;VALUE=DATE:20251020") {
			t.Error("ICS output does not contain correct start date for Refuse")
		}
	})
}