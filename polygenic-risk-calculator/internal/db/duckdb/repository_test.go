package duckdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
)

// testRecord represents a test table record
type testRecord struct {
	ID    int32
	Name  string
	Value float32
}

func setupTestDB(t *testing.T) dbinterface.Repository {
	t.Helper()

	// Create a test database
	db, err := OpenDB(":memory:")
	require.NoError(t, err)

	// Create test table
	repo := NewRepository(db)
	_, err = repo.Query(context.Background(), `
		CREATE TABLE test_records (
			id INTEGER PRIMARY KEY,
			name TEXT,
			value FLOAT
		)
	`)
	require.NoError(t, err)

	return repo
}

func TestRepository_Query(t *testing.T) {
	repo := setupTestDB(t)

	// Insert test data
	_, err := repo.Query(context.Background(), `
		INSERT INTO test_records VALUES
		(1, 'test1', 0.1),
		(2, 'test2', 0.2)
	`)
	require.NoError(t, err)

	// Test basic query
	results, err := repo.Query(context.Background(), "SELECT id, name, value FROM test_records ORDER BY id")
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Verify first row
	assert.Equal(t, int32(1), results[0]["id"])
	assert.Equal(t, "test1", results[0]["name"])
	assert.Equal(t, float32(0.1), results[0]["value"])

	// Verify second row
	assert.Equal(t, int32(2), results[1]["id"])
	assert.Equal(t, "test2", results[1]["name"])
	assert.Equal(t, float32(0.2), results[1]["value"])

	// Test empty result set
	results, err = repo.Query(context.Background(), "SELECT * FROM test_records WHERE id > 100")
	require.NoError(t, err)
	assert.Empty(t, results)

	// Test query with parameters
	results, err = repo.Query(context.Background(), "SELECT * FROM test_records WHERE id = ?", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, int32(1), results[0]["id"])
}

func TestRepository_Insert(t *testing.T) {
	repo := setupTestDB(t)

	// Test single row insert
	err := repo.Insert(context.Background(), "test_records", []map[string]interface{}{
		{"id": 1, "name": "test1", "value": float32(0.1)},
	})
	require.NoError(t, err)

	// Verify insert
	results, err := repo.Query(context.Background(), "SELECT COUNT(*) as count FROM test_records")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, int64(1), results[0]["count"])

	// Test multiple row insert
	err = repo.Insert(context.Background(), "test_records", []map[string]interface{}{
		{"id": 2, "name": "test2", "value": float32(0.2)},
		{"id": 3, "name": "test3", "value": float32(0.3)},
	})
	require.NoError(t, err)

	// Verify multiple insert
	results, err = repo.Query(context.Background(), "SELECT COUNT(*) as count FROM test_records")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, int64(3), results[0]["count"])

	// Test empty insert
	err = repo.Insert(context.Background(), "test_records", []map[string]interface{}{})
	require.NoError(t, err)
}

func TestRepository_TestConnection(t *testing.T) {
	repo := setupTestDB(t)

	// Test valid table
	err := repo.TestConnection(context.Background(), "test_records")
	require.NoError(t, err)

	// Test non-existent table
	err = repo.TestConnection(context.Background(), "nonexistent")
	require.Error(t, err)
}

func TestRepository_ValidateTable(t *testing.T) {
	repo := setupTestDB(t)

	// Test valid table with no required columns
	err := repo.ValidateTable(context.Background(), "test_records", nil)
	require.NoError(t, err)

	// Test valid table with required columns
	err = repo.ValidateTable(context.Background(), "test_records", []string{"id", "name", "value"})
	require.NoError(t, err)

	// Test valid table with missing required column
	err = repo.ValidateTable(context.Background(), "test_records", []string{"id", "name", "value", "extra"})
	require.Error(t, err)

	// Test non-existent table
	err = repo.ValidateTable(context.Background(), "nonexistent", nil)
	require.Error(t, err)
}
