package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const (
	projectID      = "sbcwaste"
	datasetID      = "bindays"
	billingCSVFile = "static/data/billing_data.csv"
)

// CostData represents the monthly cost data.
type CostData struct {
	YearMonth string  `json:"year_month" bigquery:"year_month"`
	TotalCost float64 `json:"total_cost" bigquery:"total_cost"`
	Note      string  `json:"note,omitempty"`
}

// billingCache holds the cached billing data and its expiry time.
// It's guarded by billingCacheMu since BillingHandler can be called concurrently.
var billingCacheMu sync.Mutex
var billingCache struct {
	data       []CostData
	lastFetch  time.Time
	cacheValid bool
}

// BillingHandler handles requests to the /api/costs endpoint.
func BillingHandler(w http.ResponseWriter, r *http.Request) {
	billingCacheMu.Lock()
	defer billingCacheMu.Unlock()

	// Use a 24-hour cache.
	if !billingCache.cacheValid || time.Since(billingCache.lastFetch) > 24*time.Hour {
		log.Println("Billing cache expired or invalid, fetching and merging new data...")
		data, err := fetchAndMergeBillingData(r.Context())
		if err != nil {
			log.Printf("ERROR: Failed to fetch and merge billing data: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch billing data: %v", err), http.StatusInternalServerError)
			return
		}
		billingCache.data = data
		billingCache.lastFetch = time.Now()
		billingCache.cacheValid = true
		log.Println("Successfully fetched, merged, and cached new billing data.")
	} else {
		log.Println("Serving billing data from cache.")
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(billingCache.data); err != nil {
		http.Error(w, "Failed to encode billing data to JSON", http.StatusInternalServerError)
	}
}

// fetchAndMergeBillingData fetches historical data from CSV and current data from BigQuery,
// then merges them into a single, sorted list. The CSV is the durable source of truth: once
// a month has been recorded there (typically by the CSV-update workflow, after the month has
// closed), BigQuery is no longer consulted for it. BigQuery only fills in months the CSV
// doesn't have yet - in practice the current, still-open month.
var fetchAndMergeBillingData = func(ctx context.Context) ([]CostData, error) {
	csvData, err := fetchBillingDataFromCSV()
	if err != nil {
		// Log the error but continue, as we might still get data from BigQuery.
		log.Printf("WARN: Could not fetch billing data from CSV, proceeding with BigQuery only: %v", err)
	}

	bqData, err := fetchBillingData(ctx)
	if err != nil {
		// If CSV data is also empty, this is a fatal error.
		if len(csvData) == 0 {
			return nil, fmt.Errorf("failed to fetch billing data from BigQuery and no CSV data available: %w", err)
		}
		// Otherwise, log the error and proceed with just the CSV data.
		log.Printf("WARN: Could not fetch billing data from BigQuery, proceeding with CSV data only: %v", err)
	}

	return mergeCostData(csvData, bqData), nil
}

// mergeCostData merges CSV and BigQuery cost data into a single list sorted by month,
// descending. The CSV always wins on a month present in both, since it's the durable,
// version-controlled record; BigQuery only supplies months the CSV hasn't recorded yet.
func mergeCostData(csvData, bqData []CostData) []CostData {
	mergedData := make(map[string]CostData)
	for _, item := range bqData {
		mergedData[item.YearMonth] = item
	}
	for _, item := range csvData {
		mergedData[item.YearMonth] = item
	}

	results := make([]CostData, 0, len(mergedData))
	for _, item := range mergedData {
		results = append(results, item)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].YearMonth > results[j].YearMonth
	})

	return results
}

// fetchBillingDataFromCSV reads billing data from the local CSV file.
// If the file does not exist, it returns an empty slice and no error.
func fetchBillingDataFromCSV() ([]CostData, error) {
	file, err := os.Open(billingCSVFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("No '%s' file found, skipping CSV data load.", billingCSVFile)
			return nil, nil // No file is not an error in this case.
		}
		return nil, fmt.Errorf("could not open %s: %w", billingCSVFile, err)
	}
	defer file.Close()

	log.Printf("Reading historical data from '%s'.", billingCSVFile)
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("could not read header from csv: %w", err)
	}

	// The note column is optional, and rows may omit trailing empty fields entirely
	// (e.g. "2024-08,0.31" with no third field), so ragged rows must be allowed.
	reader.FieldsPerRecord = -1

	yearMonthIndex, totalCostIndex, noteIndex := -1, -1, -1
	for i, colName := range header {
		switch colName {
		case "year_month":
			yearMonthIndex = i
		case "total_cost":
			totalCostIndex = i
		case "note":
			noteIndex = i
		}
	}

	if yearMonthIndex == -1 || totalCostIndex == -1 {
		return nil, fmt.Errorf("csv file must contain 'year_month' and 'total_cost' columns")
	}

	var results []CostData
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading csv record: %w", err)
		}

		cost, err := strconv.ParseFloat(record[totalCostIndex], 64)
		if err != nil {
			log.Printf("Skipping record with invalid cost: %s", strings.ReplaceAll(fmt.Sprintf("%v", record), "\n", "")) // #nosec G706 -- CSV record printed for diagnostics, newlines stripped
			continue
		}

		item := CostData{
			YearMonth: record[yearMonthIndex],
			TotalCost: cost,
		}
		if noteIndex != -1 && noteIndex < len(record) {
			item.Note = record[noteIndex]
		}
		results = append(results, item)
	}
	return results, nil
}

// fetchBillingData queries BigQuery for all available monthly cost data. It is not date-filtered:
// the CSV is the priority source (see mergeCostData), so BigQuery is only ever used for months
// the CSV doesn't have yet, and the billing export dataset for a small project is cheap to scan in full.
func fetchBillingData(ctx context.Context) ([]CostData, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	tableName, err := findBillingTableName(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("could not find billing table: %v", err)
	}

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
	log.Printf("Fetched %d records from BigQuery.", len(results))
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
		log.Printf("WARN: No billing export table found in dataset %s. Costs will not be fetched from BigQuery.", datasetID)
		// Return a fake table name to prevent the query from failing,
		// but this will result in zero rows.
		return "gcp_billing_export_v1_XXXXXXXXXXXXXXXX", nil
	}
	if err != nil {
		return "", fmt.Errorf("iterator.Next for table name: %v", err)
	}

	log.Printf("Found billing table: %s", row.TableName)
	return row.TableName, nil
}
