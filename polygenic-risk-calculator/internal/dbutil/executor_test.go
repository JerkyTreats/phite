package dbutil

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

func TestExecuteDuckDBQueryWithPath(t *testing.T) {
	logging.SetSilentLoggingForTest()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_path_query.db")

	// Set up test database
	db, err := OpenDuckDB(dbPath)
	require.NoError(t, err)

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE test_path_records (
			id INTEGER PRIMARY KEY,
			name TEXT,
			value FLOAT
		);

		INSERT INTO test_path_records VALUES
		(1, 'record1', 0.1),
		(2, 'record2', 0.2),
		(3, 'record3', 0.3);
	`)
	require.NoError(t, err)
	db.Close()

	type testRecord struct {
		ID    int
		Name  string
		Value float64
	}

	// Row scanner function for test_records
	testRecordScanner := func(rows *sql.Rows) (*testRecord, error) {
		var record testRecord
		err := rows.Scan(&record.ID, &record.Name, &record.Value)
		return &record, err
	}

	t.Run("successful query execution", func(t *testing.T) {
		query := "SELECT id, name, value FROM test_path_records WHERE id > ? ORDER BY id"

		records, err := ExecuteDuckDBQueryWithPath(dbPath, query, testRecordScanner, 1)
		require.NoError(t, err)

		assert.Len(t, records, 2)
		assert.Equal(t, 2, records[0].ID)
		assert.Equal(t, "record2", records[0].Name)
		assert.Equal(t, 0.2, records[0].Value)
		assert.Equal(t, 3, records[1].ID)
	})

	t.Run("empty result set", func(t *testing.T) {
		query := "SELECT id, name, value FROM test_path_records WHERE id > ?"

		records, err := ExecuteDuckDBQueryWithPath(dbPath, query, testRecordScanner, 100)
		require.NoError(t, err)
		assert.Empty(t, records)
	})

	t.Run("handles db connection error", func(t *testing.T) {
		_, err := ExecuteDuckDBQueryWithPath("/nonexistent/path.db", "SELECT 1", testRecordScanner)
		assert.Error(t, err)
	})

	t.Run("handles query execution error", func(t *testing.T) {
		_, err := ExecuteDuckDBQueryWithPath(dbPath, "SELECT * FROM nonexistent_table", testRecordScanner)
		assert.Error(t, err)
	})

	t.Run("handles scanner error", func(t *testing.T) {
		badScanner := func(rows *sql.Rows) (*testRecord, error) {
			return nil, errors.New("scanner error")
		}

		_, err := ExecuteDuckDBQueryWithPath(dbPath, "SELECT id, name, value FROM test_path_records", badScanner)
		assert.Error(t, err)
	})
}
