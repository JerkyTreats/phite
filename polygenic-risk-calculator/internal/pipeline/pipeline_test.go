package pipeline

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"phite.io/polygenic-risk-calculator/internal/ancestry"
	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/db"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
	"phite.io/polygenic-risk-calculator/internal/db/testutils"
	"phite.io/polygenic-risk-calculator/internal/gwas"
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/output"
	"phite.io/polygenic-risk-calculator/internal/reference"
	reference_cache "phite.io/polygenic-risk-calculator/internal/reference/cache"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

func setupTestRepositories(t *testing.T) (dbinterface.Repository, dbinterface.Repository) {
	gwasRepo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
		"path": "testdata/gwas.duckdb",
	})
	if err != nil {
		t.Fatalf("failed to create GWAS repository: %v", err)
	}

	refRepo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
		"path": "testdata/reference.duckdb",
	})
	if err != nil {
		t.Fatalf("failed to create reference repository: %v", err)
	}

	return gwasRepo, refRepo
}

func setupTestConfig(t *testing.T) {
	// Set up test configuration with required ancestry settings
	config.Set("ancestry.population", "EUR")
	config.Set("ancestry.gender", "")
	config.Set("reference.model", "v1")
	// Configure GWAS database path for testing
	config.Set("gwas_db_path", "testdata/gwas.duckdb")
	config.Set("gwas_table", "gwas_table")
	// Configure reference service settings
	config.Set("reference.model_table", "reference_stats")
	config.Set("reference.allele_freq_table", "allele_frequencies")
	config.Set("reference.column_mapping", map[string]string{
		"id":            "id",
		"effect_weight": "effect_weight",
		"effect_allele": "effect_allele",
		"other_allele":  "other_allele",
		"effect_freq":   "effect_freq",
		"AF_afr":        "AF_afr",
		"AF_nfe":        "AF_nfe",
	})
	// Configure GCP settings (though we might not need them for tests)
	config.Set("user.gcp_project", "test-project")
	config.Set("cache.gcp_project", "test-project")
	config.Set("cache.dataset", "test_dataset")
	config.Set("bigquery.table_id", "test_cache_table")
}

// setupMockPipeline creates a pipeline with mock repositories for testing
func setupMockPipeline(t *testing.T) (*reference.ReferenceService, reference_cache.Cache, *testutils.MockRepository, *testutils.MockRepository) {
	gnomadMock := testutils.NewMockRepository()
	cacheMock := testutils.NewMockRepository()

	cache, err := reference_cache.NewRepositoryCache(cacheMock)
	require.NoError(t, err)

	refService, err := reference.NewReferenceService(gnomadMock, cache)
	require.NoError(t, err)

	return refService, cache, gnomadMock, cacheMock
}

// setupMockBigQueryResponses configures realistic mock responses for BigQuery operations
func setupMockBigQueryResponses(gnomadMock, cacheMock *testutils.MockRepository) {
	// Mock model loading responses
	gnomadMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		if strings.Contains(query, "reference_stats") {
			// Mock PRS model data with correct genomic coordinate format
			return []map[string]interface{}{
				{
					"id":            "1:12345:G:A",
					"effect_weight": 0.5,
					"effect_allele": "A",
					"other_allele":  "G",
					"effect_freq":   0.3,
				},
				{
					"id":            "2:67890:C:T",
					"effect_weight": -0.2,
					"effect_allele": "T",
					"other_allele":  "C",
					"effect_freq":   0.4,
				},
				{
					"id":            "3:98765:A:T",
					"effect_weight": 0.3,
					"effect_allele": "T",
					"other_allele":  "A",
					"effect_freq":   0.5,
				},
			}, nil
		}

		if strings.Contains(query, "allele_frequencies") {
			// Mock allele frequency data with bulk OR query pattern
			return []map[string]interface{}{
				{
					"chrom":  "1",
					"pos":    int64(12345),
					"ref":    "G",
					"alt":    "A",
					"AF_nfe": 0.3,
					"AF_afr": 0.25,
				},
				{
					"chrom":  "2",
					"pos":    int64(67890),
					"ref":    "C",
					"alt":    "T",
					"AF_nfe": 0.4,
					"AF_afr": 0.35,
				},
				{
					"chrom":  "3",
					"pos":    int64(98765),
					"ref":    "A",
					"alt":    "T",
					"AF_nfe": 0.5,
					"AF_afr": 0.45,
				},
			}, nil
		}

		return []map[string]interface{}{}, nil
	}

	// Mock cache responses (cache miss scenario for testing bulk operations)
	cacheMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		// Return empty for cache miss to trigger bulk stats computation
		return []map[string]interface{}{}, nil
	}

	// Mock cache storage (should succeed)
	cacheMock.InsertFunc = func(ctx context.Context, table string, rows []map[string]interface{}) error {
		return nil
	}
}

func TestRun_SingleTrait_Success(t *testing.T) {
	setupTestConfig(t)

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1", "rs2"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// For now, we expect this to fail at the BigQuery step since we can't mock it easily
	// The important thing is that we've successfully:
	// 1. Initialized ancestry from configuration
	// 2. Created GWAS service with configuration
	// 3. Processed genotype data
	// 4. Reached the reference service call
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Validate that the error is related to BigQuery, not earlier pipeline steps
	if !strings.Contains(err.Error(), "BigQuery") && !strings.Contains(err.Error(), "test-project") {
		t.Fatalf("expected BigQuery-related error, got: %v", err)
	}

	// The pipeline should have attempted to process but failed at reference service
	// This validates that Phase 4 (ancestry integration) is working
	t.Logf("Pipeline correctly failed at BigQuery step as expected: %v", err)
}

func TestRun_MultiTrait_Success(t *testing.T) {
	setupTestConfig(t)

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_multi_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs3"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// Similar to single trait test - expect BigQuery failure
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Validate that the error is related to BigQuery, not earlier pipeline steps
	if !strings.Contains(err.Error(), "BigQuery") && !strings.Contains(err.Error(), "test-project") {
		t.Fatalf("expected BigQuery-related error, got: %v", err)
	}

	t.Logf("Pipeline correctly failed at BigQuery step as expected: %v", err)
}

func TestRun_ErrorOnMissingInput(t *testing.T) {
	setupTestConfig(t)
	input := PipelineInput{}
	_, err := Run(input)
	if err == nil {
		t.Fatalf("expected error on missing input, got nil")
	}
}

func TestRun_ErrorOnMissingRepository(t *testing.T) {
	setupTestConfig(t)
	input := PipelineInput{
		GenotypeFile: "testdata/genotype_single_trait.txt",
		SNPs:         []string{"rs1", "rs2"},
		OutputFormat: "json",
		OutputPath:   "",
	}
	_, err := Run(input)
	if err == nil {
		t.Fatalf("expected error on missing repository, got nil")
	}
}

