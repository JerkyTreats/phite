package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
)

// mockRepository implements the Repository interface for testing
type mockRepository struct {
	queryFunc         func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
	insertFunc        func(ctx context.Context, table string, rows []map[string]interface{}) error
	testConnFunc      func(ctx context.Context, table string) error
	validateTableFunc func(ctx context.Context, table string, requiredColumns []string) error
}

func (m *mockRepository) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *mockRepository) Insert(ctx context.Context, table string, rows []map[string]interface{}) error {
	if m.insertFunc != nil {
		return m.insertFunc(ctx, table, rows)
	}
	return nil
}

func (m *mockRepository) TestConnection(ctx context.Context, table string) error {
	if m.testConnFunc != nil {
		return m.testConnFunc(ctx, table)
	}
	return nil
}

func (m *mockRepository) ValidateTable(ctx context.Context, table string, requiredColumns []string) error {
	if m.validateTableFunc != nil {
		return m.validateTableFunc(ctx, table, requiredColumns)
	}
	return nil
}

// Mock constructors for testing
var (
	mockBQConstructor     RepositoryConstructor
	mockDuckDBConstructor RepositoryConstructor
	originalConstructors  map[string]RepositoryConstructor
)

func setupMockConstructors(t *testing.T) {
	t.Helper()

	// Store original constructors
	originalConstructors = make(map[string]RepositoryConstructor)
	originalConstructors["bq"] = constructors["bq"]
	originalConstructors["duckdb"] = constructors["duckdb"]

	// Set up mock constructors
	mockBQConstructor = func(ctx context.Context, params map[string]string) (dbinterface.Repository, error) {
		return &mockRepository{}, nil
	}

	mockDuckDBConstructor = func(ctx context.Context, params map[string]string) (dbinterface.Repository, error) {
		return &mockRepository{}, nil
	}

	constructors["bq"] = mockBQConstructor
	constructors["duckdb"] = mockDuckDBConstructor
}

func teardownMockConstructors(t *testing.T) {
	t.Helper()

	// Restore original constructors
	for dbType, constructor := range originalConstructors {
		constructors[dbType] = constructor
	}
}

func TestGetRepository_BasicUsage(t *testing.T) {
	setupMockConstructors(t)
	defer teardownMockConstructors(t)

	ctx := context.Background()

	tests := []struct {
		name   string
		dbType string
	}{
		{"BigQuery", "bq"},
		{"DuckDB", "duckdb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := GetRepository(ctx, tt.dbType)
			require.NoError(t, err)
			assert.NotNil(t, repo)
			assert.IsType(t, &mockRepository{}, repo)
		})
	}
}

func TestGetRepository_WithParameters(t *testing.T) {
	setupMockConstructors(t)
	defer teardownMockConstructors(t)

	ctx := context.Background()

	tests := []struct {
		name   string
		dbType string
		params map[string]string
	}{
		{
			name:   "BigQuery with parameters",
			dbType: "bq",
			params: map[string]string{
				"project_id":      "test-project",
				"dataset_id":      "test-dataset",
				"billing_project": "billing-project",
			},
		},
		{
			name:   "DuckDB with parameters",
			dbType: "duckdb",
			params: map[string]string{
				"path": "/tmp/test.db",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := GetRepository(ctx, tt.dbType, tt.params)
			require.NoError(t, err)
			assert.NotNil(t, repo)
		})
	}
}

func TestGetRepository_BackwardCompatibility(t *testing.T) {
	setupMockConstructors(t)
	defer teardownMockConstructors(t)

	ctx := context.Background()

	// Test that old-style calls still work
	repo, err := GetRepository(ctx, "bq")
	require.NoError(t, err)
	assert.NotNil(t, repo)

	repo, err = GetRepository(ctx, "duckdb")
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestGetRepository_UnsupportedType(t *testing.T) {
	ctx := context.Background()

	repo, err := GetRepository(ctx, "unsupported")
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "unsupported database type")
}

func TestGetRepository_EmptyType(t *testing.T) {
	ctx := context.Background()

	repo, err := GetRepository(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "database type cannot be empty")
}

func TestGetRepository_ParameterPrecedence(t *testing.T) {
	// Test that explicit parameters take precedence over config
	setupMockConstructors(t)
	defer teardownMockConstructors(t)

	// Create a constructor that captures the parameters
	var capturedParams map[string]string
	testConstructor := func(ctx context.Context, params map[string]string) (dbinterface.Repository, error) {
		capturedParams = make(map[string]string)
		for k, v := range params {
			capturedParams[k] = v
		}
		return &mockRepository{}, nil
	}

	constructors["test"] = testConstructor

	ctx := context.Background()
	expectedParams := map[string]string{
		"project_id": "explicit-project",
		"dataset_id": "explicit-dataset",
	}

	_, err := GetRepository(ctx, "test", expectedParams)
	require.NoError(t, err)

	assert.Equal(t, expectedParams["project_id"], capturedParams["project_id"])
	assert.Equal(t, expectedParams["dataset_id"], capturedParams["dataset_id"])
}

func TestGetRepository_MultipleParams(t *testing.T) {
	setupMockConstructors(t)
	defer teardownMockConstructors(t)

	ctx := context.Background()

	// Test with multiple parameter maps (should use first non-nil)
	params1 := map[string]string{"project_id": "project1"}
	params2 := map[string]string{"project_id": "project2"}

	repo, err := GetRepository(ctx, "bq", params1, params2)
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestGetRepository_NilParams(t *testing.T) {
	setupMockConstructors(t)
	defer teardownMockConstructors(t)

	ctx := context.Background()

	// Test with nil params (should work like no params)
	repo, err := GetRepository(ctx, "bq", nil)
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestGetRepository_ConstructorError(t *testing.T) {
	// Test constructor returning error
	errorConstructor := func(ctx context.Context, params map[string]string) (dbinterface.Repository, error) {
		return nil, assert.AnError
	}

	constructors["error-type"] = errorConstructor
	defer delete(constructors, "error-type")

	ctx := context.Background()

	repo, err := GetRepository(ctx, "error-type")
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Equal(t, assert.AnError, err)
}

// Integration-style tests with real constructors
func TestGetRepository_Integration_DuckDB(t *testing.T) {
	ctx := context.Background()

	// Test DuckDB with memory database
	params := map[string]string{
		"path": ":memory:",
	}

	repo, err := GetRepository(ctx, "duckdb", params)
	require.NoError(t, err)
	assert.NotNil(t, repo)

	// Test basic functionality
	err = repo.TestConnection(ctx, "test_table")
	// This should error since table doesn't exist, but connection should work
	assert.Error(t, err)
}

// Benchmark tests
func BenchmarkGetRepository_WithoutParams(b *testing.B) {
	setupMockConstructors(&testing.T{})
	defer teardownMockConstructors(&testing.T{})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetRepository(ctx, "bq")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetRepository_WithParams(b *testing.B) {
	setupMockConstructors(&testing.T{})
	defer teardownMockConstructors(&testing.T{})

	ctx := context.Background()
	params := map[string]string{
		"project_id": "test-project",
		"dataset_id": "test-dataset",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetRepository(ctx, "bq", params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test helper functions
func TestRepositoryConstructor_Interface(t *testing.T) {
	// Verify that RepositoryConstructor has the right signature
	var constructor RepositoryConstructor = func(ctx context.Context, params map[string]string) (dbinterface.Repository, error) {
		return nil, nil
	}

	assert.NotNil(t, constructor)

	ctx := context.Background()
	params := map[string]string{}

	repo, err := constructor(ctx, params)
	assert.Nil(t, repo)
	assert.Nil(t, err)
}
