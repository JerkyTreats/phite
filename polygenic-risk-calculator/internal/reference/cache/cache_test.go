package reference_cache

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"phite.io/polygenic-risk-calculator/internal/ancestry"
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

	// Create ancestry object for testing
	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	_, _ = cache.GetReferenceStats(context.Background(), eur, "Height", "test_model")
	assert.True(t, called)
}

func TestRepositoryCache_GetReferenceStats_WithGenderedAncestry(t *testing.T) {
	var capturedArgs []interface{}
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			capturedArgs = args
			return []map[string]interface{}{}, nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	// Test with gendered ancestry to ensure correct code generation
	eurMale, err := ancestry.New("EUR", "MALE")
	assert.NoError(t, err)

	_, _ = cache.GetReferenceStats(context.Background(), eurMale, "Height", "test_model")

	// Verify that the ancestry code "EUR_MALE" was used in the query
	assert.Len(t, capturedArgs, 3)
	assert.Equal(t, "EUR_MALE", capturedArgs[0])
	assert.Equal(t, "Height", capturedArgs[1])
	assert.Equal(t, "test_model", capturedArgs[2])
}

func TestRepositoryCache_GetReferenceStats_CacheKeyGeneration(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{}, nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	// Test different ancestry gender combinations
	ancestries := []struct {
		code        string
		description string
	}{
		{"EUR", "European"},
		{"EUR_MALE", "European Male"},
		{"EUR_FEMALE", "European Female"},
		{"AFR", "African"},
		{"AFR_MALE", "African Male"},
		{"AFR_FEMALE", "African Female"},
		{"AMR", "Admixed American"},
		{"AMR_MALE", "Admixed American Male"},
		{"AMR_FEMALE", "Admixed American Female"},
		{"EAS", "East Asian"},
		{"EAS_MALE", "East Asian Male"},
		{"EAS_FEMALE", "East Asian Female"},
		{"SAS", "South Asian"},
		{"SAS_MALE", "South Asian Male"},
		{"SAS_FEMALE", "South Asian Female"},
	}

	for _, anc := range ancestries {
		// Parse the code to get population and gender
		var population, gender string
		if len(anc.code) == 3 {
			population = anc.code
			gender = ""
		} else {
			parts := strings.Split(anc.code, "_")
			population = parts[0]
			gender = parts[1]
		}

		ancestry, err := ancestry.New(population, gender)
		assert.NoError(t, err)
		_, _ = cache.GetReferenceStats(context.Background(), ancestry, "Height", "test_model")
	}
}

// Test cases for batch cache operations

func TestRepositoryCache_GetBatch_MultipleCacheHits(t *testing.T) {
	defer config.ResetForTest()

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
				{
					"mean":     0.6,
					"std":      1.1,
					"min":      0.1,
					"max":      1.1,
					"ancestry": "EUR",
					"trait":    "BMI",
					"model":    "test_model",
				},
			}, nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	requests := []StatsRequest{
		{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
		{Ancestry: "EUR", Trait: "BMI", ModelID: "test_model"},
	}

	results, err := cache.GetBatch(context.Background(), requests)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	heightKey := "EUR|Height|test_model"
	bmiKey := "EUR|BMI|test_model"

	assert.Contains(t, results, heightKey)
	assert.Contains(t, results, bmiKey)
	assert.Equal(t, 0.5, results[heightKey].Mean)
	assert.Equal(t, 0.6, results[bmiKey].Mean)
}

