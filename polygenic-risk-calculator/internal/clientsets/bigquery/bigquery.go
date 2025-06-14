// Package bigquery provides a BigQuery clientset for reference stats and other backend needs.
package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/logging"

	dbconfig "phite.io/polygenic-risk-calculator/internal/db/config"
)

// Client encapsulates config, connection, and query logic for reference stats in BigQuery.

func init() {
	config.RegisterRequiredKey("bq_project")         // Project where the data resides
	config.RegisterRequiredKey("bq_billing_project") // Project used for API calls and billing
	config.RegisterRequiredKey("bq_dataset")
	config.RegisterRequiredKey("bq_table")
	// bq_credentials is optional
}

type BQClient struct {
	ProjectID string
	Dataset   string
	Table     string
	CredsPath string
	Client    *bigquery.Client
	Config    *dbconfig.BigQueryConfig
}

// NewClient initializes a BigQuery client using config.go.
func NewClient(ctx context.Context) (*BQClient, error) {
	billingProject := config.GetString("bq_billing_project")
	dataProject := config.GetString("bq_project") // This is the project where the data actually lives
	dataset := config.GetString("bq_dataset")
	table := config.GetString("bq_table")
	creds := config.GetString("bq_credentials")
	// Validation for project, dataset, and table is now handled by config.Validate() in main.

	var client *bigquery.Client
	var err error
	if creds != "" {
		logging.Info("Creating BigQuery client with credentials file '%s' for billing project '%s'", creds, billingProject)
		client, err = bigquery.NewClient(ctx, billingProject, option.WithCredentialsFile(creds))
	} else {
		logging.Info("Creating BigQuery client with default credentials for billing project: %s", billingProject)
		client, err = bigquery.NewClient(ctx, billingProject)
	}
	if err != nil {
		logging.Error("Failed to create BigQuery client for billing project %s: %v", billingProject, err)
		return nil, fmt.Errorf("failed to create BigQuery client for billing project %s: %w", billingProject, err)
	}
	logging.Info("BigQuery client configured for data project=%s, dataset=%s, table=%s (using billing project %s)", dataProject, dataset, table, billingProject)
	return &BQClient{
		ProjectID: dataProject, // This remains the project where the data is located
		Dataset:   dataset,
		Table:     table,
		CredsPath: creds,
		Client:    client,
	}, nil
}

// NewClientWithConfig creates client with specific configuration
func NewClientWithConfig(ctx context.Context, config *dbconfig.BigQueryConfig) (*BQClient, error) {
	var client *bigquery.Client
	var err error

	billingProject := config.BillingProject
	if billingProject == "" {
		billingProject = config.ProjectID
	}

	if config.CredentialsPath != "" {
		client, err = bigquery.NewClient(ctx, billingProject,
			option.WithCredentialsFile(config.CredentialsPath))
	} else {
		client, err = bigquery.NewClient(ctx, billingProject)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create BigQuery client: %w", err)
	}

	return &BQClient{
		ProjectID: config.ProjectID,
		Dataset:   config.DatasetID,
		Client:    client,
		Config:    config,
	}, nil
}

// Close releases resources held by the client.
func (c *BQClient) Close() error {
	return c.Client.Close()
}
