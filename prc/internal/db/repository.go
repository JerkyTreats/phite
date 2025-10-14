package db

import (
	"context"
	"fmt"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/db/bq"
	"phite.io/polygenic-risk-calculator/internal/db/duckdb"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
)

func init() {
	// Register required infrastructure configuration keys for repository creation
	config.RegisterRequiredKey(config.GCPDataProjectKey)        // For BigQuery data project
	config.RegisterRequiredKey(config.GCPBillingProjectKey)     // For BigQuery billing project
	config.RegisterRequiredKey(config.BigQueryGnomadDatasetKey) // For BigQuery dataset fallback

	// Note: DuckDB fallback path would be domain-specific if implemented
}

// RepositoryConstructor is a function type for creating new repository instances
type RepositoryConstructor func(ctx context.Context, params map[string]string) (dbinterface.Repository, error)

var constructors = map[string]RepositoryConstructor{
	"duckdb": newDuckDBRepository,
	"bq":     newBQRepository,
}

// GetRepository creates a repository instance of the specified type with optional parameters
func GetRepository(ctx context.Context, dbType string, params ...map[string]string) (dbinterface.Repository, error) {
	if dbType == "" {
		return nil, fmt.Errorf("database type cannot be empty")
	}

	constructor, exists := constructors[dbType]
	if !exists {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Use first non-nil parameter map, or empty map if none provided
	var paramMap map[string]string
	for _, p := range params {
		if p != nil {
			paramMap = p
			break
		}
	}
	if paramMap == nil {
		paramMap = make(map[string]string)
	}

	return constructor(ctx, paramMap)
}

// newDuckDBRepository creates a new DuckDB repository instance
func newDuckDBRepository(ctx context.Context, params map[string]string) (dbinterface.Repository, error) {
	// DuckDB requires explicit path parameter (no fallback config)
	path := params["path"]
	if path == "" {
		return nil, fmt.Errorf("path parameter is required for DuckDB repository")
	}
	db, err := duckdb.OpenDB(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	return duckdb.NewRepository(db), nil
}

// newBQRepository creates a new BigQuery repository instance
func newBQRepository(ctx context.Context, params map[string]string) (dbinterface.Repository, error) {
	// Use params if provided, otherwise fall back to infrastructure config
	dataProjectID := params["project_id"]
	if dataProjectID == "" {
		dataProjectID = config.GetString(config.GCPDataProjectKey)
	}
	if dataProjectID == "" {
		return nil, fmt.Errorf("data project_id is required for BigQuery repository")
	}

	// Extract BigQuery dataset
	datasetID := params["dataset_id"]
	if datasetID == "" {
		datasetID = config.GetString(config.BigQueryGnomadDatasetKey)
	}

	// Extract billing project (required for public dataset queries)
	billingProject := params["billing_project"]
	if billingProject == "" {
		billingProject = config.GetString(config.GCPBillingProjectKey)
	}
	if billingProject == "" {
		return nil, fmt.Errorf("billing project is required for BigQuery repository")
	}

	return bq.NewRepository(dataProjectID, datasetID, billingProject)
}
