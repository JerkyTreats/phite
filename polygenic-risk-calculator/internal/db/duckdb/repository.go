package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// Repository implements the dbinterface.Repository interface for DuckDB
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new DuckDB repository
func NewRepository(db *sql.DB) dbinterface.Repository {
	return &Repository{db: db}
}

// Query executes a SQL query and returns the results as a slice of maps
func (r *Repository) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	logging.Debug("Executing DuckDB query with %d args: %s", len(args), query)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for the row
		row := make(map[string]interface{})
		for i, col := range columns {
			// Convert any database-specific types to standard Go types
			switch v := values[i].(type) {
			case int32:
				row[col] = v
			case int64:
				row[col] = v
			case float32:
				row[col] = v
			case float64:
				row[col] = v
			case string:
				row[col] = v
			case bool:
				row[col] = v
			case []byte:
				row[col] = string(v)
			case nil:
				row[col] = nil
			default:
				// For any other types, store as is
				row[col] = v
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	logging.Debug("Successfully executed query and scanned %d rows for query: %s", len(results), query)
	return results, nil
}

// Insert inserts multiple rows into a table
func (r *Repository) Insert(ctx context.Context, table string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	logging.Debug("Inserting %d rows into table %s", len(rows), table)

	// Get the columns from the first row
	columns := make([]string, 0, len(rows[0]))
	for col := range rows[0] {
		columns = append(columns, col)
	}

	// Build the INSERT statement
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// Prepare the statement
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Insert each row
	for _, row := range rows {
		args := make([]interface{}, len(columns))
		for i, col := range columns {
			args[i] = row[col]
		}

		if _, err := stmt.ExecContext(ctx, args...); err != nil {
			return fmt.Errorf("failed to insert row: %w", err)
		}
	}

	logging.Debug("Successfully inserted %d rows into table %s", len(rows), table)
	return nil
}

// TestConnection tests the database connection and validates the given table
func (r *Repository) TestConnection(ctx context.Context, table string) error {
	return r.ValidateTable(ctx, table, nil)
}

// ValidateTable validates that a table exists and has the required columns
func (r *Repository) ValidateTable(ctx context.Context, table string, requiredColumns []string) error {
	logging.Info("Validating table %q for required columns", table)

	// Check if table exists
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 0", table)
	if _, err := r.db.QueryContext(ctx, query); err != nil {
		return fmt.Errorf("table %q does not exist", table)
	}

	// If no required columns, we're done
	if len(requiredColumns) == 0 {
		logging.Info("Table %q validation passed", table)
		return nil
	}

	// Get table columns
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("failed to get table info: %w", err)
	}
	defer rows.Close()

	// Map of existing columns
	existingColumns := make(map[string]bool)
	for rows.Next() {
		var (
			cid     int
			name    string
			typ     string
			notnull bool
			dflt    interface{}
			pk      bool
		)
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return fmt.Errorf("failed to scan table info: %w", err)
		}
		existingColumns[name] = true
	}

	// Check required columns
	var missingColumns []string
	for _, col := range requiredColumns {
		if !existingColumns[col] {
			missingColumns = append(missingColumns, col)
		}
	}

	if len(missingColumns) > 0 {
		return fmt.Errorf("table %q is missing required columns: %v", table, missingColumns)
	}

	logging.Info("Table %q validation passed", table)
	return nil
}
