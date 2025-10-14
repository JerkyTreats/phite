package dbinterface

import (
	"context"
)

// Repository defines the common interface for database operations.
type Repository interface {
	// Query executes a query and returns results as a slice of maps.
	Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)

	// Insert inserts rows into the specified table.
	Insert(ctx context.Context, table string, rows []map[string]interface{}) error

	// TestConnection checks that the database is reachable and the table exists.
	TestConnection(ctx context.Context, table string) error

	// ValidateTable validates that a table exists and has the required columns
	ValidateTable(ctx context.Context, table string, requiredColumns []string) error
}
