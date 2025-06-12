package bq

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// VerifySchema checks the schema of the specified BigQuery table against a map of required fields.
// It prints the schema, checks for missing required fields, and prints a sample row.
func VerifySchema(ctx context.Context, client *bigquery.Client, projectID, datasetID, tableID string, requiredFields map[string]bool) error {
	logging.Info("Verifying schema for table: %s.%s.%s", projectID, datasetID, tableID)
	// Get table metadata
	table := client.Dataset(datasetID).Table(tableID)
	metadata, err := table.Metadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to get table metadata: %w", err)
	}

	// Print schema information
	fmt.Printf("Table: %s.%s.%s\n", projectID, datasetID, tableID)
	fmt.Println("\nSchema:")
	for _, field := range metadata.Schema {
		printField(field, 0)
	}

	// Check for required fields
	for _, field := range metadata.Schema {
		checkRequiredField(field, requiredFields)
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

	// Print sample data (first row)
	fmt.Println("\nSample data (first row):")
	it := table.Read(ctx)
	var row map[string]bigquery.Value
	if err := it.Next(&row); err != nil && err != iterator.Done {
		return fmt.Errorf("failed to read row: %w", err)
	}
	if err == nil {
		for key, value := range row {
			fmt.Printf("%s: %v\n", key, value)
		}
	}
	return nil
}

// printField prints a BigQuery field and its nested fields (if any) with indentation.
func printField(field *bigquery.FieldSchema, indent int) {
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}
	fmt.Printf("%s%s (%s)", indentStr, field.Name, field.Type)
	if field.Repeated {
		fmt.Print(" [REPEATED]")
	}
	fmt.Println()
	for _, f := range field.Schema {
		printField(f, indent+1)
	}
}

// checkRequiredField marks a field as found if it matches a required field name.
func checkRequiredField(field *bigquery.FieldSchema, requiredFields map[string]bool) {
	if found, ok := requiredFields[field.Name]; ok && !found {
		requiredFields[field.Name] = true
	}
	for _, f := range field.Schema {
		checkRequiredField(f, requiredFields)
	}
}
