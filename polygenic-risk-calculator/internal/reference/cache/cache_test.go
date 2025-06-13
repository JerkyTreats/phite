package reference_cache

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"phite.io/polygenic-risk-calculator/internal/config"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

type mockRepo struct {
	queryFunc  func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
	insertFunc func(ctx context.Context, table string, rows []map[string]interface{}) error
}

func (m *mockRepo) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	return m.queryFunc(ctx, query, args...)
}
func (m *mockRepo) Insert(ctx context.Context, table string, rows []map[string]interface{}) error {
	return m.insertFunc(ctx, table, rows)
}
func (m *mockRepo) TestConnection(ctx context.Context, table string) error {
	return nil
}
func (m *mockRepo) ValidateTable(ctx context.Context, table string, requiredColumns []string) error {
	return nil
}

func TestMain(m *testing.M) {
	// Reset config for testing
	config.ResetForTest()

	// Create a temporary config file
	tmpDir, err := os.MkdirTemp("", "phite-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	config.SetConfigPath(configPath)

	// Create config file with test values
	configContent := `{
		"reference_stats": {
			"table_id": "test_table"
		}
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		panic(err)
	}

	// Run tests
	os.Exit(m.Run())
}

func TestRepositoryCache_Get_CacheHit(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{
					"mean":     0.5,
					"std":      1.0,
					"min":      0.0,
					"max":      1.0,
					"ancestry": "EUR",
					"trait":    "Height",
					"model":    "test_model",
				},
			}, nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}
	stats, err := cache.Get(context.Background(), StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"})
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0.5, stats.Mean)
	assert.Equal(t, 1.0, stats.Std)
	assert.Equal(t, "EUR", stats.Ancestry)
}

func TestRepositoryCache_Get_CacheMiss(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{}, nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}
	stats, err := cache.Get(context.Background(), StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"})
	assert.NoError(t, err)
	assert.Nil(t, stats)
}

func TestRepositoryCache_Get_MultipleRows(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"mean": 0.5, "std": 1.0, "min": 0.0, "max": 1.0, "ancestry": "EUR", "trait": "Height", "model": "test_model"},
				{"mean": 0.6, "std": 1.1, "min": 0.1, "max": 1.1, "ancestry": "EUR", "trait": "Height", "model": "test_model"},
			}, nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}
	stats, err := cache.Get(context.Background(), StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"})
	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestRepositoryCache_Get_InvalidStats(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"mean": 0.5, "std": -1.0, "min": 0.0, "max": 1.0, "ancestry": "EUR", "trait": "Height", "model": "test_model"},
			}, nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}
	stats, err := cache.Get(context.Background(), StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"})
	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestRepositoryCache_Get_QueryError(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return nil, errors.New("db error")
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}
	stats, err := cache.Get(context.Background(), StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"})
	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestRepositoryCache_Store_Valid(t *testing.T) {
	inserted := false
	repo := &mockRepo{
		insertFunc: func(ctx context.Context, table string, rows []map[string]interface{}) error {
			inserted = true
			return nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}
	stats := &reference_stats.ReferenceStats{Mean: 0.5, Std: 1.0, Min: 0.0, Max: 1.0}
	err := cache.Store(context.Background(), StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"}, stats)
	assert.NoError(t, err)
	assert.True(t, inserted)
}

func TestRepositoryCache_Store_InvalidStats(t *testing.T) {
	repo := &mockRepo{
		insertFunc: func(ctx context.Context, table string, rows []map[string]interface{}) error {
			return nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}
	stats := &reference_stats.ReferenceStats{Mean: 0.5, Std: -1.0, Min: 0.0, Max: 1.0}
	err := cache.Store(context.Background(), StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"}, stats)
	assert.Error(t, err)
}

func TestRepositoryCache_Store_InsertError(t *testing.T) {
	repo := &mockRepo{
		insertFunc: func(ctx context.Context, table string, rows []map[string]interface{}) error {
			return errors.New("insert error")
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}
	stats := &reference_stats.ReferenceStats{Mean: 0.5, Std: 1.0, Min: 0.0, Max: 1.0}
	err := cache.Store(context.Background(), StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"}, stats)
	assert.Error(t, err)
}

func TestRepositoryCache_GetReferenceStats(t *testing.T) {
	called := false
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			called = true
			return []map[string]interface{}{}, nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}
	_, _ = cache.GetReferenceStats(context.Background(), "EUR", "Height", "test_model")
	assert.True(t, called)
}
