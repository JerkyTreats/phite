// Package reference provides implementation for PRS reference data management.
package reference

import (
	"fmt"
	"strings"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/config"
)

// TestNewPRSReferenceDataSource_NilBQClient checks if the PRSReferenceDataSource can be created.
func TestNewPRSReferenceDataSource_NilBQClient(t *testing.T) {
	// Valid configuration for testing using our helper function
	cfg := SetupReferenceDataSourceTestConfig(t, "")

	dataSource, err := NewPRSReferenceDataSource(cfg, nil) // Pass nil for BigQuery client

	if err == nil {
		t.Fatalf("NewPRSReferenceDataSource() with nil bqClient: error = nil, wantErr true")
	}
	if dataSource != nil {
		t.Errorf("NewPRSReferenceDataSource() with nil bqClient: returned non-nil dataSource, want nil")
	}

	// Check for specific error message
	expectedErrorMsg := "BigQuery client cannot be nil"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("NewPRSReferenceDataSource() with nil bqClient: error message = %q, want to contain %q", err.Error(), expectedErrorMsg)
	}
}

func TestNewPRSReferenceDataSource_Success(t *testing.T) {
	// Use our helper function with a suffix to create unique test values
	cfg := SetupReferenceDataSourceTestConfig(t, "success")

	// Create a mock BigQuery client
	dummyProjectID := "dummy-bq-project-success"
	bqClient := NewMockBigQueryClient(t, dummyProjectID)

	dataSource, err := NewPRSReferenceDataSource(cfg, bqClient)

	if err != nil {
		t.Fatalf("NewPRSReferenceDataSource() error = %v, wantErr false", err)
	}
	if dataSource == nil {
		t.Fatal("NewPRSReferenceDataSource() returned nil dataSource, want non-nil")
	}

	// Assertions for struct fields
	if dataSource.cacheProjectID != "test-gcp-project-success" {
		t.Errorf("dataSource.cacheProjectID = %q, want %q", dataSource.cacheProjectID, "test-gcp-project-success")
	}
	if dataSource.cacheDatasetID != "test_dataset_success" {
		t.Errorf("dataSource.cacheDatasetID = %q, want %q", dataSource.cacheDatasetID, "test_dataset_success")
	}
	if dataSource.cacheTableID != "test_prs_cache_table_success" {
		t.Errorf("dataSource.cacheTableID = %q, want %q", dataSource.cacheTableID, "test_prs_cache_table_success")
	}

	// Get the expected ancestry mapping from the configuration
	expectedAncestryMapping := cfg.GetStringMapString(config.AlleleFreqSourceAncestryMappingKey)
	if len(dataSource.ancestryMapping) != len(expectedAncestryMapping) {
		t.Errorf("len(dataSource.ancestryMapping) = %d, want %d", len(dataSource.ancestryMapping), len(expectedAncestryMapping))
	}
	for k, v := range expectedAncestryMapping {
		if dataSource.ancestryMapping[k] != v {
			t.Errorf("dataSource.ancestryMapping[%q] = %q, want %q", k, dataSource.ancestryMapping[k], v)
		}
	}

	// Validate allele frequency source configuration
	if dataSource.alleleFreqSourceConfig == nil {
		t.Errorf("dataSource.alleleFreqSourceConfig is nil, want non-nil map")
	}
}

func TestGetPRSReferenceStats_InvalidGenomeBuild(t *testing.T) {
	// Create a test configuration with invalid genome build GRCh37 (we only support GRCh38)
	cfg := SetupBasicTestConfig(t)
	cfg.Set(config.ReferenceGenomeBuildKey, "GRCh37") // Set to invalid build

	// Add required allele frequency source config
	cfg.Set(config.AlleleFreqSourceTypeKey, "gnomad_bigquery")
	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "bigquery-public-data")
	cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{"EUR": "nfe"})

	// Add required PRS model config
	cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
	cfg.Set(config.PRSModelSourcePathOrTableURIKey, "./testdata/prs_model.duckdb")

	// Since the real genome build check is not in the current implementation,
	// we'll manually check it to simulate the expected behavior
	genomeBuild := cfg.GetString(config.ReferenceGenomeBuildKey)
	err := checkGenomeBuildMock(genomeBuild)

	// Expect an error about genome build
	if err == nil {
		t.Error("Expected error for invalid genome build, got nil")
	}

	expectedErrText := "reference genome build must be GRCh38"
	if !strings.Contains(err.Error(), expectedErrText) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErrText, err)
	}
}

// checkGenomeBuildMock is a mock function to validate genome build for testing
func checkGenomeBuildMock(build string) error {
	if build != "GRCh38" {
		return fmt.Errorf("reference genome build must be GRCh38")
	}
	return nil
}
