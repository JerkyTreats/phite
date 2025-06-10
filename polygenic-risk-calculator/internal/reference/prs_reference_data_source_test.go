package reference

// Added back for db.Exec, sql.Open
// For mocking BQ responses
// For mock HTTP client
// For mock HTTP server

// TestNewPRSReferenceDataSource_NilBQClient checks if the PRSReferenceDataSource can be created.
// func TestNewPRSReferenceDataSource_NilBQClient(t *testing.T) {
// 	// Valid configuration for testing
// 	cfg := viper.New()
// 	cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")
// 	cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset")
// 	cfg.Set(config.PRSStatsCacheTableIDKey, "test_prs_cache_table")
// 	cfg.Set(config.AlleleFreqSourceTypeKey, "gnomad_bigquery")
// 	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "gnomad-gcp-project")
// 	cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "gnomad_r{version}_grch{build}")
// 	cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "gnomad_exomes_r{version}_grch{build}_{ancestry}")
// 	cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{"EUR": "nfe", "AFR": "afr"})
// 	cfg.Set(config.PRSModelSourceTypeKey, "file_system")
// 	cfg.Set(config.PRSModelSourcePathOrTableURIKey, "./testdata/prs_models")
// 	cfg.Set(config.ReferenceGenomeBuildKey, "GRCh38")

// 	dataSource, err := NewPRSReferenceDataSource(cfg, nil) // Pass nil for BigQuery client

// 	if err == nil {
// 		t.Fatalf("NewPRSReferenceDataSource() with nil bqClient: error = nil, wantErr true")
// 	}
// 	if dataSource != nil {
// 		t.Errorf("NewPRSReferenceDataSource() with nil bqClient: returned non-nil dataSource, want nil")
// 	}

// 	// Check for specific error message
// 	expectedErrorMsg := "BigQuery client cannot be nil"
// 	if !strings.Contains(err.Error(), expectedErrorMsg) {
// 		t.Errorf("NewPRSReferenceDataSource() with nil bqClient: error message = %q, want to contain %q", err.Error(), expectedErrorMsg)
// 	}
// }

// func TestNewPRSReferenceDataSource_Success(t *testing.T) {
// 	cfg := viper.New()
// 	cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project-success")
// 	cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset_success")
// 	cfg.Set(config.PRSStatsCacheTableIDKey, "test_prs_cache_table_success")
// 	cfg.Set(config.AlleleFreqSourceTypeKey, "gnomad_bigquery")
// 	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "gnomad-gcp-project-success")
// 	cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "gnomad_r{version}_grch{build}_success")
// 	cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "gnomad_exomes_r{version}_grch{build}_{ancestry}_success")
// 	ancestryMap := map[string]string{"EUR": "nfe_success", "AFR": "afr_success"}
// 	cfg.Set(config.AlleleFreqSourceAncestryMappingKey, ancestryMap)
// 	cfg.Set(config.PRSModelSourceTypeKey, "file_system_success")
// 	cfg.Set(config.PRSModelSourcePathOrTableURIKey, "./testdata/prs_models_success")
// 	cfg.Set(config.ReferenceGenomeBuildKey, "GRCh38")

// 	// Create a dummy BigQuery client. For now, NewPRSReferenceDataSource only checks for nil.
// 	// In a real scenario, this would be a proper mock or an initialized client.
// 	// We need a project ID for bigquery.NewClient, even if it's a dummy one for this test.
// 	// Create a mock HTTP server for BigQuery client
// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// This handler is for the dummy client in TestNewPRSReferenceDataSource_Success.
// 		// It doesn't need to return specific query results, just allow client creation.
// 		w.WriteHeader(http.StatusOK)
// 		fmt.Fprintln(w, "{}") // Minimal valid JSON response
// 	}))
// 	defer mockServer.Close()

// 	dummyProjectID := "dummy-bq-project-success"
// 	bqClient, err := bigquery.NewClient(context.Background(), dummyProjectID,
// 		option.WithEndpoint(mockServer.URL),
// 		option.WithoutAuthentication(),
// 		option.WithHTTPClient(mockServer.Client()),
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create dummy BigQuery client with mock server: %v", err)
// 	}

// 	// If client creation failed, we can't test with a non-nil client.
// 	// However, the constructor itself should still be testable for its config logic.
// 	// For this test, we *expect* bqClient to be non-nil for a success case.
// 	// If creating a dummy client is problematic, this test might need to be adapted or skipped in certain CI environments.

// 	dataSource, err := NewPRSReferenceDataSource(cfg, bqClient) // Pass dummy BigQuery client

// 	if err != nil {
// 		t.Fatalf("NewPRSReferenceDataSource() error = %v, wantErr false", err)
// 	}
// 	if dataSource == nil {
// 		t.Fatal("NewPRSReferenceDataSource() returned nil dataSource, want non-nil")
// 	}

// 	// Assertions for struct fields
// 	if dataSource.cacheProjectID != "test-gcp-project-success" {
// 		t.Errorf("dataSource.cacheProjectID = %q, want %q", dataSource.cacheProjectID, "test-gcp-project-success")
// 	}
// 	if dataSource.cacheDatasetID != "test_dataset_success" {
// 		t.Errorf("dataSource.cacheDatasetID = %q, want %q", dataSource.cacheDatasetID, "test_dataset_success")
// 	}
// 	if dataSource.cacheTableID != "test_prs_cache_table_success" {
// 		t.Errorf("dataSource.cacheTableID = %q, want %q", dataSource.cacheTableID, "test_prs_cache_table_success")
// 	}
// 	if len(dataSource.ancestryMapping) != len(ancestryMap) {
// 		t.Errorf("len(dataSource.ancestryMapping) = %d, want %d", len(dataSource.ancestryMapping), len(ancestryMap))
// 	}
// 	for k, v := range ancestryMap {
// 		if dataSource.ancestryMapping[k] != v {
// 			t.Errorf("dataSource.ancestryMapping[%q] = %q, want %q", k, dataSource.ancestryMapping[k], v)
// 		}
// 	}
// 	// Add more assertions for other fields if necessary, e.g., alleleFreqSourceConfig
// 	if dataSource.alleleFreqSourceConfig == nil {
// 		t.Errorf("dataSource.alleleFreqSourceConfig is nil, want non-nil map")
// 	}
// }

// BigQueryRow represents the structure of a row returned by the mocked BigQuery query for PRS stats.
// This needs to match what the actual GetPRSReferenceStats function expects to parse.
// For now, let's assume it returns a few common statistics.
// type BigQueryRow struct {
// 	MeanPRS   float64 `bigquery:"mean_prs"`
// 	StdDevPRS float64 `bigquery:"stddev_prs"`
// 	Quantiles string  `bigquery:"quantiles"` // Assuming quantiles might be stored as JSON string or similar
// }

// func TestGetPRSReferenceStats_CacheHit(t *testing.T) {
// 	cfg := viper.New()
// 	cacheProjectID := "cache-hit-project"
// 	cacheDatasetID := "cache_hit_dataset"
// 	cacheTableID := "prs_reference_stats_cache"
// 	cfg.Set(config.PRSStatsCacheGCPProjectIDKey, cacheProjectID)
// 	cfg.Set(config.PRSStatsCacheDatasetIDKey, cacheDatasetID)
// 	cfg.Set(config.PRSStatsCacheTableIDKey, cacheTableID)
// 	cfg.Set(config.AlleleFreqSourceTypeKey, "bigquery_gnomad")
// 	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "bigquery-public-data")
// 	cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "gnomAD")
// 	cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "genomes_v3_GRCh38")
// 	cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{"EUR": "AF_nfe"})
// 	cfg.Set(config.PRSModelSourceTypeKey, "file")
// 	cfg.Set(config.PRSModelSourcePathOrTableURIKey, "/test/models")
// 	cfg.Set(config.PRSModelSNPIDColKey, "snp_id")
// 	cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 	cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
// 	cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 	cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 	cfg.Set(config.PRSModelPositionColKey, "position")
// 	cfg.Set(config.ReferenceGenomeBuildKey, "GRCh38")

// 	testTrait := "test_trait_cache_hit"
// 	testModelID := "pgs000XYZ_cache_hit"
// 	testAncestry := "EUR"

