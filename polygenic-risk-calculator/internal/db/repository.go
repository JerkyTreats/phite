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
	config.RegisterRequiredKey("db.path")       // For DuckDB
	config.RegisterRequiredKey("db.project_id") // For BigQuery
}

// RepositoryConstructor is a function type for creating new repository instances
type RepositoryConstructor func(ctx context.Context) (dbinterface.Repository, error)

var constructors = map[string]RepositoryConstructor{
	"duckdb": newDuckDBRepository,
	"bq":     newBQRepository,
}

// GetRepository creates a new repository instance of the specified type
func GetRepository(ctx context.Context, dbType string) (dbinterface.Repository, error) {
	if constructor, ok := constructors[dbType]; ok {
		return constructor(ctx)
	}
	return nil, fmt.Errorf("unsupported database type: %s", dbType)
}

// newDuckDBRepository creates a new DuckDB repository instance
func newDuckDBRepository(ctx context.Context) (dbinterface.Repository, error) {
	path := config.GetString("db.path")
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
func newBQRepository(ctx context.Context) (dbinterface.Repository, error) {
	projectID := config.GetString("db.project_id")
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required for BigQuery repository")
	}
	return bq.NewRepository()
}
