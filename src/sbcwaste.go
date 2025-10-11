// sbcwaste.go
// Date: 2024-07-15
// Version: 0.2.1
// License: GPL-3.0
// License Details: https://www.gnu.org/licenses/gpl-3.0.en.html
//

package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/yaml.v2"
)

// A single collection
type Collection struct {
	Type            string   `json:"type" xml:"type" yaml:"type"`
	CollectionDates []string `json:"CollectionDates" xml:"CollectionDates" yaml:"CollectionDates"`
	IconURL         string   `json:"iconURL" xml:"iconURL" yaml:"iconURL"`
	IconDataURI     string   `json:"iconDataURI,omitempty" xml:"iconDataURI,omitempty" yaml:"iconDataURI,omitempty"`
}

// All the collections
type Collections struct {
	XMLName     xml.Name     `json:"-" xml:"collections" yaml:"-"`
	Collections []Collection `json:"collections" xml:"collection" yaml:"collections"`
	Address     string       `json:"address" xml:"address" yaml:"address"`
}

type requestParams struct {
	uprn      string
	output    string
	debugging bool
	showIcons bool
}

var iconCache = struct {
	sync.RWMutex
	data map[string]string
}{data: make(map[string]string)}

var lastCacheRefresh time.Time

const cacheDuration = 7 * 24 * time.Hour

func getIcon(url string) (string, error) {
	iconCache.RLock()
	if time.Since(lastCacheRefresh) > cacheDuration {
		iconCache.RUnlock()
		if err := refreshIconCache(); err != nil {
			return "", err
		}
		iconCache.RLock()
	}
	cachedIcon, ok := iconCache.data[url]
	iconCache.RUnlock()
	if ok {
		return cachedIcon, nil
	}

	return "", fmt.Errorf("icon not found in cache")
}

func refreshIconCache() error {
	log.Println("Refreshing icon cache")
	// In a real application, you would fetch the URLs from a reliable source.
	// For this example, we'll assume a static list of known icon URLs.
	knownIconURLs := []string{
		"https://www.swindon.gov.uk/recycling_icon.png",
		"https://www.swindon.gov.uk/rubbish_icon.png",
	}

	newCache := make(map[string]string)
	for _, url := range knownIconURLs {
		dataURI, err := convertImageToBase64URI(url)
		if err != nil {
			log.Printf("Failed to fetch and convert icon %s: %v", url, err)
			continue
		}
		newCache[url] = dataURI
	}

	iconCache.Lock()
	iconCache.data = newCache
	lastCacheRefresh = time.Now()
	iconCache.Unlock()
	log.Println("Icon cache refreshed")
	return nil
}

func parseRequestParams(r *http.Request) (*requestParams, error) {
	params := &requestParams{}
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathSegments) > 0 && pathSegments[0] != "" {
		params.uprn = pathSegments[0]
	}

	if params.uprn == "" {
		params.uprn = r.URL.Query().Get("uprn")
	}

	if params.uprn == "" {
		return nil, errors.New("UPRN not provided")
	}

	// Validate that the UPRN is a numeric value
	if matched, _ := regexp.MatchString("^[0-9]+$", params.uprn); !matched {
		return nil, errors.New("invalid UPRN format")
	}

	if len(pathSegments) >= 2 {
		params.output = pathSegments[1]
	} else {
		params.output = "json"
	}

	params.debugging = r.URL.Query().Get("debug") == "yes"
	params.showIcons = r.URL.Query().Get("icons") == "yes"

	return params, nil
}

// fetchCollectionsFromSBC fetches waste collection data from the SBC website using HTTP requests.
var fetchCollectionsFromSBC = func(ctx context.Context, params *requestParams) (*Collections, error) {
	if params.debugging {
		log.Printf("Fetching URL: https://www.swindon.gov.uk/info/20122/rubbish_and_recycling_collection_days?addressList=%s&uprnSubmit=Yes", params.uprn)
	}

	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.swindon.gov.uk/info/20122/rubbish_and_recycling_collection_days?addressList="+params.uprn+"&uprnSubmit=Yes", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	collections, err := parseCollections(doc)
	if err != nil {
		return nil, err
	}

	if params.showIcons {
		for i := range collections.Collections {
			if collections.Collections[i].IconURL != "" {
				iconDataURI, err := getIcon(collections.Collections[i].IconURL)
				if err != nil {
					log.Printf("Failed to get icon %s from cache: %v", collections.Collections[i].IconURL, err)
					continue
				}
				collections.Collections[i].IconDataURI = iconDataURI
			}
		}
	}

	address, err := getAddressFromUPRN(params.uprn, params.debugging)
	if err != nil {
		log.Printf("Failed to get address from UPRN: %v\n", err)
	} else {
		collections.Address = address
	}

	return collections, nil
}