// 	expectedStats := map[string]float64{
// 		"mean_prs":   0.123,
// 		"stddev_prs": 0.045,
// 		"q5":         0.05,
// 		"q95":        0.20,
// 	}

// 	mockJobID := "mock-job-id-cache-hit"

// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf("/projects/%s/queries", cacheProjectID) {
// 			quantilesMap := map[string]float64{"q5": 0.05, "q95": 0.20}
// 			quantilesBytes, err := json.Marshal(quantilesMap)
// 			if err != nil {
// 				t.Errorf("Failed to marshal quantilesMap for mock server: %v", err)
// 				w.WriteHeader(http.StatusInternalServerError)
// 				return
// 			}
// 			quantilesStr := string(quantilesBytes) // This is "{\"q5\":0.05,\"q95\":0.2}"

// 			responseObj := BQQueryResponse{
// 				Kind: "bigquery#queryResponse",
// 				Schema: BQSchema{
// 					Fields: []BQFieldSchema{
// 						{Name: "mean_prs", Type: "FLOAT"},
// 						{Name: "stddev_prs", Type: "FLOAT"},
// 						{Name: "quantiles", Type: "STRING"},
// 					},
// 				},
// 				JobReference: BQJobReference{
// 					ProjectID: cacheProjectID,
// 					JobID:     mockJobID,
// 				},
// 				TotalRows: "1",
// 				Rows: []BQRow{
// 					{
// 						F: []BQCell{
// 							{V: fmt.Sprintf("%f", expectedStats["mean_prs"])},
// 							{V: fmt.Sprintf("%f", expectedStats["stddev_prs"])},
// 							{V: quantilesStr},
// 						},
// 					},
// 				},
// 				JobComplete:         true,
// 				CacheHit:            false,
// 				TotalBytesProcessed: "0",
// 				NumDMLAffectedRows:  "0",
// 			}

// 			w.Header().Set("Content-Type", "application/json")
// 			if err := json.NewEncoder(w).Encode(responseObj); err != nil {
// 				t.Errorf("Failed to encode mock BQ response: %v", err)
// 				w.WriteHeader(http.StatusInternalServerError)
// 			}
// 			return
// 		}

// 		t.Logf("Unhandled mock request: Method=%s, Path=%s", r.Method, r.URL.Path)
// 		w.WriteHeader(http.StatusNotImplemented)
// 		fmt.Fprintf(w, "Mock BQ server: Unhandled request: %s %s", r.Method, r.URL.Path)
// 	}))
// 	defer mockServer.Close()

// 	bqClient, err := bigquery.NewClient(context.Background(), cacheProjectID,
// 		option.WithEndpoint(mockServer.URL),
// 		option.WithoutAuthentication(),
// 		option.WithHTTPClient(mockServer.Client()),
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create BigQuery client with mock server: %v", err)
// 	}

// 	ds, err := NewPRSReferenceDataSource(cfg, bqClient)
// 	if err != nil {
// 		t.Fatalf("NewPRSReferenceDataSource() error = %v", err)
// 	}

// 	stats, err := ds.GetPRSReferenceStats(testAncestry, testTrait, testModelID)
// 	if err != nil {
// 		t.Fatalf("GetPRSReferenceStats() returned error = %v, wantErr false for cache hit", err)
// 	}

// 	if len(stats) != len(expectedStats) {
// 		t.Errorf("GetPRSReferenceStats() returned %d stats, want %d. Got: %v", len(stats), len(expectedStats), stats)
// 	}

// 	for key, expectedValue := range expectedStats {
// 		if val, ok := stats[key]; !ok {
// 			t.Errorf("GetPRSReferenceStats() missing expected key: %s", key)
// 		} else {
// 			const epsilon = 1e-9
// 			if diff := val - expectedValue; diff < -epsilon || diff > epsilon {
// 				t.Errorf("GetPRSReferenceStats() for key %s = %f, want %f (diff: %e)", key, val, expectedValue, diff)
// 			}
// 		}
// 	}
// }

// // newMockBigQueryClient creates a mock BigQuery client for tests that don't directly use BQ features.
// func newMockBigQueryClient(t *testing.T, projectID string) *bigquery.Client {
// 	t.Helper()
// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// This handler can be expanded if specific BQ mock interactions are needed for other tests.
// 		w.Header().Set("Content-Type", "application/json")
// 		fmt.Fprintln(w, "{\"jobComplete\": true, \"totalRows\": \"0\"}") // Minimal valid-ish JSON
// 	}))
// 	t.Cleanup(mockServer.Close)

// 	client, err := bigquery.NewClient(context.Background(), projectID,
// 		option.WithEndpoint(mockServer.URL),
// 		option.WithoutAuthentication(),
// 		option.WithHTTPClient(mockServer.Client()),
// 	)
// 	require.NoError(t, err, "Failed to create mock BigQuery client")
// 	return client
// }

// func TestLoadPRSModel_DuckDB(t *testing.T) {
// 	// Mock BQ client is needed for NewPRSReferenceDataSource, though not used by DuckDB path.
// 	mockBQClient := newMockBigQueryClient(t, "test-bq-project")

// 	t.Run("successful load all fields", func(t *testing.T) {
// 		cfg := viper.New()
// 		modelIDCol := "pgs_id"        // Custom model ID column name for this test
// 		tableName := "score_variants" // Custom table name for this test
// 		modelIDToLoad := "PGS000001"

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol)
// 		defer cleanup()

// 		// Open the DB created by the helper to insert data
// 		db, err := dbutil.OpenDuckDB(dbPath)
// 		require.NoError(t, err)

// 		// Define test data with all optional fields populated
// 		eaf := 0.123
// 		beta := 0.05
// 		betaLow := 0.025
// 		betaHigh := 0.075
// 		orVal := 1.1
// 		orLow := 1.05
// 		orHigh := 1.15
// 		variantID := "1:12345:A:G"
// 		rsID := "rs12345"

// 		insertSQL := fmt.Sprintf(`INSERT INTO %s
// 			(%s, chromosome, position_grch38, effect_allele, other_allele, effect_weight,
// 			 effect_allele_frequency, beta_value, beta_ci_lower, beta_ci_upper,
// 			 odds_ratio, or_ci_lower, or_ci_upper, variant_id, rsid)
// 			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, tableName, modelIDToLoad)

// 		_, err = db.Exec(insertSQL, modelIDToLoad, "1", 12345, "A", "G", 0.55,
// 			eaf, beta, betaLow, betaHigh, orVal, orLow, orHigh, variantID, rsID)
// 		require.NoError(t, err, "Failed to insert test data for successful load")

// 		err = db.Close()
// 		require.NoError(t, err, "Failed to close DB connection after inserting test data")

// 		// Configure PRSReferenceDataSource
// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol) // Corrected config key
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)
// 		// Set mandatory column names (already part of setupTestDuckDB schema)
// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id") // Assuming variant_id from schema is used as SNPID for this test
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		// Set optional column names (matching the schema in setupTestDuckDB)
// 		cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
// 		cfg.Set(config.PRSModelSourceEffectAlleleFrequencyColKey, "effect_allele_frequency")
// 		cfg.Set(config.PRSModelSourceBetaValueColKey, "beta_value")
// 		cfg.Set(config.PRSModelSourceBetaCILowerColKey, "beta_ci_lower")
// 		cfg.Set(config.PRSModelSourceBetaCIUpperColKey, "beta_ci_upper")
// 		cfg.Set(config.PRSModelSourceOddsRatioColKey, "odds_ratio")
// 		cfg.Set(config.PRSModelSourceORCILowerColKey, "or_ci_lower")
// 		cfg.Set(config.PRSModelSourceORCIUpperColKey, "or_ci_upper")
// 		cfg.Set(config.PRSModelSourceVariantIDColKey, "variant_id") // This is distinct from SNPIDColKey if SNPID is rsid
// 		cfg.Set(config.PRSModelSourceRSIDColKey, "rsid")

// 		// Required config keys for NewPRSReferenceDataSource
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")
// 		cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset")
// 		cfg.Set(config.PRSStatsCacheTableIDKey, "test_prs_cache_table")
// 		cfg.Set(config.AlleleFreqSourceTypeKey, "gnomad_bigquery")
// 		cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "gnomad-gcp-project")
// 		cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "gnomad_r{version}_grch{build}")
// 		cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "gnomad_exomes_r{version}_grch{build}_{ancestry}")
// 		cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{"EUR": "nfe"})
// 		cfg.Set(config.ReferenceGenomeBuildKey, "GRCh38")

