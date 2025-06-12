package reference

// import (
// 	"context"
// 	"testing"

// 	"phite.io/polygenic-risk-calculator/internal/config"
// )

// // Make sure we run this test with the entire package so we have access to other functions
// // that are not explicitly imported:
// // go test -v ./internal/reference/...

// // Using the common test helpers instead of local helper functions

// func TestLoadPRSModel_DuckDB(t *testing.T) {
// 	// Skip this test as DuckDB loading is not fully implemented yet
// 	t.Skip("Skipping test that requires DuckDB loading implementation")

// 	// Mock BQ client is needed for NewPRSReferenceDataSource, though not used by DuckDB path.
// 	mockBQClient := NewMockBigQueryClient(t, "test-bq-project")

// 	t.Run("successful load all fields", func(t *testing.T) {
// 		dbPath, cleanup := SetupPRSModelDuckDB(t)
// 		defer cleanup()

// 		// Use the new helper function to set up the configuration
// 		cfg := SetupDuckDBPRSModelTestConfig(t, dbPath, "")

// 		// Add specific settings for this test case
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position")

// 		ds, err := NewPRSReferenceDataSource(cfg, mockBQClient)
// 		if err != nil {
// 			t.Fatalf("Failed to create PRSReferenceDataSource: %v", err)
// 		}

// 		model, err := ds.loadPRSModel(context.Background(), "test_model")
// 		if err != nil {
// 			t.Fatalf("Failed to load PRS model: %v", err)
// 		}

// 		// Verify the model contains the expected records
// 		if len(model) != 3 {
// 			t.Fatalf("Expected 3 SNPs in model, got %d", len(model))
// 		}

// 		// Check specific SNPs
// 		for _, snp := range model {
// 			switch snp.SNPID {
// 			case "rs123":
// 				if snp.EffectAllele != "A" || snp.OtherAllele != "G" || snp.EffectWeight != 0.1 {
// 					t.Errorf("Incorrect data for rs123: %+v", snp)
// 				}
// 			case "rs456":
// 				if snp.EffectAllele != "T" || snp.OtherAllele != "C" || snp.EffectWeight != -0.2 {
// 					t.Errorf("Incorrect data for rs456: %+v", snp)
// 				}
// 			case "rs789":
// 				if snp.EffectAllele != "G" || snp.OtherAllele != "A" || snp.EffectWeight != 0.3 {
// 					t.Errorf("Incorrect data for rs789: %+v", snp)
// 				}
// 			default:
// 				t.Errorf("Unexpected SNP in model: %s", snp.SNPID)
// 			}
// 		}
// 	})

// 	t.Run("missing table returns error", func(t *testing.T) {
// 		dbPath, cleanup := SetupPRSModelDuckDB(t)
// 		defer cleanup()

// 		// Use the helper function with a nonexistent table name
// 		cfg := SetupDuckDBPRSModelTestConfig(t, dbPath, "nonexistent_table")

// 		ds, err := NewPRSReferenceDataSource(cfg, mockBQClient)
// 		if err != nil {
// 			t.Fatalf("Failed to create PRSReferenceDataSource: %v", err)
// 		}

// 		_, err = ds.loadPRSModel(context.Background(), "test_model")
// 		if err == nil {
// 			t.Fatal("Expected error for nonexistent table, got nil")
// 		}
// 	})

// 	t.Run("missing required columns returns error", func(t *testing.T) {
// 		// Use the helper function to create a DB with a table missing required columns
// 		dbPath, cleanup := SetupIncompleteModelDuckDB(t)
// 		defer cleanup()

// 		// Set up the configuration
// 		cfg := SetupDuckDBPRSModelTestConfig(t, dbPath, "incomplete_model")

// 		ds, err := NewPRSReferenceDataSource(cfg, mockBQClient)
// 		if err != nil {
// 			t.Fatalf("Failed to create PRSReferenceDataSource: %v", err)
// 		}

// 		_, err = ds.loadPRSModel(context.Background(), "test_model")
// 		if err == nil {
// 			t.Fatal("Expected error for missing required column, got nil")
// 		}
// 	})
// }