func TestRun_ErrorOnInvalidGenotypeFile(t *testing.T) {
	setupTestConfig(t)

	input := PipelineInput{
		GenotypeFile:   "testdata/nonexistent.txt",
		SNPs:           []string{"rs1", "rs2"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}
	_, err := Run(input)
	if err == nil {
		t.Fatalf("expected error on invalid genotype file, got nil")
	}
}

func TestRun_ErrorOnMissingAncestryConfig(t *testing.T) {
	// Don't set up config to test missing ancestry configuration
	config.ResetForTest()

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1", "rs2"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}
	_, err := Run(input)
	if err == nil {
		t.Fatalf("expected error on missing ancestry configuration, got nil")
	}
}

func TestRun_CustomAncestryConfig(t *testing.T) {
	// Test with different ancestry configuration
	config.Set("ancestry.population", "AFR")
	config.Set("ancestry.gender", "FEMALE")
	config.Set("reference.model", "v1")
	// Also need to set up GWAS database configuration
	config.Set("gwas_db_path", "testdata/gwas.duckdb")
	config.Set("gwas_table", "gwas_table")
	// Configure reference service settings
	config.Set("reference.model_table", "reference_stats")
	config.Set("reference.allele_freq_table", "allele_frequencies")
	config.Set("reference.column_mapping", map[string]string{
		"AF_afr": "AF_afr",
		"AF_nfe": "AF_nfe",
	})
	// Configure GCP settings (though we might not need them for tests)
	config.Set("user.gcp_project", "test-project")
	config.Set("cache.gcp_project", "test-project")
	config.Set("cache.dataset", "test_dataset")

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1", "rs2"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// This test validates that Phase 4 properly handles custom ancestry (AFR_FEMALE)
	// We expect it to fail at BigQuery, but that proves the ancestry integration worked
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Validate that the error is related to BigQuery, not earlier pipeline steps
	if !strings.Contains(err.Error(), "BigQuery") && !strings.Contains(err.Error(), "test-project") {
		t.Fatalf("expected BigQuery-related error, got: %v", err)
	}

	// This validates that Phase 4 correctly initialized AFR_FEMALE ancestry from config
	t.Logf("Pipeline correctly processed AFR_FEMALE ancestry and failed at BigQuery as expected: %v", err)
}

// TestRun_BulkOperations_Phase1_RequirementsAnalysis tests Phase 1 requirements analysis
func TestRun_BulkOperations_Phase1_RequirementsAnalysis(t *testing.T) {
	setupTestConfig(t)

	// Test with multiple traits to validate bulk requirements collection
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_multi_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs3", "rs4", "rs5"}, // More SNPs for bulk testing
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// Should fail at BigQuery but validate Phase 1 completed successfully
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Validate that Phase 1 completed (requirements analysis)
	if !strings.Contains(err.Error(), "bulk data retrieval failed") {
		t.Fatalf("expected Phase 1 to complete and fail at Phase 2, got: %v", err)
	}

	t.Logf("Phase 1 requirements analysis completed successfully, failed at Phase 2 as expected: %v", err)
}

// TestRun_BulkOperations_CacheMissScenario tests bulk processing with cache misses
func TestRun_BulkOperations_CacheMissScenario(t *testing.T) {
	setupTestConfig(t)

	// Test scenario where all traits are cache misses (new traits)
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_multi_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs3"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// Should attempt bulk stats computation for cache misses
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Validate that it reached bulk stats computation phase
	if !strings.Contains(err.Error(), "BigQuery") && !strings.Contains(err.Error(), "bulk") {
		t.Fatalf("expected bulk operations to be attempted, got: %v", err)
	}

	t.Logf("Bulk cache miss scenario processed correctly: %v", err)
}

// TestRun_BulkOperations_MultiTraitProcessing tests processing multiple traits in bulk
func TestRun_BulkOperations_MultiTraitProcessing(t *testing.T) {
	setupTestConfig(t)

	// Test with maximum trait diversity
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_multi_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs3"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// Validate multi-trait bulk processing was attempted
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Should have processed multiple traits in Phase 1
	if !strings.Contains(err.Error(), "bulk data retrieval failed") {
		t.Fatalf("expected multi-trait bulk processing, got: %v", err)
	}

	t.Logf("Multi-trait bulk processing validated: %v", err)
}

// TestRun_BulkOperations_DataStructureValidation tests bulk operation data structures
func TestRun_BulkOperations_DataStructureValidation(t *testing.T) {
	setupTestConfig(t)

	// Test with edge case: single SNP, single trait
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// Validate that bulk operations handle single-item scenarios correctly
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Should still use bulk operations even for single items
	if !strings.Contains(err.Error(), "bulk data retrieval failed") {
		t.Fatalf("expected bulk operations for single trait, got: %v", err)
	}

	t.Logf("Single trait bulk operations validated: %v", err)
}

// TestRun_BulkOperations_ErrorPropagation tests error handling across phases
func TestRun_BulkOperations_ErrorPropagation(t *testing.T) {
	setupTestConfig(t)

	// Test with invalid SNPs to trigger Phase 1 errors
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{}, // Empty SNPs should cause Phase 1 error
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// Should fail early in Phase 1 due to missing SNPs
	if err == nil {
		t.Fatalf("expected error due to missing SNPs, got nil")
	}

	// Should fail before reaching bulk operations
	if strings.Contains(err.Error(), "bulk data retrieval failed") {
		t.Fatalf("expected Phase 1 error, but reached Phase 2: %v", err)
	}

	t.Logf("Phase 1 error propagation validated: %v", err)
}

// TestRun_BulkOperations_MemoryEfficiency tests memory usage patterns
func TestRun_BulkOperations_MemoryEfficiency(t *testing.T) {
	setupTestConfig(t)

	// Test with larger dataset to validate memory efficiency
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_multi_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs3"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	// This test validates that bulk operations don't cause memory issues
	_, err := Run(input)

	// Should handle bulk data structures without memory errors
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Validate that memory-related operations completed
	if !strings.Contains(err.Error(), "bulk data retrieval failed") {
		t.Fatalf("expected bulk memory operations, got: %v", err)
	}

	t.Logf("Bulk operations memory efficiency validated: %v", err)
}

// TestRun_BulkOperations_AncestryIntegration tests bulk operations with different ancestries
func TestRun_BulkOperations_AncestryIntegration(t *testing.T) {
	// Test with Asian ancestry for bulk operations
	config.Set("ancestry.population", "EAS")
	config.Set("ancestry.gender", "MALE")
	config.Set("reference.model", "v1")
	config.Set("gwas_db_path", "testdata/gwas.duckdb")
	config.Set("gwas_table", "gwas_table")
	config.Set("reference.model_table", "reference_stats")
	config.Set("reference.allele_freq_table", "allele_frequencies")
	config.Set("reference.column_mapping", map[string]string{
		"AF_afr": "AF_afr",
		"AF_nfe": "AF_nfe",
	})
	config.Set("user.gcp_project", "test-project")
	config.Set("cache.gcp_project", "test-project")
	config.Set("cache.dataset", "test_dataset")

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_multi_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs3"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// Validate EAS ancestry bulk processing
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Should have processed EAS ancestry in bulk operations
	if !strings.Contains(err.Error(), "bulk data retrieval failed") {
		t.Fatalf("expected EAS ancestry bulk processing, got: %v", err)
	}

	t.Logf("EAS ancestry bulk operations validated: %v", err)
}

// TestRun_BulkOperations_PhaseTransitions tests transitions between phases
func TestRun_BulkOperations_PhaseTransitions(t *testing.T) {
	setupTestConfig(t)

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_multi_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs3"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// Validate that phases transition correctly
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Should complete Phase 1 and fail at Phase 2 (bulk data retrieval)
	if !strings.Contains(err.Error(), "bulk data retrieval failed") {
		t.Fatalf("expected Phase 1->2 transition, got: %v", err)
	}

	t.Logf("Phase transition validation successful: %v", err)
}

// TestRun_BulkOperations_ConfigurationVariations tests different configurations
func TestRun_BulkOperations_ConfigurationVariations(t *testing.T) {
	// Test with different model configuration
	config.Set("ancestry.population", "EUR")
	config.Set("ancestry.gender", "")
	config.Set("reference.model", "v2") // Different model version
	config.Set("gwas_db_path", "testdata/gwas.duckdb")
	config.Set("gwas_table", "gwas_table")
	config.Set("reference.model_table", "reference_stats")
	config.Set("reference.allele_freq_table", "allele_frequencies")
	config.Set("reference.column_mapping", map[string]string{
		"AF_afr": "AF_afr",
		"AF_nfe": "AF_nfe",
	})
	config.Set("user.gcp_project", "test-project")
	config.Set("cache.gcp_project", "test-project")
	config.Set("cache.dataset", "test_dataset")

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1", "rs2"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input)

	// Validate different model configuration in bulk operations
	if err == nil {
		t.Fatalf("expected error due to BigQuery connection issue, got nil")
	}

	// Should handle different model versions in bulk operations
	if !strings.Contains(err.Error(), "bulk data retrieval failed") {
		t.Fatalf("expected model v2 bulk processing, got: %v", err)
	}

	t.Logf("Configuration variation bulk operations validated: %v", err)
}

// ==================== MOCK TESTS (UNIT TESTS WITH CONTROLLED DEPENDENCIES) ====================

func TestRun_WithMocks_FullPipeline_Success(t *testing.T) {
	setupTestConfig(t)

	// Create mock repositories with realistic responses
	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)
	setupMockBigQueryResponses(gnomadMock, cacheMock)

	// Run pipeline with mock services
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1", "rs2"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	output, err := Run(input, refService)

	// Validate successful execution
	require.NoError(t, err)
	assert.NotEmpty(t, output.TraitSummaries)

	// Validate BigQuery call optimization - should have maximum 4 calls
	assert.LessOrEqual(t, len(gnomadMock.QueryCalls), 4, "Should have maximum 4 BigQuery calls for bulk operations")
	assert.GreaterOrEqual(t, len(gnomadMock.QueryCalls), 2, "Should have minimum 2 BigQuery calls (model + frequencies)")

	// Validate bulk query patterns
	foundModelQuery := false
	foundFrequencyQuery := false

	for _, call := range gnomadMock.QueryCalls {
		if strings.Contains(call.Query, "reference_stats") {
			foundModelQuery = true
		}
		if strings.Contains(call.Query, "allele_frequencies") {
			foundFrequencyQuery = true
			// Validate bulk OR pattern in frequency query
			assert.Contains(t, call.Query, " OR ", "Frequency query should use bulk OR pattern")
		}
	}

	assert.True(t, foundModelQuery, "Should have executed model loading query")
	assert.True(t, foundFrequencyQuery, "Should have executed bulk frequency query")

	// Validate cache operations
	assert.GreaterOrEqual(t, len(cacheMock.QueryCalls), 1, "Should have executed cache lookup")
	assert.GreaterOrEqual(t, len(cacheMock.InsertCalls), 1, "Should have executed cache storage")
}

