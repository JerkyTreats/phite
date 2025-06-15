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
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"

	"phite.io/polygenic-risk-calculator/internal/ancestry"
	"phite.io/polygenic-risk-calculator/internal/model"
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
	getFunc        func(ctx context.Context, req reference_cache.StatsRequest) (*reference_stats.ReferenceStats, error)
	storeFunc      func(ctx context.Context, req reference_cache.StatsRequest, stats *reference_stats.ReferenceStats) error
	getBatchFunc   func(ctx context.Context, reqs []reference_cache.StatsRequest) (map[string]*reference_stats.ReferenceStats, error)
	storeBatchFunc func(ctx context.Context, entries []reference_cache.CacheEntry) error
}

func (m *mockCache) Get(ctx context.Context, req reference_cache.StatsRequest) (*reference_stats.ReferenceStats, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockCache) Store(ctx context.Context, req reference_cache.StatsRequest, stats *reference_stats.ReferenceStats) error {
	if m.storeFunc != nil {
		return m.storeFunc(ctx, req, stats)
	}
	return nil
}

func (m *mockCache) GetBatch(ctx context.Context, reqs []reference_cache.StatsRequest) (map[string]*reference_stats.ReferenceStats, error) {
	if m.getBatchFunc != nil {
		return m.getBatchFunc(ctx, reqs)
	}
	return make(map[string]*reference_stats.ReferenceStats), nil
}

func (m *mockCache) StoreBatch(ctx context.Context, entries []reference_cache.CacheEntry) error {
	if m.storeBatchFunc != nil {
		return m.storeBatchFunc(ctx, entries)
	}
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

	// Create config file with test values (removed ancestry_mapping)
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
	}
	model, err := service.LoadModel(context.Background(), "test_model")
	assert.Error(t, err)
	assert.Nil(t, model)
}

func TestReferenceService_GetAlleleFrequenciesForTraits_Success(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"chrom": "1", "pos": int64(1000), "ref": "A", "alt": "G", "AF_nfe": 0.2},
				{"chrom": "2", "pos": int64(2000), "ref": "C", "alt": "T", "AF_nfe": 0.3},
			}, nil
		},
	}
	service := &ReferenceService{
		gnomadDB:        repo,
		referenceCache:  &mockCache{},
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
	}

	// Create ancestry object for testing
	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	// Test multi-trait scenario
	traitVariants := map[string][]model.Variant{
		"Height": {{ID: "1:1000:A:G"}},
		"BMI":    {{ID: "2:2000:C:T"}},
	}

	results, err := service.GetAlleleFrequenciesForTraits(context.Background(), traitVariants, eur)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 0.2, results["Height"]["1:1000:A:G"])
	assert.Equal(t, 0.3, results["BMI"]["2:2000:C:T"])
}

func TestReferenceService_GetAlleleFrequenciesForTraits_UnsupportedAncestry(t *testing.T) {
	// Test with invalid ancestry - should fail during creation
	_, err := ancestry.New("INVALID", "")
	assert.Error(t, err) // This should fail, confirming validation works
}

func TestReferenceService_GetAlleleFrequenciesForTraits_NoFrequencyData(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"chrom": "1", "pos": int64(1000), "ref": "A", "alt": "G", "AF_nfe": 0.0}, // Zero frequency
			}, nil
		},
	}
	service := &ReferenceService{
		gnomadDB:        repo,
		referenceCache:  &mockCache{},
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
	}

	// Create ancestry object for testing
	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	traitVariants := map[string][]model.Variant{
		"Height": {{ID: "1:1000:A:G"}},
	}
	results, err := service.GetAlleleFrequenciesForTraits(context.Background(), traitVariants, eur)
	assert.NoError(t, err)
	// Should return empty trait map when no frequency data available
	assert.Empty(t, results["Height"])
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
	}

	// Create ancestry object for testing
	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	stats, err := service.GetReferenceStats(context.Background(), eur, "Height", "test_model")
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
				{"chrom": "1", "pos": int64(1000), "ref": "A", "alt": "G", "AF_nfe": 0.2},
			}, nil
		},
	}
	service := &ReferenceService{
		gnomadDB:        repo,
		referenceCache:  cache,
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
	}

	// Create ancestry object for testing
	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	stats, err := service.GetReferenceStats(context.Background(), eur, "Height", "test_model")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestReferenceService_GetAlleleFrequenciesForTraits_EmptyInput(t *testing.T) {
	service := &ReferenceService{
		gnomadDB:        &mockRepo{},
		referenceCache:  &mockCache{},
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
	}

	// Create ancestry object for testing
	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	// Test empty trait variants map
	results, err := service.GetAlleleFrequenciesForTraits(context.Background(), map[string][]model.Variant{}, eur)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestReferenceService_GetAlleleFrequenciesForTraits_VariantDeduplication(t *testing.T) {
	queryCalled := 0
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			queryCalled++
			// Should only be called once despite duplicate variants across traits
			return []map[string]interface{}{
				{"chrom": "1", "pos": int64(1000), "ref": "A", "alt": "G", "AF_nfe": 0.2},
			}, nil
		},
	}
	service := &ReferenceService{
		gnomadDB:        repo,
		referenceCache:  &mockCache{},
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
	}

	// Create ancestry object for testing
	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	// Test with same variant in multiple traits (should deduplicate)
	traitVariants := map[string][]model.Variant{
		"Height":   {{ID: "1:1000:A:G"}},
		"BMI":      {{ID: "1:1000:A:G"}}, // Same variant
		"Diabetes": {{ID: "1:1000:A:G"}}, // Same variant again
	}

	results, err := service.GetAlleleFrequenciesForTraits(context.Background(), traitVariants, eur)
	assert.NoError(t, err)
	assert.Len(t, results, 3) // All three traits should get results
	assert.Equal(t, 0.2, results["Height"]["1:1000:A:G"])
	assert.Equal(t, 0.2, results["BMI"]["1:1000:A:G"])
	assert.Equal(t, 0.2, results["Diabetes"]["1:1000:A:G"])

	// Verify query was only called once (deduplication worked)
	assert.Equal(t, 1, queryCalled)
}

func TestReferenceService_GetAlleleFrequenciesForTraits_DBError(t *testing.T) {
	repo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return nil, errors.New("database error")
		},
	}
	service := &ReferenceService{
		gnomadDB:        repo,
		referenceCache:  &mockCache{},
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
	}

	// Create ancestry object for testing
	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	traitVariants := map[string][]model.Variant{
		"Height": {{ID: "1:1000:A:G"}},
	}

	results, err := service.GetAlleleFrequenciesForTraits(context.Background(), traitVariants, eur)
	assert.Error(t, err)
	assert.Nil(t, results)
}