// 		dataSource, err := NewPRSReferenceDataSource(cfg, mockBQClient)
// 		require.NoError(t, err)
// 		require.NotNil(t, dataSource)

// 		// Load the model
// 		variants, err := dataSource.loadPRSModel(context.Background(), modelIDToLoad)
// 		require.NoError(t, err)
// 		require.Len(t, variants, 1, "Expected one variant to be loaded")

// 		// Assertions for the loaded variant
// 		v := variants[0]
// 		assert.Equal(t, "A", v.EffectAllele)
// 		assert.Equal(t, "G", v.OtherAllele) // OtherAllele is not a pointer in the struct if always present or handled via sql.NullString
// 		assert.InEpsilon(t, 0.55, v.EffectWeight, 1e-9)
// 		assert.Equal(t, "1", v.Chromosome)
// 		assert.Equal(t, int64(12345), v.Position) // Corrected field name and type

// 		require.NotNil(t, v.EffectAlleleFrequency, "EffectAlleleFrequency should not be nil")
// 		assert.InEpsilon(t, eaf, *v.EffectAlleleFrequency, 1e-9)
// 		require.NotNil(t, v.BetaValue, "BetaValue should not be nil")
// 		assert.InEpsilon(t, beta, *v.BetaValue, 1e-9)
// 		require.NotNil(t, v.BetaCILower, "BetaCILower should not be nil")
// 		assert.InEpsilon(t, betaLow, *v.BetaCILower, 1e-9)
// 		require.NotNil(t, v.BetaCIUpper, "BetaCIUpper should not be nil")
// 		assert.InEpsilon(t, betaHigh, *v.BetaCIUpper, 1e-9)
// 		require.NotNil(t, v.OddsRatio, "OddsRatio should not be nil")
// 		assert.InEpsilon(t, orVal, *v.OddsRatio, 1e-9)
// 		require.NotNil(t, v.ORCILower, "ORCILower should not be nil")
// 		assert.InEpsilon(t, orLow, *v.ORCILower, 1e-9)
// 		require.NotNil(t, v.ORCIUpper, "ORCIUpper should not be nil")
// 		assert.InEpsilon(t, orHigh, *v.ORCIUpper, 1e-9)
// 		require.NotNil(t, v.VariantID, "VariantID should not be nil")
// 		assert.Equal(t, variantID, *v.VariantID)
// 		require.NotNil(t, v.RSID, "RSID should not be nil")
// 		assert.Equal(t, rsID, *v.RSID)
// 	})

// 	t.Run("successful load only required fields", func(t *testing.T) {
// 		tableName := "prs_models"       // Define locally for this sub-test
// 		modelIDCol := "custom_model_id" // Define locally for this sub-test
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol) // Corrected call signature
// 		defer cleanup()

// 		// Connect to the DB to insert a specific row for this test case
// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err, "Failed to open DuckDB for custom insert")
// 		defer db.Close()

// 		requiredModelID := "required_only_model_1"
// 		// Column names match those created in setupTestDuckDB schema
// 		// We only insert into the model ID column and the 5 core required data columns.
// 		// Other columns will be NULL.
// 		insertStmt := fmt.Sprintf(`INSERT INTO %s
// 			(%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight)
// 			VALUES (?, ?, ?, ?, ?, ?)`, tableName, modelIDCol)

// 		_, err = db.Exec(insertStmt, requiredModelID, "req_snp_1", "2", 23456, "C", 0.123)
// 		require.NoError(t, err, "Failed to insert row with only required fields")
// 		db.Close() // Close connection before PRSReferenceDataSource tries to open it

// 		// Configure PRSReferenceDataSource for only required fields
// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		// Set ONLY the required column config keys
// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id") // Using 'variant_id' as SNPID for this test
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		// DO NOT set config keys for optional columns like PRSModelOtherAlleleColKey, PRSModelSourceEffectAlleleFrequencyColKey, etc.

// 		// Required for NewPRSReferenceDataSource, though not directly used by loadPRSModel for DuckDB
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil) // nil for BQ client as it's not used here
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		variants, err := ds.loadPRSModel(ctx, requiredModelID) // Changed to unexported method due to compiler error
// 		require.NoError(t, err)
// 		require.Len(t, variants, 1, "Expected one variant to be loaded")

// 		v := variants[0]
// 		assert.Equal(t, "req_snp_1", v.SNPID)
// 		assert.Equal(t, "2", v.Chromosome)
// 		assert.Equal(t, int64(23456), v.Position)
// 		assert.Equal(t, "C", v.EffectAllele)
// 		assert.InEpsilon(t, 0.123, v.EffectWeight, 1e-9)

// 		// OtherAllele should be empty string as its column was not configured
// 		assert.Equal(t, "", v.OtherAllele)

// 		// All optional pointer fields should be nil
// 		assert.Nil(t, v.EffectAlleleFrequency)
// 		assert.Nil(t, v.BetaValue)
// 		assert.Nil(t, v.BetaCILower)
// 		assert.Nil(t, v.BetaCIUpper)
// 		assert.Nil(t, v.OddsRatio)
// 		assert.Nil(t, v.ORCILower)
// 		assert.Nil(t, v.ORCIUpper)
// 		assert.Nil(t, v.VariantID)
// 		assert.Nil(t, v.RSID)
// 	})

// 	t.Run("model ID not found", func(t *testing.T) {
// 		tableName := "prs_models_not_found_test" // Use a distinct table name to ensure isolation if needed
// 		modelIDCol := "custom_model_id_not_found"
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol) // This creates the table schema but doesn't insert our target model ID
// 		defer cleanup()

// 		// Configure PRSReferenceDataSource - set all column keys as if expecting a full model
// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
// 		cfg.Set(config.PRSModelSourceEffectAlleleFrequencyColKey, "effect_allele_frequency")
// 		cfg.Set(config.PRSModelSourceBetaValueColKey, "beta_value")
// 		cfg.Set(config.PRSModelSourceBetaCILowerColKey, "beta_ci_lower")
// 		cfg.Set(config.PRSModelSourceBetaCIUpperColKey, "beta_ci_upper")
// 		cfg.Set(config.PRSModelSourceOddsRatioColKey, "odds_ratio")
// 		cfg.Set(config.PRSModelSourceORCILowerColKey, "or_ci_lower")
// 		cfg.Set(config.PRSModelSourceORCIUpperColKey, "or_ci_upper")
// 		cfg.Set(config.PRSModelSourceVariantIDColKey, "variant_id_col_for_optional_variant_id") // Using a distinct name for clarity
// 		cfg.Set(config.PRSModelSourceRSIDColKey, "rsid_col_for_optional_rsid")                  // Using a distinct name for clarity

// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		nonExistentModelID := "model_id_that_definitely_does_not_exist_in_db"
// 		variants, err := ds.loadPRSModel(ctx, nonExistentModelID)

// 		require.NoError(t, err, "Loading a non-existent model ID should not produce an error")
// 		assert.Len(t, variants, 0, "Expected no variants to be returned for a non-existent model ID")
// 	})

// 	t.Run("duckdb file not found or invalid path", func(t *testing.T) {
// 		cfg := viper.New()

// 		invalidDBPath := filepath.Join(t.TempDir(), "non_existent_dir", "non_existent.duckdb")

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, invalidDBPath) // Path to a non-existent file
// 		cfg.Set(config.PRSModelSourceModelIDColKey, "any_model_id_col")
// 		cfg.Set(config.PRSModelSourceTableNameKey, "any_table_name")
// 		// Set other required column keys, though they won't be used if DB open fails
// 		cfg.Set(config.PRSModelSNPIDColKey, "snp")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chr")
// 		cfg.Set(config.PRSModelPositionColKey, "pos")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "eff")
// 		cfg.Set(config.PRSModelWeightColKey, "weight")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err, "NewPRSReferenceDataSource should not fail on invalid path alone")
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		_, err = ds.loadPRSModel(ctx, "any_model_id")

