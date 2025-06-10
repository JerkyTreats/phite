// Package reference provides test helpers for the reference package.
package reference

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/dbutil"
)

// Structs for BQ QueryResponse, shared across tests
type BQFieldSchema struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
type BQSchema struct {
	Fields []BQFieldSchema `json:"fields"`
}
type BQJobReference struct {
	ProjectID string `json:"projectId"`
	JobID     string `json:"jobId"`
}
type BQCell struct {
	V string `json:"v"` // Values are typically strings in the raw API response
}
type BQRow struct {
	F []BQCell `json:"f"`
}
type BQQueryResponse struct {
	Kind                string         `json:"kind"`
	Schema              BQSchema       `json:"schema"`
	JobReference        BQJobReference `json:"jobReference"`
	TotalRows           string         `json:"totalRows"`
	Rows                []BQRow        `json:"rows,omitempty"`
	JobComplete         bool           `json:"jobComplete"`
	CacheHit            bool           `json:"cacheHit"`
	TotalBytesProcessed string         `json:"totalBytesProcessed"`
	NumDMLAffectedRows  string         `json:"numDmlAffectedRows,omitempty"`
}

// NewMockBigQueryClient creates a mock BigQuery client for testing.
// It sets up a mock HTTP server that returns a simple JSON response.
func NewMockBigQueryClient(t *testing.T, projectID string) *bigquery.Client {
	t.Helper()
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "{}") // Minimal valid JSON response
	}))
	t.Cleanup(func() { mockServer.Close() })

	bqClient, err := bigquery.NewClient(context.Background(), projectID,
		option.WithEndpoint(mockServer.URL),
		option.WithoutAuthentication(),
		option.WithHTTPClient(mockServer.Client()),
	)
	require.NoError(t, err, "Failed to create mock BigQuery client")
	return bqClient
}

// NewMockBigQueryServer creates a mock HTTP server for BigQuery testing.
// It returns the server and a cleanup function.
func NewMockBigQueryServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(func() { server.Close() })
	return server
}

// SetupBasicTestConfig creates a basic viper configuration with common test values.
func SetupBasicTestConfig(t *testing.T) *viper.Viper {
	t.Helper()
	cfg := viper.New()

	// Common configuration values used in tests
	cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")
	cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset")
	cfg.Set(config.PRSStatsCacheTableIDKey, "test_prs_cache_table")
	cfg.Set(config.ReferenceGenomeBuildKey, "GRCh38")

	return cfg
}

// SetupPRSModelTestConfig extends the basic configuration with PRS model specific settings.
func SetupPRSModelTestConfig(t *testing.T, baseConfig *viper.Viper) *viper.Viper {
	t.Helper()

	// If no base config is provided, create a new one with basic settings
	cfg := baseConfig
	if cfg == nil {
		cfg = SetupBasicTestConfig(t)
	}

	// Add PRS model specific configuration
	cfg.Set(config.AlleleFreqSourceTypeKey, "bigquery_gnomad")
	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "bigquery-public-data")
	cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "gnomAD")
	cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "genomes_v3_GRCh38")
	cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{"EUR": "AF_nfe", "AFR": "AF_afr"})
	cfg.Set(config.PRSModelSourceTypeKey, "file")
	cfg.Set(config.PRSModelSourcePathOrTableURIKey, "./testdata/test_prs_model.tsv")
	cfg.Set(config.PRSModelSNPIDColKey, "snp_id")
	cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
	cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
	cfg.Set(config.PRSModelWeightColKey, "effect_weight")

	return cfg
}

// SetupPRSModelDuckDB creates a temporary DuckDB database with a PRS model table for testing.
// It returns the path to the database and a cleanup function.
func SetupPRSModelDuckDB(t *testing.T) (string, func()) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_prs_model.duckdb")

	// Create a temporary DuckDB database with a PRS model table
	db, err := dbutil.OpenDuckDB(dbPath)
	require.NoError(t, err, "Failed to open DuckDB")
	defer db.Close()

	// Create a PRS model table with all required columns
	_, err = db.Exec(`
		CREATE TABLE prs_model (
			snp_id VARCHAR,
			effect_allele CHAR(1),
			other_allele CHAR(1),
			effect_weight DOUBLE,
			chromosome VARCHAR,
			position INTEGER
		);

		INSERT INTO prs_model VALUES
			('rs123', 'A', 'G', 0.1, '1', 1000000),
			('rs456', 'T', 'C', -0.2, '2', 2000000),
			('rs789', 'G', 'A', 0.3, '3', 3000000);
	`)
	require.NoError(t, err, "Failed to create and populate PRS model table")

	return dbPath, func() {
		// tempDir is cleaned up automatically by t.TempDir()
	}
}

// CreateMockBQResponse creates a mock BigQuery response for testing.
func CreateMockBQResponse(stats map[string]float64) BQQueryResponse {
	return BQQueryResponse{
		Kind:        "bigquery#queryResponse",
		JobComplete: true,
		TotalRows:   "1",
		Schema: BQSchema{
			Fields: []BQFieldSchema{
				{Name: "mean_prs", Type: "FLOAT"},
				{Name: "stddev_prs", Type: "FLOAT"},
				{Name: "min_prs", Type: "FLOAT"},
				{Name: "max_prs", Type: "FLOAT"},
				{Name: "quantiles", Type: "STRING"},
			},
		},
		Rows: []BQRow{
			{
				F: []BQCell{
					{V: fmt.Sprintf("%f", stats["mean_prs"])},
					{V: fmt.Sprintf("%f", stats["stddev_prs"])},
					{V: fmt.Sprintf("%f", stats["min_prs"])},
					{V: fmt.Sprintf("%f", stats["max_prs"])},
					{V: fmt.Sprintf(`{"q5":%f,"q95":%f}`, stats["q5"], stats["q95"])},
				},
			},
		},
	}
}

// NewMockBigQueryClientWithResponse creates a mock BigQuery client that returns a predefined response.
// This is useful for testing functions that query BigQuery and expect a specific result.
func NewMockBigQueryClientWithResponse(t *testing.T, projectID string, stats map[string]float64) *bigquery.Client {
	t.Helper()

	// Create the mock response
	mockResponse := CreateMockBQResponse(stats)

	// Set up a mock server that returns the response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		responseBytes, err := json.Marshal(mockResponse)
		require.NoError(t, err, "Failed to marshal mock BQ response")
		w.Write(responseBytes)
	}))
	t.Cleanup(func() { mockServer.Close() })

	// Create a BigQuery client that uses the mock server
	bqClient, err := bigquery.NewClient(context.Background(), projectID,
		option.WithEndpoint(mockServer.URL),
		option.WithoutAuthentication(),
		option.WithHTTPClient(mockServer.Client()),
	)
	require.NoError(t, err, "Failed to create mock BigQuery client with response")
	return bqClient
}
