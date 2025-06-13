package duckdb

import (
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// OpenDB opens a connection to a DuckDB database
func OpenDB(path string) (*sql.DB, error) {
	logging.Info("Opening DuckDB database at %s", path)

	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping DuckDB database: %w", err)
	}

	logging.Info("DuckDB connection established at %s", path)
	return db, nil
}
