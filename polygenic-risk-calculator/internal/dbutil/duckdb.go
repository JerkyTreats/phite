// Package dbutil provides shared utilities for interacting with DuckDB databases.
// It includes connection management, schema validation, and standardized error handling.
package dbutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"phite.io/polygenic-risk-calculator/internal/logging"

	_ "github.com/marcboeker/go-duckdb"
)

// OpenDuckDB opens a connection to a DuckDB database at the specified path.
// The caller is responsible for closing the connection.
// OpenDuckDB opens a connection to a DuckDB database at the specified path.
// The caller is responsible for closing the database.
func OpenDuckDB(dbPath string) (*sql.DB, error) {
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
// WithConnection opens a DuckDB connection, runs the provided function with it,
// and ensures the connection is properly closed afterward.
func WithConnection(dbPath string, fn func(*sql.DB) error) error {
	db, err := OpenDuckDB(dbPath)
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

// IsTableEmpty checks if the specified table is empty.
// IsTableEmpty checks if the specified table is empty.
func IsTableEmpty(db *sql.DB, tableName string) (bool, error) {
	var count int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, tableName)
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error counting rows in table %q: %w", tableName, err)
	}
	return count == 0, nil
}

// TableExists checks if a table exists in the database.
func TableExists(db *sql.DB, tableName string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)`, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if table exists: %w", err)
	}
	return exists, nil
}

// ExecInTransaction executes a function within a transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func ExecInTransaction(db *sql.DB, fn func(tx *sql.Tx) error) error {
	logging.Info("Beginning transaction")
	tx, err := db.Begin()
	if err != nil {
		logging.Error("error beginning transaction: %v", err)
		return err
	}
	var success bool
	defer func() {
		if !success {
			logging.Error("transaction rollback due to error")
			tx.Rollback()
		}
	}()
	if err := fn(tx); err != nil {
		logging.Error("error in transaction function: %v", err)
		return err
	}
	if err := tx.Commit(); err != nil {
		logging.Error("error committing transaction: %v", err)
		return err
	}
	success = true
	logging.Info("Transaction committed successfully")
	return nil
}

// RowScanner is a generic function type that defines how to scan a single sql.Row
// into a pointer to an instance of type T.
// It's used by ExecuteDuckDBQuery to map query results to specific structs.
type RowScanner[T any] func(rows *sql.Rows) (*T, error)

// ExecuteDuckDBQuery executes a given SQL query with the provided arguments
// and uses the supplied RowScanner to map the results to a slice of *T.
// It handles context cancellation and ensures rows are closed.
func ExecuteDuckDBQuery[T any](ctx context.Context, db *sql.DB, query string, scanner RowScanner[T], args ...any) ([]*T, error) {
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
			// Decide if we should continue or return error immediately.
			// For now, let's return immediately to avoid partial results that might be misleading.
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

// CloseDB closes a database and logs any errors. Safe for defer.
func CloseDB(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
			logging.Error("error closing database: %v", err)
		}
		logging.Info("Database connection closed")
	}
}
