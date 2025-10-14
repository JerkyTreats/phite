package testutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
)

func TestMockRepository_Interface(t *testing.T) {
	// Test that MockRepository implements the Repository interface
	var _ dbinterface.Repository = (*MockRepository)(nil)
}

func TestMockRepository_CallTracking(t *testing.T) {
	mock := NewMockRepository()
	ctx := context.Background()

	// Test Query call tracking
	_, _ = mock.Query(ctx, "SELECT * FROM test", "arg1", "arg2")
	assert.Len(t, mock.QueryCalls, 1)
	assert.Equal(t, "SELECT * FROM test", mock.QueryCalls[0].Query)
	assert.Equal(t, []interface{}{"arg1", "arg2"}, mock.QueryCalls[0].Args)

	// Test Insert call tracking
	rows := []map[string]interface{}{{"id": 1, "name": "test"}}
	_ = mock.Insert(ctx, "test_table", rows)
	assert.Len(t, mock.InsertCalls, 1)
	assert.Equal(t, "test_table", mock.InsertCalls[0].Table)
	assert.Equal(t, rows, mock.InsertCalls[0].Rows)

	// Test TestConnection call tracking
	_ = mock.TestConnection(ctx, "test_table")
	assert.Len(t, mock.TestConnectionCalls, 1)
	assert.Equal(t, "test_table", mock.TestConnectionCalls[0].Table)

	// Test ValidateTable call tracking
	columns := []string{"id", "name"}
	_ = mock.ValidateTable(ctx, "test_table", columns)
	assert.Len(t, mock.ValidateTableCalls, 1)
	assert.Equal(t, "test_table", mock.ValidateTableCalls[0].Table)
	assert.Equal(t, columns, mock.ValidateTableCalls[0].RequiredColumns)
}

func TestMockRepository_CustomFunctions(t *testing.T) {
	mock := NewMockRepository()
	ctx := context.Background()

	// Test custom Query function
	expectedResult := []map[string]interface{}{{"id": 1, "name": "test"}}
	mock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		return expectedResult, nil
	}

	result, err := mock.Query(ctx, "SELECT * FROM test")
	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	// Test custom Insert function
	insertError := assert.AnError
	mock.InsertFunc = func(ctx context.Context, table string, rows []map[string]interface{}) error {
		return insertError
	}

	err = mock.Insert(ctx, "test_table", []map[string]interface{}{})
	assert.Equal(t, insertError, err)
}

func TestBigQueryTestConfig(t *testing.T) {
	config := BigQueryTestConfig()

	assert.Equal(t, "bq", config.Type)
	assert.Equal(t, "test-project", config.Params["project_id"])
	assert.Equal(t, "test-dataset", config.Params["dataset_id"])
	assert.Equal(t, "billing-project", config.Params["billing_project"])
	assert.False(t, config.ExpectedError)
}

func TestGnomADTestConfig(t *testing.T) {
	config := GnomADTestConfig()

	assert.Equal(t, "bq", config.Type)
	assert.Equal(t, "bigquery-public-data", config.Params["project_id"])
	assert.Equal(t, "gnomad", config.Params["dataset_id"])
	assert.Equal(t, "user-billing-project", config.Params["billing_project"])
	assert.False(t, config.ExpectedError)
}

func TestCacheTestConfig(t *testing.T) {
	config := CacheTestConfig()

	assert.Equal(t, "bq", config.Type)
	assert.Equal(t, "user-cache-project", config.Params["project_id"])
	assert.Equal(t, "phite_reference_cache", config.Params["dataset_id"])
	assert.Equal(t, "user-billing-project", config.Params["billing_project"])
	assert.False(t, config.ExpectedError)
}

func TestDuckDBTestConfig(t *testing.T) {
	config := DuckDBTestConfig()

	assert.Equal(t, "duckdb", config.Type)
	assert.Equal(t, ":memory:", config.Params["path"])
	assert.False(t, config.ExpectedError)
}