func parseDate(dateStr string) (string, error) {
	parsed, err := time.Parse("Monday, 2 January 2006", dateStr)
	if err != nil {
		return "", err
	}
	return parsed.Format("2006-01-02"), nil
}

// parseCollections parses the HTML document and extracts collection data.
func parseCollections(doc *goquery.Document) (*Collections, error) {
	var collections Collections
	var types, dates []string

	doc.Find("div.bin-collection-content h3").Each(func(i int, s *goquery.Selection) {
		types = append(types, s.Text())
	})

	doc.Find("span.nextCollectionDate").Each(func(i int, s *goquery.Selection) {
		dates = append(dates, s.Text())
	})

	if len(types) < 2 || len(dates) < 2 {
		return nil, errors.New("failed to find all required nodes on the page")
	}

	var collectionsList []Collection
	for i := 0; i < 2; i++ {
		parsedDate, err := parseDate(dates[i])
		if err != nil {
			log.Printf("Could not parse date %s: %v", dates[i], err)
			continue
		}
		collectionsList = append(collectionsList, Collection{
			Type:            types[i],
			CollectionDates: []string{parsedDate},
		})
	}

	doc.Find("div.row.collection-next > div.row").Each(func(i int, s *goquery.Selection) {
		s.Find("p").Each(func(j int, p *goquery.Selection) {
			dateStr := strings.TrimSpace(p.Text())
			if parsedDate, err := parseDate(dateStr); err == nil {
				if i < len(collectionsList) {
					collectionsList[i].CollectionDates = append(collectionsList[i].CollectionDates, parsedDate)
				}
			}
		})
	})

	doc.Find(".bin-icons").Each(func(i int, s *goquery.Selection) {
		style, _ := s.Attr("style")
		re := regexp.MustCompile(`url\(['"]?([^'"]+)['"]?\)`)
		matches := re.FindStringSubmatch(style)
		if len(matches) > 1 {
			if i < len(collectionsList) {
				collectionsList[i].IconURL = matches[1]
			}
		}
	})

	collections.Collections = collectionsList
	return &collections, nil
}

func WasteCollection(w http.ResponseWriter, r *http.Request) {
	params, err := parseRequestParams(r)
	if err != nil {
		if err.Error() == "UPRN not provided" {
			showHelp(w)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	cache, err := NewCache(ctx)
	if err != nil {
		http.Error(w, "Failed to initialize cache", http.StatusInternalServerError)
		log.Printf("Failed to initialize cache: %v", err)
		return
	}
	defer cache.Close()

	collections, err := cache.Get(params.uprn)
	if err != nil || collections == nil {
		if params.debugging {
			log.Printf("Cache miss for UPRN: %s", params.uprn)
		}
		collections, err = fetchCollectionsFromSBC(ctx, params)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch collections: %v", err), http.StatusInternalServerError)
			return
		}

		cacheExpirySecondsStr := os.Getenv("CACHE_EXPIRY_SECONDS")
		cacheExpirySeconds, _ := strconv.Atoi(cacheExpirySecondsStr)
		if cacheExpirySeconds <= 0 {
			cacheExpirySeconds = 259200 // 3 days
		}
		if err := cache.Set(params.uprn, collections, time.Duration(cacheExpirySeconds)*time.Second); err != nil {
			log.Printf("Failed to cache collections for UPRN %s: %v", params.uprn, err)
		}
	} else {
		if params.debugging {
			log.Printf("Cache hit for UPRN: %s", params.uprn)
		}
	}

	w.Header().Set("Cache-Control", "max-age=3600")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	switch params.output {
	case "json", "":
		formatAsJSON(w, collections)
	case "ics":
		formatAsICS(w, collections, params)
	case "xml":
		formatAsXML(w, collections)
	case "yaml":
		formatAsYAML(w, collections)
	default:
		http.Error(w, "Invalid output format", http.StatusBadRequest)
	}
}

func formatAsJSON(w http.ResponseWriter, collections *Collections) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(collections); err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
	}
}

