// Package bigquery provides a BigQuery clientset for reference stats and other backend needs.
package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// Client encapsulates config, connection, and query logic for reference stats in BigQuery.

func init() {
	config.RegisterRequiredKey("bq_project")
	config.RegisterRequiredKey("bq_dataset")
	config.RegisterRequiredKey("bq_table")
	// bq_credentials is optional
}

type Client struct {
	ProjectID string
	Dataset   string
	Table     string
	CredsPath string
	bqClient  *bigquery.Client
}

// NewClient initializes a BigQuery client using config.go.
func NewClient(ctx context.Context) (*Client, error) {
	project := config.GetString("bq_project")
	dataset := config.GetString("bq_dataset")
	table := config.GetString("bq_table")
	creds := config.GetString("bq_credentials")
	// Validation for project, dataset, and table is now handled by config.Validate() in main.

	var client *bigquery.Client
	var err error
	if creds != "" {
		logging.Info("Creating BigQuery client with credentials file: %s", creds)
		client, err = bigquery.NewClient(ctx, project, option.WithCredentialsFile(creds))
	} else {
		logging.Info("Creating BigQuery client with default credentials for project: %s", project)
		client, err = bigquery.NewClient(ctx, project)
	}
	if err != nil {
		logging.Error("Failed to create BigQuery client for project %s: %v", project, err)
		return nil, fmt.Errorf("failed to create BigQuery client: %w", err)
	}
	logging.Info("BigQuery client created for project=%s, dataset=%s, table=%s", project, dataset, table)
	return &Client{
		ProjectID: project,
		Dataset:   dataset,
		Table:     table,
		CredsPath: creds,
		bqClient:  client,
	}, nil
}

// BigQuery returns the underlying BigQuery client.
func (c *Client) BigQuery() *bigquery.Client {
	return c.bqClient
}

// Close releases resources held by the client.
func (c *Client) Close() error {
	return c.bqClient.Close()
}
