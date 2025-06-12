package duckdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"phite.io/polygenic-risk-calculator/internal/logging"

	_ "github.com/marcboeker/go-duckdb"
)

// OpenDB opens a connection to a DuckDB database at the specified path.
// The caller is responsible for closing the connection.
func OpenDB(dbPath string) (*sql.DB, error) {
	logging.Info("Opening DuckDB database at %s", dbPath)
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		logging.Error("failed to open DuckDB database: %v", err)
		return nil, err
	}
	logging.Info("DuckDB connection established at %s", dbPath)
	return db, nil
}

// WithConnection opens a DuckDB connection, runs the provided function with it,
// and ensures the connection is properly closed afterward.
func WithConnection(dbPath string, fn func(*sql.DB) error) error {
	db, err := OpenDB(dbPath)
	if err != nil {
		logging.Error("failed to open DuckDB connection: %v", err)
		return err
	}
	defer func() {
		logging.Info("Closing DuckDB connection at %s", dbPath)
		db.Close()
	}()
	return fn(db)
}

// ValidateTable checks if the specified table exists and contains all required columns.
func ValidateTable(db *sql.DB, tableName string, requiredColumns []string) error {
	logging.Info("Validating table %q for required columns", tableName)
	// Check if table exists
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)`, tableName).Scan(&exists)
	if err != nil {
		logging.Error("error checking if table exists: %v", err)
		return err
	}
	if !exists {
		logging.Error("table %q does not exist", tableName)
		return errors.New("table does not exist")
	}
	// Get actual columns from the table
	rows, err := db.Query(
		`SELECT column_name FROM information_schema.columns WHERE table_name = ?`, tableName)
	if err != nil {
		logging.Error("error querying table columns: %v", err)
		return err
	}
	defer rows.Close()
	// Create a set of existing columns
	existingColumns := make(map[string]bool)
	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			logging.Error("error scanning column name: %v", err)
			return err
		}
		existingColumns[colName] = true
	}
	if err := rows.Err(); err != nil {
		logging.Error("error iterating columns: %v", err)
		return err
	}
	// Check for missing columns
	var missing []string
	for _, col := range requiredColumns {
		if !existingColumns[col] {
			missing = append(missing, col)
		}
	}
	if len(missing) > 0 {
		logging.Error("missing required columns in table %q: %v", tableName, missing)
		return errors.New("missing required columns")
	}
	logging.Info("Table %q validation passed", tableName)
	return nil
}

// ExecuteDuckDBQuery executes a query and uses the provided scanner to map results.
func ExecuteDuckDBQuery[T any](ctx context.Context, db *sql.DB, query string, scanner func(*sql.Rows) (*T, error)) ([]*T, error) {
	logging.Debug("Executing DuckDB query: %s", query)

	rows, err := db.QueryContext(ctx, query)
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
