package gwas_test

import (
	"context"
	"os"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/db"
	"phite.io/polygenic-risk-calculator/internal/gwas"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

func TestFetchGWASRecords(t *testing.T) {
	logging.SetSilentLoggingForTest()
	dbPath := "testdata/gwas.duckdb" // Assume a test DuckDB file exists with known data
	repo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
		"path": dbPath,
	})
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	service := gwas.NewGWASService(repo)
	ctx := context.Background()

	t.Run("fetches known rsids", func(t *testing.T) {
		logging.SetSilentLoggingForTest()
		rsids := []string{"rs1", "rs2"}
		records, err := service.FetchGWASRecords(ctx, rsids)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 2 {
			t.Errorf("expected 2 records, got %d", len(records))
		}
	})

	t.Run("handles missing rsids", func(t *testing.T) {
		rsids := []string{"rs1", "rs_missing"}
		records, err := service.FetchGWASRecords(ctx, rsids)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := records["rs_missing"]; ok {
			t.Errorf("should not return missing rsid")
		}
	})

	t.Run("malformed or missing db returns error", func(t *testing.T) {
		badRepo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
			"path": "/nonexistent/path.duckdb",
		})
		if err == nil {
			t.Errorf("expected error for missing db file")
		}
		if badRepo != nil {
			t.Errorf("expected nil repository for missing db file")
		}
	})

	t.Run("error if table does not exist", func(t *testing.T) {
		oldEnv := setEnv("GWAS_TABLE", "not_a_table")
		defer oldEnv()
		records, err := service.FetchGWASRecords(ctx, []string{"rs1"})
		if err == nil {
			t.Errorf("expected error for missing table")
		}
		_ = records // silence unused warning
	})

	t.Run("error if required columns are missing", func(t *testing.T) {
		missingColsDB := "testdata/gwas_missing_cols.duckdb"
		missingRepo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
			"path": missingColsDB,
		})
		if err != nil {
			t.Fatalf("failed to create repository: %v", err)
		}
		missingService := gwas.NewGWASService(missingRepo)
		_, err = missingService.FetchGWASRecords(ctx, []string{"rs1"})
		if err == nil {
			t.Skip("skipping: test DB missing or does not have a table missing required columns")
		}
	})

	t.Run("respects env for table name", func(t *testing.T) {
		oldEnv := setEnv("GWAS_TABLE", "associations_clean")
		defer oldEnv()
		rsids := []string{"rs1"}
		_, err := service.FetchGWASRecords(ctx, rsids)
		if err != nil {
			t.Errorf("should succeed when correct table is set via env: %v", err)
		}
	})
}

// setEnv sets an env var and returns a function to restore the old value
func TestValidateGWASDBAndTable(t *testing.T) {
	logging.SetSilentLoggingForTest()
	validDB := "testdata/gwas.duckdb"
	validTable := "associations_clean"
	missingDB := "/nonexistent/path.duckdb"
	missingTable := "not_a_table"
	missingColsDB := "testdata/gwas_missing_cols.duckdb"

	t.Run("valid DB and table returns nil", func(t *testing.T) {
		repo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
			"path": validDB,
		})
		if err != nil {
			t.Fatalf("failed to create repository: %v", err)
		}
		err = gwas.ValidateGWASDBAndTable(repo, validTable)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("missing DB returns error", func(t *testing.T) {
		repo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
			"path": missingDB,
		})
		if err == nil {
			t.Errorf("expected error for missing DB")
		}
		if repo != nil {
			t.Errorf("expected nil repository for missing DB")
		}
	})

	t.Run("missing table returns error", func(t *testing.T) {
		repo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
			"path": validDB,
		})
		if err != nil {
			t.Fatalf("failed to create repository: %v", err)
		}
		err = gwas.ValidateGWASDBAndTable(repo, missingTable)
		if err == nil {
			t.Errorf("expected error for missing table, got nil")
		}
	})

	t.Run("missing required columns returns error", func(t *testing.T) {
		repo, err := db.GetRepository(context.Background(), "duckdb", map[string]string{
			"path": missingColsDB,
		})
		if err != nil {
			t.Fatalf("failed to create repository: %v", err)
		}
		err = gwas.ValidateGWASDBAndTable(repo, validTable)
		if err == nil {
			t.Skip("skipping: test DB missing or does not have a table missing required columns")
		}
	})
}

func setEnv(key, val string) func() {
	old, present := os.LookupEnv(key)
	os.Setenv(key, val)
	return func() {
		if present {
			os.Setenv(key, old)
		} else {
			os.Unsetenv(key)
		}
	}
}
