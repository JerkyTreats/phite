package testutils

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
)

// MockRepository implements the Repository interface for testing
type MockRepository struct {
	QueryFunc          func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
	InsertFunc         func(ctx context.Context, table string, rows []map[string]interface{}) error
	TestConnectionFunc func(ctx context.Context, table string) error
	ValidateTableFunc  func(ctx context.Context, table string, requiredColumns []string) error

	// Call tracking
	QueryCalls          []QueryCall
	InsertCalls         []InsertCall
	TestConnectionCalls []TestConnectionCall
	ValidateTableCalls  []ValidateTableCall
}

type QueryCall struct {
	Query string
	Args  []interface{}
}

type InsertCall struct {
	Table string
	Rows  []map[string]interface{}
}

type TestConnectionCall struct {
	Table string
}

type ValidateTableCall struct {
	Table           string
	RequiredColumns []string
}

func (m *MockRepository) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	m.QueryCalls = append(m.QueryCalls, QueryCall{Query: query, Args: args})
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockRepository) Insert(ctx context.Context, table string, rows []map[string]interface{}) error {
	m.InsertCalls = append(m.InsertCalls, InsertCall{Table: table, Rows: rows})
	if m.InsertFunc != nil {
		return m.InsertFunc(ctx, table, rows)
	}
	return nil
}

func (m *MockRepository) TestConnection(ctx context.Context, table string) error {
	m.TestConnectionCalls = append(m.TestConnectionCalls, TestConnectionCall{Table: table})
	if m.TestConnectionFunc != nil {
		return m.TestConnectionFunc(ctx, table)
	}
	return nil
}

func (m *MockRepository) ValidateTable(ctx context.Context, table string, requiredColumns []string) error {
	m.ValidateTableCalls = append(m.ValidateTableCalls, ValidateTableCall{Table: table, RequiredColumns: requiredColumns})
	if m.ValidateTableFunc != nil {
		return m.ValidateTableFunc(ctx, table, requiredColumns)
	}
	return nil
}

// NewMockRepository creates a new mock repository with default behavior
func NewMockRepository() *MockRepository {
	return &MockRepository{
		QueryCalls:          make([]QueryCall, 0),
		InsertCalls:         make([]InsertCall, 0),
		TestConnectionCalls: make([]TestConnectionCall, 0),
		ValidateTableCalls:  make([]ValidateTableCall, 0),
	}
}

// RepositoryTestConfig holds configuration for repository tests
type RepositoryTestConfig struct {
	Type                  string
	Params                map[string]string
	ExpectedError         bool
	ExpectedErrorContains string
}

// BigQueryTestConfig returns a standard BigQuery test configuration
func BigQueryTestConfig() RepositoryTestConfig {
	return RepositoryTestConfig{
		Type: "bq",
		Params: map[string]string{
			"project_id":      "test-project",
			"dataset_id":      "test-dataset",
			"billing_project": "billing-project",
		},
		ExpectedError: false,
	}
}

// GnomADTestConfig returns configuration for gnomAD public data access
func GnomADTestConfig() RepositoryTestConfig {
	return RepositoryTestConfig{
		Type: "bq",
		Params: map[string]string{
			"project_id":      "bigquery-public-data",
			"dataset_id":      "gnomad",
			"billing_project": "user-billing-project",
		},
		ExpectedError: false,
	}
}

// CacheTestConfig returns configuration for cache storage
func CacheTestConfig() RepositoryTestConfig {
	return RepositoryTestConfig{
		Type: "bq",
		Params: map[string]string{
			"project_id":      "user-cache-project",
			"dataset_id":      "phite_reference_cache",
			"billing_project": "user-billing-project",
		},
		ExpectedError: false,
	}
}

// DuckDBTestConfig returns a standard DuckDB test configuration
func DuckDBTestConfig() RepositoryTestConfig {
	return RepositoryTestConfig{
		Type: "duckdb",
		Params: map[string]string{
			"path": ":memory:",
		},
		ExpectedError: false,
	}
}

// DuckDBFileTestConfig returns DuckDB configuration with file storage
func DuckDBFileTestConfig(path string) RepositoryTestConfig {
	return RepositoryTestConfig{
		Type: "duckdb",
		Params: map[string]string{
			"path": path,
		},
		ExpectedError: false,
	}
}

// InvalidTestConfig returns configuration that should cause errors
func InvalidTestConfig() RepositoryTestConfig {
	return RepositoryTestConfig{
		Type: "invalid-type",
		Params: map[string]string{
			"invalid": "params",
		},
		ExpectedError:         true,
		ExpectedErrorContains: "unsupported database type",
	}
}

// MockRepositoryFactory creates a mock repository factory for testing
type MockRepositoryFactory struct {
	Repositories    map[string]dbinterface.Repository
	ConstructorFunc func(dbType string, params map[string]string) (dbinterface.Repository, error)
}

