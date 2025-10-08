package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	results, err := searchAddress(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search for address: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func searchAddress(query string) ([]AddressSearchResult, error) {
	url := fmt.Sprintf("https://maps.swindon.gov.uk/getdata.aspx?callback=my_callback&type=jsonp&service=LocationSearch&RequestType=LocationSearch&location=%s&pagesize=100&startnum=1&gettotals=false&axuid=1&mapsource=mapsources/MyHouse", query)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`\(([\s\S]*?)\);?$`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to extract JSON from JSONP response")
	}
	jsonString := string(matches[1])

	var addressResponse AddressResponse
	err = json.Unmarshal([]byte(jsonString), &addressResponse)
	if err != nil {
		return nil, err
	}

	var results []AddressSearchResult
	for _, data := range addressResponse.Data {
		if len(data) > 2 {
			results = append(results, AddressSearchResult{
				Address: data[2],
				UPRN:    data[0],
			})
		}
	}

	return results, nil
}