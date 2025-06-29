package reference

import (
	"context"
	"errors"
	"fmt"
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

	// Create config file with test values
	configContent := `{
		"gcp": {
			"data_project": "test-data-project",
			"billing_project": "test-billing-project",
			"cache_project": "test-cache-project"
		},
		"bigquery": {
			"cache_dataset": "test-cache-dataset"
		},
		"reference": {
			"model_table": "model_table",
			"allele_freq_table": "allele_freq_table"
		}
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		panic(err)
	}

	// Run tests
	os.Exit(m.Run())
}

func TestReferenceService_LoadModel_Success(t *testing.T) {
	mockModelRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			assert.Contains(t, query, "WHERE trait = ?")
			assert.Equal(t, "Height", args[0])
			return []map[string]interface{}{
				{"rsid": "rs123", "beta": 0.5, "risk_allele": "A", "chr": "1", "chr_pos": int64(1000), "ref": "A", "alt": "G"},
			}, nil
		},
	}

	service, err := NewReferenceService(&mockRepo{}, mockModelRepo, &mockCache{})
	assert.NoError(t, err)

	model, err := service.LoadModel(context.Background(), "Height")
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, "Height", model.ID)
	assert.Len(t, model.Variants, 1)
	assert.Equal(t, "1:1000:A:G", model.Variants[0].ID)
}

func TestReferenceService_LoadModel_DBError(t *testing.T) {
	mockModelRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return nil, errors.New("db error")
		},
	}
	service, err := NewReferenceService(&mockRepo{}, mockModelRepo, &mockCache{})
	assert.NoError(t, err)

	model, err := service.LoadModel(context.Background(), "Height")
	assert.Error(t, err)
	assert.Nil(t, model)
}

func TestReferenceService_LoadModel_NoRows(t *testing.T) {
	mockModelRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{}, nil
		},
	}
	service, err := NewReferenceService(&mockRepo{}, mockModelRepo, &mockCache{})
	assert.NoError(t, err)

	model, err := service.LoadModel(context.Background(), "Height")
	assert.Error(t, err)
	assert.Nil(t, model)
}

func TestReferenceService_GetAlleleFrequenciesForTraits_Success(t *testing.T) {
	mockGnomadRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"chrom": "1", "pos": int64(1000), "ref": "A", "alt": "G", "AF_nfe": 0.2},
				{"chrom": "2", "pos": int64(2000), "ref": "C", "alt": "T", "AF_nfe": 0.3},
			}, nil
		},
	}
	service, err := NewReferenceService(mockGnomadRepo, &mockRepo{}, &mockCache{})
	assert.NoError(t, err)

	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

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
	_, err := ancestry.New("INVALID", "")
	assert.Error(t, err)
}

func TestReferenceService_GetAlleleFrequenciesForTraits_NoFrequencyData(t *testing.T) {
	mockGnomadRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"chrom": "1", "pos": int64(1000), "ref": "A", "alt": "G", "AF_nfe": 0.0},
			}, nil
		},
	}
	service, err := NewReferenceService(mockGnomadRepo, &mockRepo{}, &mockCache{})
	assert.NoError(t, err)

	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	traitVariants := map[string][]model.Variant{
		"Height": {{ID: "1:1000:A:G"}},
	}
	results, err := service.GetAlleleFrequenciesForTraits(context.Background(), traitVariants, eur)
	assert.NoError(t, err)
	assert.Empty(t, results["Height"])
}

func TestReferenceService_GetReferenceStats_CacheHit(t *testing.T) {
	cache := &mockCache{
		getFunc: func(ctx context.Context, req reference_cache.StatsRequest) (*reference_stats.ReferenceStats, error) {
			assert.Equal(t, "Height", req.Trait)
			assert.Equal(t, "Height", req.ModelID)
			return &reference_stats.ReferenceStats{Mean: 0.5, Std: 1.0}, nil
		},
	}
	service, err := NewReferenceService(&mockRepo{}, &mockRepo{}, cache)
	assert.NoError(t, err)

	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	stats, err := service.GetReferenceStats(context.Background(), eur, "Height")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0.5, stats.Mean)
}

func TestReferenceService_GetReferenceStats_CacheMissAndCompute(t *testing.T) {
	mockGnomadRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"chrom": "1", "pos": int64(123), "ref": "A", "alt": "G", "AF_nfe": 0.1},
			}, nil
		},
	}
	mockModelRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"rsid": "rs123", "beta": 0.5, "risk_allele": "A", "chr": "1", "chr_pos": int64(123), "ref": "A", "alt": "G"},
			}, nil
		},
	}
	cache := &mockCache{
		getFunc: func(ctx context.Context, req reference_cache.StatsRequest) (*reference_stats.ReferenceStats, error) {
			return nil, errors.New("cache miss")
		},
		storeFunc: func(ctx context.Context, req reference_cache.StatsRequest, stats *reference_stats.ReferenceStats) error {
			assert.Equal(t, "Height", req.Trait)
			assert.Equal(t, "Height", req.ModelID)
			return nil
		},
	}

	service, err := NewReferenceService(mockGnomadRepo, mockModelRepo, cache)
	assert.NoError(t, err)

	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	stats, err := service.GetReferenceStats(context.Background(), eur, "Height")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.True(t, stats.Mean > 0)
}

func TestReferenceService_GetAlleleFrequenciesForTraits_EmptyInput(t *testing.T) {
	service, err := NewReferenceService(&mockRepo{}, &mockRepo{}, &mockCache{})
	assert.NoError(t, err)

	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	results, err := service.GetAlleleFrequenciesForTraits(context.Background(), make(map[string][]model.Variant), eur)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestReferenceService_GetAlleleFrequenciesForTraits_VariantDeduplication(t *testing.T) {
	callCount := 0
	mockGnomadRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			callCount++
			assert.Equal(t, 1, strings.Count(query, "chrom = ?"))
			return []map[string]interface{}{
				{"chrom": "1", "pos": int64(1000), "ref": "A", "alt": "G", "AF_nfe": 0.2},
			}, nil
		},
	}
	service, err := NewReferenceService(mockGnomadRepo, &mockRepo{}, &mockCache{})
	assert.NoError(t, err)

	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	traitVariants := map[string][]model.Variant{
		"Height": {{ID: "1:1000:A:G"}},
		"BMI":    {{ID: "1:1000:A:G"}},
	}
	results, err := service.GetAlleleFrequenciesForTraits(context.Background(), traitVariants, eur)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Equal(t, 0.2, results["Height"]["1:1000:A:G"])
	assert.Equal(t, 0.2, results["BMI"]["1:1000:A:G"])
}

func TestReferenceService_GetAlleleFrequenciesForTraits_DBError(t *testing.T) {
	mockGnomadRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return nil, errors.New("db error")
		},
	}
	service, err := NewReferenceService(mockGnomadRepo, &mockRepo{}, &mockCache{})
	assert.NoError(t, err)

	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	traitVariants := map[string][]model.Variant{
		"Height": {{ID: "1:1000:A:G"}},
	}
	_, err = service.GetAlleleFrequenciesForTraits(context.Background(), traitVariants, eur)
	assert.Error(t, err)
}

func TestReferenceService_GetReferenceStatsBatch_Success(t *testing.T) {
	mockGnomadRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"chrom": "1", "pos": int64(100), "ref": "A", "alt": "G", "AF_nfe": 0.1},
				{"chrom": "2", "pos": int64(200), "ref": "C", "alt": "T", "AF_nfe": 0.2},
			}, nil
		},
	}
	mockModelRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			trait := args[0].(string)
			if trait == "Height" {
				return []map[string]interface{}{
					{"rsid": "rs1", "beta": 0.1, "risk_allele": "G", "chr": "1", "chr_pos": int64(100), "ref": "A", "alt": "G"},
				}, nil
			}
			if trait == "BMI" {
				return []map[string]interface{}{
					{"rsid": "rs2", "beta": -0.2, "risk_allele": "T", "chr": "2", "chr_pos": int64(200), "ref": "C", "alt": "T"},
				}, nil
			}
			return nil, fmt.Errorf("unexpected trait: %s", trait)
		},
	}
	service, err := NewReferenceService(mockGnomadRepo, mockModelRepo, &mockCache{})
	assert.NoError(t, err)

	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	requests := []ReferenceStatsRequest{
		{Ancestry: eur, Trait: "Height"},
		{Ancestry: eur, Trait: "BMI"},
	}

	results, err := service.GetReferenceStatsBatch(context.Background(), requests)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	heightKey := fmt.Sprintf("%s|%s|%s", "EUR", "Height", "Height")
	bmiKey := fmt.Sprintf("%s|%s|%s", "EUR", "BMI", "BMI")

	assert.Contains(t, results, heightKey)
	assert.Contains(t, results, bmiKey)
	assert.NotNil(t, results[heightKey])
	assert.NotNil(t, results[bmiKey])
}

func TestReferenceService_GetReferenceStatsBatch_EmptyInput(t *testing.T) {
	service, err := NewReferenceService(&mockRepo{}, &mockRepo{}, &mockCache{})
	assert.NoError(t, err)

	results, err := service.GetReferenceStatsBatch(context.Background(), []ReferenceStatsRequest{})
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestReferenceService_GetReferenceStatsBatch_ModelLoadError(t *testing.T) {
	mockModelRepo := &mockRepo{
		queryFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return nil, errors.New("db error")
		},
	}
	service, err := NewReferenceService(&mockRepo{}, mockModelRepo, &mockCache{})
	assert.NoError(t, err)

	eur, err := ancestry.New("EUR", "")
	assert.NoError(t, err)

	requests := []ReferenceStatsRequest{
		{Ancestry: eur, Trait: "Height"},
	}
	_, err = service.GetReferenceStatsBatch(context.Background(), requests)
	assert.Error(t, err)
}
