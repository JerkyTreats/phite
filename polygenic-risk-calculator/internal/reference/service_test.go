package reference

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"phite.io/polygenic-risk-calculator/internal/config"
	reference_cache "phite.io/polygenic-risk-calculator/internal/reference/cache"
	model "phite.io/polygenic-risk-calculator/internal/reference/model"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

type mockRepo struct {
	queryFunc func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
}

func (m *mockRepo) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	return m.queryFunc(ctx, query, args...)
}
func (m *mockRepo) Insert(ctx context.Context, table string, rows []map[string]interface{}) error {
	return nil
}
func (m *mockRepo) TestConnection(ctx context.Context, table string) error { return nil }
func (m *mockRepo) ValidateTable(ctx context.Context, table string, requiredColumns []string) error {
	return nil
}

type mockCache struct {
	getFunc   func(ctx context.Context, req reference_cache.StatsRequest) (*reference_stats.ReferenceStats, error)
	storeFunc func(ctx context.Context, req reference_cache.StatsRequest, stats *reference_stats.ReferenceStats) error
}

func (m *mockCache) Get(ctx context.Context, req reference_cache.StatsRequest) (*reference_stats.ReferenceStats, error) {
	return m.getFunc(ctx, req)
}
func (m *mockCache) Store(ctx context.Context, req reference_cache.StatsRequest, stats *reference_stats.ReferenceStats) error {
	return m.storeFunc(ctx, req, stats)
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
		"reference": {
			"model_table": "model_table",
			"allele_freq_table": "allele_freq_table",
			"column_mapping": {
				"model_id": "model_id",
				"id": "id",
				"effect_weight": "effect_weight",
				"effect_allele": "effect_allele",
				"other_allele": "other_allele",
				"effect_freq": "effect_freq"
			},
			"ancestry_mapping": {
				"EUR": "eur_freq"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		panic(err)
	}

	// Run tests
	os.Exit(m.Run())
}

func TestReferenceService_LoadModel_Success(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"id": "1:1000:A:G", "effect_weight": 0.5, "effect_allele": "A", "other_allele": "G", "effect_freq": 0.1},
			}, nil
		},
	}
	service := &ReferenceService{
		gnomadDB:        repo,
		referenceCache:  &mockCache{},
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
		ancestryMapping: config.GetStringMapString("reference.ancestry_mapping"),
	}
	model, err := service.LoadModel(context.Background(), "test_model")
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "test_model", model.ID)
	assert.Len(t, model.Variants, 1)
}

func TestReferenceService_LoadModel_DBError(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return nil, errors.New("db error")
		},
	}
	service := &ReferenceService{
		gnomadDB:        repo,
		referenceCache:  &mockCache{},
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
		ancestryMapping: config.GetStringMapString("reference.ancestry_mapping"),
	}
	model, err := service.LoadModel(context.Background(), "test_model")
	assert.Error(t, err)
	assert.Nil(t, model)
}

func TestReferenceService_LoadModel_NoRows(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{}, nil
		},
	}
	service := &ReferenceService{
		gnomadDB:        repo,
		referenceCache:  &mockCache{},
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
		ancestryMapping: config.GetStringMapString("reference.ancestry_mapping"),
	}
	model, err := service.LoadModel(context.Background(), "test_model")
	assert.Error(t, err)
	assert.Nil(t, model)
}

func TestReferenceService_GetAlleleFrequencies_Success(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"chrom": "1", "pos": int64(1000), "ref": "A", "alt": "G", "freq": 0.2},
			}, nil
		},
	}
	service := &ReferenceService{
		gnomadDB:        repo,
		referenceCache:  &mockCache{},
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
		ancestryMapping: config.GetStringMapString("reference.ancestry_mapping"),
	}
	variants := []model.Variant{{ID: "1:1000:A:G"}}
	freqs, err := service.GetAlleleFrequencies(context.Background(), variants, "EUR")
	assert.NoError(t, err)
	assert.Equal(t, 0.2, freqs["1:1000:A:G"])
}

func TestReferenceService_GetAlleleFrequencies_UnsupportedAncestry(t *testing.T) {
	service := &ReferenceService{
		gnomadDB:        &mockRepo{},
		referenceCache:  &mockCache{},
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
		ancestryMapping: config.GetStringMapString("reference.ancestry_mapping"),
	}
	variants := []model.Variant{{ID: "1:1000:A:G"}}
	_, err := service.GetAlleleFrequencies(context.Background(), variants, "AFR")
	assert.Error(t, err)
}

func TestReferenceService_GetReferenceStats_CacheHit(t *testing.T) {
	cache := &mockCache{
		getFunc: func(ctx context.Context, req reference_cache.StatsRequest) (*reference_stats.ReferenceStats, error) {
			return &reference_stats.ReferenceStats{Mean: 0.5, Std: 1.0, Min: 0.0, Max: 1.0}, nil
		},
	}
	service := &ReferenceService{
		gnomadDB:        &mockRepo{},
		referenceCache:  cache,
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
		ancestryMapping: config.GetStringMapString("reference.ancestry_mapping"),
	}
	stats, err := service.GetReferenceStats(context.Background(), "EUR", "Height", "test_model")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0.5, stats.Mean)
}

func TestReferenceService_GetReferenceStats_CacheMissAndCompute(t *testing.T) {
	cache := &mockCache{
		getFunc: func(ctx context.Context, req reference_cache.StatsRequest) (*reference_stats.ReferenceStats, error) {
			return nil, nil // cache miss
		},
		storeFunc: func(ctx context.Context, req reference_cache.StatsRequest, stats *reference_stats.ReferenceStats) error {
			return nil
		},
	}
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			// For LoadModel and GetAlleleFrequencies
			if strings.Contains(query, "model_table") {
				return []map[string]interface{}{
					{"id": "1:1000:A:G", "effect_weight": 0.5, "effect_allele": "A", "other_allele": "G", "effect_freq": 0.1},
				}, nil
			}
			return []map[string]interface{}{
				{"chrom": "1", "pos": int64(1000), "ref": "A", "alt": "G", "freq": 0.2},
			}, nil
		},
	}
	service := &ReferenceService{
		gnomadDB:        repo,
		referenceCache:  cache,
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
		ancestryMapping: config.GetStringMapString("reference.ancestry_mapping"),
	}
	stats, err := service.GetReferenceStats(context.Background(), "EUR", "Height", "test_model")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}