func TestRepositoryCache_GetBatch_CacheMisses(t *testing.T) {
	defer config.ResetForTest()

	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{}, nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	requests := []StatsRequest{
		{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
		{Ancestry: "EUR", Trait: "BMI", ModelID: "test_model"},
	}

	results, err := cache.GetBatch(context.Background(), requests)
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestRepositoryCache_GetBatch_EmptyRequests(t *testing.T) {
	defer config.ResetForTest()

	repo := &mockRepo{}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	results, err := cache.GetBatch(context.Background(), []StatsRequest{})
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestRepositoryCache_GetBatch_MixedResults(t *testing.T) {
	defer config.ResetForTest()

	queryCount := 0
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			queryCount++
			// Only return one result to test mixed hit/miss scenario
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

	requests := []StatsRequest{
		{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
		{Ancestry: "EUR", Trait: "BMI", ModelID: "test_model"},
	}

	results, err := cache.GetBatch(context.Background(), requests)
	assert.NoError(t, err)
	assert.Len(t, results, 1)      // Only one result should be found
	assert.Equal(t, 1, queryCount) // Should make single bulk query
}

func TestRepositoryCache_GetBatch_QueryError(t *testing.T) {
	defer config.ResetForTest()

	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return nil, errors.New("database error")
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	requests := []StatsRequest{
		{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
	}

	results, err := cache.GetBatch(context.Background(), requests)
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestRepositoryCache_GetBatch_InvalidStats(t *testing.T) {
	defer config.ResetForTest()

	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{
					"mean":     0.5,
					"std":      -1.0, // Invalid std
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

	requests := []StatsRequest{
		{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
	}

	results, err := cache.GetBatch(context.Background(), requests)
	assert.NoError(t, err)
	assert.Len(t, results, 0) // Invalid stats should be skipped
}

func TestRepositoryCache_StoreBatch_ValidEntries(t *testing.T) {
	config.SetForTest(CacheBatchSizeKey, 100)
	defer config.ResetForTest()

	insertCount := 0
	repo := &mockRepo{
		insertFunc: func(ctx context.Context, table string, rows []map[string]interface{}) error {
			insertCount++
			assert.Len(t, rows, 2)
			return nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	entries := []CacheEntry{
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.5, Std: 1.0, Min: 0.0, Max: 1.0},
		},
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "BMI", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.6, Std: 1.1, Min: 0.1, Max: 1.1},
		},
	}

	err := cache.StoreBatch(context.Background(), entries)
	assert.NoError(t, err)
	assert.Equal(t, 1, insertCount)
}

func TestRepositoryCache_StoreBatch_EmptyEntries(t *testing.T) {
	defer config.ResetForTest()

	repo := &mockRepo{}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	err := cache.StoreBatch(context.Background(), []CacheEntry{})
	assert.NoError(t, err)
}

func TestRepositoryCache_StoreBatch_SingleBatch(t *testing.T) {
	defer config.ResetForTest()

	insertCount := 0
	repo := &mockRepo{
		insertFunc: func(ctx context.Context, table string, rows []map[string]interface{}) error {
			insertCount++
			assert.Len(t, rows, 2) // Bulk insert
			return nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	entries := []CacheEntry{
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.5, Std: 1.0, Min: 0.0, Max: 1.0},
		},
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "BMI", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.6, Std: 1.1, Min: 0.1, Max: 1.1},
		},
	}

	err := cache.StoreBatch(context.Background(), entries)
	assert.NoError(t, err)
	assert.Equal(t, 1, insertCount) // Should make single bulk insert
}

func TestRepositoryCache_StoreBatch_BatchSizeLimit(t *testing.T) {
	config.SetForTest(CacheBatchSizeKey, 2)
	defer config.ResetForTest()

	insertCount := 0
	repo := &mockRepo{
		insertFunc: func(ctx context.Context, table string, rows []map[string]interface{}) error {
			insertCount++
			assert.LessOrEqual(t, len(rows), 2)
			return nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	entries := []CacheEntry{
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.5, Std: 1.0, Min: 0.0, Max: 1.0},
		},
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "BMI", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.6, Std: 1.1, Min: 0.1, Max: 1.1},
		},
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "Weight", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.7, Std: 1.2, Min: 0.2, Max: 1.2},
		},
	}

	err := cache.StoreBatch(context.Background(), entries)
	assert.NoError(t, err)
	assert.Equal(t, 2, insertCount) // Should make 2 batch inserts
}

func TestRepositoryCache_StoreBatch_InvalidStats(t *testing.T) {
	defer config.ResetForTest()

	repo := &mockRepo{
		insertFunc: func(ctx context.Context, table string, rows []map[string]interface{}) error {
			return nil
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	entries := []CacheEntry{
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.5, Std: -1.0, Min: 0.0, Max: 1.0}, // Invalid std
		},
	}

	err := cache.StoreBatch(context.Background(), entries)
	assert.Error(t, err)
}

func TestRepositoryCache_StoreBatch_InsertError(t *testing.T) {
	defer config.ResetForTest()

	repo := &mockRepo{
		insertFunc: func(ctx context.Context, table string, rows []map[string]interface{}) error {
			return errors.New("insert error")
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	entries := []CacheEntry{
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.5, Std: 1.0, Min: 0.0, Max: 1.0},
		},
	}

	err := cache.StoreBatch(context.Background(), entries)
	assert.Error(t, err)
}

func TestRepositoryCache_StoreBatch_PartialBatchFailure(t *testing.T) {
	config.SetForTest(CacheBatchSizeKey, 2)
	defer config.ResetForTest()

	insertCount := 0
	repo := &mockRepo{
		insertFunc: func(ctx context.Context, table string, rows []map[string]interface{}) error {
			insertCount++
			if insertCount == 1 {
				return nil // First batch succeeds
			}
			return errors.New("second batch fails")
		},
	}
	cache := &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}

	entries := []CacheEntry{
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "Height", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.5, Std: 1.0, Min: 0.0, Max: 1.0},
		},
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "BMI", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.6, Std: 1.1, Min: 0.1, Max: 1.1},
		},
		{
			Request: StatsRequest{Ancestry: "EUR", Trait: "Weight", ModelID: "test_model"},
			Stats:   &reference_stats.ReferenceStats{Mean: 0.7, Std: 1.2, Min: 0.2, Max: 1.2},
		},
	}

	err := cache.StoreBatch(context.Background(), entries)
	assert.Error(t, err)
	assert.Equal(t, 2, insertCount) // Should attempt both batches
}
