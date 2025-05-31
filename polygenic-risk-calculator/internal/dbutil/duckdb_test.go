package dbutil

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDuckDB(t *testing.T) {
	t.Run("opens a new database connection", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")

		db, err := OpenDuckDB(dbPath)
		require.NoError(t, err)
		defer db.Close()

		// Verify connection is usable
		var result int
		err = db.QueryRow("SELECT 42").Scan(&result)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		_, err := OpenDuckDB("/invalid/path/to/db.duckdb")
		assert.Error(t, err)
	})
}

func TestValidateTable(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_validate.db")

	// Set up test database
	db, err := OpenDuckDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT,
			value FLOAT
		);
	`)
	require.NoError(t, err)

	tests := []struct {
		name        string
		table       string
		columns     []string
		expectError bool
	}{
		{
			name:        "valid table and columns",
			table:       "test_table",
			columns:     []string{"id", "name", "value"},
			expectError: false,
		},
		{
			name:        "missing column",
			table:       "test_table",
			columns:     []string{"id", "nonexistent"},
			expectError: true,
		},
		{
			name:        "nonexistent table",
			table:       "nonexistent_table",
			columns:     []string{"id"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTable(db, tt.table, tt.columns)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWithConnection(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_connection.db")

	// Test successful operation
	err := WithConnection(dbPath, func(db *sql.DB) error {
		var result int
		return db.QueryRow("SELECT 42").Scan(&result)
	})
	require.NoError(t, err)

	// Test error handling
	err = WithConnection("/invalid/path", func(db *sql.DB) error {
		return nil
	})
	assert.Error(t, err)
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}
