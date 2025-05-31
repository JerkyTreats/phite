// Package dbutil provides shared utilities for interacting with DuckDB databases.
// It includes connection management, schema validation, and standardized error handling.
package dbutil

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
)

// OpenDuckDB opens a connection to a DuckDB database at the specified path.
// The caller is responsible for closing the connection.
// OpenDuckDB opens a connection to a DuckDB database at the specified path.
// The caller is responsible for closing the database.
func OpenDuckDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

// WithConnection opens a DuckDB connection, runs the provided function with it,
// and ensures the connection is properly closed afterward.
// WithConnection opens a DuckDB connection, runs the provided function with it,
// and ensures the connection is properly closed afterward.
func WithConnection(dbPath string, fn func(*sql.DB) error) error {
	db, err := OpenDuckDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	return fn(db)
}

// ValidateTable checks if the specified table exists and contains all required columns.
// ValidateTable checks if the specified table exists and contains all required columns.
func ValidateTable(db *sql.DB, tableName string, requiredColumns []string) error {
	// Check if table exists
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)`, tableName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking if table exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("table %q does not exist", tableName)
	}
	// Get actual columns from the table
	rows, err := db.Query(
		`SELECT column_name FROM information_schema.columns WHERE table_name = ?`, tableName)
	if err != nil {
		return fmt.Errorf("error querying table columns: %w", err)
	}
	defer rows.Close()
	// Create a set of existing columns
	existingColumns := make(map[string]bool)
	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			return fmt.Errorf("error scanning column name: %w", err)
		}
		existingColumns[colName] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating columns: %w", err)
	}
	// Check for missing columns
	var missing []string
	for _, col := range requiredColumns {
		if !existingColumns[col] {
			missing = append(missing, col)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required columns in table %q: %v", tableName, missing)
	}
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
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	var success bool
	defer func() {
		if !success {
			tx.Rollback()
		}
	}()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	success = true
	return nil
}

// CloseDB closes a database and logs any errors. Safe for defer.
func CloseDB(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
			// Optionally log error here
		}
	}
}
