// sbcwaste.go
// Date: 2024-07-15
// Version: 0.2.0
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
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
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

// rawData holds the raw data scraped from the SBC website.
type rawData struct {
	h3Nodes        []*cdp.Node
	dateNodes      []*cdp.Node
	nextThreeNodes []*cdp.Node
	imageURLs      string
}

func parseRequestParams(r *http.Request) (*requestParams, error) {
	params := &requestParams{}
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathSegments) > 0 {
		params.uprn = pathSegments[0]
	} else {
		params.uprn = r.URL.Query().Get("uprn")
	}

	if params.uprn == "" {
		return nil, errors.New("UPRN not provided")
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

func extractCollectionData(taskCtx context.Context, url string) (*rawData, error) {
	data := &rawData{}

	tasks := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(`div.bin-collection-content`, chromedp.ByQuery),
		chromedp.Nodes(`div.content-left h3`, &data.h3Nodes, chromedp.ByQueryAll),
		chromedp.Nodes(`span.nextCollectionDate`, &data.dateNodes, chromedp.ByQueryAll),
		chromedp.Nodes(`div.row.collection-next > div.row`, &data.nextThreeNodes, chromedp.ByQueryAll),
		chromedp.Evaluate(`
			(function() {
				let images = [];
				let bins = document.querySelectorAll('.bin-icons');
				bins.forEach(bin => {
					let style = window.getComputedStyle(bin);
					let bgImage = style.getPropertyValue('background-image');
					if (bgImage && bgImage.startsWith('url("')) {
						bgImage = bgImage.slice(5, -2);
						images.push(bgImage);
					}
				});
				return images.join('\n');
			})()
		`, &data.imageURLs),
	}

	if err := chromedp.Run(taskCtx, tasks); err != nil {
		return nil, fmt.Errorf("failed to run chromedp tasks: %w", err)
	}

	if len(data.h3Nodes) < 2 || len(data.dateNodes) < 2 {
		return nil, errors.New("failed to find all required nodes on the page")
	}

	return data, nil
}

func parseDate(dateStr string) (string, error) {
	parsed, err := time.Parse("Monday, 2 January 2006", dateStr)
	if err != nil {
		return "", err
	}
	return parsed.Format("2006-01-02"), nil
}

func parseCollectionData(taskCtx context.Context, data *rawData) (*Collections, error) {
	var collection1, collection2 Collection
	var firstDate1, firstDate2 string

	if err := chromedp.Run(taskCtx,
		chromedp.Text(data.h3Nodes[0].FullXPath(), &collection1.Type),
		chromedp.Text(data.dateNodes[0].FullXPath(), &firstDate1),
		chromedp.Text(data.h3Nodes[1].FullXPath(), &collection2.Type),
		chromedp.Text(data.dateNodes[1].FullXPath(), &firstDate2),
	); err != nil {
		return nil, fmt.Errorf("failed to extract initial collection data: %w", err)
	}

	if d, err := parseDate(firstDate1); err == nil {
		collection1.CollectionDates = append(collection1.CollectionDates, d)
	}
	if d, err := parseDate(firstDate2); err == nil {
		collection2.CollectionDates = append(collection2.CollectionDates, d)
	}

	// Simplified date parsing for the next three dates
	for i, node := range data.nextThreeNodes {
		var text string
		if err := chromedp.Run(taskCtx, chromedp.Text(node.FullXPath(), &text, chromedp.BySearch)); err != nil {
			log.Printf("could not get text for node: %v", err)
			continue
		}
		for _, dateStr := range strings.Split(text, "\n") {
			if trimmed := strings.TrimSpace(dateStr); trimmed != "" {
				if d, err := parseDate(trimmed); err == nil {
					if i == 0 {
						collection1.CollectionDates = append(collection1.CollectionDates, d)
					} else {
						collection2.CollectionDates = append(collection2.CollectionDates, d)
					}
				}
			}
		}
	}

	imageUrls := strings.Split(strings.TrimSpace(data.imageURLs), "\n")
	if len(imageUrls) >= 2 {
		collection1.IconURL = imageUrls[0]
		collection2.IconURL = imageUrls[1]
	}

	return &Collections{
		Collections: []Collection{collection1, collection2},
	}, nil
}

func fetchAndProcessIcons(collections *Collections) {
	for i := range collections.Collections {
		if collections.Collections[i].IconURL != "" {
			var err error
			collections.Collections[i].IconDataURI, err = convertImageToBase64URI(collections.Collections[i].IconURL)
			if err != nil {
				log.Printf("Failed to convert image %s: %v\n", collections.Collections[i].IconURL, err)
			}
		}
	}
}

func fetchCollectionsFromSBC(ctx context.Context, params *requestParams) (*Collections, error) {
	url := "https://www.swindon.gov.uk/info/20122/rubbish_and_recycling_collection_days?addressList=" + params.uprn + "&uprnSubmit=Yes"

	taskCtx, cancel := chromedp.NewContext(allocatorContext)
	defer cancel()
	taskCtx, cancel = context.WithTimeout(taskCtx, 15*time.Second)
	defer cancel()

	if params.debugging {
		log.Printf("Fetching URL: %s", url)
	}

	rawData, err := extractCollectionData(taskCtx, url)
	if err != nil {
		return nil, err
	}

	collections, err := parseCollectionData(taskCtx, rawData)
	if err != nil {
		return nil, err
	}

	if params.showIcons {
		fetchAndProcessIcons(collections)
	}

	address, err := getAddressFromUPRN(params.uprn, params.debugging)
	if err != nil {
		log.Printf("Failed to get address from UPRN: %v\n", err)
	} else {
		collections.Address = address
	}

	return collections, nil
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
		cache.Set(params.uprn, collections, time.Duration(cacheExpirySeconds)*time.Second)
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
	w.Write(yamlData)
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
	w.Write([]byte(icsBuilder.String()))
}