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
	// BigQuery client uses infrastructure constants for shared GCP resources
	config.RegisterRequiredKey(config.GCPDataProjectKey)        // Project where the data resides
	config.RegisterRequiredKey(config.GCPBillingProjectKey)     // Project used for API calls and billing
	config.RegisterRequiredKey(config.BigQueryGnomadDatasetKey) // Dataset reference
	// Note: Table references are now handled by infrastructure constants (config.Table*)
	// bq_credentials remains optional (no constant needed)
}

type BQClient struct {
	ProjectID string
	Dataset   string
	Table     string
	CredsPath string
	Client    *bigquery.Client
	Config    *dbconfig.BigQueryConfig
}

// NewClient initializes a BigQuery client using infrastructure configuration.
func NewClient(ctx context.Context) (*BQClient, error) {
	billingProject := config.GetString(config.GCPBillingProjectKey)
	dataProject := config.GetString(config.GCPDataProjectKey) // This is the project where the data actually lives
	dataset := config.GetString(config.BigQueryGnomadDatasetKey)
	table := config.GetString(config.TableAlleleFreqTableKey) // Default to allele freq table
	creds := config.GetString("bq_credentials")               // Still using string key as this is optional
	// Validation for project, dataset, and table is now handled by infrastructure constants

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
