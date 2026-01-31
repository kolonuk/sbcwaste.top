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
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const (
	projectID      = "sbcwaste"
	datasetID      = "bindays"
	billingCSVFile = "billing_data.csv"
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
// then merges them into a single, sorted list.
var fetchAndMergeBillingData = func(ctx context.Context) ([]CostData, error) {
	// 1. Fetch historical data from CSV.
	csvData, err := fetchBillingDataFromCSV()
	if err != nil {
		// Log the error but continue, as we might still get data from BigQuery.
		log.Printf("WARN: Could not fetch billing data from CSV, proceeding with BigQuery only: %v", err)
	}

	// 2. Determine the start date for the BigQuery query.
	var startDate time.Time
	if len(csvData) > 0 {
		// Find the latest month in the CSV data.
		latestMonth := ""
		for _, item := range csvData {
			if item.YearMonth > latestMonth {
				latestMonth = item.YearMonth
			}
		}

		// Start the BigQuery query from the month *after* the latest one in the CSV.
		latestDate, err := time.Parse("2006-01", latestMonth)
		if err != nil {
			return nil, fmt.Errorf("could not parse latest month from CSV: %w", err)
		}
		startDate = latestDate.AddDate(0, 1, 0)
	} else {
		// If there's no CSV data, fetch everything from the beginning of the project.
		// Replace with a more appropriate start date if needed.
		startDate = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	// 3. Fetch current data from BigQuery from the calculated start date onwards.
	bqData, err := fetchBillingData(ctx, startDate)
	if err != nil {
		// If CSV data is also empty, this is a fatal error.
		if len(csvData) == 0 {
			return nil, fmt.Errorf("failed to fetch billing data from BigQuery and no CSV data available: %w", err)
		}
		// Otherwise, log the error and proceed with just the CSV data.
		log.Printf("WARN: Could not fetch billing data from BigQuery, proceeding with CSV data only: %v", err)
	}

	// 4. Merge the two datasets.
	// Use a map to handle overwrites, ensuring BigQuery data for a given month replaces CSV data.
	mergedData := make(map[string]CostData)
	for _, item := range csvData {
		mergedData[item.YearMonth] = item
	}
	for _, item := range bqData {
		mergedData[item.YearMonth] = item
	}

	// 5. Convert map back to a slice.
	var results []CostData
	for _, item := range mergedData {
		results = append(results, item)
	}

	// 6. Sort the results by month in descending order.
	sort.Slice(results, func(i, j int) bool {
		return results[i].YearMonth > results[j].YearMonth
	})

	return results, nil
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

	yearMonthIndex, totalCostIndex := -1, -1
	for i, colName := range header {
		if colName == "year_month" {
			yearMonthIndex = i
		} else if colName == "total_cost" {
			totalCostIndex = i
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
			log.Printf("Skipping record with invalid cost: %v", record)
			continue
		}

		results = append(results, CostData{
			YearMonth: record[yearMonthIndex],
			TotalCost: cost,
		})
	}
	return results, nil
}

// fetchBillingData queries BigQuery for monthly cost data from a specific start date.
func fetchBillingData(ctx context.Context, startDate time.Time) ([]CostData, error) {
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
		WHERE usage_start_time >= @startDate
		GROUP BY 1
		ORDER BY 1 DESC
	`, projectID, datasetID, tableName)

	q := client.Query(queryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "startDate", Value: startDate},
	}

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
	log.Printf("Fetched %d records from BigQuery for dates >= %s.", len(results), startDate.Format("2006-01-02"))
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