func TestDuckDBFileTestConfig(t *testing.T) {
	testPath := "/tmp/test.db"
	config := DuckDBFileTestConfig(testPath)

	assert.Equal(t, "duckdb", config.Type)
	assert.Equal(t, testPath, config.Params["path"])
	assert.False(t, config.ExpectedError)
}

func TestInvalidTestConfig(t *testing.T) {
	config := InvalidTestConfig()

	assert.Equal(t, "invalid-type", config.Type)
	assert.True(t, config.ExpectedError)
	assert.Equal(t, "unsupported database type", config.ExpectedErrorContains)
}

func TestMockRepositoryFactory(t *testing.T) {
	factory := NewMockRepositoryFactory()
	ctx := context.Background()

	// Test with no repositories set
	repo, err := factory.GetRepository(ctx, "bq")
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "unsupported database type")

	// Test with repository set
	mockRepo := NewMockRepository()
	factory.SetRepository("bq", mockRepo)

	repo, err = factory.GetRepository(ctx, "bq")
	require.NoError(t, err)
	assert.Equal(t, mockRepo, repo)
}

func TestMockRepositoryFactory_WithConstructor(t *testing.T) {
	factory := NewMockRepositoryFactory()
	ctx := context.Background()

	// Test with custom constructor
	var capturedType string
	var capturedParams map[string]string

	factory.ConstructorFunc = func(dbType string, params map[string]string) (dbinterface.Repository, error) {
		capturedType = dbType
		capturedParams = params
		return NewMockRepository(), nil
	}

	params := map[string]string{"project_id": "test-project"}
	repo, err := factory.GetRepository(ctx, "bq", params)

	require.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, "bq", capturedType)
	assert.Equal(t, params, capturedParams)
}

func TestValidateRepositoryParams(t *testing.T) {
	expected := map[string]string{
		"project_id": "test-project",
		"dataset_id": "test-dataset",
	}

	// Test with matching params
	actual := map[string]string{
		"project_id":  "test-project",
		"dataset_id":  "test-dataset",
		"extra_param": "extra-value",
	}

	// This should not panic or fail
	ValidateRepositoryParams(t, expected, actual)
}

func TestAssertMockRepositoryCalls(t *testing.T) {
	mock := NewMockRepository()
	ctx := context.Background()

	// Make some calls
	_, _ = mock.Query(ctx, "SELECT 1")
	_, _ = mock.Query(ctx, "SELECT 2")
	_ = mock.Insert(ctx, "test", []map[string]interface{}{})

	// Test assertions
	expectedCalls := NewMockRepositoryExpectedCalls().
		WithQueryCalls(2).
		WithInsertCalls(1).
		WithTestConnectionCalls(0).
		WithValidateTableCalls(0)

	AssertMockRepositoryCalls(t, mock, expectedCalls)
}

func TestMockRepositoryExpectedCalls_Builders(t *testing.T) {
	calls := NewMockRepositoryExpectedCalls()

	// Test initial state
	assert.Equal(t, -1, calls.QueryCallCount)
	assert.Equal(t, -1, calls.InsertCallCount)
	assert.Equal(t, -1, calls.TestConnectionCallCount)
	assert.Equal(t, -1, calls.ValidateTableCallCount)

	// Test builders
	calls = calls.WithQueryCalls(5).
		WithInsertCalls(3).
		WithTestConnectionCalls(1).
		WithValidateTableCalls(2)

	assert.Equal(t, 5, calls.QueryCallCount)
	assert.Equal(t, 3, calls.InsertCallCount)
	assert.Equal(t, 1, calls.TestConnectionCallCount)
	assert.Equal(t, 2, calls.ValidateTableCallCount)
}

func TestTestRepositoryInterface_WithMock(t *testing.T) {
	mock := NewMockRepository()

	// This should not panic
	TestRepositoryInterface(t, mock)

	// Verify that all methods were called
	assert.Len(t, mock.QueryCalls, 1)
	assert.Len(t, mock.InsertCalls, 1)
	assert.Len(t, mock.TestConnectionCalls, 1)
	assert.Len(t, mock.ValidateTableCalls, 1)
}