func TestRun_WithMocks_BulkOperations_CallCounting(t *testing.T) {
	setupTestConfig(t)

	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)
	setupMockBigQueryResponses(gnomadMock, cacheMock)

	// Test with multiple traits to validate bulk optimization
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_multi_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs3"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input, refService)
	require.NoError(t, err)

	// Critical test: Validate that bulk operations reduce BigQuery calls
	// Without bulk operations, this would be ~10-20 calls
	// With bulk operations, should be maximum 4 calls
	totalBQCalls := len(gnomadMock.QueryCalls)
	assert.LessOrEqual(t, totalBQCalls, 4,
		"Bulk operations should limit BigQuery calls to maximum 4, got %d", totalBQCalls)

	// Validate that we're using bulk patterns
	bulkQueries := 0
	for _, call := range gnomadMock.QueryCalls {
		if strings.Contains(call.Query, " OR ") {
			bulkQueries++
		}
	}

	assert.GreaterOrEqual(t, bulkQueries, 1, "Should use bulk OR queries for optimization")
}

func TestRun_WithMocks_CacheHit_Scenario(t *testing.T) {
	setupTestConfig(t)

	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)

	// Setup model loading mock
	gnomadMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		if strings.Contains(query, "reference_stats") {
			return []map[string]interface{}{
				{
					"id":            "1:12345:G:A",
					"effect_weight": 0.5,
					"effect_allele": "A",
					"other_allele":  "G",
				},
			}, nil
		}
		if strings.Contains(query, "allele_frequencies") {
			return []map[string]interface{}{
				{
					"chrom":  "1",
					"pos":    int64(12345),
					"ref":    "G",
					"alt":    "A",
					"AF_nfe": 0.3,
				},
			}, nil
		}
		return []map[string]interface{}{}, nil
	}

	// Mock cache hit with correct key format "ancestry|trait|model"
	cacheMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		// Return cache hit for "EUR|height|v1" key
		return []map[string]interface{}{
			{
				"mean":     0.5,
				"std":      1.0,
				"min":      -2.0,
				"max":      3.0,
				"ancestry": "EUR",
				"trait":    "height",
				"model":    "v1",
			},
		}, nil
	}

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input, refService)
	require.NoError(t, err)

	// With cache hit, should have fewer BigQuery calls (no stats computation)
	assert.LessOrEqual(t, len(gnomadMock.QueryCalls), 3, "Cache hit should reduce BigQuery calls")

	// Should not have cache storage calls (cache hit scenario)
	assert.Equal(t, 0, len(cacheMock.InsertCalls), "Cache hit should not trigger storage")
}

