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
	// Register required configuration keys
	config.RegisterRequiredKey("db.type")
	config.RegisterRequiredKey("db.path")       // For DuckDB fallback
	config.RegisterRequiredKey("db.project_id") // For BigQuery fallback

	// BigQuery-specific keys used as fallbacks
	config.RegisterRequiredKey("bigquery.dataset_id") // For BigQuery dataset fallback
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
	// Use params if provided, otherwise fall back to config
	path := params["path"]
	if path == "" {
		path = config.GetString("db.path")
	}
	if path == "" {
		return nil, fmt.Errorf("path is required for DuckDB repository")
	}
	db, err := duckdb.OpenDB(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	return duckdb.NewRepository(db), nil
}

// newBQRepository creates a new BigQuery repository instance
func newBQRepository(ctx context.Context, params map[string]string) (dbinterface.Repository, error) {
	// Use params if provided, otherwise fall back to config
	projectID := params["project_id"]
	if projectID == "" {
		projectID = config.GetString("db.project_id")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required for BigQuery repository")
	}

	// Extract other BQ-specific parameters
	datasetID := params["dataset_id"]
	if datasetID == "" {
		datasetID = config.GetString("bigquery.dataset_id")
	}

	billingProject := params["billing_project"]
	if billingProject == "" {
		billingProject = projectID // Default to same project
	}

	return bq.NewRepository(projectID, datasetID, billingProject)
}
