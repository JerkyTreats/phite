package duckdb

import (
	"context"
	"database/sql"
	"fmt"

	"phite.io/polygenic-risk-calculator/internal/logging"
)

// RowScanner is a generic function type that defines how to scan a single sql.Row
// into a pointer to an instance of type T.
type RowScanner[T any] func(rows *sql.Rows) (*T, error)

// ExecuteQuery executes a given SQL query with the provided arguments
// and uses the supplied RowScanner to map the results to a slice of *T.
// It handles context cancellation and ensures rows are closed.
func ExecuteQuery[T any](ctx context.Context, db *sql.DB, query string, scanner RowScanner[T], args ...any) ([]*T, error) {
	logging.Debug("Executing DuckDB query: %s with args: %v", query, args)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		logging.Error("Error executing query '%s': %v", query, err)
		return nil, fmt.Errorf("error executing query '%s': %w", query, err)
	}
	defer rows.Close()

	var results []*T
	for rows.Next() {
		select {
		case <-ctx.Done():
			logging.Warn("Context cancelled during query execution for query: %s", query)
			return nil, ctx.Err()
		default:
		}

		item, err := scanner(rows)
		if err != nil {
			logging.Error("Error scanning row for query '%s': %v", query, err)
			return nil, fmt.Errorf("error scanning row for query '%s': %w", query, err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		logging.Error("Error iterating rows for query '%s': %v", query, err)
		return nil, fmt.Errorf("error iterating rows for query '%s': %w", query, err)
	}

	logging.Debug("Successfully executed query and scanned %d rows for query: %s", len(results), query)
	return results, nil
}

// ExecuteQueryWithPath provides a generic way to query a DuckDB database by path.
// It's a wrapper around ExecuteQuery that handles opening the database connection.
func ExecuteQueryWithPath[T any](dbPath string, query string, scanner RowScanner[T], args ...interface{}) ([]*T, error) {
	logging.Info("Executing DuckDB query on %s", dbPath)

	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		logging.Error("Failed to open DuckDB at %s: %v", dbPath, err)
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	defer db.Close()

	return ExecuteQuery(context.Background(), db, query, scanner, args...)
}