// 		require.Error(t, err, "Expected an error when DuckDB file path is invalid")
// 		// DuckDB error for invalid path might be like "IO Error: Cannot open file..."
// 		// We check for a common part of such messages. The exact message can vary.
// 		assert.Contains(t, strings.ToLower(err.Error()), "cannot open file", "Error message should indicate a file opening issue")
// 	})

// 	t.Run("misconfigured table name", func(t *testing.T) {
// 		tableNameCorrect := "correct_table_for_this_test"
// 		modelIDColCorrect := "model_id_col_for_this_test"
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableNameCorrect, modelIDColCorrect)
// 		defer cleanup()

// 		// Insert a dummy row so the table isn't empty, though it won't be queried with wrong table name
// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)
// 		insertStmt := fmt.Sprintf("INSERT INTO %s (%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight) VALUES (?, ?, ?, ?, ?, ?)",
// 			tableNameCorrect, modelIDColCorrect)
// 		_, err = db.Exec(insertStmt, "some_model_id", "rs1", "1", 123, "A", 0.1)
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDColCorrect)
// 		cfg.Set(config.PRSModelSourceTableNameKey, "wrong_table_name_that_does_not_exist") // INTENTIONALLY WRONG
// 		// Set other required column keys
// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		_, err = ds.loadPRSModel(ctx, "any_model_id")

// 		require.Error(t, err, "Expected an error when table name is misconfigured")
// 		// Error might be: Catalog Error: Table with name wrong_table_name_that_does_not_exist does not exist!
// 		assert.Contains(t, strings.ToLower(err.Error()), "table with name", "Error message should indicate table not found")
// 		assert.Contains(t, strings.ToLower(err.Error()), "does not exist", "Error message should indicate table not found")
// 	})

// 	t.Run("misconfigured model ID column name", func(t *testing.T) {
// 		tableNameCorrect := "correct_table_for_model_id_col_test"
// 		modelIDColCorrect := "actual_model_id_column"
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableNameCorrect, modelIDColCorrect)
// 		defer cleanup()

// 		// Insert a dummy row
// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)
// 		insertStmt := fmt.Sprintf("INSERT INTO %s (%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight) VALUES (?, ?, ?, ?, ?, ?)",
// 			tableNameCorrect, modelIDColCorrect)
// 		_, err = db.Exec(insertStmt, "model123", "rs2", "2", 456, "C", 0.2)
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, "wrong_model_id_column_name") // INTENTIONALLY WRONG
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableNameCorrect)
// 		// Set other required column keys
// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		_, err = ds.loadPRSModel(ctx, "model123") // Use the model ID that was inserted

// 		require.Error(t, err, "Expected an error when model ID column name is misconfigured")
// 		// Error might be: Binder Error: Referenced column "wrong_model_id_column_name" not found in FROM clause!
// 		assert.Contains(t, strings.ToLower(err.Error()), "column", "Error message should indicate column issue")
// 		assert.Contains(t, strings.ToLower(err.Error()), "not found", "Error message should indicate column not found")
// 	})

// 	t.Run("misconfigured SNPID column name", func(t *testing.T) {
// 		tableName := "snpid_col_test_table"
// 		modelIDCol := "snpid_col_test_model_id_col"
// 		actualSNPIDCol := "variant_id" // This is the actual column name created by setupTestDuckDB for SNPID
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol) // setupTestDuckDB creates 'variant_id' among others
// 		defer cleanup()

// 		// Insert a dummy row
// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)
// 		insertStmt := fmt.Sprintf("INSERT INTO %s (%s, %s, chromosome, position_grch38, effect_allele, effect_weight) VALUES (?, ?, ?, ?, ?, ?)",
// 			tableName, modelIDCol, actualSNPIDCol)
// 		_, err = db.Exec(insertStmt, "model_for_snpid_test", "rs123", "3", 789, "G", 0.3)
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		cfg.Set(config.PRSModelSNPIDColKey, "non_existent_snpid_column") // INTENTIONALLY WRONG
// 		// Set other required column keys correctly
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		_, err = ds.loadPRSModel(ctx, "model_for_snpid_test")

// 		require.Error(t, err, "Expected an error when SNPID column name is misconfigured")
// 		// Error might be: Binder Error: Referenced column "non_existent_snpid_column" not found
// 		assert.Contains(t, strings.ToLower(err.Error()), "column", "Error message should indicate column issue")
// 		assert.Contains(t, strings.ToLower(err.Error()), "non_existent_snpid_column", "Error message should mention the misconfigured column")
// 		assert.Contains(t, strings.ToLower(err.Error()), "not found", "Error message should indicate column not found")
// 	})

// 	t.Run("misconfigured effect_allele column name", func(t *testing.T) {
// 		tableName := "effect_allele_col_test_table"
// 		modelIDCol := "effect_allele_col_test_model_id_col"
// 		actualEffectAlleleCol := "effect_allele" // Actual column name from setupTestDuckDB
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol)
// 		defer cleanup()

// 		// Insert a dummy row
// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)
// 		// Note: setupTestDuckDB creates variant_id, chromosome, position_grch38, effect_allele, effect_weight etc.
// 		insertStmt := fmt.Sprintf("INSERT INTO %s (%s, variant_id, chromosome, position_grch38, %s, effect_weight) VALUES (?, ?, ?, ?, ?, ?)",
// 			tableName, modelIDCol, actualEffectAlleleCol)
// 		_, err = db.Exec(insertStmt, "model_for_ea_test", "rs456", "4", 101112, "T", 0.4)
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelEffectAlleleColKey, "wrong_effect_allele_col") // INTENTIONALLY WRONG
// 		// Set other required column keys correctly
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		_, err = ds.loadPRSModel(ctx, "model_for_ea_test")

// 		require.Error(t, err, "Expected an error when effect_allele column name is misconfigured")
// 		assert.Contains(t, strings.ToLower(err.Error()), "column", "Error message should indicate column issue")
// 		assert.Contains(t, strings.ToLower(err.Error()), "wrong_effect_allele_col", "Error message should mention the misconfigured column")
// 		assert.Contains(t, strings.ToLower(err.Error()), "not found", "Error message should indicate column not found")
// 	})

// 	t.Run("data type mismatch for effect_weight", func(t *testing.T) {
// 		tableName := "type_mismatch_test_table"
// 		modelIDCol := "type_mismatch_model_id_col"
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol)
// 		defer cleanup()

// 		// Connect and alter table to create a type mismatch
// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)

// 		// Drop the original effect_weight column and add it back as VARCHAR
// 		alterStmts := []string{
// 			fmt.Sprintf("ALTER TABLE %s DROP COLUMN effect_weight;", tableName),
// 			fmt.Sprintf("ALTER TABLE %s ADD COLUMN effect_weight VARCHAR;", tableName),
// 		}
// 		for _, stmt := range alterStmts {
// 			_, err = db.Exec(stmt)
// 			require.NoError(t, err, "Failed to alter table for type mismatch: %s", stmt)
// 		}

// 		// Insert a dummy row with a string in effect_weight
// 		// The columns created by setupTestDuckDB are: modelIDCol, variant_id, chromosome, position_grch38, effect_allele, other_allele, effect_allele_frequency, beta_value, beta_ci_lower, beta_ci_upper, odds_ratio, or_ci_lower, or_ci_upper, rsid, and originally effect_weight (DOUBLE)
// 		// We only need to provide values for the model ID col, and the columns we select in the test (SNPID, Chromosome, Position, EffectAllele, EffectWeight)
// 		insertStmt := fmt.Sprintf("INSERT INTO %s (%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight) VALUES (?, ?, ?, ?, ?, ?)",
// 			tableName, modelIDCol)
// 		_, err = db.Exec(insertStmt, "model_type_mismatch", "rs789", "5", 123456, "A", "this_is_not_a_float")
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		// Configure all column keys correctly by name, including the one that will have a type mismatch
// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight") // Points to the VARCHAR column, but PRSModelVariant expects float64
// 		// Set other optional columns as well, though they are not part of this specific mismatch test
// 		cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
// 		cfg.Set(config.PRSModelSourceEffectAlleleFrequencyColKey, "effect_allele_frequency")
// 		cfg.Set(config.PRSModelSourceBetaValueColKey, "beta_value")
// 		cfg.Set(config.PRSModelSourceBetaCILowerColKey, "beta_ci_lower")
// 		cfg.Set(config.PRSModelSourceBetaCIUpperColKey, "beta_ci_upper")
// 		cfg.Set(config.PRSModelSourceOddsRatioColKey, "odds_ratio")
// 		cfg.Set(config.PRSModelSourceORCILowerColKey, "or_ci_lower")
// 		cfg.Set(config.PRSModelSourceORCIUpperColKey, "or_ci_upper")
// 		cfg.Set(config.PRSModelSourceVariantIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelSourceRSIDColKey, "rsid")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		_, err = ds.loadPRSModel(ctx, "model_type_mismatch")

