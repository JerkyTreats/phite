package bq

import (
	"context"

	"cloud.google.com/go/bigquery"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// Insert operations for BigQuery

// TODO: Consider using dependency injection (e.g., a struct-based approach) for the BigQuery client.
// TODO: Ensure proper context cancellation and timeout handling for long-running insert operations.

// InsertRows inserts a slice of rows into a BigQuery table.
func (c *BQController) InsertRows(ctx context.Context, projectID, datasetID, tableID string, rows []map[string]bigquery.Value) error {
	logging.Info("Inserting %d rows into %s.%s.%s", len(rows), projectID, datasetID, tableID)
	// TODO: Implement row insertion
	inserter := c.client.DatasetInProject(projectID, datasetID).Table(tableID).Inserter()
	if err := inserter.Put(ctx, rows); err != nil {
		logging.Error("Failed to insert rows: %v", err)
		return err
	}
	logging.Info("Successfully inserted %d rows", len(rows))
	return nil
}
