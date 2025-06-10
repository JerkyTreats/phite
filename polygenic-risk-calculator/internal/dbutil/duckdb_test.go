package dbutil

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"phite.io/polygenic-risk-calculator/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDuckDB(t *testing.T) {
	logging.SetSilentLoggingForTest()
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
	logging.SetSilentLoggingForTest()
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
	logging.SetSilentLoggingForTest()
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

// Define a simple struct for testing ExecuteDuckDBQuery
type testItem struct {
	ID   int
	Name string
	Rate float64
}

// Define a RowScanner for testItem
func testItemScanner(rows *sql.Rows) (*testItem, error) {
	var item testItem
	err := rows.Scan(&item.ID, &item.Name, &item.Rate)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func TestExecuteDuckDBQuery(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Helper function to set up an in-memory DB with a test table
	setupInMemoryDBWithTestData := func(t *testing.T) *sql.DB {
		db, err := OpenDuckDB(":memory:") // Use in-memory DuckDB
		require.NoError(t, err, "Failed to open in-memory DuckDB")

		_, err = db.Exec(`
			CREATE TABLE items (
				id INTEGER PRIMARY KEY,
				name TEXT,
				rate DOUBLE
			);
		`)
		require.NoError(t, err, "Failed to create test table 'items'")

		// Insert test data
		stmt, err := db.Prepare("INSERT INTO items (id, name, rate) VALUES (?, ?, ?)")
		require.NoError(t, err, "Failed to prepare insert statement")
		defer stmt.Close()

		testData := []testItem{
			{ID: 1, Name: "Item One", Rate: 1.23},
			{ID: 2, Name: "Item Two", Rate: 4.56},
			{ID: 3, Name: "Item Three", Rate: 7.89},
		}
		for _, item := range testData {
			_, err = stmt.Exec(item.ID, item.Name, item.Rate)
			require.NoError(t, err, "Failed to insert test data: %+v", item)
		}
		return db
	}

	t.Run("successful execution with data", func(t *testing.T) {
		db := setupInMemoryDBWithTestData(t)
		defer db.Close()

		query := "SELECT id, name, rate FROM items ORDER BY id"
		expectedResults := []*testItem{
			{ID: 1, Name: "Item One", Rate: 1.23},
			{ID: 2, Name: "Item Two", Rate: 4.56},
			{ID: 3, Name: "Item Three", Rate: 7.89},
		}

		results, err := ExecuteDuckDBQuery(context.Background(), db, query, testItemScanner)
		require.NoError(t, err)
		require.Len(t, results, len(expectedResults))
		assert.Equal(t, expectedResults, results)
	})

	t.Run("successful execution with no rows", func(t *testing.T) {
		db := setupInMemoryDBWithTestData(t) // Sets up table with data, but we query for non-existent
		defer db.Close()

		query := "SELECT id, name, rate FROM items WHERE name = 'NonExistentItem'"
		results, err := ExecuteDuckDBQuery(context.Background(), db, query, testItemScanner)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("query syntax error", func(t *testing.T) {
		db, err := OpenDuckDB(":memory:")
		require.NoError(t, err)
		defer db.Close()

		query := "SELEKT id, name FROM items" // Intentional syntax error
		_, err = ExecuteDuckDBQuery(context.Background(), db, query, testItemScanner)
		assert.Error(t, err)
		// We could be more specific about the error type if DuckDB driver provides it
	})

	t.Run("row scanner error", func(t *testing.T) {
		db := setupInMemoryDBWithTestData(t)
		defer db.Close()

		// This scanner will fail if it encounters "Item Two"
		erroringScanner := func(rows *sql.Rows) (*testItem, error) {
			var item testItem
			err := rows.Scan(&item.ID, &item.Name, &item.Rate)
			if err != nil {
				return nil, err
			}
			if item.Name == "Item Two" {
				return nil, errors.New("intentional scanner error for Item Two")
			}
			return &item, nil
		}

		query := "SELECT id, name, rate FROM items ORDER BY id" // Query will fetch "Item Two"
		_, err := ExecuteDuckDBQuery(context.Background(), db, query, erroringScanner)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "intentional scanner error for Item Two")
	})

	t.Run("context cancellation", func(t *testing.T) {
		db := setupInMemoryDBWithTestData(t)
		defer db.Close()

		// Create a context that will be cancelled quickly
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond) // Very short timeout
		defer cancel()

		query := "SELECT id, name, rate FROM items" 
		// For DuckDB in-memory, queries are very fast.
		// To reliably test cancellation, we might need a query that takes longer,
		// or a way to inject a delay. For now, we rely on the timeout.
		// If the query is faster than the timeout, this test might not catch cancellation.
		// A more robust test might involve a mock DB or a custom DuckDB function that sleeps.

		time.Sleep(5 * time.Millisecond) // Ensure timeout has likely occurred

		_, err := ExecuteDuckDBQuery(ctx, db, query, testItemScanner)
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded, "Expected context.DeadlineExceeded error")
	})
}

func TestMain(m *testing.M) {
	logging.SetSilentLoggingForTest()
	// Run tests
	code := m.Run()
	os.Exit(code)
}