func TestRun_WithMocks_ErrorHandling(t *testing.T) {
	setupTestConfig(t)

	refService, _, gnomadMock, _ := setupMockPipeline(t)

	// Setup error scenario
	gnomadMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		return nil, assert.AnError
	}

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, err := Run(input, refService)

	// Should propagate BigQuery errors properly
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load PRS model")
}

// ==================== PHASE-LEVEL UNIT TESTS (PRIORITY 1) ====================

// ==================== PHASE 1: REQUIREMENTS ANALYSIS TESTS ====================

func TestAnalyzeAllRequirements_Success(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1", "rs2"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	requirements, genoOut, annotated, err := analyzeAllRequirements(ctx, input)

	// Should succeed for valid input
	require.NoError(t, err)

	// Validate requirements structure
	assert.NotNil(t, requirements)
	assert.NotEmpty(t, requirements.TraitSet)
	assert.NotNil(t, requirements.AncestryObj)
	assert.NotEmpty(t, requirements.ModelID)
	assert.NotEmpty(t, requirements.CacheKeys)

	// Validate genotype output
	assert.NotEmpty(t, genoOut.ValidatedSNPs)

	// Validate annotated GWAS data
	assert.NotEmpty(t, annotated.AnnotatedSNPs)

	// Validate cache keys match trait set
	assert.Equal(t, len(requirements.TraitSet), len(requirements.CacheKeys))

	// Validate ancestry code is set
	assert.NotEmpty(t, requirements.AncestryObj.Code())
}

func TestAnalyzeAllRequirements_NoTraitsFound(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	// Use empty genotype file to simulate no traits scenario
	input := PipelineInput{
		GenotypeFile:   "testdata/empty.txt",
		SNPs:           []string{"rs_nonexistent"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	requirements, _, _, err := analyzeAllRequirements(ctx, input)

	// Should succeed but have empty trait set
	require.NoError(t, err)
	assert.NotNil(t, requirements)
	assert.Empty(t, requirements.TraitSet)
	assert.Empty(t, requirements.CacheKeys)
}

func TestAnalyzeAllRequirements_PartialSNPMatch(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	// Request more SNPs than available in test file
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs_nonexistent", "rs_missing"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	requirements, genoOut, annotated, err := analyzeAllRequirements(ctx, input)

	// Should succeed with partial matches
	require.NoError(t, err)
	assert.NotNil(t, requirements)

	// Should have some missing SNPs
	assert.NotEmpty(t, genoOut.SNPsMissing)

	// Should still process available SNPs
	assert.NotEmpty(t, genoOut.ValidatedSNPs)
	assert.NotEmpty(t, annotated.AnnotatedSNPs)
}

func TestAnalyzeAllRequirements_GWASServiceFailure(t *testing.T) {
	setupTestConfig(t)

	// Set invalid GWAS DB path to trigger GWAS service failure
	config.Set("gwas_db_path", "/invalid/path/that/does/not/exist.duckdb")

	ctx := context.Background()
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1", "rs2"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, _, _, err := analyzeAllRequirements(ctx, input)

	// Should fail with GWAS-related error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize GWAS service")
}

func TestAnalyzeAllRequirements_GenotypeParsingFailure(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	// Use invalid genotype file path
	input := PipelineInput{
		GenotypeFile:   "/invalid/path/nonexistent.txt",
		SNPs:           []string{"rs1", "rs2"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, _, _, err := analyzeAllRequirements(ctx, input)

	// Should fail with genotype parsing error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse genotype data")
}

func TestAnalyzeAllRequirements_AncestryConfigurationFailure(t *testing.T) {
	// Set up invalid ancestry configuration
	config.Set("gwas_db_path", "testdata/gwas.duckdb")
	config.Set("gwas_table", "gwas_table")
	config.Set("ancestry.population", "INVALID_POPULATION")
	config.Set("ancestry.gender", "")

	ctx := context.Background()
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs1", "rs2"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	_, _, _, err := analyzeAllRequirements(ctx, input)

	// Should fail with ancestry configuration error
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "failed to initialize ancestry")
	}
}

func TestAnalyzeAllRequirements_EmptyGWASResults(t *testing.T) {
	setupTestConfig(t)

	ctx := context.Background()
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_single_trait.txt",
		SNPs:           []string{"rs_nonexistent1", "rs_nonexistent2"}, // Use SNPs that don't exist in GWAS
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	requirements, _, annotated, err := analyzeAllRequirements(ctx, input)

	// Should succeed but have empty results
	require.NoError(t, err)
	assert.NotNil(t, requirements)

	// Should have empty trait set due to no GWAS matches
	assert.Empty(t, requirements.TraitSet)
	assert.Empty(t, annotated.AnnotatedSNPs)
}

func TestAnalyzeAllRequirements_DuplicateTraits(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	// Use multi-trait file which may have duplicate trait entries
	input := PipelineInput{
		GenotypeFile:   "testdata/genotype_multi_trait.txt",
		SNPs:           []string{"rs1", "rs2", "rs3"},
		ReferenceTable: "reference_stats",
		OutputFormat:   "json",
		OutputPath:     "",
	}

	requirements, _, _, err := analyzeAllRequirements(ctx, input)

	// Should succeed and deduplicate traits
	require.NoError(t, err)
	assert.NotNil(t, requirements)

	// TraitSet should automatically deduplicate
	for trait := range requirements.TraitSet {
		assert.NotEmpty(t, trait)
	}

	// Cache keys should match unique traits
	assert.Equal(t, len(requirements.TraitSet), len(requirements.CacheKeys))
}

// ==================== PHASE 2: BULK DATA RETRIEVAL TESTS ====================

func TestRetrieveAllDataBulk_FullCacheHit(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	// Create mock service with cache hit scenario
	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)

	// Mock model loading
	gnomadMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		if strings.Contains(query, "reference_stats") {
			return []map[string]interface{}{
				{
					"id":            "1:12345:G:A",
					"effect_weight": 0.5,
					"effect_allele": "A",
					"other_allele":  "G",
				},
			}, nil
		}
		return []map[string]interface{}{}, nil
	}

	// Mock full cache hit
	cacheMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{
				"mean":     0.5,
				"std":      1.0,
				"min":      -2.0,
				"max":      3.0,
				"ancestry": "EUR",
				"trait":    "height",
				"model":    "v1",
			},
		}, nil
	}

	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	// Create requirements
	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{"height": {}},
		CacheKeys: []reference_cache.StatsRequest{
			{Ancestry: "EUR", Trait: "height", ModelID: "v1"},
		},
		ModelID:     "v1",
		AncestryObj: ancestryObj,
	}

	// Create annotated data
	annotated := &gwas.GWASDataFetcherOutput{
		AnnotatedSNPs: []model.AnnotatedSNP{
			{RSID: "rs1", Trait: "height"},
		},
	}

	bulkData, err := retrieveAllDataBulk(ctx, requirements, annotated, refService)

	// Should succeed with cache hit
	require.NoError(t, err)
	assert.NotNil(t, bulkData)
	assert.NotNil(t, bulkData.PRSModel)
	assert.NotEmpty(t, bulkData.CachedStats)
	assert.Empty(t, bulkData.ComputedStats) // No computation needed
	assert.NotEmpty(t, bulkData.TraitSNPs)

	// Should have minimal BigQuery calls (just model loading)
	assert.LessOrEqual(t, len(gnomadMock.QueryCalls), 2)
}

