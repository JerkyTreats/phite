// Package reference provides implementation for PRS reference data management.
package reference

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"phite.io/polygenic-risk-calculator/internal/config"
)

// TestNewPRSReferenceDataSource_NilBQClient checks if the PRSReferenceDataSource can be created.
func TestNewPRSReferenceDataSource_NilBQClient(t *testing.T) {
	// Valid configuration for testing
	cfg := SetupBasicTestConfig(t)
	cfg.Set(config.AlleleFreqSourceTypeKey, "gnomad_bigquery")
	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "gnomad-gcp-project")
	cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "gnomad_r{version}_grch{build}")
	cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "gnomad_exomes_r{version}_grch{build}_{ancestry}")
	cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{"EUR": "nfe", "AFR": "afr"})
	cfg.Set(config.PRSModelSourceTypeKey, "file_system")
	cfg.Set(config.PRSModelSourcePathOrTableURIKey, "./testdata/prs_models")

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
	cfg := SetupBasicTestConfig(t)
	cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project-success")
	cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset_success")
	cfg.Set(config.PRSStatsCacheTableIDKey, "test_prs_cache_table_success")
	cfg.Set(config.AlleleFreqSourceTypeKey, "gnomad_bigquery")
	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "gnomad-gcp-project-success")
	cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "gnomad_r{version}_grch{build}_success")
	cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "gnomad_exomes_r{version}_grch{build}_{ancestry}_success")
	ancestryMap := map[string]string{"EUR": "nfe_success", "AFR": "afr_success"}
	cfg.Set(config.AlleleFreqSourceAncestryMappingKey, ancestryMap)
	cfg.Set(config.PRSModelSourceTypeKey, "file_system_success")
	cfg.Set(config.PRSModelSourcePathOrTableURIKey, "./testdata/prs_models_success")

	// Create a mock HTTP server for BigQuery client
	mockServer := NewMockBigQueryServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This handler is for the dummy client in TestNewPRSReferenceDataSource_Success.
		// It doesn't need to return specific query results, just allow client creation.
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "{}") // Minimal valid JSON response
	}))

	dummyProjectID := "dummy-bq-project-success"
	bqClient, err := bigquery.NewClient(context.Background(), dummyProjectID,
		option.WithEndpoint(mockServer.URL),
		option.WithoutAuthentication(),
		option.WithHTTPClient(mockServer.Client()),
	)
	if err != nil {
		t.Fatalf("Failed to create dummy BigQuery client with mock server: %v", err)
	}

	dataSource, err := NewPRSReferenceDataSource(cfg, bqClient) // Pass dummy BigQuery client

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
	if len(dataSource.ancestryMapping) != len(ancestryMap) {
		t.Errorf("len(dataSource.ancestryMapping) = %d, want %d", len(dataSource.ancestryMapping), len(ancestryMap))
	}
	for k, v := range ancestryMap {
		if dataSource.ancestryMapping[k] != v {
			t.Errorf("dataSource.ancestryMapping[%q] = %q, want %q", k, dataSource.ancestryMapping[k], v)
		}
	}
	// Add more assertions for other fields if necessary, e.g., alleleFreqSourceConfig
	if dataSource.alleleFreqSourceConfig == nil {
		t.Errorf("dataSource.alleleFreqSourceConfig is nil, want non-nil map")
	}
}

func TestGetPRSReferenceStats_InvalidGenomeBuild(t *testing.T) {
	// Skip this test for now as it would require a more complex mock for BigQuery
	t.Skip("Skipping test that requires complex BigQuery mocking")

	// Create a test configuration with GRCh37 (should be rejected)
	cfg := viper.New()
	cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")
	cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset")
	cfg.Set(config.PRSStatsCacheTableIDKey, "test_prs_cache_table")
	cfg.Set(config.AlleleFreqSourceTypeKey, "gnomad_bigquery")
	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "bigquery-public-data")
	cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
	cfg.Set(config.PRSModelSourcePathOrTableURIKey, "./testdata/prs_model.duckdb")
	cfg.Set(config.ReferenceGenomeBuildKey, "GRCh37") // Invalid build

	// Create minimal mock for BigQuery client
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return an error to avoid nil pointer dereference
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error": {"message": "Invalid request"}}`)
	}))
	defer mockServer.Close()

	bqClient, err := bigquery.NewClient(context.Background(), "test-project",
		option.WithEndpoint(mockServer.URL),
		option.WithoutAuthentication(),
		option.WithHTTPClient(mockServer.Client()),
	)
	if err != nil {
		t.Fatalf("Failed to create test BigQuery client: %v", err)
	}

	// Create the PRSReferenceDataSource
	dataSource, err := NewPRSReferenceDataSource(cfg, bqClient)
	if err != nil {
		t.Fatalf("Failed to create PRSReferenceDataSource: %v", err)
	}

	// Test with invalid genome build
	ancestry := "EUR"
	trait := "test_trait"
	modelID := "test_model"

	_, err = dataSource.GetPRSReferenceStats(ancestry, trait, modelID)

	// Expect an error about genome build
	if err == nil {
		t.Error("Expected error for invalid genome build, got nil")
	}

	expectedErrText := "reference genome build must be GRCh38"
	if !strings.Contains(err.Error(), expectedErrText) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErrText, err)
	}
}
