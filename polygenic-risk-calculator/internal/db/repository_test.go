package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Query(t *testing.T) {
	// Create a test database
	repo, err := GetRepository(context.Background(), "duckdb", map[string]string{
		"path": ":memory:",
	})
	require.NoError(t, err)

	// Create a test table
	_, err = repo.Query(context.Background(), "CREATE TABLE test (id INTEGER, name VARCHAR)")
	require.NoError(t, err)

	// Insert test data
	_, err = repo.Query(context.Background(), "INSERT INTO test VALUES (1, 'test1'), (2, 'test2')")
	require.NoError(t, err)

	// Query data
	results, err := repo.Query(context.Background(), "SELECT id, name FROM test ORDER BY id")
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Verify first row
	assert.Equal(t, float64(1), results[0]["id"])
	assert.Equal(t, "test1", results[0]["name"])

	// Verify second row
	assert.Equal(t, float64(2), results[1]["id"])
	assert.Equal(t, "test2", results[1]["name"])
}

func TestRepository_Insert(t *testing.T) {
	// Create a test database
	repo, err := GetRepository(context.Background(), "duckdb", map[string]string{
		"path": ":memory:",
	})
	require.NoError(t, err)

	// Create a test table
	_, err = repo.Query(context.Background(), "CREATE TABLE test (id INTEGER, name VARCHAR)")
	require.NoError(t, err)

	// Insert test data
	err = repo.Insert(context.Background(), "test", []map[string]interface{}{
		{"id": 1, "name": "test1"},
		{"id": 2, "name": "test2"},
	})
	require.NoError(t, err)

	// Verify data was inserted
	results, err := repo.Query(context.Background(), "SELECT COUNT(*) as count FROM test")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, float64(2), results[0]["count"])
}

func TestRepository_TestConnection(t *testing.T) {
	// Create a test database
	repo, err := GetRepository(context.Background(), "duckdb", map[string]string{
		"path": ":memory:",
	})
	require.NoError(t, err)

	// Create a test table
	_, err = repo.Query(context.Background(), "CREATE TABLE test (id INTEGER)")
	require.NoError(t, err)

	// Test connection
	err = repo.TestConnection(context.Background(), "test")
	require.NoError(t, err)

	// Test non-existent table
	err = repo.TestConnection(context.Background(), "nonexistent")
	require.Error(t, err)
}