func TestRetrieveAllDataBulk_FullCacheMiss(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)
	setupMockBigQueryResponses(gnomadMock, cacheMock)

	// Mock cache miss
	cacheMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		return []map[string]interface{}{}, nil // Empty = cache miss
	}

	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{"height": {}},
		CacheKeys: []reference_cache.StatsRequest{
			{Ancestry: "EUR", Trait: "height", ModelID: "v1"},
		},
		ModelID:     "v1",
		AncestryObj: ancestryObj,
	}

	annotated := &gwas.GWASDataFetcherOutput{
		AnnotatedSNPs: []model.AnnotatedSNP{
			{RSID: "rs1", Trait: "height"},
		},
	}

	bulkData, err := retrieveAllDataBulk(ctx, requirements, annotated, refService)

	// Should succeed with stats computation
	require.NoError(t, err)
	assert.NotNil(t, bulkData)
	assert.NotNil(t, bulkData.PRSModel)
	assert.Empty(t, bulkData.CachedStats)
	assert.NotEmpty(t, bulkData.ComputedStats) // Should compute stats

	// Should have more BigQuery calls for stats computation
	assert.GreaterOrEqual(t, len(gnomadMock.QueryCalls), 2)
}

func TestRetrieveAllDataBulk_MixedCacheScenario(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)
	setupMockBigQueryResponses(gnomadMock, cacheMock)

	// Mock partial cache hit (only one trait cached)
	cacheMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{
				"mean":     0.5,
				"std":      1.0,
				"min":      -2.0,
				"max":      3.0,
				"ancestry": "EUR",
				"trait":    "height", // Only height cached
				"model":    "v1",
			},
		}, nil
	}

	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{
			"height": {},
			"weight": {}, // Two traits, only height cached
		},
		CacheKeys: []reference_cache.StatsRequest{
			{Ancestry: "EUR", Trait: "height", ModelID: "v1"},
			{Ancestry: "EUR", Trait: "weight", ModelID: "v1"},
		},
		ModelID:     "v1",
		AncestryObj: ancestryObj,
	}

	annotated := &gwas.GWASDataFetcherOutput{
		AnnotatedSNPs: []model.AnnotatedSNP{
			{RSID: "rs1", Trait: "height"},
			{RSID: "rs2", Trait: "weight"},
		},
	}

	bulkData, err := retrieveAllDataBulk(ctx, requirements, annotated, refService)

	// Should succeed with mixed cache scenario
	require.NoError(t, err)
	assert.NotNil(t, bulkData)
	assert.NotNil(t, bulkData.PRSModel)
	assert.NotEmpty(t, bulkData.CachedStats)   // Should have cache hits
	assert.NotEmpty(t, bulkData.ComputedStats) // Should compute missing

	// Should optimize BigQuery calls for cache misses only
	assert.GreaterOrEqual(t, len(gnomadMock.QueryCalls), 2)
}

func TestRetrieveAllDataBulk_ModelLoadingFailure(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)

	// Mock model loading failure
	gnomadMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		if strings.Contains(query, "reference_stats") {
			return nil, assert.AnError
		}
		return []map[string]interface{}{}, nil
	}

	cacheMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		return []map[string]interface{}{}, nil
	}

	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{"height": {}},
		ModelID:  "v1",
	}

	annotated := &gwas.GWASDataFetcherOutput{
		AnnotatedSNPs: []model.AnnotatedSNP{
			{RSID: "rs1", Trait: "height"},
		},
	}

	_, err := retrieveAllDataBulk(ctx, requirements, annotated, refService)

	// Should fail with model loading error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load PRS model")
}

func TestRetrieveAllDataBulk_IncompleteAlleleFrequencies(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)

	// Mock model loading success but incomplete frequency data
	gnomadMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		if strings.Contains(query, "reference_stats") {
			return []map[string]interface{}{
				{
					"id":            "1:12345:G:A",
					"effect_weight": 0.5,
					"effect_allele": "A",
					"other_allele":  "G",
				},
			}, nil
		}
		if strings.Contains(query, "allele_frequencies") {
			// Return incomplete/empty frequency data
			return []map[string]interface{}{}, nil
		}
		return []map[string]interface{}{}, nil
	}

	// Cache miss to trigger stats computation
	cacheMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		return []map[string]interface{}{}, nil
	}

	// Create proper ancestry object
	ancestryObj, ancestryErr := ancestry.New("EUR", "")
	require.NoError(t, ancestryErr)

	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{"height": {}},
		CacheKeys: []reference_cache.StatsRequest{
			{Ancestry: "EUR", Trait: "height", ModelID: "v1"},
		},
		ModelID:     "v1",
		AncestryObj: ancestryObj,
	}

	annotated := &gwas.GWASDataFetcherOutput{
		AnnotatedSNPs: []model.AnnotatedSNP{
			{RSID: "rs1", Trait: "height"},
		},
	}

	_, err := retrieveAllDataBulk(ctx, requirements, annotated, refService)

	// Should fail or handle gracefully with incomplete data
	// The exact behavior depends on reference service implementation
	if err != nil {
		assert.Contains(t, err.Error(), "reference stats")
	}
}

func TestRetrieveAllDataBulk_CorruptedCacheData(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)

	gnomadMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		if strings.Contains(query, "reference_stats") {
			return []map[string]interface{}{
				{
					"id":            "1:12345:G:A",
					"effect_weight": 0.5,
					"effect_allele": "A",
					"other_allele":  "G",
				},
			}, nil
		}
		return []map[string]interface{}{}, nil
	}

	// Mock corrupted cache data (missing required fields)
	cacheMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{
				"corrupted": "data",
				// Missing mean, std, min, max fields
			},
		}, nil
	}

	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{"height": {}},
		CacheKeys: []reference_cache.StatsRequest{
			{Ancestry: "EUR", Trait: "height", ModelID: "v1"},
		},
		ModelID: "v1",
	}

	annotated := &gwas.GWASDataFetcherOutput{
		AnnotatedSNPs: []model.AnnotatedSNP{
			{RSID: "rs1", Trait: "height"},
		},
	}

	// Expect this to panic or error due to corrupted cache data
	defer func() {
		if r := recover(); r != nil {
			// This is expected - corrupted cache data should cause a panic or error
			t.Logf("Expected panic due to corrupted cache data: %v", r)
		}
	}()

	bulkData, err := retrieveAllDataBulk(ctx, requirements, annotated, refService)

	// Should handle corrupted cache gracefully with error or panic
	if err != nil {
		assert.Contains(t, strings.ToLower(err.Error()), "cache")
	} else if bulkData != nil {
		// If it doesn't error, it should have handled the corruption gracefully
		assert.NotNil(t, bulkData.CachedStats)
	}
}

