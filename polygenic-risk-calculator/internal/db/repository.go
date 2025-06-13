package db

import (
	"context"
	"fmt"

	"phite.io/polygenic-risk-calculator/internal/db/duckdb"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
)

// RepositoryConstructor is a function type for creating new repository instances
type RepositoryConstructor func(ctx context.Context, config map[string]string) (dbinterface.Repository, error)

var constructors = map[string]RepositoryConstructor{
	"duckdb": newDuckDBRepository,
}

// GetRepository creates a new repository instance of the specified type
func GetRepository(ctx context.Context, dbType string, config map[string]string) (dbinterface.Repository, error) {
	if constructor, ok := constructors[dbType]; ok {
		return constructor(ctx, config)
	}
	return nil, fmt.Errorf("unsupported database type: %s", dbType)
}

// newDuckDBRepository creates a new DuckDB repository instance
func newDuckDBRepository(ctx context.Context, config map[string]string) (dbinterface.Repository, error) {
	path, ok := config["path"]
	if !ok {
		return nil, fmt.Errorf("path is required for DuckDB repository")
	}
	db, err := duckdb.OpenDB(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	return duckdb.NewRepository(db), nil
}
