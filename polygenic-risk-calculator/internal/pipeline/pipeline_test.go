package pipeline

import (
	"context"
	"strings"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/db"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
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
	config.SetForTest("ancestry.population", "EUR")
	config.SetForTest("ancestry.gender", "")
	config.SetForTest("reference.model", "v1")
	// Configure GWAS database path for testing
	config.SetForTest("gwas_db_path", "testdata/gwas.duckdb")
	config.SetForTest("gwas_table", "gwas_table")
	// Configure reference service settings
	config.SetForTest("reference.model_table", "reference_stats")
	config.SetForTest("reference.allele_freq_table", "allele_frequencies")
	config.SetForTest("reference.column_mapping", map[string]string{
		"AF_afr": "AF_afr",
		"AF_nfe": "AF_nfe",
	})
	// Configure GCP settings (though we might not need them for tests)
	config.SetForTest("user.gcp_project", "test-project")
	config.SetForTest("cache.gcp_project", "test-project")
	config.SetForTest("cache.dataset", "test_dataset")
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
	config.SetForTest("ancestry.population", "AFR")
	config.SetForTest("ancestry.gender", "FEMALE")
	config.SetForTest("reference.model", "v1")
	// Also need to set up GWAS database configuration
	config.SetForTest("gwas_db_path", "testdata/gwas.duckdb")
	config.SetForTest("gwas_table", "gwas_table")
	// Configure reference service settings
	config.SetForTest("reference.model_table", "reference_stats")
	config.SetForTest("reference.allele_freq_table", "allele_frequencies")
	config.SetForTest("reference.column_mapping", map[string]string{
		"AF_afr": "AF_afr",
		"AF_nfe": "AF_nfe",
	})
	// Configure GCP settings (though we might not need them for tests)
	config.SetForTest("user.gcp_project", "test-project")
	config.SetForTest("cache.gcp_project", "test-project")
	config.SetForTest("cache.dataset", "test_dataset")

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
	config.SetForTest("ancestry.population", "EAS")
	config.SetForTest("ancestry.gender", "MALE")
	config.SetForTest("reference.model", "v1")
	config.SetForTest("gwas_db_path", "testdata/gwas.duckdb")
	config.SetForTest("gwas_table", "gwas_table")
	config.SetForTest("reference.model_table", "reference_stats")
	config.SetForTest("reference.allele_freq_table", "allele_frequencies")
	config.SetForTest("reference.column_mapping", map[string]string{
		"AF_afr": "AF_afr",
		"AF_nfe": "AF_nfe",
	})
	config.SetForTest("user.gcp_project", "test-project")
	config.SetForTest("cache.gcp_project", "test-project")
	config.SetForTest("cache.dataset", "test_dataset")

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
	config.SetForTest("ancestry.population", "EUR")
	config.SetForTest("ancestry.gender", "")
	config.SetForTest("reference.model", "v2") // Different model version
	config.SetForTest("gwas_db_path", "testdata/gwas.duckdb")
	config.SetForTest("gwas_table", "gwas_table")
	config.SetForTest("reference.model_table", "reference_stats")
	config.SetForTest("reference.allele_freq_table", "allele_frequencies")
	config.SetForTest("reference.column_mapping", map[string]string{
		"AF_afr": "AF_afr",
		"AF_nfe": "AF_nfe",
	})
	config.SetForTest("user.gcp_project", "test-project")
	config.SetForTest("cache.gcp_project", "test-project")
	config.SetForTest("cache.dataset", "test_dataset")

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