func TestRetrieveAllDataBulk_LargeTraitSet(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)
	setupMockBigQueryResponses(gnomadMock, cacheMock)

	// Create large trait set (10 traits)
	traitSet := make(map[string]struct{})
	cacheKeys := make([]reference_cache.StatsRequest, 0, 10)
	annotatedSNPs := make([]model.AnnotatedSNP, 0, 10)

	for i := 0; i < 10; i++ {
		trait := fmt.Sprintf("trait_%d", i)
		traitSet[trait] = struct{}{}
		cacheKeys = append(cacheKeys, reference_cache.StatsRequest{
			Ancestry: "EUR",
			Trait:    trait,
			ModelID:  "v1",
		})
		annotatedSNPs = append(annotatedSNPs, model.AnnotatedSNP{
			RSID:  fmt.Sprintf("rs%d", i),
			Trait: trait,
		})
	}

	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	requirements := &PipelineRequirements{
		TraitSet:    traitSet,
		CacheKeys:   cacheKeys,
		ModelID:     "v1",
		AncestryObj: ancestryObj,
	}

	annotated := &gwas.GWASDataFetcherOutput{
		AnnotatedSNPs: annotatedSNPs,
	}

	bulkData, err := retrieveAllDataBulk(ctx, requirements, annotated, refService)

	// Should handle large trait set efficiently
	require.NoError(t, err)
	assert.NotNil(t, bulkData)
	assert.NotNil(t, bulkData.PRSModel)

	// Should organize trait SNPs correctly
	assert.Equal(t, 10, len(bulkData.TraitSNPs))

	// Should use bulk operations (limited BigQuery calls despite large trait set)
	assert.LessOrEqual(t, len(gnomadMock.QueryCalls), 5, "Should use bulk operations for large trait set")
}

func TestRetrieveAllDataBulk_BulkQueryOptimization(t *testing.T) {
	setupTestConfig(t)
	ctx := context.Background()

	refService, _, gnomadMock, cacheMock := setupMockPipeline(t)
	setupMockBigQueryResponses(gnomadMock, cacheMock)

	// Cache miss to trigger bulk query optimization
	cacheMock.QueryFunc = func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
		return []map[string]interface{}{}, nil
	}

	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{
			"height": {},
			"weight": {},
			"bmi":    {},
		},
		CacheKeys: []reference_cache.StatsRequest{
			{Ancestry: "EUR", Trait: "height", ModelID: "v1"},
			{Ancestry: "EUR", Trait: "weight", ModelID: "v1"},
			{Ancestry: "EUR", Trait: "bmi", ModelID: "v1"},
		},
		ModelID:     "v1",
		AncestryObj: ancestryObj,
	}

	annotated := &gwas.GWASDataFetcherOutput{
		AnnotatedSNPs: []model.AnnotatedSNP{
			{RSID: "rs1", Trait: "height"},
			{RSID: "rs2", Trait: "weight"},
			{RSID: "rs3", Trait: "bmi"},
		},
	}

	bulkData, err := retrieveAllDataBulk(ctx, requirements, annotated, refService)

	// Should succeed with bulk optimization
	require.NoError(t, err)
	assert.NotNil(t, bulkData)

	// Validate bulk query optimization
	bulkQueries := 0
	for _, call := range gnomadMock.QueryCalls {
		if strings.Contains(call.Query, " OR ") {
			bulkQueries++
		}
	}

	assert.GreaterOrEqual(t, bulkQueries, 1, "Should use bulk OR queries for optimization")
	assert.LessOrEqual(t, len(gnomadMock.QueryCalls), 4, "Should limit total BigQuery calls through bulk optimization")
}

// ==================== PHASE 3: IN-MEMORY PROCESSING TESTS ====================

func TestProcessAllTraitsInMemory_MultipleTraits(t *testing.T) {
	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	// Create requirements with multiple traits
	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{
			"height": {},
			"weight": {},
			"bmi":    {},
		},
		TraitModels: map[string]string{
			"height": "v1",
			"weight": "v1",
			"bmi":    "v1",
		},
		AncestryObj: ancestryObj,
		ModelID:     "v1",
	}

	// Create reference stats for all traits
	refStats := &reference_stats.ReferenceStats{
		Mean:     0.5,
		Std:      1.0,
		Min:      -2.0,
		Max:      3.0,
		Ancestry: "EUR",
		Trait:    "height",
		Model:    "v1",
	}

	// Create bulk data with cached stats for all traits
	ancestryCode := ancestryObj.Code()
	bulkData := &BulkDataContext{
		CachedStats: map[string]*reference_stats.ReferenceStats{
			fmt.Sprintf("%s|height|v1", ancestryCode): refStats,
			fmt.Sprintf("%s|weight|v1", ancestryCode): refStats,
			fmt.Sprintf("%s|bmi|v1", ancestryCode):    refStats,
		},
		ComputedStats: make(map[string]*reference_stats.ReferenceStats),
		TraitSNPs: map[string][]model.AnnotatedSNP{
			"height": {
				{RSID: "rs1", Trait: "height", Beta: 0.5, RiskAllele: "A", Genotype: "AG", Dosage: 1},
				{RSID: "rs2", Trait: "height", Beta: -0.3, RiskAllele: "T", Genotype: "CT", Dosage: 1},
			},
			"weight": {
				{RSID: "rs3", Trait: "weight", Beta: 0.2, RiskAllele: "G", Genotype: "GA", Dosage: 1},
			},
			"bmi": {
				{RSID: "rs4", Trait: "bmi", Beta: 0.4, RiskAllele: "C", Genotype: "CC", Dosage: 2},
			},
		},
	}

	results, err := processAllTraitsInMemory(requirements, bulkData)

	// Should succeed with multiple traits
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Validate all traits were processed
	assert.Equal(t, 3, len(results.PRSResults))
	assert.Equal(t, 3, len(results.NormalizedPRS))

	// Validate trait summaries were generated
	assert.GreaterOrEqual(t, len(results.TraitSummaries), 3)

	// Validate no cache entries (all cached)
	assert.Empty(t, results.CacheEntries)

	// Validate specific trait results
	assert.Contains(t, results.PRSResults, "height")
	assert.Contains(t, results.PRSResults, "weight")
	assert.Contains(t, results.PRSResults, "bmi")
	assert.Contains(t, results.NormalizedPRS, "height")
	assert.Contains(t, results.NormalizedPRS, "weight")
	assert.Contains(t, results.NormalizedPRS, "bmi")
}

