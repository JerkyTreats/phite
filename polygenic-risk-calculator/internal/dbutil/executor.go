// Package dbutil provides database utility functions for working with DuckDB
package dbutil

import (
	"context"
	"fmt"

	"phite.io/polygenic-risk-calculator/internal/logging"
)

// ExecuteDuckDBQueryWithPath provides a generic way to query a DuckDB database by path.
// It's a wrapper around ExecuteDuckDBQuery that handles opening the database connection.
//
// Parameters:
//   - dbPath: The file path to the DuckDB database
//   - query: The SQL query to execute
//   - scanner: A function that maps a database row to a typed result
//   - args: The arguments to pass to the query
//
// Returns:
//   - A slice of typed results
//   - An error if the query fails
func ExecuteDuckDBQueryWithPath[T any](dbPath string, query string, scanner RowScanner[T], args ...interface{}) ([]*T, error) {
	logging.Info("Executing DuckDB query on %s", dbPath)

	// Open database connection
	db, err := OpenDuckDB(dbPath)
	if err != nil {
		logging.Error("Failed to open DuckDB at %s: %v", dbPath, err)
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	defer db.Close()

	// Execute the query with background context
	return ExecuteDuckDBQuery(context.Background(), db, query, scanner, args...)
}
