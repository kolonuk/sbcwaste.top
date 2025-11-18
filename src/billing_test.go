package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