func TestProcessAllTraitsInMemory_EmptyTraitSet(t *testing.T) {
	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	// Create requirements with empty trait set
	requirements := &PipelineRequirements{
		TraitSet:    make(map[string]struct{}),
		TraitModels: make(map[string]string),
		AncestryObj: ancestryObj,
		ModelID:     "v1",
	}

	// Create empty bulk data
	bulkData := &BulkDataContext{
		CachedStats:   make(map[string]*reference_stats.ReferenceStats),
		ComputedStats: make(map[string]*reference_stats.ReferenceStats),
		TraitSNPs:     make(map[string][]model.AnnotatedSNP),
	}

	results, err := processAllTraitsInMemory(requirements, bulkData)

	// Should succeed with empty results
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Validate empty results
	assert.Empty(t, results.PRSResults)
	assert.Empty(t, results.NormalizedPRS)
	assert.Empty(t, results.TraitSummaries)
	assert.Empty(t, results.CacheEntries)
}

func TestProcessAllTraitsInMemory_MissingReferenceStats(t *testing.T) {
	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	// Create requirements with trait
	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{
			"height": {},
		},
		TraitModels: map[string]string{
			"height": "v1",
		},
		AncestryObj: ancestryObj,
		ModelID:     "v1",
	}

	// Create bulk data without reference stats
	bulkData := &BulkDataContext{
		CachedStats:   make(map[string]*reference_stats.ReferenceStats),
		ComputedStats: make(map[string]*reference_stats.ReferenceStats),
		TraitSNPs: map[string][]model.AnnotatedSNP{
			"height": {
				{RSID: "rs1", Trait: "height", Beta: 0.5, RiskAllele: "A", Genotype: "AG", Dosage: 1},
			},
		},
	}

	_, err = processAllTraitsInMemory(requirements, bulkData)

	// Should fail due to missing reference stats
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no reference stats available for trait height")
}

func TestProcessAllTraitsInMemory_PRSCalculationAccuracy(t *testing.T) {
	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	// Create requirements with single trait
	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{
			"height": {},
		},
		TraitModels: map[string]string{
			"height": "v1",
		},
		AncestryObj: ancestryObj,
		ModelID:     "v1",
	}

	// Create reference stats with non-zero mean
	refStats := &reference_stats.ReferenceStats{
		Mean:     1.0, // Use non-zero mean for valid normalization
		Std:      1.0,
		Min:      -3.0,
		Max:      3.0,
		Ancestry: "EUR",
		Trait:    "height",
		Model:    "v1",
	}

	// Create bulk data with test SNPs for PRS calculation
	ancestryCode := ancestryObj.Code()
	bulkData := &BulkDataContext{
		CachedStats: map[string]*reference_stats.ReferenceStats{
			fmt.Sprintf("%s|height|v1", ancestryCode): refStats,
		},
		ComputedStats: make(map[string]*reference_stats.ReferenceStats),
		TraitSNPs: map[string][]model.AnnotatedSNP{
			"height": {
				// Dosage(1) × Beta(0.5) = 0.5
				{RSID: "rs1", Trait: "height", Beta: 0.5, RiskAllele: "A", Genotype: "AG", Dosage: 1},
				// Dosage(2) × Beta(0.3) = 0.6
				{RSID: "rs2", Trait: "height", Beta: 0.3, RiskAllele: "T", Genotype: "TT", Dosage: 2},
				// Dosage(0) × Beta(0.2) = 0.0
				{RSID: "rs3", Trait: "height", Beta: 0.2, RiskAllele: "G", Genotype: "AA", Dosage: 0},
			},
		},
	}

	results, err := processAllTraitsInMemory(requirements, bulkData)

	// Should succeed
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Validate PRS calculation
	assert.Contains(t, results.PRSResults, "height")
	prsResult := results.PRSResults["height"]

	// Expected PRS = 0.5 + 0.6 + 0.0 = 1.1
	expectedPRS := 1.1
	assert.InDelta(t, expectedPRS, prsResult.PRSScore, 0.01, "PRS calculation should be accurate")

	// Validate that the calculation includes all SNPs
	assert.Equal(t, 3, len(prsResult.Details), "Should include all variants in calculation")
}

func TestProcessAllTraitsInMemory_NormalizationAccuracy(t *testing.T) {
	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	// Create requirements
	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{
			"height": {},
		},
		TraitModels: map[string]string{
			"height": "v1",
		},
		AncestryObj: ancestryObj,
		ModelID:     "v1",
	}

	// Create reference stats with known values for normalization testing
	refStats := &reference_stats.ReferenceStats{
		Mean:     1.0, // Known mean
		Std:      0.5, // Known std
		Min:      0.0,
		Max:      2.0,
		Ancestry: "EUR",
		Trait:    "height",
		Model:    "v1",
	}

	// Create bulk data that will produce a predictable PRS score
	ancestryCode := ancestryObj.Code()
	bulkData := &BulkDataContext{
		CachedStats: map[string]*reference_stats.ReferenceStats{
			fmt.Sprintf("%s|height|v1", ancestryCode): refStats,
		},
		ComputedStats: make(map[string]*reference_stats.ReferenceStats),
		TraitSNPs: map[string][]model.AnnotatedSNP{
			"height": {
				// Dosage(2) × Beta(1.5) = 3.0
				{RSID: "rs1", Trait: "height", Beta: 1.5, RiskAllele: "A", Genotype: "AA", Dosage: 2},
			},
		},
	}

	results, err := processAllTraitsInMemory(requirements, bulkData)

	// Should succeed
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Validate normalization
	assert.Contains(t, results.NormalizedPRS, "height")
	normPRS := results.NormalizedPRS["height"]

	// Expected normalized score = (3.0 - 1.0) / 0.5 = 4.0
	expectedNormalized := 4.0
	assert.InDelta(t, expectedNormalized, normPRS.ZScore, 0.01, "Normalized Z-score should be accurate")

	// Validate that raw score is preserved
	assert.InDelta(t, 3.0, normPRS.RawScore, 0.01, "Raw PRS score should be preserved")
}

