package reference

import (
	"testing"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

func TestLoadReferenceStatsFromDuckDB(t *testing.T) {
	logging.SetSilentLoggingForTest()
	dbPath := "testdata/reference_stats.duckdb" // Assume this file exists for tests

	t.Run("loads valid stats", func(t *testing.T) {
		ref, err := LoadReferenceStatsFromDuckDB(dbPath, "EUR", "height", "v1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref == nil {
			t.Fatalf("expected stats, got nil")
		}
		if ref.Ancestry != "EUR" || ref.Trait != "height" || ref.Model != "v1" {
			t.Errorf("unexpected stats: %+v", ref)
		}
	})

	t.Run("returns nil if no match", func(t *testing.T) {
		ref, err := LoadReferenceStatsFromDuckDB(dbPath, "AFR", "height", "v1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref != nil {
			t.Errorf("expected nil for missing stats, got %+v", ref)
		}
	})

	t.Run("malformed or missing DB returns error", func(t *testing.T) {
		_, err := LoadReferenceStatsFromDuckDB("/nonexistent/path.duckdb", "EUR", "height", "v1")
		if err == nil {
			t.Errorf("expected error for missing db file")
		}
	})
}
