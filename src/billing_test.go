package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestBillingHandler(t *testing.T) {
	// Mock the fetchAndMergeBillingData function
	originalFetchAndMergeBillingData := fetchAndMergeBillingData
	fetchAndMergeBillingData = func(ctx context.Context) ([]CostData, error) {
		return []CostData{
			{YearMonth: "2023-03", TotalCost: 15.75},
			{YearMonth: "2023-02", TotalCost: 12.00},
			{YearMonth: "2023-01", TotalCost: 10.50},
		}, nil
	}
	defer func() { fetchAndMergeBillingData = originalFetchAndMergeBillingData }()

	// Reset the cache
	billingCache.cacheValid = false

	req, err := http.NewRequest("GET", "/api/costs", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(BillingHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var costs []CostData
	if err := json.NewDecoder(rr.Body).Decode(&costs); err != nil {
		t.Errorf("handler returned invalid JSON: %v", err)
	}

	if len(costs) != 3 {
		t.Errorf("expected 3 cost items, got %d", len(costs))
	}
}

func TestBillingHandler_FetchError(t *testing.T) {
	// Mock the fetchAndMergeBillingData function to return an error
	originalFetchAndMergeBillingData := fetchAndMergeBillingData
	fetchAndMergeBillingData = func(ctx context.Context) ([]CostData, error) {
		return nil, fmt.Errorf("mock fetch error")
	}
	defer func() { fetchAndMergeBillingData = originalFetchAndMergeBillingData }()

	// Reset the cache
	billingCache.cacheValid = false

	req, err := http.NewRequest("GET", "/api/costs", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(BillingHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestMergeCostData_CSVTakesPriorityOverBigQuery(t *testing.T) {
	csvData := []CostData{
		{YearMonth: "2024-07", TotalCost: 0.09, Note: "Experimenting"},
		{YearMonth: "2024-08", TotalCost: 0.31},
	}
	// BigQuery returns a conflicting value for a month the CSV already has, plus
	// the current, not-yet-finalized month the CSV doesn't have.
	bqData := []CostData{
		{YearMonth: "2024-08", TotalCost: 999.99},
		{YearMonth: "2024-09", TotalCost: 1.23},
	}

	got := mergeCostData(csvData, bqData)

	want := []CostData{
		{YearMonth: "2024-09", TotalCost: 1.23},
		{YearMonth: "2024-08", TotalCost: 0.31},
		{YearMonth: "2024-07", TotalCost: 0.09, Note: "Experimenting"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeCostData() = %+v, want %+v", got, want)
	}
}

func TestFetchBillingDataFromCSV_ParsesRaggedRowsAndNotes(t *testing.T) {
	// billingCSVFile is relative to the repo root (where the app runs from), but
	// `go test` sets the working directory to this package's own directory.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() failed: %v", err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("os.Chdir(\"..\") failed: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	results, err := fetchBillingDataFromCSV()
	if err != nil {
		t.Fatalf("fetchBillingDataFromCSV() returned error: %v", err)
	}

	byMonth := make(map[string]CostData)
	for _, item := range results {
		byMonth[item.YearMonth] = item
	}

	// A row with no trailing note field at all (e.g. "2024-08,0.31").
	if got, ok := byMonth["2024-08"]; !ok || got.TotalCost != 0.31 || got.Note != "" {
		t.Errorf("2024-08 = %+v, want {TotalCost: 0.31, Note: \"\"}", got)
	}

	// A row with a note.
	if got, ok := byMonth["2024-11"]; !ok || got.TotalCost != 6.56 || got.Note != "Initial release" {
		t.Errorf("2024-11 = %+v, want {TotalCost: 6.56, Note: \"Initial release\"}", got)
	}
}