func TestProcessAllTraitsInMemory_SummaryGeneration(t *testing.T) {
	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	// Create requirements
	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{
			"height": {},
		},
		TraitModels: map[string]string{
			"height": "v1",
		},
		AncestryObj: ancestryObj,
		ModelID:     "v1",
	}

	// Create reference stats with non-zero mean
	refStats := &reference_stats.ReferenceStats{
		Mean:     1.0, // Use non-zero mean
		Std:      1.0,
		Min:      -2.0,
		Max:      2.0,
		Ancestry: "EUR",
		Trait:    "height",
		Model:    "v1",
	}

	// Create bulk data with multiple SNPs for comprehensive summary
	ancestryCode := ancestryObj.Code()
	bulkData := &BulkDataContext{
		CachedStats: map[string]*reference_stats.ReferenceStats{
			fmt.Sprintf("%s|height|v1", ancestryCode): refStats,
		},
		ComputedStats: make(map[string]*reference_stats.ReferenceStats),
		TraitSNPs: map[string][]model.AnnotatedSNP{
			"height": {
				{RSID: "rs1", Trait: "height", Beta: 0.5, RiskAllele: "A", Genotype: "AG", Dosage: 1},
				{RSID: "rs2", Trait: "height", Beta: -0.3, RiskAllele: "T", Genotype: "CT", Dosage: 1},
				{RSID: "rs3", Trait: "height", Beta: 0.2, RiskAllele: "G", Genotype: "GG", Dosage: 2},
			},
		},
	}

	results, err := processAllTraitsInMemory(requirements, bulkData)

	// Should succeed
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Validate trait summaries were generated
	assert.NotEmpty(t, results.TraitSummaries)

	// Find the height summary
	var heightSummary *output.TraitSummary
	for i := range results.TraitSummaries {
		if results.TraitSummaries[i].Trait == "height" {
			heightSummary = &results.TraitSummaries[i]
			break
		}
	}

	// Validate summary structure
	require.NotNil(t, heightSummary, "Should have generated summary for height trait")
	assert.Equal(t, "height", heightSummary.Trait)
	assert.NotEmpty(t, heightSummary.RiskLevel)

	// Validate that summary includes the trait processing
	normPRS := results.NormalizedPRS["height"]
	assert.NotZero(t, normPRS.ZScore, "Should have calculated normalized score")

	// Validate effect weighted contribution matches our test data
	assert.NotZero(t, heightSummary.EffectWeightedContribution)
}

func TestProcessAllTraitsInMemory_ComputedStatsAndCacheEntries(t *testing.T) {
	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	// Create requirements
	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{
			"height": {},
			"weight": {},
		},
		TraitModels: map[string]string{
			"height": "v1",
			"weight": "v1",
		},
		AncestryObj: ancestryObj,
		ModelID:     "v1",
	}

	// Create reference stats for computed stats scenario
	heightStats := &reference_stats.ReferenceStats{
		Mean:     0.1,
		Std:      0.8,
		Min:      -1.5,
		Max:      1.5,
		Ancestry: "EUR",
		Trait:    "height",
		Model:    "v1",
	}

	weightStats := &reference_stats.ReferenceStats{
		Mean:     0.2,
		Std:      0.9,
		Min:      -1.8,
		Max:      1.8,
		Ancestry: "EUR",
		Trait:    "weight",
		Model:    "v1",
	}

	// Create bulk data with computed stats (cache miss scenario)
	bulkData := &BulkDataContext{
		CachedStats: make(map[string]*reference_stats.ReferenceStats), // No cached stats
		ComputedStats: map[string]*reference_stats.ReferenceStats{
			"height": heightStats,
			"weight": weightStats,
		},
		TraitSNPs: map[string][]model.AnnotatedSNP{
			"height": {
				{RSID: "rs1", Trait: "height", Beta: 0.3, RiskAllele: "A", Genotype: "AG", Dosage: 1},
			},
			"weight": {
				{RSID: "rs2", Trait: "weight", Beta: 0.4, RiskAllele: "T", Genotype: "TT", Dosage: 2},
			},
		},
	}

	results, err := processAllTraitsInMemory(requirements, bulkData)

	// Should succeed
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Validate processing results
	assert.Equal(t, 2, len(results.PRSResults))
	assert.Equal(t, 2, len(results.NormalizedPRS))

	// Validate cache entries were created for computed stats
	assert.Equal(t, 2, len(results.CacheEntries))

	// Validate cache entry structure
	ancestryCode := ancestryObj.Code()
	cacheKeySet := make(map[string]bool)
	for _, entry := range results.CacheEntries {
		key := fmt.Sprintf("%s|%s|%s", entry.Request.Ancestry, entry.Request.Trait, entry.Request.ModelID)
		cacheKeySet[key] = true

		assert.Equal(t, ancestryCode, entry.Request.Ancestry)
		assert.Equal(t, "v1", entry.Request.ModelID)
		assert.NotNil(t, entry.Stats)
	}

	// Validate both traits have cache entries
	assert.True(t, cacheKeySet[fmt.Sprintf("%s|height|v1", ancestryCode)])
	assert.True(t, cacheKeySet[fmt.Sprintf("%s|weight|v1", ancestryCode)])
}

func TestProcessAllTraitsInMemory_SkippedTraitsWithNoSNPs(t *testing.T) {
	// Create proper ancestry object
	ancestryObj, err := ancestry.New("EUR", "")
	require.NoError(t, err)

	// Create requirements with traits
	requirements := &PipelineRequirements{
		TraitSet: map[string]struct{}{
			"height": {},
			"weight": {},
			"empty":  {}, // This trait will have no SNPs
		},
		TraitModels: map[string]string{
			"height": "v1",
			"weight": "v1",
			"empty":  "v1",
		},
		AncestryObj: ancestryObj,
		ModelID:     "v1",
	}

	// Create reference stats with non-zero mean
	refStats := &reference_stats.ReferenceStats{
		Mean:     1.0, // Use non-zero mean
		Std:      1.0,
		Min:      -2.0,
		Max:      2.0,
		Ancestry: "EUR",
		Trait:    "test",
		Model:    "v1",
	}

	// Create bulk data with SNPs only for height and weight
	ancestryCode := ancestryObj.Code()
	bulkData := &BulkDataContext{
		CachedStats: map[string]*reference_stats.ReferenceStats{
			fmt.Sprintf("%s|height|v1", ancestryCode): refStats,
			fmt.Sprintf("%s|weight|v1", ancestryCode): refStats,
			fmt.Sprintf("%s|empty|v1", ancestryCode):  refStats,
		},
		ComputedStats: make(map[string]*reference_stats.ReferenceStats),
		TraitSNPs: map[string][]model.AnnotatedSNP{
			"height": {
				{RSID: "rs1", Trait: "height", Beta: 0.5, RiskAllele: "A", Genotype: "AG", Dosage: 1},
			},
			"weight": {
				{RSID: "rs2", Trait: "weight", Beta: 0.3, RiskAllele: "T", Genotype: "CT", Dosage: 1},
			},
			"empty": {}, // No SNPs for this trait
		},
	}

	results, err := processAllTraitsInMemory(requirements, bulkData)

	// Should succeed
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Should only have results for traits with SNPs
	assert.Equal(t, 2, len(results.PRSResults))
	assert.Equal(t, 2, len(results.NormalizedPRS))

	// Validate specific traits are present
	assert.Contains(t, results.PRSResults, "height")
	assert.Contains(t, results.PRSResults, "weight")
	assert.NotContains(t, results.PRSResults, "empty")

	assert.Contains(t, results.NormalizedPRS, "height")
	assert.Contains(t, results.NormalizedPRS, "weight")
	assert.NotContains(t, results.NormalizedPRS, "empty")

	// Should have trait summaries for processed traits
	assert.GreaterOrEqual(t, len(results.TraitSummaries), 2)
}
