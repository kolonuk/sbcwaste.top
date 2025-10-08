// sbcwaste.go
// Date: 2024-07-15
// Version: 0.1.5
// License: GPL-3.0
// License Details: https://www.gnu.org/licenses/gpl-3.0.en.html
//

package main

import (
	"context"
	"encoding/json"
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
)

// A single collection
type Collection struct {
	Type            string   `json:"type"`
	CollectionDates []string `json:"CollectionDates"`
	IconURL         string   `json:"iconURL"`
	IconDataURI     string   `json:"iconDataURI,omitempty"`
}

// All the collections
type Collections struct {
	Collections []Collection `json:"collections"`
	Address     string       `json:"address"`
}

// Define a struct to match the JSON structure
type AddressResponse struct {
	Name    string     `json:"name"`
	Columns []string   `json:"columns"`
	Data    [][]string `json:"data"`
	Total   int        `json:"total"`
}

func WasteCollection(w http.ResponseWriter, r *http.Request) {
	var err error
	var collections *Collections
	var collection1 Collection
	var collection2 Collection
	var nextThreenodes []*cdp.Node
	var h3Nodes []*cdp.Node
	var dateNodes []*cdp.Node
	var Firstdate1 string
	var Firstdate2 string
	var images string
	var outputFormat string
	var debuggingEnable bool
	var showIcons bool

	// Get the application environment from the environment variable
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development" // Default to development
	}

	log.Default().Printf("URL: %s", r.URL)

	// Are we in debug mode?
	debuggingEnable = (r.URL.Query().Get("debug") == "yes")
	log.Default().Printf("Debugging: %t\n", debuggingEnable)

	// Are we showing icons?
	showIcons = (r.URL.Query().Get("icons") == "yes")
	log.Default().Printf("Showing Icons: %t\n", showIcons)

	// Trim the leading slash and then split the path into segments
	pathSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if debuggingEnable {
		log.Default().Printf("%v", pathSegments)
	}

	// Get UPRN from the query parameters. Check for both upper and lower case UPRN. Return an error if UPRN is not provided.
	// UPRN := r.URL.Query().Get("UPRN")
	UPRN := pathSegments[0]
	if UPRN == "" {
		UPRN = r.URL.Query().Get("uprn")
	}
	if UPRN == "" {
		log.Default().Printf("UPRN not provided, showing help page\n")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, "<h1>sbcwaste - Swindon Borough Council Waste Collection API</h1>")
		fmt.Fprintln(w, "<p>This service provides waste collection dates for properties in Swindon.</p>")
		fmt.Fprintln(w, "<h2>Usage:</h2>")
		fmt.Fprintln(w, "<p><code>/[UPRN]/[format]</code></p>")
		fmt.Fprintln(w, "<ul>")
		fmt.Fprintln(w, "<li><b>UPRN</b>: The Unique Property Reference Number for your address.</li>")
		fmt.Fprintln(w, "<li><b>format</b>: The output format. Can be <code>json</code> (default) or <code>ics</code>.</li>")
		fmt.Fprintln(w, "</ul>")
		fmt.Fprintln(w, "<h2>Optional Parameters:</h2>")
		fmt.Fprintln(w, "<ul>")
		fmt.Fprintln(w, "<li><b>?debug=yes</b>: Enable debug logging.</li>")
		fmt.Fprintln(w, "<li><b>?icons=yes</b>: Include icon data in the JSON output.</li>")
		fmt.Fprintln(w, "</ul>")
		return
	}
	if debuggingEnable {
		log.Default().Printf("UPRN: %s", UPRN)
	}

	// // Get the output format from the query parameters. Return an error if the output format is not provided.
	// outputFormat := r.URL.Query().Get("format")
	if len(pathSegments) >= 2 {
		outputFormat = pathSegments[1]
	}
	if debuggingEnable {
		log.Default().Printf("outputFormat: %s", outputFormat)
	}

	url := "https://www.swindon.gov.uk/info/20122/rubbish_and_recycling_collection_days?addressList=" + UPRN + "&uprnSubmit=Yes"

	// Initialize cache
	ctx := context.Background()
	cache, err := NewCache(ctx)
	if err != nil {
		http.Error(w, "Failed to initialize cache", http.StatusInternalServerError)
		log.Printf("Failed to initialize cache: %v", err)
		return
	}
	defer cache.Close()

	// Check cache first
	cachedCollections, err := cache.Get(UPRN)
	if err != nil {
		log.Printf("Error getting from cache: %v", err)
	}

	if cachedCollections != nil {
		log.Printf("Cache hit for UPRN: %s", UPRN)
		collections = cachedCollections
	} else {
		log.Printf("Cache miss for UPRN: %s", UPRN)
		collections = &Collections{}
		// Create a new chromedp context, directing it to use the non-snap version of chrome
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath("/usr/bin/chromium"),
			chromedp.Flag("no-sandbox", true), // Running as root requires this
			chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36`),
		)
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancel()

		// create context
		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()
		ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		// Do the navigate and process the web page.
		if debuggingEnable {
			log.Default().Printf("URL: %s", url)
		}
		err = chromedp.Run(ctx,
			chromedp.Navigate(url),
		)
		if errors.Is(err, context.DeadlineExceeded) {
			// Log the timeout
			log.Default().Printf("Operation timed out for UPRN: %s", UPRN)
			// Return an HTTP 408 Request Timeout status code
			http.Error(w, "Operation timed out", http.StatusRequestTimeout)
			return
		} else if err != nil {
			http.Error(w, "Unable to contact SBC website", http.StatusBadGateway)
			log.Default().Print(err)
			return
		}

		// Set tasks to run.
		tasks := chromedp.Tasks{
			chromedp.WaitVisible(`div.bin-collection-content`, chromedp.ByQuery),                      // The div that contains all the data we need
			chromedp.Nodes(`div.content-left h3`, &h3Nodes, chromedp.ByQueryAll),                      // The h3 is the type of collection
			chromedp.Nodes(`span.nextCollectionDate`, &dateNodes, chromedp.ByQueryAll),                // This is date 1 of 4 needed
			chromedp.Nodes(`div.row.collection-next > div.row`, &nextThreenodes, chromedp.ByQueryAll), // This is date 2, 3, 4 of 4 needed
			// Evaluate JavaScript to get the background-image URLs. These are controlled by javascript, so we can only get these by javascript,
			// and only once the whole webpage is constructed and completely loaded
			chromedp.Evaluate(`
		(function() {
			let images = [];
			let bins = document.querySelectorAll('.bin-icons');
			bins.forEach(bin => {
				let style = window.getComputedStyle(bin);
				let bgImage = style.getPropertyValue('background-image');
				if (bgImage && bgImage.startsWith('url("')) {
					bgImage = bgImage.slice(5, -2); // Remove 'url("' and '")'
					images.push(bgImage);
				}
			});
			return images.join('\n');
		})()
		`, &images),
		}

		// Run the tasks.
		err = chromedp.Run(ctx,
			tasks,
		)
		if err != nil {
			http.Error(w, "Unable to parse SBC website", http.StatusInternalServerError)
			log.Default().Print(err)
			return
		}

		// Grab each of the entities.
		if len(h3Nodes) >= 2 && len(dateNodes) >= 2 {
			err = chromedp.Run(ctx,
				chromedp.Text(h3Nodes[0].FullXPath(), &collection1.Type),
				chromedp.Text(dateNodes[0].FullXPath(), &Firstdate1),
				chromedp.Text(h3Nodes[1].FullXPath(), &collection2.Type),
				chromedp.Text(dateNodes[1].FullXPath(), &Firstdate2),
			)
			if err != nil {
				http.Error(w, "Unable to parse SBC website data nodes", http.StatusInternalServerError)
				log.Default().Print(err)
				return
			}

			if Firstdate1 != "" { // Check if the trimmed date is not an empty string
				parsedDate, err := time.Parse("Monday, 2 January 2006", Firstdate1)
				if err != nil {
					// Handle parsing error, maybe log or continue
					http.Error(w, "Unable to parse SBC website dates 1", http.StatusInternalServerError)
					log.Default().Print(err)
					return
				}
				numericalDate := parsedDate.Format("2006-01-02")
				collection1.CollectionDates = append(collection1.CollectionDates, numericalDate)
			}
			if Firstdate2 != "" { // Check if the trimmed date is not an empty string
				parsedDate, err := time.Parse("Monday, 2 January 2006", Firstdate2)
				if err != nil {
					// Handle parsing error, maybe log or continue
					http.Error(w, "Unable to parse SBC website dates 2", http.StatusInternalServerError)
					log.Default().Print(err)
					return
				}
				numericalDate := parsedDate.Format("2006-01-02")
				collection2.CollectionDates = append(collection2.CollectionDates, numericalDate)
			}
		}

		// grab and assign the next three dates for each entity
		if len(nextThreenodes) > 1 {
			var text string
			err = chromedp.Run(ctx, chromedp.Text(nextThreenodes[0].FullXPath(), &text, chromedp.BySearch))
			if err != nil {
				http.Error(w, "Unable to parse SBC website dates 3", http.StatusInternalServerError)
				log.Default().Print(err)
				return
			}
			splitText := strings.Split(text, "\n")
			for _, date := range splitText {
				// Trim space to remove any leading or trailing whitespace
				trimmedDate := strings.TrimSpace(date)
				if trimmedDate != "" { // Check if the trimmed date is not an empty string
					parsedDate, err := time.Parse("Monday, 2 January 2006", trimmedDate)
					if err != nil {
						// Handle parsing error, maybe log or continue
						http.Error(w, "Unable to parse SBC website dates 4", http.StatusInternalServerError)
						log.Default().Print(err)
						continue
					}
					numericalDate := parsedDate.Format("2006-01-02")
					collection1.CollectionDates = append(collection1.CollectionDates, numericalDate)
				}
			}

			err = chromedp.Run(ctx, chromedp.Text(nextThreenodes[1].FullXPath(), &text, chromedp.BySearch))
			if err != nil {
				http.Error(w, "Unable to parse SBC website dates 5", http.StatusInternalServerError)
				log.Default().Print(err)
				return
			}
			splitText = strings.Split(text, "\n")
			for _, date := range splitText {
				// Trim space to remove any leading or trailing whitespace
				trimmedDate := strings.TrimSpace(date)
				if trimmedDate != "" { // Check if the trimmed date is not an empty string
					parsedDate, err := time.Parse("Monday, 2 January 2006", trimmedDate)
					if err != nil {
						// Handle parsing error, maybe log or continue
						http.Error(w, "Unable to parse SBC website dates 6", http.StatusInternalServerError)
						log.Default().Print(err)
						continue
					}
					numericalDate := parsedDate.Format("2006-01-02")
					collection2.CollectionDates = append(collection2.CollectionDates, numericalDate)
				}
			}
		}

		// Do the background images found in .bin-icons divs
		imageUrls := strings.Split(strings.TrimSpace(images), "\n")

		collection1.IconURL = imageUrls[0]
		collection2.IconURL = imageUrls[1]

		if showIcons {
			collection1.IconDataURI, err = convertImageToBase64URI(collection1.IconURL)
			if err != nil {
				http.Error(w, "Unable to parse SBC website icons 1", http.StatusInternalServerError)
				log.Default().Printf("Failed to convert image %s: %v\n", collection1.IconURL, err)
			}
			collection2.IconDataURI, err = convertImageToBase64URI(collection2.IconURL)
			if err != nil {
				http.Error(w, "Unable to parse SBC website icons 1", http.StatusInternalServerError)
				log.Default().Printf("Failed to convert image %s: %v\n", collection2.IconURL, err)
			}
		}

		// Query the address API to get the address string using UPRN variable
		address, err := queryAddressAPI("https://maps.swindon.gov.uk/getdata.aspx?callback=jQuery16406504322666596749_1721033956585&type=jsonp&service=LocationSearch&RequestType=LocationSearch&location="+UPRN+"&pagesize=13&startnum=1&gettotals=false&axuid=1721033978935&mapsource=mapsources/MyHouse&_=1721033978935", debuggingEnable)
		if err != nil {
			http.Error(w, "Unable to parse SBC map location JSON", http.StatusBadGateway)
			log.Default().Printf("Failed to query address API: %v\n", err)
			//		return
		} else {
			if debuggingEnable {
				log.Default().Printf("Queried Address: %s\n", address)
			}
			collections.Address = address
		}

		// save each entity into the json array
		collections.Collections = append(collections.Collections, collection1, collection2)

		if debuggingEnable {
			log.Default().Printf("Address: %s", collections.Address)
			log.Default().Printf("Type: %s", collection1.Type)
			log.Default().Printf("CollectionDates: %s", collection1.CollectionDates)
			log.Default().Printf("IconURL: %s", collection1.IconURL)
			log.Default().Printf("Type: %s", collection2.Type)
			log.Default().Printf("CollectionDates: %s", collection2.CollectionDates)
			log.Default().Printf("IconURL: %s", collection2.IconURL)
			log.Default().Printf("IconDataURI: %s", collection1.IconDataURI)
			log.Default().Printf("IconDataURI: %s", collection2.IconDataURI)
		}

		// Get cache expiry from env var, default to 3 days
		cacheExpirySecondsStr := os.Getenv("CACHE_EXPIRY_SECONDS")
		cacheExpirySeconds, err := strconv.Atoi(cacheExpirySecondsStr)
		if err != nil || cacheExpirySeconds <= 0 {
			cacheExpirySeconds = 259200 // 3 days
		}
		cacheExpiry := time.Duration(cacheExpirySeconds) * time.Second

		// Set data in cache
		err = cache.Set(UPRN, collections, cacheExpiry)
		if err != nil {
			log.Printf("Error setting cache: %v", err)
		}
	}

	// Set more http headers
	w.Header().Set("Cache-Control", "max-age=3600")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	//check the URL parameter output format, and if json, output the json (json is the default)
	switch outputFormat {
	case "json", "":
		jsonOutput, err := json.MarshalIndent(collections, "", "  ")
		if err != nil {
			http.Error(w, "Unable to process JSON output data", http.StatusInternalServerError)
			log.Default().Printf("Failed to marshal JSON: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonOutput)

	case "ics":
		// Generate an iCalendar file
		prodID := fmt.Sprintf("-//Swindon Borough Council Waste Collections//sbcwaste-%s//EN", appEnv)
		icsContent := "BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:" + prodID + "\n"

		// for each entry in collections
		for _, collection := range collections.Collections {
			var uid string
			var start string
			var end string

			// for each date in CollectionDates
			for _, dateString := range collection.CollectionDates {
				eventDate, _ := time.Parse("2006-01-02", dateString)
				dtStamp := time.Now().UTC().Format("20060102T150405Z")
				uid = foldLine("UID:" + generateUID(collection.Type, dateString, collections.Address))
				start = eventDate.Format("20060102")
				end = eventDate.Add((24 * time.Hour)).Format("20060102")
				summary := foldLine(fmt.Sprintf("SUMMARY:%s", collection.Type))
				location := foldLine(fmt.Sprintf("LOCATION:%s", collections.Address))
				attach := foldLine(fmt.Sprintf("ATTACH;VALUE=URI:%s", collection.IconDataURI))
				urlLine := foldLine(fmt.Sprintf("URL:%s", url))
				icsContent += fmt.Sprintf("BEGIN:VEVENT\r\n%s\r\nDTSTAMP:%s\r\nDTSTART;VALUE=DATE:%s\r\nDTEND;VALUE=DATE:%s\r\n%s\r\n%s\r\n%s\r\nTRANSP:TRANSPARENT\r\n%s\r\nEND:VEVENT\r\n",
					uid, dtStamp, start, end, summary, location, attach, urlLine)
			}
		}
		icsContent += "END:VCALENDAR"
		w.Header().Set("Content-Type", "text/calendar")
		w.Write([]byte(icsContent))
	}
}
