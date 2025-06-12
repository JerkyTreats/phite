package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/db/duckdb"
)

// DBRepository defines the common interface for database operations.
type DBRepository interface {
	// Query executes a query and returns results as a slice of maps.
	Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)

	// Insert inserts rows into the specified table.
	Insert(ctx context.Context, table string, rows []map[string]interface{}) error

	// TestConnection checks that the database is reachable and the table exists.
	TestConnection(ctx context.Context, table string) error
}

// RepositoryConstructor is a function type for creating new repository instances
type RepositoryConstructor func(ctx context.Context, config map[string]string) (DBRepository, error)

var constructors = map[string]RepositoryConstructor{
	"duckdb": newDuckDBRepository,
}

// GetRepository creates a new repository instance of the specified type
func GetRepository(ctx context.Context, dbType string, config map[string]string) (DBRepository, error) {
	if constructor, ok := constructors[dbType]; ok {
		return constructor(ctx, config)
	}
	return nil, fmt.Errorf("unsupported database type: %s", dbType)
}

// newDuckDBRepository creates a new DuckDB repository instance
func newDuckDBRepository(ctx context.Context, config map[string]string) (DBRepository, error) {
	path, ok := config["path"]
	if !ok {
		return nil, fmt.Errorf("path is required for DuckDB repository")
	}
	db, err := duckdb.OpenDB(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	return &duckdbRepository{db: db}, nil
}

// duckdbRepository implements DBRepository for DuckDB
type duckdbRepository struct {
	db *sql.DB
}

func (r *duckdbRepository) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *duckdbRepository) Insert(ctx context.Context, table string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// Get columns from first row
	columns := make([]string, 0, len(rows[0]))
	for col := range rows[0] {
		columns = append(columns, col)
	}

	// Build query
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", table, strings.Join(columns, ", "))
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}
	query += "(" + strings.Join(placeholders, ", ") + ")"

	// Prepare statement
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert rows
	for _, row := range rows {
		args := make([]interface{}, len(columns))
		for i, col := range columns {
			args[i] = row[col]
		}
		if _, err := stmt.ExecContext(ctx, args...); err != nil {
			return err
		}
	}

	return nil
}

func (r *duckdbRepository) TestConnection(ctx context.Context, table string) error {
	return duckdb.ValidateTable(r.db, table, []string{})
}