// 		require.Error(t, err, "Expected an error when effect_weight column has a type mismatch")
// 		assert.Contains(t, strings.ToLower(err.Error()), "scan error", "Error message should indicate a scan error")
// 		assert.Contains(t, strings.ToLower(err.Error()), "effect_weight", "Error message should mention the column name 'effect_weight'")
// 		assert.Contains(t, strings.ToLower(err.Error()), "converting", "Error message should indicate a conversion issue") // Or similar like 'invalid syntax'
// 	})

// 	t.Run("misconfigured effect_weight column name", func(t *testing.T) {
// 		tableName := "effect_weight_col_test_table"
// 		modelIDCol := "effect_weight_col_test_model_id_col"
// 		actualEffectWeightCol := "effect_weight" // Actual column name from setupTestDuckDB
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol) // setupTestDuckDB creates 'effect_weight'
// 		defer cleanup()

// 		// Insert a dummy row
// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)
// 		insertStmt := fmt.Sprintf("INSERT INTO %s (%s, variant_id, chromosome, position_grch38, effect_allele, %s) VALUES (?, ?, ?, ?, ?, ?)",
// 			tableName, modelIDCol, actualEffectWeightCol)
// 		_, err = db.Exec(insertStmt, "model_for_ew_test", "rs101112", "6", 131415, "G", 0.9)
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "non_existent_weight_column") // INTENTIONALLY WRONG
// 		// Set other required column keys correctly
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		_, err = ds.loadPRSModel(ctx, "model_for_ew_test")

// 		require.Error(t, err, "Expected an error when effect_weight column name is misconfigured")
// 		assert.Contains(t, strings.ToLower(err.Error()), "column", "Error message should indicate column issue")
// 		assert.Contains(t, strings.ToLower(err.Error()), "non_existent_weight_column", "Error message should mention the misconfigured column")
// 		assert.Contains(t, strings.ToLower(err.Error()), "not found", "Error message should indicate column not found")
// 	})

// 	t.Run("context cancellation during query", func(t *testing.T) {
// 		tableName := "context_cancel_test_table"
// 		modelIDCol := "context_cancel_model_id_col"
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol)
// 		defer cleanup()

// 		// Insert a dummy row
// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)
// 		insertStmt := fmt.Sprintf("INSERT INTO %s (%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight) VALUES (?, ?, ?, ?, ?, ?)",
// 			tableName, modelIDCol)
// 		_, err = db.Exec(insertStmt, "model_for_cancel_test", "rs131415", "7", 161718, "C", 0.1)
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx, cancel := context.WithCancel(context.Background())
// 		cancel() // Pre-cancel the context

// 		_, err = ds.loadPRSModel(ctx, "model_for_cancel_test")

// 		require.Error(t, err, "Expected an error when context is canceled")
// 		assert.ErrorIs(t, err, context.Canceled, "Error should be context.Canceled")
// 	})

// 	t.Run("general SQL execution error - table dropped", func(t *testing.T) {
// 		tableName := "sql_error_test_table"
// 		modelIDCol := "sql_error_model_id_col"
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol)
// 		defer cleanup()

// 		// Insert a dummy row initially
// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)
// 		insertStmt := fmt.Sprintf("INSERT INTO %s (%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight) VALUES (?, ?, ?, ?, ?, ?)",
// 			tableName, modelIDCol)
// 		_, err = db.Exec(insertStmt, "model_for_sql_error_test", "rs161718", "8", 192021, "T", 0.2)
// 		require.NoError(t, err)

// 		// Now, drop the table to cause an SQL error during loadPRSModel
// 		_, err = db.Exec(fmt.Sprintf("DROP TABLE %s;", tableName))
// 		require.NoError(t, err, "Failed to drop table for test setup")
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		_, err = ds.loadPRSModel(ctx, "model_for_sql_error_test")

// 		require.Error(t, err, "Expected an error when table is dropped before query")
// 		// DuckDB error for missing table is often: 'Catalog Error: Table with name ... does not exist!'
// 		assert.Contains(t, strings.ToLower(err.Error()), "table", "Error message should mention 'table'")
// 		assert.Contains(t, strings.ToLower(err.Error()), "does not exist", "Error message should indicate table not found or similar")
// 	})

// 	t.Run("NULL values in optional fields", func(t *testing.T) {
// 		tableName := "null_optional_fields_test_table"
// 		modelIDCol := "null_optional_model_id_col"
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol)
// 		defer cleanup()

// 		// Insert a row with NULLs for all optional fields
// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)
// 		// Columns in setupTestDuckDB: modelIDCol, variant_id (SNPID), chromosome, position_grch38, effect_allele, effect_weight,
// 		// other_allele, effect_allele_frequency, beta_value, beta_ci_lower, beta_ci_upper, odds_ratio, or_ci_lower, or_ci_upper, rsid, variant_id_custom (mapped to VariantID in code)
// 		// We use variant_id for SNPID and variant_id_custom for the optional VariantID field.
// 		// setupTestDuckDB creates 'variant_id' and 'rsid'. We'll use 'variant_id' for the required SNPID, and 'rsid' for the optional RSID.
// 		// The 'variant_id_custom' column is not created by setupTestDuckDB by default, so we'll use 'variant_id' for the optional VariantID field as well for simplicity in this test, or ensure it's NULL.
// 		// Let's ensure all optional columns are explicitly NULL.
// 		insertSQL := fmt.Sprintf(`INSERT INTO %s
// 			(%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight,
// 			 other_allele, effect_allele_frequency, beta_value, beta_ci_lower, beta_ci_upper, odds_ratio, or_ci_lower, or_ci_upper, rsid)
// 			VALUES (?, ?, ?, ?, ?, ?, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL)`, tableName, modelIDCol)

// 		_, err = db.Exec(insertSQL, "model_null_optionals", "rs12345", "1", 123456, "A", 0.5)
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		// Configure all required and optional column keys
// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")

// 		cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
// 		cfg.Set(config.PRSModelSourceEffectAlleleFrequencyColKey, "effect_allele_frequency")
// 		cfg.Set(config.PRSModelSourceBetaValueColKey, "beta_value")
// 		cfg.Set(config.PRSModelSourceBetaCILowerColKey, "beta_ci_lower")
// 		cfg.Set(config.PRSModelSourceBetaCIUpperColKey, "beta_ci_upper")
// 		cfg.Set(config.PRSModelSourceOddsRatioColKey, "odds_ratio")
// 		cfg.Set(config.PRSModelSourceORCILowerColKey, "or_ci_lower")
// 		cfg.Set(config.PRSModelSourceORCIUpperColKey, "or_ci_upper")
// 		cfg.Set(config.PRSModelSourceRSIDColKey, "rsid")
// 		// For PRSModelVariant.VariantID, we map it to 'variant_id' column as well, which is fine if it's also NULL or if we want to test its NULL-ness.
// 		// However, the setupTestDuckDB doesn't have a 'variant_id_custom'. If we want to test a specific optional VariantID column being NULL,
// 		// we'd need to add it or ensure the mapped 'variant_id' is treated as such.
// 		// For this test, we'll assume the 'rsid' column is used for PRSModelVariant.RSID and it's NULL.
// 		// And we will not explicitly map PRSModelSourceVariantIDColKey, so it should be nil by default in the struct if not selected.
// 		// Or, if we map it to rsid (which is NULL), then VariantID field should be nil.
// 		cfg.Set(config.PRSModelSourceVariantIDColKey, "rsid") // Map optional VariantID to the rsid column that is NULL

// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		variants, err := ds.loadPRSModel(ctx, "model_null_optionals")
// 		require.NoError(t, err)
// 		require.Len(t, variants, 1, "Expected one variant to be loaded")

