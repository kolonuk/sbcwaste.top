package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

const mockHTML = `
<!DOCTYPE html>
<html>
<body>
    <div class="bin-collection-content">
        <h3>Recycling</h3>
        <span class="nextCollectionDate">Monday, 1 January 2024</span>
        <div class="row collection-next">
            <div class="row">
                <p>Monday, 8 January 2024</p>
                <p>Monday, 15 January 2024</p>
            </div>
        </div>
        <div class="bin-icons" style="background-image: url('https://www.swindon.gov.uk/recycling_icon.png');"></div>
    </div>
    <div class="bin-collection-content">
        <h3>Rubbish</h3>
        <span class="nextCollectionDate">Tuesday, 2 January 2024</span>
        <div class="row collection-next">
            <div class="row">
                <p>Tuesday, 9 January 2024</p>
                <p>Tuesday, 16 January 2024</p>
            </div>
        </div>
        <div class="bin-icons" style="background-image: url('https://www.swindon.gov.uk/rubbish_icon.png');"></div>
    </div>
</body>
</html>
`

func TestFetchCollectionsFromSBC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(mockHTML))
	}))
	defer server.Close()

	originalFetch := fetchCollectionsFromSBC
	fetchCollectionsFromSBC = func(ctx context.Context, params *requestParams) (*Collections, error) {
		client := &http.Client{}
		req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
		if err != nil {
			return nil, err
		}

		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, err
		}

		return parseCollections(doc)
	}
	defer func() { fetchCollectionsFromSBC = originalFetch }()

	params := &requestParams{uprn: "12345", showIcons: false}
	collections, err := fetchCollectionsFromSBC(context.Background(), params)
	if err != nil {
		t.Fatalf("fetchCollectionsFromSBC failed: %v", err)
	}

	if len(collections.Collections) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(collections.Collections))
	}

	if collections.Collections[0].Type != "Recycling" {
		t.Errorf("expected collection type 'Recycling', got '%s'", collections.Collections[0].Type)
	}
	if collections.Collections[1].Type != "Rubbish" {
		t.Errorf("expected collection type 'Rubbish', got '%s'", collections.Collections[1].Type)
	}
	if len(collections.Collections[0].CollectionDates) != 3 {
		t.Errorf("expected 3 recycling dates, got %d", len(collections.Collections[0].CollectionDates))
	}
}

func TestOutputFormats(t *testing.T) {
	collections := &Collections{
		Collections: []Collection{
			{
				Type:            "Recycling",
				CollectionDates: []string{"2024-01-01", "2024-01-08"},
				IconURL:         "http://example.com/recycling.png",
			},
			{
				Type:            "Rubbish",
				CollectionDates: []string{"2024-01-02", "2024-01-09"},
				IconURL:         "http://example.com/rubbish.png",
			},
		},
		Address: "123 Test Street, Swindon",
	}

	testCases := []struct {
		format      string
		contentType string
		validator   func(string) bool
	}{
		{"json", "application/json", func(body string) bool {
			return strings.Contains(body, `"type":"Recycling"`) && strings.Contains(body, `"CollectionDates":["2024-01-01","2024-01-08"]`)
		}},
		{"xml", "application/xml", func(body string) bool {
			return strings.Contains(body, "<type>Recycling</type>") && strings.Contains(body, "<CollectionDates>2024-01-08</CollectionDates>")
		}},
		{"yaml", "application/x-yaml", func(body string) bool {
			return strings.Contains(body, "type: Recycling") && strings.Contains(body, "- \"2024-01-01\"") && strings.Contains(body, "- \"2024-01-08\"")
		}},
		{"ics", "text/calendar", func(body string) bool {
			return strings.Contains(body, "SUMMARY:Recycling") && strings.Contains(body, "DTSTART;VALUE=DATE:20240108")
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			rr := httptest.NewRecorder()

			switch tc.format {
			case "json":
				formatAsJSON(rr, collections)
			case "xml":
				formatAsXML(rr, collections)
			case "yaml":
				formatAsYAML(rr, collections)
			case "ics":
				formatAsICS(rr, collections, &requestParams{uprn: "12345"})
			}

			res := rr.Result()
			defer res.Body.Close()

			if contentType := res.Header.Get("Content-Type"); contentType != tc.contentType {
				t.Errorf("expected content type %s, got %s", tc.contentType, contentType)
			}

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response body: %v", err)
			}

			if !tc.validator(string(body)) {
				t.Errorf("output validation failed for format %s. Body:\n%s", tc.format, body)
			}
		})
	}
}

func TestIconCachingDynamic(t *testing.T) {
	// Mock the image conversion
	originalConvert := convertImageToBase64URI
	var convertCount int
	convertImageToBase64URI = func(url string) (string, error) {
		convertCount++
		return "data:image/png;base64,mocked_base64_data", nil
	}
	defer func() { convertImageToBase64URI = originalConvert }()

	iconURL := "http://example.com/new_icon.png"

	// 1. First call, should fetch and cache
	cachedIcon, err := getIcon(iconURL)
	if err != nil {
		t.Fatalf("Failed to get icon from cache: %v", err)
	}
	if cachedIcon != "data:image/png;base64,mocked_base64_data" {
		t.Errorf("Unexpected cached icon data: got %s", cachedIcon)
	}
	if convertCount != 1 {
		t.Errorf("Expected convert function to be called once, but was called %d times", convertCount)
	}

	// 2. Second call, should use cache
	cachedIcon, err = getIcon(iconURL)
	if err != nil {
		t.Fatalf("Failed to get icon from cache on second call: %v", err)
	}
	if cachedIcon != "data:image/png;base64,mocked_base64_data" {
		t.Errorf("Unexpected cached icon data on second call: got %s", cachedIcon)
	}
	if convertCount != 1 {
		t.Errorf("Expected convert function to still be called only once, but was called %d times", convertCount)
	}
}