func formatAsXML(w http.ResponseWriter, collections *Collections) {
	w.Header().Set("Content-Type", "application/xml")
	if err := xml.NewEncoder(w).Encode(collections); err != nil {
		http.Error(w, "Failed to marshal XML", http.StatusInternalServerError)
	}
}

func formatAsYAML(w http.ResponseWriter, collections *Collections) {
	w.Header().Set("Content-Type", "application/x-yaml")
	yamlData, err := yaml.Marshal(collections)
	if err != nil {
		http.Error(w, "Failed to marshal YAML", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(yamlData); err != nil {
		log.Printf("Failed to write YAML response: %v", err)
	}
}

func showHelp(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<h1>sbcwaste - Swindon Borough Council Waste Collection API</h1>")
	fmt.Fprintln(w, "<p>This service provides waste collection dates for properties in Swindon.</p>")
	fmt.Fprintln(w, "<h2>Usage:</h2>")
	fmt.Fprintln(w, "<p><code>/[UPRN]/[format]</code></p>")
	fmt.Fprintln(w, "<ul>")
	fmt.Fprintln(w, "<li><b>UPRN</b>: The Unique Property Reference Number for your address.</li>")
	fmt.Fprintln(w, "<li><b>format</b>: The output format. Can be <code>json</code> (default), <code>ics</code>, <code>xml</code>, or <code>yaml</code>.</li>")
	fmt.Fprintln(w, "</ul>")
	fmt.Fprintln(w, "<h2>Optional Parameters:</h2>")
	fmt.Fprintln(w, "<ul>")
	fmt.Fprintln(w, "<li><b>?debug=yes</b>: Enable debug logging.</li>")
	fmt.Fprintln(w, "<li><b>?icons=yes</b>: Include base64-encoded icon data in the output (JSON, XML, YAML only).</li>")
	fmt.Fprintln(w, "</ul>")
}

func formatAsICS(w http.ResponseWriter, collections *Collections, params *requestParams) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}
	prodID := fmt.Sprintf("-//Swindon Borough Council Waste Collections//sbcwaste-%s//EN", appEnv)
	var icsBuilder strings.Builder
	icsBuilder.WriteString("BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:" + prodID + "\n")

	url := "https://www.swindon.gov.uk/info/20122/rubbish_and_recycling_collection_days?addressList=" + params.uprn + "&uprnSubmit=Yes"

	for _, collection := range collections.Collections {
		for _, dateString := range collection.CollectionDates {
			eventDate, _ := time.Parse("2006-01-02", dateString)
			dtStamp := time.Now().UTC().Format("20060102T150405Z")
			uid := foldLine("UID:" + generateUID(collection.Type, dateString, collections.Address))
			start := eventDate.Format("20060102")
			end := eventDate.Add(24 * time.Hour).Format("20060102")
			summary := foldLine(fmt.Sprintf("SUMMARY:%s", collection.Type))
			location := foldLine(fmt.Sprintf("LOCATION:%s", collections.Address))
			attach := foldLine(fmt.Sprintf("ATTACH;VALUE=URI:%s", collection.IconDataURI))
			urlLine := foldLine(fmt.Sprintf("URL:%s", url))
			fmt.Fprintf(&icsBuilder, "BEGIN:VEVENT\r\n%s\r\nDTSTAMP:%s\r\nDTSTART;VALUE=DATE:%s\r\nDTEND;VALUE=DATE:%s\r\n%s\r\n%s\r\n%s\r\nTRANSP:TRANSPARENT\r\n%s\r\nEND:VEVENT\r\n",
				uid, dtStamp, start, end, summary, location, attach, urlLine)
		}
	}
	icsBuilder.WriteString("END:VCALENDAR")
	w.Header().Set("Content-Type", "text/calendar")
	if _, err := w.Write([]byte(icsBuilder.String())); err != nil {
		log.Printf("Failed to write ICS response: %v", err)
	}
}