package bq

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/spf13/viper"
	"google.golang.org/api/iterator"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// Repository implements the dbinterface.Repository interface for BigQuery
type Repository struct {
	client *bigquery.Client
	config *viper.Viper
}

// NewRepository creates a new BigQuery repository
func NewRepository(client *bigquery.Client, config *viper.Viper) dbinterface.Repository {
	return &Repository{
		client: client,
		config: config,
	}
}

// Query executes a SQL query and returns the results as a slice of maps
func (r *Repository) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	logging.Debug("Executing BigQuery query: %s", query)

	it, err := r.client.Query(query).Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var results []map[string]interface{}
	for {
		var row map[string]bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch row: %w", err)
		}

		// Convert BigQuery values to standard Go types
		convertedRow := make(map[string]interface{})
		for k, v := range row {
			convertedRow[k] = v
		}
		results = append(results, convertedRow)
	}

	logging.Debug("Successfully executed query and scanned %d rows for query: %s", len(results), query)
	return results, nil
}

// Insert inserts multiple rows into a table
func (r *Repository) Insert(ctx context.Context, table string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	logging.Debug("Inserting %d rows into table %s", len(rows), table)

	// Get dataset and table references
	datasetID := r.config.GetString("bigquery.dataset_id")
	tableRef := r.client.Dataset(datasetID).Table(table)
	inserter := tableRef.Inserter()

	// Convert rows to BigQuery values
	var bqRows []map[string]bigquery.Value
	for _, row := range rows {
		bqRow := make(map[string]bigquery.Value)
		for k, v := range row {
			bqRow[k] = bigquery.Value(v)
		}
		bqRows = append(bqRows, bqRow)
	}

	if err := inserter.Put(ctx, bqRows); err != nil {
		return fmt.Errorf("failed to insert rows: %w", err)
	}

	logging.Debug("Successfully inserted %d rows into table %s", len(rows), table)
	return nil
}

// TestConnection tests the database connection and validates the given table
func (r *Repository) TestConnection(ctx context.Context, table string) error {
	return r.ValidateTable(ctx, table, nil)
}

// ValidateTable validates that a table exists and has the required columns
func (r *Repository) ValidateTable(ctx context.Context, table string, requiredColumns []string) error {
	logging.Info("Validating table %q for required columns", table)

	// Get dataset and table references
	datasetID := r.config.GetString("bigquery.dataset_id")
	tableRef := r.client.Dataset(datasetID).Table(table)

	// Check if table exists
	_, err := tableRef.Metadata(ctx)
	if err != nil {
		return fmt.Errorf("table %q does not exist: %w", table, err)
	}

	// If no required columns, we're done
	if len(requiredColumns) == 0 {
		logging.Info("Table %q validation passed", table)
		return nil
	}

	// Get table metadata
	metadata, err := tableRef.Metadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to get table metadata: %w", err)
	}

	// Map of existing columns
	existingColumns := make(map[string]bool)
	for _, field := range metadata.Schema {
		existingColumns[field.Name] = true
	}

	// Check required columns
	var missingColumns []string
	for _, col := range requiredColumns {
		if !existingColumns[col] {
			missingColumns = append(missingColumns, col)
		}
	}

	if len(missingColumns) > 0 {
		return fmt.Errorf("table %q is missing required columns: %v", table, missingColumns)
	}

	logging.Info("Table %q validation passed", table)
	return nil
}
