package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestFetchBillingData_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	// Use a date far in the past to try and get some data.
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	data, err := fetchBillingData(context.Background(), startDate)

	if err != nil {
		t.Fatalf("fetchBillingData returned an error: %v", err)
	}

	// We can't guarantee data, but we can log what we find.
	// The primary goal is to ensure the connection and query don't fail.
	t.Logf("Successfully fetched %d records from BigQuery.", len(data))
}