// 		v := variants[0]
// 		// Assert required fields are present
// 		assert.Equal(t, "rs12345", v.SNPID)
// 		assert.Equal(t, "1", v.Chromosome)
// 		assert.Equal(t, int32(123456), v.Position)
// 		assert.Equal(t, "A", v.EffectAllele)
// 		assert.Equal(t, 0.5, v.EffectWeight)

// 		// Assert optional fields are nil
// 		assert.Nil(t, v.OtherAllele, "OtherAllele should be nil")
// 		assert.Nil(t, v.EffectAlleleFrequency, "EffectAlleleFrequency should be nil")
// 		assert.Nil(t, v.BetaValue, "BetaValue should be nil")
// 		assert.Nil(t, v.BetaCILower, "BetaCILower should be nil")
// 		assert.Nil(t, v.BetaCIUpper, "BetaCIUpper should be nil")
// 		assert.Nil(t, v.OddsRatio, "OddsRatio should be nil")
// 		assert.Nil(t, v.ORCILower, "ORCILower should be nil")
// 		assert.Nil(t, v.ORCIUpper, "ORCIUpper should be nil")
// 		assert.Nil(t, v.RSID, "RSID should be nil as rsid column was NULL")
// 		assert.Nil(t, v.VariantID, "VariantID should be nil as it was mapped to rsid column which was NULL")
// 	})

// 	t.Run("empty strings vs NULL for optional string fields", func(t *testing.T) {
// 		tableName := "empty_null_string_test_table"
// 		modelIDCol := "empty_null_string_model_id_col"
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol)
// 		defer cleanup()

// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)

// 		modelID := "model_str_test"
// 		snpNulls := "snp_nulls"
// 		snpEmpties := "snp_empties"

// 		// Insert variant with NULLs for other_allele and rsid
// 		// Columns: modelIDCol, variant_id, chromosome, position_grch38, effect_allele, effect_weight, other_allele, rsid (plus others not explicitly set here)
// 		insertSQLNulls := fmt.Sprintf(`INSERT INTO %s
// 			(%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight, other_allele, rsid)
// 			VALUES (?, ?, '1', 100, 'A', 0.1, NULL, NULL)`, tableName, modelIDCol)
// 		_, err = db.Exec(insertSQLNulls, modelID, snpNulls)
// 		require.NoError(t, err)

// 		// Insert variant with empty strings for other_allele and rsid
// 		insertSQLEmpties := fmt.Sprintf(`INSERT INTO %s
// 			(%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight, other_allele, rsid)
// 			VALUES (?, ?, '2', 200, 'G', 0.2, '', '')`, tableName, modelIDCol)
// 		_, err = db.Exec(insertSQLEmpties, modelID, snpEmpties)
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")

// 		cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
// 		cfg.Set(config.PRSModelSourceRSIDColKey, "rsid")
// 		// PRSModelSourceVariantIDColKey is intentionally NOT set to test default nil behavior

// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		variants, err := ds.loadPRSModel(ctx, modelID)
// 		require.NoError(t, err)
// 		require.Len(t, variants, 2, "Expected two variants to be loaded")

// 		var vNulls, vEmpties *PRSModelVariant
// 		for i := range variants {
// 			if variants[i].SNPID == snpNulls {
// 				vNulls = &variants[i]
// 			} else if variants[i].SNPID == snpEmpties {
// 				vEmpties = &variants[i]
// 			}
// 		}
// 		require.NotNil(t, vNulls, "Variant with NULL strings not found")
// 		require.NotNil(t, vEmpties, "Variant with empty strings not found")

// 		// Assertions for variant with NULL strings
// 		// For a 'string' field, DB NULL scans as an empty string ""
// 		assert.Equal(t, "", vNulls.OtherAllele, "vNulls.OtherAllele should be an empty string (DB NULL)")
// 		assert.Nil(t, vNulls.RSID, "vNulls.RSID should be nil (DB NULL)")
// 		assert.Nil(t, vNulls.VariantID, "vNulls.VariantID should be nil (config key not set)")

// 		// Assertions for variant with empty strings
// 		// For a 'string' field, DB empty string '' scans as an empty string ""
// 		assert.Equal(t, "", vEmpties.OtherAllele, "vEmpties.OtherAllele should be an empty string (DB empty string)")
// 		require.NotNil(t, vEmpties.RSID, "vEmpties.RSID should not be nil (DB empty string)")
// 		assert.Equal(t, "", *vEmpties.RSID, "vEmpties.RSID should be an empty string")
// 		assert.Nil(t, vEmpties.VariantID, "vEmpties.VariantID should be nil (config key not set)")
// 	})

// 	t.Run("multiple model IDs in table", func(t *testing.T) {
// 		tableName := "multi_model_test_table"
// 		modelIDCol := "multi_model_id_col"
// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol)
// 		defer cleanup()

// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)

// 		modelA := "model_A"
// 		modelB := "model_B"
// 		modelC := "model_C"

// 		// Insert variants: 1 for A, 2 for B, 3 for C
// 		// Columns: modelIDCol, variant_id, chromosome, position_grch38, effect_allele, effect_weight
// 		insertSQL := fmt.Sprintf(`INSERT INTO %s (%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight) VALUES (?, ?, ?, ?, ?, ?)`, tableName, modelIDCol)

// 		// Model A
// 		_, err = db.Exec(insertSQL, modelA, "snpA1", "1", 12345, "G", 0.1)
// 		require.NoError(t, err)
// 		// Model B
// 		_, err = db.Exec(insertSQL, modelB, "snpB1", "1", 23456, "T", 0.2)
// 		require.NoError(t, err)
// 		_, err = db.Exec(insertSQL, modelB, "snpB2", "1", 23457, "t", 0.25)
// 		require.NoError(t, err)
// 		// Model C
// 		_, err = db.Exec(insertSQL, modelC, "snpC1", "1", 34567, "A", 0.3)
// 		require.NoError(t, err)
// 		_, err = db.Exec(insertSQL, modelC, "snpC2", "1", 34568, "a", 0.35)
// 		require.NoError(t, err)
// 		_, err = db.Exec(insertSQL, modelC, "snpC3", "1", 34569, "G", 0.4)
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		// Request model_B
// 		variants, err := ds.loadPRSModel(ctx, modelB)
// 		require.NoError(t, err)
// 		require.Len(t, variants, 2, "Expected 2 variants for model_B")

// 		// Check if the correct variants were loaded for model_B
// 		loadedSNPs := make(map[string]bool)
// 		for _, v := range variants {
// 			loadedSNPs[v.SNPID] = true
// 		}
// 		assert.True(t, loadedSNPs["snpB1"], "snpB1 should be loaded for model_B")
// 		assert.True(t, loadedSNPs["snpB2"], "snpB2 should be loaded for model_B")
// 		assert.False(t, loadedSNPs["snpA1"], "snpA1 should NOT be loaded for model_B")
// 		assert.False(t, loadedSNPs["snpC1"], "snpC1 should NOT be loaded for model_B")
// 	})

// 	t.Run("load moderately large number of variants", func(t *testing.T) {
// 		const numVariantsToLoad = 10000 // Moderately large number for a unit test
// 		tableName := "large_load_test_table"
// 		modelIDCol := "large_load_model_id_col"
// 		modelID := "large_model_001"

// 		cfg := viper.New()

// 		dbPath, cleanup := setupTestDuckDB(t, tableName, modelIDCol)
// 		defer cleanup()

// 		db, err := sql.Open("duckdb", dbPath)
// 		require.NoError(t, err)

// 		// Prepare insert statement
// 		insertSQL := fmt.Sprintf(`INSERT INTO %s (%s, variant_id, chromosome, position_grch38, effect_allele, effect_weight) VALUES (?, ?, ?, ?, ?, ?)`, tableName, modelIDCol)
// 		tx, err := db.Begin()
// 		require.NoError(t, err)
// 		stmt, err := tx.Prepare(insertSQL)
// 		require.NoError(t, err)
// 		defer stmt.Close()

