package gwas_test

import (
	"testing"
	"phite.io/polygenic-risk-calculator/internal/gwas"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

func TestFetchGWASRecords(t *testing.T) {
	logging.SetSilentLoggingForTest()
	dbPath := "testdata/gwas.duckdb" // Assume a test DuckDB file exists with known data
	t.Run("fetches known rsids", func(t *testing.T) {
		logging.SetSilentLoggingForTest()
		rsids := []string{"rs1", "rs2"}
		records, err := gwas.FetchGWASRecords(dbPath, rsids)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 2 {
			t.Errorf("expected 2 records, got %d", len(records))
		}
	})

	t.Run("handles missing rsids", func(t *testing.T) {
		rsids := []string{"rs1", "rs_missing"}
		records, err := gwas.FetchGWASRecords(dbPath, rsids)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := records["rs_missing"]; ok {
			t.Errorf("should not return missing rsid")
		}
	})

	t.Run("malformed or missing db returns error", func(t *testing.T) {
		_, err := gwas.FetchGWASRecords("/nonexistent/path.duckdb", []string{"rs1"})
		if err == nil {
			t.Errorf("expected error for missing db file")
		}
	})
}
