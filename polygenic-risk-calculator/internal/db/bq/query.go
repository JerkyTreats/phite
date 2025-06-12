package bq

import (
	"context"

	"cloud.google.com/go/bigquery"
	"github.com/spf13/viper"
	"google.golang.org/api/iterator"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// Query operations for BigQuery

// TODO: Consider adding a generic or reflection-based mapper to convert query results into strongly-typed structs.
// TODO: Ensure proper context cancellation and timeout handling for long-running queries.

// BQController encapsulates the BigQuery client and configuration.
type BQController struct {
	client *bigquery.Client
	config *viper.Viper
}

// NewBQController creates a new BQController.
func NewBQController(client *bigquery.Client, config *viper.Viper) *BQController {
	return &BQController{
		client: client,
		config: config,
	}
}

// Query executes a parameterized SQL query and returns a RowIterator.
func (c *BQController) Query(ctx context.Context, queryString string, params []bigquery.QueryParameter) (*bigquery.RowIterator, error) {
	logging.Info("Executing query: %s", queryString)
	// TODO: Implement parameterized query execution
	it, err := c.client.Query(queryString).Read(ctx)
	if err != nil {
		logging.Error("Query execution failed: %v", err)
		return nil, err
	}
	logging.Info("Query executed successfully")
	return it, nil
}

// QueryAll executes a parameterized SQL query and returns all results as a slice of maps.
func (c *BQController) QueryAll(ctx context.Context, queryString string, params []bigquery.QueryParameter) ([]map[string]bigquery.Value, error) {
	logging.Info("Executing query and fetching all results: %s", queryString)
	// TODO: Implement fetching all results
	it, err := c.client.Query(queryString).Read(ctx)
	if err != nil {
		logging.Error("Query execution failed: %v", err)
		return nil, err
	}
	var results []map[string]bigquery.Value
	for {
		var row map[string]bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			logging.Error("Error fetching row: %v", err)
			return nil, err
		}
		results = append(results, row)
	}
	logging.Info("Query executed successfully, fetched %d results", len(results))
	return results, nil
}
