// Package reference provides test helpers for the reference package.
package reference

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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

// BigQueryRow represents the structure of a row returned by the mocked BigQuery query for PRS stats.
// This needs to match what the actual GetPRSReferenceStats function expects to parse.
type BigQueryRow struct {
	MeanPRS   float64 `bigquery:"mean_prs"`
	StdDevPRS float64 `bigquery:"stddev_prs"`
	Quantiles string  `bigquery:"quantiles"` // Assuming quantiles might be stored as JSON string or similar
}

// SetupReferenceDuckDB creates a DuckDB database with reference stats for testing.
// It returns the path to the database and a cleanup function.
func SetupReferenceDuckDB(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory for the test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "reference_test.duckdb")

	// Open database connection
	db, err := dbutil.OpenDuckDB(dbPath)
	require.NoError(t, err, "Failed to create test DuckDB database")

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE reference_stats (
			ancestry TEXT,
			trait TEXT,
			model TEXT,
			mean DOUBLE,
			std DOUBLE,
			min DOUBLE,
			max DOUBLE,
			quantiles TEXT
		);
	`)
	require.NoError(t, err, "Failed to create reference_stats table")

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO reference_stats
		(ancestry, trait, model, mean, std, min, max, quantiles)
		VALUES
		('EUR', 'height', 'v1', 1.2, 0.5, -2.0, 2.0, '{"q5":0.05,"q95":0.95}'),
		('AFR', 'height', 'v1', 1.0, 0.4, -1.8, 1.8, '{"q5":0.03,"q95":0.93}');
	`)
	require.NoError(t, err, "Failed to insert test data")

	db.Close()

	return dbPath, func() {
		// Database will be automatically deleted when the temporary directory is cleaned up
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