// NewMockRepositoryFactory creates a new mock factory
func NewMockRepositoryFactory() *MockRepositoryFactory {
	return &MockRepositoryFactory{
		Repositories: make(map[string]dbinterface.Repository),
	}
}

// SetRepository sets a mock repository for a specific database type
func (f *MockRepositoryFactory) SetRepository(dbType string, repo dbinterface.Repository) {
	f.Repositories[dbType] = repo
}

// GetRepository returns the mock repository for the given type
func (f *MockRepositoryFactory) GetRepository(ctx context.Context, dbType string, params ...map[string]string) (dbinterface.Repository, error) {
	if f.ConstructorFunc != nil {
		var paramMap map[string]string
		if len(params) > 0 {
			paramMap = params[0]
		}
		return f.ConstructorFunc(dbType, paramMap)
	}

	if repo, exists := f.Repositories[dbType]; exists {
		return repo, nil
	}

	return nil, fmt.Errorf("unsupported database type: %s", dbType)
}

// TestRepositoryInterface tests that a repository properly implements the interface
func TestRepositoryInterface(t *testing.T, repo dbinterface.Repository) {
	t.Helper()

	ctx := context.Background()

	// Test Query method exists
	_, _ = repo.Query(ctx, "SELECT 1")
	// Don't assert error - just ensure method exists

	// Test Insert method exists
	_ = repo.Insert(ctx, "test_table", []map[string]interface{}{})
	// Don't assert error - just ensure method exists

	// Test TestConnection method exists
	_ = repo.TestConnection(ctx, "test_table")
	// Don't assert error - just ensure method exists

	// Test ValidateTable method exists
	_ = repo.ValidateTable(ctx, "test_table", []string{})
	// Don't assert error - just ensure method exists

	t.Log("Repository interface methods are properly implemented")
}

// ValidateRepositoryParams validates that repository parameters are handled correctly
func ValidateRepositoryParams(t *testing.T, expectedParams, actualParams map[string]string) {
	t.Helper()

	for key, expectedValue := range expectedParams {
		actualValue, exists := actualParams[key]
		assert.True(t, exists, "Expected parameter %s not found", key)
		assert.Equal(t, expectedValue, actualValue, "Parameter %s has incorrect value", key)
	}
}

// AssertMockRepositoryCalls validates that mock repository was called as expected
func AssertMockRepositoryCalls(t *testing.T, mock *MockRepository, expectedCalls MockRepositoryExpectedCalls) {
	t.Helper()

	if expectedCalls.QueryCallCount >= 0 {
		assert.Len(t, mock.QueryCalls, expectedCalls.QueryCallCount, "Unexpected number of Query calls")
	}

	if expectedCalls.InsertCallCount >= 0 {
		assert.Len(t, mock.InsertCalls, expectedCalls.InsertCallCount, "Unexpected number of Insert calls")
	}

	if expectedCalls.TestConnectionCallCount >= 0 {
		assert.Len(t, mock.TestConnectionCalls, expectedCalls.TestConnectionCallCount, "Unexpected number of TestConnection calls")
	}

	if expectedCalls.ValidateTableCallCount >= 0 {
		assert.Len(t, mock.ValidateTableCalls, expectedCalls.ValidateTableCallCount, "Unexpected number of ValidateTable calls")
	}
}

type MockRepositoryExpectedCalls struct {
	QueryCallCount          int
	InsertCallCount         int
	TestConnectionCallCount int
	ValidateTableCallCount  int
}

// NewMockRepositoryExpectedCalls creates expected calls with all counts set to -1 (don't check)
func NewMockRepositoryExpectedCalls() MockRepositoryExpectedCalls {
	return MockRepositoryExpectedCalls{
		QueryCallCount:          -1,
		InsertCallCount:         -1,
		TestConnectionCallCount: -1,
		ValidateTableCallCount:  -1,
	}
}

// WithQueryCalls sets the expected query call count
func (e MockRepositoryExpectedCalls) WithQueryCalls(count int) MockRepositoryExpectedCalls {
	e.QueryCallCount = count
	return e
}

// WithInsertCalls sets the expected insert call count
func (e MockRepositoryExpectedCalls) WithInsertCalls(count int) MockRepositoryExpectedCalls {
	e.InsertCallCount = count
	return e
}

// WithTestConnectionCalls sets the expected test connection call count
func (e MockRepositoryExpectedCalls) WithTestConnectionCalls(count int) MockRepositoryExpectedCalls {
	e.TestConnectionCallCount = count
	return e
}

// WithValidateTableCalls sets the expected validate table call count
func (e MockRepositoryExpectedCalls) WithValidateTableCalls(count int) MockRepositoryExpectedCalls {
	e.ValidateTableCallCount = count
	return e
}
