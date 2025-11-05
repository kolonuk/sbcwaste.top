package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"gopkg.in/yaml.v2"
)

var chromeAvailable bool

// TestMain sets up and tears down the test environment.
func TestMain(m *testing.M) {
	// Check if Chrome is available by trying to create a new context.
	// We use a short timeout to avoid hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	allocCtx, cancel2 := chromedp.NewExecAllocator(ctx)
	defer cancel2()
	_, err := chromedp.NewContext(allocCtx)
	chromeAvailable = err == nil

	// Run the tests.
	exitVal := m.Run()

	// Gracefully shut down the browser.
	shutdownSbcwasteChromedp()

	os.Exit(exitVal)
}

// TestLiveWasteCollectionAPI is a comprehensive integration test that queries the live SBC Waste API.
// It checks all output formats, with and without the icon scraping feature enabled.
func TestLiveWasteCollectionAPI(t *testing.T) {
	// Set the environment to development to use the SQLite cache for tests.
	t.Setenv("APP_ENV", "development")

	// Skip this test in short mode as it makes live network requests.
	if testing.Short() {
		t.Skip("Skipping live API test in short mode")
	}

	uprns := []string{
		"200001860227", // Crickhollow, High Street
		"200001814797", // Dormers, High Street
		"10008541132",  // 36 Langton Park
		"100121129753", // 89 Dulverton Avenue
		"200001615122", // 105 County Road
	}

	formats := []string{"json", "xml", "yaml", "ics"}
	iconSettings := []bool{false, true} // false = no icons, true = with icons

	for _, uprn := range uprns {
		for _, format := range formats {
			for _, showIcons := range iconSettings {
				// ICS format does not support icons, so skip that combination.
				if format == "ics" && showIcons {
					continue
				}

				testName := fmt.Sprintf("UPRN_%s_Format_%s_Icons_%t", uprn, format, showIcons)

				t.Run(testName, func(t *testing.T) {
					// If icons are requested but Chrome is not available, skip the test.
					if showIcons && !chromeAvailable {
						t.Skip("google-chrome not found or not functional, skipping icon-related test")
					}

					// Create a request to our handler.
					url := fmt.Sprintf("/%s/%s", uprn, format)
					if showIcons {
						url += "?icons=yes"
					}
					req := httptest.NewRequest(http.MethodGet, url, nil)
					w := httptest.NewRecorder()

					// We need to use the actual WasteCollection handler.
					WasteCollection(w, req)

					resp := w.Result()
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						t.Fatalf("Failed to read response body: %v", err)
					}

					if resp.StatusCode != http.StatusOK {
						t.Fatalf("Expected status OK, got %v. Body: %s", resp.Status, string(body))
					}

					// Validate the response based on the format.
					validateResponse(t, body, format, showIcons)
				})
			}
		}
	}
}

// validateResponse checks the structure and content of the API response.
func validateResponse(t *testing.T, body []byte, format string, showIcons bool) {
	switch format {
	case "json":
		var collections Collections
		if err := json.Unmarshal(body, &collections); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v. Body: %s", err, string(body))
		}
		validateCollectionsStruct(t, &collections, showIcons)
	case "xml":
		var collections Collections
		if err := xml.Unmarshal(body, &collections); err != nil {
			t.Fatalf("Failed to unmarshal XML: %v. Body: %s", err, string(body))
		}
		validateCollectionsStruct(t, &collections, showIcons)
	case "yaml":
		var collections Collections
		if err := yaml.Unmarshal(body, &collections); err != nil {
			t.Fatalf("Failed to unmarshal YAML: %v. Body: %s", err, string(body))
		}
		validateCollectionsStruct(t, &collections, showIcons)
	case "ics":
		bodyStr := string(body)
		if !strings.HasPrefix(bodyStr, "BEGIN:VCALENDAR") {
			t.Error("ICS response does not start with BEGIN:VCALENDAR")
		}
		if !strings.Contains(bodyStr, "BEGIN:VEVENT") {
			t.Error("ICS response does not contain any VEVENT")
		}
	}
}

// validateCollectionsStruct checks the common structure of the Collections object.
func validateCollectionsStruct(t *testing.T, c *Collections, showIcons bool) {
	if c.Address == "" {
		t.Error("Expected Address to be populated, but it was empty")
	}

	if len(c.Collections) == 0 {
		t.Logf("Address: %s", c.Address)
		t.Fatal("Expected at least one collection, but got none")
	}

	for _, coll := range c.Collections {
		if coll.Type == "" {
			t.Error("Expected collection Type to be populated, but it was empty")
		}

		if len(coll.CollectionDates) == 0 {
			t.Error("Expected at least one CollectionDate, but got none")
		}

		for _, dateStr := range coll.CollectionDates {
			if len(dateStr) != 10 || strings.Count(dateStr, "-") != 2 {
				t.Errorf("Expected date to be in YYYY-MM-DD format, but got %s", dateStr)
			}
		}

		if showIcons {
			if coll.IconURL == "" {
				t.Error("Expected IconURL to be populated when icons are requested")
			}
			if !strings.HasPrefix(coll.IconURL, "http") {
				t.Errorf("Expected IconURL to be a valid URL, but got %s", coll.IconURL)
			}
			if coll.IconDataURI == "" {
				t.Error("Expected IconDataURI to be populated when icons are requested")
			}
			if !strings.HasPrefix(coll.IconDataURI, "data:image/") {
				t.Errorf("Expected IconDataURI to be a valid data URI, but got %s", coll.IconDataURI)
			}
		} else {
			// With omitempty, the fields should not even be present.
			// Unmarshaling into a struct will leave them as empty strings,
			// so we can still check for that.
			if coll.IconURL != "" {
				t.Error("Expected IconURL to be empty when icons are not requested")
			}
			if coll.IconDataURI != "" {
				t.Error("Expected IconDataURI to be empty when icons are not requested")
			}
		}
	}
}
