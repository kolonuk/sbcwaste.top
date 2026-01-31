package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

type AddressSearchResult struct {
	Address string `json:"address"`
	UPRN    string `json:"uprn"`
}

func SearchAddressHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	insecure := r.URL.Query().Get("insecure") == "true"
	var client *http.Client
	if insecure {
		client = InsecureHTTPClient
	} else {
		client = HTTPClient
	}

	results, err := searchAddress(client, query)
	if err != nil {
		log.Printf("Failed to search for address: %v", err)
		http.Error(w, "The address lookup service is temporarily unavailable. Please try again later.", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode address search results", http.StatusInternalServerError)
	}
}

func searchAddress(client *http.Client, query string) ([]AddressSearchResult, error) {
	escapedQuery := url.QueryEscape(query)
	url := fmt.Sprintf("https://maps.swindon.gov.uk/getdata.aspx?callback=my_callback&type=jsonp&service=LocationSearch&RequestType=LocationSearch&location=%s&pagesize=100&startnum=1&gettotals=false&axuid=1&mapsource=mapsources/MyHouse", escapedQuery)

	addressResponse, err := fetchAddressData(client, url)
	if err != nil {
		return nil, err
	}

	var results []AddressSearchResult
	// This regex will match any HTML tags.
	re := regexp.MustCompile("<(.|\n)*?>")

	for _, data := range addressResponse.Data {
		if len(data) > 2 {
			// Strip HTML tags from the address.
			address := re.ReplaceAllString(data[2], "")
			results = append(results, AddressSearchResult{
				Address: address,
				UPRN:    data[0],
			})
		}
	}

	return results, nil
}