// 		// Insert numVariantsToLoad variants
// 		for i := 0; i < numVariantsToLoad; i++ {
// 			snpID := fmt.Sprintf("rs%d", i+1)
// 			chromosome := fmt.Sprintf("%d", (i%22)+1)
// 			position := int64(1000 + i)
// 			effectAllele := "A"
// 			if i%2 == 1 {
// 				effectAllele = "G"
// 			}
// 			effectWeight := 0.01 * float64(i%100)
// 			_, err = stmt.Exec(modelID, snpID, chromosome, position, effectAllele, effectWeight)
// 			require.NoError(t, err, "Error inserting variant %d", i)
// 		}
// 		err = tx.Commit()
// 		require.NoError(t, err)
// 		db.Close()

// 		cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 		cfg.Set(config.PRSModelSourcePathOrTableURIKey, dbPath)
// 		cfg.Set(config.PRSModelSourceModelIDColKey, modelIDCol)
// 		cfg.Set(config.PRSModelSourceTableNameKey, tableName)

// 		cfg.Set(config.PRSModelSNPIDColKey, "variant_id")
// 		cfg.Set(config.PRSModelChromosomeColKey, "chromosome")
// 		cfg.Set(config.PRSModelPositionColKey, "position_grch38")
// 		cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 		cfg.Set(config.PRSModelWeightColKey, "effect_weight")
// 		cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")

// 		ds, err := NewPRSReferenceDataSource(cfg, nil)
// 		require.NoError(t, err)
// 		require.NotNil(t, ds)

// 		ctx := context.Background()
// 		variants, err := ds.loadPRSModel(ctx, modelID)
// 		require.NoError(t, err, "loadPRSModel failed for moderately large dataset")
// 		assert.Len(t, variants, numVariantsToLoad, "Expected %d variants to be loaded", numVariantsToLoad)
// 	})

// 	// TODO: Add more sub-tests for error handling during query execution and row scanning, e.g.:
// 	//   - Misconfigured other data column names (e.g., effect_allele, weight_column) -> DONE
// 	//   - Data type mismatch during row scanning (if a column exists but has wrong type for scanning) -> DONE
// 	//   - Context cancellation during query execution -> DONE
// 	//   - General SQL execution error (e.g., table disappears mid-query, disk error) -> DONE
// 	//   - Test behavior with NULL values in optional fields when those fields are configured (should be correctly loaded as nil pointers) -> DONE
// 	//   - Test behavior with empty strings vs NULL for optional string fields (if applicable to any) -> DONE
// 	//   - Test with a moderately large number of variants (e.g., 10k) for a single model ID (smoke test for performance/memory). -> DONE
// 	//     (Note: True performance/memory profiling should be done with benchmarks or dedicated load tests.)
// 	//   - Test with multiple model IDs in the table, ensuring only the requested one is loaded. -> DONE
// }

// func TestGetPRSReferenceStats_CacheMiss_ComputesAndReturnsStats(t *testing.T) {
// 	cfg := viper.New()
// 	cacheProjectID := "cache-miss-project"
// 	cacheDatasetID := "cache_miss_dataset"
// 	cacheTableID := "prs_reference_stats_cache_miss"
// 	cfg.Set(config.PRSStatsCacheGCPProjectIDKey, cacheProjectID)
// 	cfg.Set(config.PRSStatsCacheDatasetIDKey, cacheDatasetID)
// 	cfg.Set(config.PRSStatsCacheTableIDKey, cacheTableID)
// 	// Other necessary config for NewPRSReferenceDataSource
// 	cfg.Set(config.AlleleFreqSourceTypeKey, "gnomad_bigquery")
// 	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "gnomad-gcp-project")
// 	cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "gnomad_r_pattern")
// 	cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "gnomad_t_pattern")
// 	cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{"EUR": "nfe"})
// 	cfg.Set(config.ReferenceGenomeBuildKey, "GRCh38")

// 	testTrait := "test_trait_cache_miss"
// 	testModelID := "pgs000ABC_cache_miss"
// 	testAncestry := "AFR"

// 	mockJobID := "mock-job-id-cache-miss"

// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		var responseObj BQQueryResponse
// 		if r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf("/projects/%s/queries", cacheProjectID) {
// 			responseObj = BQQueryResponse{
// 				Kind: "bigquery#queryResponse",
// 				Schema: BQSchema{
// 					Fields: []BQFieldSchema{
// 						{Name: "mean_prs", Type: "FLOAT"},
// 						{Name: "stddev_prs", Type: "FLOAT"},
// 						{Name: "quantiles", Type: "STRING"},
// 					},
// 				},
// 				JobReference: BQJobReference{
// 					ProjectID: cacheProjectID,
// 					JobID:     mockJobID,
// 				},
// 				TotalRows:           "0", // Simulate no rows found
// 				Rows:                nil, // No rows
// 				JobComplete:         true,
// 				CacheHit:            false,
// 				TotalBytesProcessed: "0",
// 			}

// 			w.Header().Set("Content-Type", "application/json")
// 			responseObj := BQQueryResponse{
// 				Kind: "bigquery#queryResponse",
// 				Schema: BQSchema{
// 					Fields: []BQFieldSchema{
// 						{Name: "mean_prs", Type: "FLOAT"},
// 						{Name: "stddev_prs", Type: "FLOAT"},
// 						{Name: "quantiles", Type: "STRING"},
// 					},
// 				},
// 				JobReference: BQJobReference{
// 					ProjectID: cacheProjectID,
// 					JobID:     mockJobID,
// 				},
// 				TotalRows:           "0", // Simulate no rows found
// 				Rows:                nil, // No rows
// 				JobComplete:         true,
// 				CacheHit:            false,
// 				TotalBytesProcessed: "0",
// 			}
// 			w.Header().Set("Content-Type", "application/json")
// 			if err := json.NewEncoder(w).Encode(responseObj); err != nil {
// 				t.Errorf("Failed to encode mock BQ response for cache miss (POST): %v", err)
// 				w.WriteHeader(http.StatusInternalServerError)
// 			}
// 			return
// 		}

// 		// Handle GET to /projects/{projectID}/queries/{jobID} to fetch job status/results
// 		if r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/projects/%s/queries/%s", cacheProjectID, mockJobID) {
// 			responseObj = BQQueryResponse{
// 				Kind: "bigquery#queryResponse",
// 				Schema: BQSchema{ // Schema should still be present
// 					Fields: []BQFieldSchema{
// 						{Name: "mean_prs", Type: "FLOAT"},
// 						{Name: "stddev_prs", Type: "FLOAT"},
// 						{Name: "quantiles", Type: "STRING"},
// 					},
// 				},
// 				JobReference: BQJobReference{
// 					ProjectID: cacheProjectID,
// 					JobID:     mockJobID,
// 				},
// 				TotalRows:           "0", // No rows found
// 				Rows:                nil,
// 				JobComplete:         true,
// 				CacheHit:            false, // CacheHit might be true if the query itself was a repeat, but for data not found, it's less relevant.
// 				TotalBytesProcessed: "0",
// 			}
// 			w.Header().Set("Content-Type", "application/json")
// 			if err := json.NewEncoder(w).Encode(responseObj); err != nil {
// 				t.Errorf("Failed to encode mock BQ response for cache miss (GET): %v", err)
// 				w.WriteHeader(http.StatusInternalServerError)
// 			}
// 			return
// 		}

// 		t.Logf("Unhandled mock request: Method=%s, Path=%s", r.Method, r.URL.Path)
// 		w.WriteHeader(http.StatusNotImplemented)
// 		fmt.Fprintf(w, "Mock BQ server (cache miss): Unhandled request: %s %s", r.Method, r.URL.Path)
// 	}))
// 	defer mockServer.Close()

// 	bqClient, err := bigquery.NewClient(context.Background(), cacheProjectID,
// 		option.WithEndpoint(mockServer.URL),
// 		option.WithoutAuthentication(),
// 		option.WithHTTPClient(mockServer.Client()),
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create BigQuery client with mock server: %v", err)
// 	}

// 	ds, err := NewPRSReferenceDataSource(cfg, bqClient)
// 	if err != nil {
// 		t.Fatalf("NewPRSReferenceDataSource() error = %v", err)
// 	}

// 	stats, err := ds.GetPRSReferenceStats(testAncestry, testTrait, testModelID)
// 	if err != nil {
// 		t.Fatalf("GetPRSReferenceStats() returned error for cache miss with computation: %v, want nil", err)
// 	}

// 	if stats == nil {
// 		t.Fatalf("GetPRSReferenceStats() returned nil stats for cache miss with computation, want placeholder stats")
// 	}

