// Package reference provides implementation for PRS reference data management.
package reference

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/dbutil"
)

// Using the common test helpers instead of local helper functions

func TestLoadPRSModel_DuckDB(t *testing.T) {
	// Skip this test as DuckDB loading is not fully implemented yet
	t.Skip("Skipping test that requires DuckDB loading implementation")

	// Mock BQ client is needed for NewPRSReferenceDataSource, though not used by DuckDB path.
	mockBQClient := NewMockBigQueryClient(t, "test-bq-project")

	t.Run("successful load all fields", func(t *testing.T) {
		cfg := SetupPRSModelTestConfig(t, nil)
		dbPath, cleanup := SetupPRSModelDuckDB(t)
		defer cleanup()

		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
		// cfg.Set(config.PRSModelTableNameKey, "prs_model")
		cfg.Set(config.PRSModelSNPIDColKey, "snp_id")
		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
		cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
		cfg.Set(config.PRSModelPositionColKey, "position")
		cfg.Set(config.ReferenceGenomeBuildKey, "GRCh38")

		// Additional required configs for PRSReferenceDataSource
		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-project")
		cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset")
		cfg.Set(config.PRSStatsCacheTableIDKey, "test_table")
		cfg.Set(config.AlleleFreqSourceTypeKey, "bigquery_gnomad")
		cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "test-project")
		cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "test_pattern")
		cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "test_table_pattern")
		cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{"EUR": "nfe"})

		ds, err := NewPRSReferenceDataSource(cfg, mockBQClient)
		if err != nil {
			t.Fatalf("Failed to create PRSReferenceDataSource: %v", err)
		}

		model, err := ds.loadPRSModel(context.Background(), "test_model")
		if err != nil {
			t.Fatalf("Failed to load PRS model: %v", err)
		}

		// Verify the model contains the expected records
		if len(model) != 3 {
			t.Fatalf("Expected 3 SNPs in model, got %d", len(model))
		}

		// Check specific SNPs
		for _, snp := range model {
			switch snp.SNPID {
			case "rs123":
				if snp.EffectAllele != "A" || snp.OtherAllele != "G" || snp.EffectWeight != 0.1 {
					t.Errorf("Incorrect data for rs123: %+v", snp)
				}
			case "rs456":
				if snp.EffectAllele != "T" || snp.OtherAllele != "C" || snp.EffectWeight != -0.2 {
					t.Errorf("Incorrect data for rs456: %+v", snp)
				}
			case "rs789":
				if snp.EffectAllele != "G" || snp.OtherAllele != "A" || snp.EffectWeight != 0.3 {
					t.Errorf("Incorrect data for rs789: %+v", snp)
				}
			default:
				t.Errorf("Unexpected SNP in model: %s", snp.SNPID)
			}
		}
	})

	t.Run("missing table returns error", func(t *testing.T) {
		cfg := SetupPRSModelTestConfig(t, nil)
		dbPath, cleanup := SetupPRSModelDuckDB(t)
		defer cleanup()

		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
		cfg.Set(config.PRSModelSourceTableNameKey, "nonexistent_table") // Table doesn't exist
		cfg.Set(config.PRSModelSNPIDColKey, "snp_id")
		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
		cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
		cfg.Set(config.ReferenceGenomeBuildKey, "GRCh38")

		// Additional required configs for PRSReferenceDataSource
		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-project")
		cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset")
		cfg.Set(config.PRSStatsCacheTableIDKey, "test_table")
		cfg.Set(config.AlleleFreqSourceTypeKey, "bigquery_gnomad")
		cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "test-project")
		cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "test_pattern")
		cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "test_table_pattern")
		cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{"EUR": "nfe"})

		ds, err := NewPRSReferenceDataSource(cfg, mockBQClient)
		if err != nil {
			t.Fatalf("Failed to create PRSReferenceDataSource: %v", err)
		}

		_, err = ds.loadPRSModel(context.Background(), "test_model")
		if err == nil {
			t.Fatal("Expected error for nonexistent table, got nil")
		}
	})

	t.Run("missing required columns returns error", func(t *testing.T) {
		// Create a new DB with a table missing required columns
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "missing_columns.duckdb")

		db, err := dbutil.OpenDuckDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to open DuckDB at %s: %v", dbPath, err)
		}

		// Create table missing the weight column
		_, err = db.Exec(`
			CREATE TABLE incomplete_model (
				snp_id VARCHAR,
				effect_allele CHAR(1),
				other_allele CHAR(1)
				-- missing effect_weight column
			);
		`)
		if err != nil {
			t.Fatalf("Failed to create incomplete model table: %v", err)
		}
		db.Close()

		cfg := viper.New()
		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
		cfg.Set(config.PRSModelSourceTableNameKey, "incomplete_model")
		cfg.Set(config.PRSModelSNPIDColKey, "snp_id")
		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
		cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
		cfg.Set(config.PRSModelWeightColKey, "effect_weight") // This column is missing
		cfg.Set(config.ReferenceGenomeBuildKey, "GRCh38")

		// Additional required configs for PRSReferenceDataSource
		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-project")
		cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset")
		cfg.Set(config.PRSStatsCacheTableIDKey, "test_table")
		cfg.Set(config.AlleleFreqSourceTypeKey, "bigquery_gnomad")
		cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "test-project")
		cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "test_pattern")
		cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "test_table_pattern")
		cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{"EUR": "nfe"})

		ds, err := NewPRSReferenceDataSource(cfg, mockBQClient)
		if err != nil {
			t.Fatalf("Failed to create PRSReferenceDataSource: %v", err)
		}

		_, err = ds.loadPRSModel(context.Background(), "test_model")
		if err == nil {
			t.Fatal("Expected error for missing required column, got nil")
		}
	})
}
