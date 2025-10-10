package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const (
	projectID = "sbcwaste"
	datasetID = "bindays"
)

// CostData represents the monthly cost data.
type CostData struct {
	YearMonth string  `json:"year_month"`
	TotalCost float64 `json:"total_cost"`
}

// billingCache holds the cached billing data and its expiry time.
var billingCache struct {
	data       []CostData
	lastFetch  time.Time
	cacheValid bool
}

// BillingHandler handles requests to the /api/costs endpoint.
func BillingHandler(w http.ResponseWriter, r *http.Request) {
	// Use a 24-hour cache to avoid hitting BigQuery on every request.
	if !billingCache.cacheValid || time.Since(billingCache.lastFetch) > 24*time.Hour {
		log.Println("Billing cache expired or invalid, fetching new data from BigQuery...")
		data, err := fetchBillingData(r.Context())
		if err != nil {
			log.Printf("ERROR: Failed to fetch billing data: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch billing data: %v", err), http.StatusInternalServerError)
			return
		}
		billingCache.data = data
		billingCache.lastFetch = time.Now()
		billingCache.cacheValid = true
		log.Println("Successfully fetched and cached new billing data.")
	} else {
		log.Println("Serving billing data from cache.")
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(billingCache.data); err != nil {
		http.Error(w, "Failed to encode billing data to JSON", http.StatusInternalServerError)
	}
}

// fetchBillingData queries BigQuery to get the monthly cost data.
func fetchBillingData(ctx context.Context) ([]CostData, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	// First, find the billing table name dynamically.
	tableName, err := findBillingTableName(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("could not find billing table: %v", err)
	}

	// Now, query the billing table for monthly costs.
	queryStr := fmt.Sprintf(`
		SELECT
		  FORMAT_DATE("%%Y-%%m", usage_start_time) AS year_month,
		  ROUND(SUM(cost), 2) AS total_cost
		FROM `+"`%s.%s.%s`"+`
		GROUP BY 1
		ORDER BY 1 DESC
	`, projectID, datasetID, tableName)

	q := client.Query(queryStr)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("query.Read: %v", err)
	}

	var results []CostData
	for {
		var row CostData
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterator.Next: %v", err)
		}
		results = append(results, row)
	}

	return results, nil
}

// findBillingTableName looks up the full name of the billing export table.
func findBillingTableName(ctx context.Context, client *bigquery.Client) (string, error) {
	queryStr := fmt.Sprintf(`
		SELECT table_name
		FROM `+"`%s.%s.INFORMATION_SCHEMA.TABLES`"+`
		WHERE STARTS_WITH(table_name, 'gcp_billing_export_v1_')
		LIMIT 1
	`, projectID, datasetID)

	q := client.Query(queryStr)
	it, err := q.Read(ctx)
	if err != nil {
		return "", fmt.Errorf("query.Read for table name: %v", err)
	}

	var row struct {
		TableName string `bigquery:"table_name"`
	}
	err = it.Next(&row)
	if err == iterator.Done {
		return "", fmt.Errorf("no billing export table found in dataset %s", datasetID)
	}
	if err != nil {
		return "", fmt.Errorf("iterator.Next for table name: %v", err)
	}

	log.Printf("Found billing table: %s", row.TableName)
	return row.TableName, nil
}