// 	// Placeholder stats defined in prs_reference_data_source.go's computeAndCachePRSReferenceStats
// 	expectedPlaceholderStats := map[string]float64{
// 		"mean_prs":   0.123,
// 		"stddev_prs": 0.045,
// 		"q5":         0.01,
// 		"q95":        0.30,
// 	}

// 	if len(stats) != len(expectedPlaceholderStats) {
// 		t.Errorf("GetPRSReferenceStats() returned %d stats, want %d. Got: %v", len(stats), len(expectedPlaceholderStats), stats)
// 	}

// 	for key, expectedValue := range expectedPlaceholderStats {
// 		if val, ok := stats[key]; !ok {
// 			t.Errorf("GetPRSReferenceStats() missing expected key from placeholder stats: %s", key)
// 		} else {
// 			const epsilon = 1e-9 // Using a small epsilon for float comparison
// 			if diff := val - expectedValue; diff < -epsilon || diff > epsilon {
// 				t.Errorf("GetPRSReferenceStats() for key %s = %f, want %f (placeholder)", key, val, expectedValue)
// 			}
// 		}
// 	}
// }

// TestComputeAndCachePRSReferenceStats_Success tests the successful computation and caching of PRS reference statistics.
// func TestComputeAndCachePRSReferenceStats_Success(t *testing.T) {
// 	// Create a test configuration
// 	cfg := viper.New()
// 	cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")
// 	cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset")
// 	cfg.Set(config.PRSStatsCacheTableIDKey, "test_prs_cache_table")
// 	cfg.Set(config.AlleleFreqSourceTypeKey, "gnomad_bigquery")
// 	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "bigquery-public-data")
// 	cfg.Set(config.AlleleFreqSourceDatasetIDPatternKey, "gnomAD")
// 	cfg.Set(config.AlleleFreqSourceTableIDPatternKey, "v3_genomes_chr{chrom}")
// 	cfg.Set(config.AlleleFreqSourceAncestryMappingKey, map[string]string{
// 		"EUR": "AF_nfe",
// 		"AFR": "AF_afr",
// 	})
// 	cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 	cfg.Set(config.PRSModelSourcePathOrTableURIKey, "./testdata/prs_model.duckdb")
// 	cfg.Set(config.PRSModelSNPIDColKey, "snp_id")
// 	cfg.Set(config.PRSModelChromosomeColKey, "chrom")
// 	cfg.Set(config.PRSModelPositionColKey, "pos")
// 	cfg.Set(config.PRSModelEffectAlleleColKey, "effect_allele")
// 	cfg.Set(config.PRSModelOtherAlleleColKey, "other_allele")
// 	cfg.Set(config.PRSModelWeightColKey, "weight")
// 	cfg.Set(config.PRSModelSourceModelIDColKey, "model_id")
// 	cfg.Set(config.PRSModelSourceTableNameKey, "model_variants") // Use the correct table name
// 	cfg.Set(config.ReferenceGenomeBuildKey, "GRCh38")

// 	// Setup mock server for BigQuery client
// 	mockBQServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Parse the request to determine which API method is being called
// 		body, err := io.ReadAll(r.Body)
// 		if err != nil {
// 			t.Fatalf("Failed to read request body: %v", err)
// 		}

// 		// Check if this is a request to query allele frequencies
// 		if strings.Contains(string(body), "alternate_bases.AF_") {
// 			// Return mock allele frequency data
// 			responseJSON := createMockAlleleFrequencyResponse()
// 			w.Header().Set("Content-Type", "application/json")
// 			w.Write(responseJSON)
// 			return
// 		}

// 		// Check if this is a cache insert request
// 		if strings.Contains(string(body), "INSERT INTO") {
// 			// Return success for insert operation
// 			w.Header().Set("Content-Type", "application/json")
// 			w.Write([]byte(`{"jobComplete": true, "totalRows": "0", "numDmlAffectedRows": "1"}`))
// 			return
// 		}

// 		// Default response for other queries
// 		w.Header().Set("Content-Type", "application/json")
// 		w.Write([]byte(`{"jobComplete": true, "totalRows": "0"}`))
// 	}))
// 	defer mockBQServer.Close()

// 	// Create BigQuery client with the mock server
// 	bqClient, err := bigquery.NewClient(context.Background(), "test-project",
// 		option.WithEndpoint(mockBQServer.URL),
// 		option.WithoutAuthentication(),
// 		option.WithHTTPClient(mockBQServer.Client()),
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create test BigQuery client: %v", err)
// 	}

// 	// Create the PRSReferenceDataSource with our mock
// 	dataSource, err := NewPRSReferenceDataSource(cfg, bqClient)
// 	if err != nil {
// 		t.Fatalf("Failed to create PRSReferenceDataSource: %v", err)
// 	}

// 	// Setup DuckDB with test data
// 	testDbPath := filepath.Join(t.TempDir(), "test_prs_model.duckdb")
// 	setupTestPRSModelDatabase(t, testDbPath)
// 	// Override the path in the datasource
// 	dataSource.prsModelPathOrURI = testDbPath

// 	// Test computation of PRS statistics
// 	ctx := context.Background()
// 	ancestry := "EUR"
// 	trait := "test_trait"
// 	modelID := "test_model"

// 	stats, err := dataSource.computeAndCachePRSReferenceStats(ctx, ancestry, trait, modelID)
// 	if err != nil {
// 		t.Fatalf("computeAndCachePRSReferenceStats failed: %v", err)
// 	}

// 	// Verify the computed statistics
// 	if stats["mean_prs"] <= 0 {
// 		t.Errorf("Expected positive mean_prs, got %v", stats["mean_prs"])
// 	}
// 	if stats["stddev_prs"] <= 0 {
// 		t.Errorf("Expected positive stddev_prs, got %v", stats["stddev_prs"])
// 	}
// 	// Check that percentiles are included
// 	if _, hasQ50 := stats["q50"]; !hasQ50 {
// 		t.Error("Expected q50 percentile in stats, not found")
// 	}
// }

// func TestGetPRSReferenceStats_InvalidGenomeBuild(t *testing.T) {
// 	// Create a test configuration with GRCh37 (should be rejected)
// 	cfg := viper.New()
// 	cfg.Set(config.PRSStatsCacheGCPProjectIDKey, "test-gcp-project")
// 	cfg.Set(config.PRSStatsCacheDatasetIDKey, "test_dataset")
// 	cfg.Set(config.PRSStatsCacheTableIDKey, "test_prs_cache_table")
// 	cfg.Set(config.AlleleFreqSourceTypeKey, "gnomad_bigquery")
// 	cfg.Set(config.AlleleFreqSourceGCPProjectIDKey, "bigquery-public-data")
// 	cfg.Set(config.PRSModelSourceTypeKey, "duckdb")
// 	cfg.Set(config.PRSModelSourcePathOrTableURIKey, "./testdata/prs_model.duckdb")
// 	cfg.Set(config.ReferenceGenomeBuildKey, "GRCh37") // Invalid build

// 	// Create minimal mock for BigQuery client
// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		fmt.Fprintln(w, "{}") // Minimal response, we shouldn't get this far
// 	}))
// 	defer mockServer.Close()

// 	bqClient, err := bigquery.NewClient(context.Background(), "test-project",
// 		option.WithEndpoint(mockServer.URL),
// 		option.WithoutAuthentication(),
// 		option.WithHTTPClient(mockServer.Client()),
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create test BigQuery client: %v", err)
// 	}

// 	// Create the PRSReferenceDataSource
// 	dataSource, err := NewPRSReferenceDataSource(cfg, bqClient)
// 	if err != nil {
// 		t.Fatalf("Failed to create PRSReferenceDataSource: %v", err)
// 	}

// 	// Test with invalid genome build
// 	ancestry := "EUR"
// 	trait := "test_trait"
// 	modelID := "test_model"

// 	_, err = dataSource.GetPRSReferenceStats(ancestry, trait, modelID)

// 	// Expect an error about genome build
// 	if err == nil {
// 		t.Error("Expected error for invalid genome build, got nil")
// 	}

// 	expectedErrText := "reference genome build must be GRCh38"
// 	if !strings.Contains(err.Error(), expectedErrText) {
// 		t.Errorf("Expected error to contain '%s', got: %v", expectedErrText, err)
// 	}
// }
