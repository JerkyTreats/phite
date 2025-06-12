package jerkytreats

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"phite.io/polygenic-risk-calculator/scripts/bigquery/internal/utils"
)

const (
	jerkytreatsProjectID = "jerkytreats"
	jerkytreatsDatasetID = "prs_stats"
	jerkytreatsTableID   = "prs_reference_stats"
)

func VerifySchema() {
	ctx := context.Background()

	// Create BigQuery client
	client, err := bigquery.NewClient(ctx, jerkytreatsProjectID)
	if err != nil {
		log.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer client.Close()

	// Get table metadata
	table := client.Dataset(jerkytreatsDatasetID).Table(jerkytreatsTableID)
	metadata, err := table.Metadata(ctx)
	if err != nil {
		log.Fatalf("Failed to get table metadata: %v", err)
	}

	// Print schema information
	fmt.Printf("Table: %s.%s.%s\n", jerkytreatsProjectID, jerkytreatsDatasetID, jerkytreatsTableID)
	fmt.Println("\nSchema:")
	for _, field := range metadata.Schema {
		utils.PrintField(field, 0)
	}

	// Verify required fields
	requiredFields := map[string]bool{
		"ancestry":     false,
		"trait":        false,
		"model_id":     false,
		"mean_prs":     false,
		"stddev_prs":   false,
		"min_prs":      false,
		"max_prs":      false,
		"sample_size":  false,
		"source":       false,
		"notes":        false,
		"last_updated": false,
	}

	// Check for required fields
	for _, field := range metadata.Schema {
		utils.CheckRequiredField(field, requiredFields)
	}

	// Print missing required fields
	fmt.Println("\nMissing required fields:")
	missing := false
	for field, found := range requiredFields {
		if !found {
			fmt.Printf("- %s\n", field)
			missing = true
		}
	}
	if !missing {
		fmt.Println("None - All required fields are present!")
	}

	// Print sample data
	fmt.Println("\nSample data (first row):")
	it := table.Read(ctx)

	var row map[string]bigquery.Value
	if err := it.Next(&row); err != nil && err != iterator.Done {
		log.Fatalf("Failed to read row: %v", err)
	}

	if err == nil {
		for key, value := range row {
			fmt.Printf("%s: %v\n", key, value)
		}
	}
}
