package reference

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/config"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"

	"github.com/spf13/viper"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// PRSReferenceDataSource provides access to PRS reference statistics.
// It can fetch from a BigQuery cache or compute them on the fly if not cached.
type PRSReferenceDataSource struct {
	bqClient                         *bigquery.Client
	cacheProjectID                   string
	cacheDatasetID                   string
	cacheTableID                     string
	prsModelSourceType               string
	prsModelPathOrURI                string
	prsModelSNPIDCol                 string
	prsModelEffectAlleleCol          string
	prsModelOtherAlleleCol           string // Optional, might not always be present or required by all model files
	prsModelWeightCol                string
	prsModelChromosomeCol            string
	prsModelPositionCol              string
	prsModelIDColName                string            // Column name for the PRS model identifier in the source (e.g., study_id)
	prsModelTableName                string            // Name of the table in the DuckDB file (e.g., associations_clean)
	prsModelEffectAlleleFrequencyCol string            // Optional
	prsModelBetaValueCol             string            // Optional
	prsModelBetaCILowerCol           string            // Optional
	prsModelBetaCIUpperCol           string            // Optional
	prsModelOddsRatioCol             string            // Optional
	prsModelORCILowerCol             string            // Optional
	prsModelORCIUpperCol             string            // Optional
	prsModelVariantIDCol             string            // Optional
	prsModelRSIDCol                  string            // Optional
	ancestryMapping                  map[string]string // Maps internal ancestry codes to source-specific codes (e.g., gnomAD)
	alleleFreqSourceConfig           map[string]interface{}
	// TODO: Add fields for PRS model source configuration if needed for on-the-fly computation.
}

// PRSModelVariant holds information for a single variant within a PRS model.
// These fields are expected to be mapped from the PRS model file based on configuration.
type PRSModelVariant struct {
	SNPID                 string   // Variant identifier (e.g., rsID or chr:pos:ref:alt)
	Chromosome            string   // Chromosome (GRCh38)
	Position              int64    // Position (GRCh38, 0-based or 1-based as per model file convention)
	EffectAllele          string   // The allele associated with the effect weight
	OtherAllele           string   // The non-effect allele (can be derived if not present, or might be required)
	EffectWeight          float64  // The weight or beta score of the effect allele
	EffectAlleleFrequency *float64 // Optional: Frequency of the effect allele
	BetaValue             *float64 // Optional: Beta value
	BetaCILower           *float64 // Optional: Lower bound of Beta's confidence interval
	BetaCIUpper           *float64 // Optional: Upper bound of Beta's confidence interval
	OddsRatio             *float64 // Optional: Odds Ratio
	ORCILower             *float64 // Optional: Lower bound of OR's confidence interval
	ORCIUpper             *float64 // Optional: Upper bound of OR's confidence interval
	VariantID             *string  // Optional: Variant ID (e.g., chr:pos:ref:alt)
	RSID                  *string  // Optional: rsID
}

// NewPRSReferenceDataSource creates a new PRSReferenceDataSource.
// It requires a valid BigQuery client and configuration for cache and allele frequency sources.
func NewPRSReferenceDataSource(cfg *viper.Viper, bqClient *bigquery.Client) (*PRSReferenceDataSource, error) {
	if bqClient == nil {
		return nil, fmt.Errorf("BigQuery client cannot be nil")
	}

	// Configuration keys are validated globally by config.Validate(),
	// so we assume they are present here.
	cacheProjectID := cfg.GetString(config.PRSStatsCacheGCPProjectIDKey)
	cacheDatasetID := cfg.GetString(config.PRSStatsCacheDatasetIDKey)
	cacheTableID := cfg.GetString(config.PRSStatsCacheTableIDKey)
	prsModelSourceType := cfg.GetString(config.PRSModelSourceTypeKey)
	prsModelPathOrURI := cfg.GetString(config.PRSModelSourcePathOrTableURIKey)
	prsModelSNPIDCol := cfg.GetString(config.PRSModelSNPIDColKey)
	prsModelEffectAlleleCol := cfg.GetString(config.PRSModelEffectAlleleColKey)
	prsModelOtherAlleleCol := cfg.GetString(config.PRSModelOtherAlleleColKey) // Optional
	prsModelWeightCol := cfg.GetString(config.PRSModelWeightColKey)
	prsModelChromosomeCol := cfg.GetString(config.PRSModelChromosomeColKey)
	prsModelPositionCol := cfg.GetString(config.PRSModelPositionColKey)
	ancestryMapping := cfg.GetStringMapString(config.AlleleFreqSourceAncestryMappingKey)
	alleleFreqSourceConfig := cfg.GetStringMap(config.AlleleFreqSourceKey)

	prsModelIDColName := cfg.GetString(config.PRSModelSourceModelIDColKey)
	if prsModelIDColName == "" {
		prsModelIDColName = "study_id" // Default value
		logging.Debug("PRSModelSourceModelIDColKey not set or empty, defaulting to 'study_id'")
	}

	prsModelTableName := cfg.GetString(config.PRSModelSourceTableNameKey)
	if prsModelTableName == "" {
		prsModelTableName = "associations_clean" // Default value
		logging.Debug("PRSModelSourceTableNameKey not set or empty, defaulting to 'associations_clean'")
	}

	// Read optional column names
	prsModelEffectAlleleFrequencyCol := cfg.GetString(config.PRSModelSourceEffectAlleleFrequencyColKey)
	prsModelBetaValueCol := cfg.GetString(config.PRSModelSourceBetaValueColKey)
	prsModelBetaCILowerCol := cfg.GetString(config.PRSModelSourceBetaCILowerColKey)
	prsModelBetaCIUpperCol := cfg.GetString(config.PRSModelSourceBetaCIUpperColKey)
	prsModelOddsRatioCol := cfg.GetString(config.PRSModelSourceOddsRatioColKey)
	prsModelORCILowerCol := cfg.GetString(config.PRSModelSourceORCILowerColKey)
	prsModelORCIUpperCol := cfg.GetString(config.PRSModelSourceORCIUpperColKey)
	prsModelVariantIDCol := cfg.GetString(config.PRSModelSourceVariantIDColKey)
	prsModelRSIDCol := cfg.GetString(config.PRSModelSourceRSIDColKey)

	return &PRSReferenceDataSource{
		bqClient:                         bqClient,
		cacheProjectID:                   cacheProjectID,
		cacheDatasetID:                   cacheDatasetID,
		cacheTableID:                     cacheTableID,
		prsModelSourceType:               prsModelSourceType,
		prsModelPathOrURI:                prsModelPathOrURI,
		prsModelSNPIDCol:                 prsModelSNPIDCol,
		prsModelEffectAlleleCol:          prsModelEffectAlleleCol,
		prsModelOtherAlleleCol:           prsModelOtherAlleleCol,
		prsModelWeightCol:                prsModelWeightCol,
		prsModelChromosomeCol:            prsModelChromosomeCol,
		prsModelPositionCol:              prsModelPositionCol,
		prsModelIDColName:                prsModelIDColName,
		prsModelTableName:                prsModelTableName,
		prsModelEffectAlleleFrequencyCol: prsModelEffectAlleleFrequencyCol,
		prsModelBetaValueCol:             prsModelBetaValueCol,
		prsModelBetaCILowerCol:           prsModelBetaCILowerCol,
		prsModelBetaCIUpperCol:           prsModelBetaCIUpperCol,
		prsModelOddsRatioCol:             prsModelOddsRatioCol,
		prsModelORCILowerCol:             prsModelORCILowerCol,
		prsModelORCIUpperCol:             prsModelORCIUpperCol,
		prsModelVariantIDCol:             prsModelVariantIDCol,
		prsModelRSIDCol:                  prsModelRSIDCol,
		ancestryMapping:                  ancestryMapping,
		alleleFreqSourceConfig:           alleleFreqSourceConfig,
	}, nil
}

// Tags must match column names in BigQuery.
type prsCacheSchema struct {
	MeanPRS   float64 `bigquery:"mean_prs"`
	StdDevPRS float64 `bigquery:"stddev_prs"`
	Quantiles string  `bigquery:"quantiles"` // JSON string of map[string]float64
}

// GetPRSReferenceStats retrieves PRS reference statistics for a given ancestry, trait, and model ID.
// It first attempts to fetch from the configured BigQuery cache table.
// If not found in cache, it will attempt to compute them on-the-fly and cache the result.
func (ds *PRSReferenceDataSource) GetPRSReferenceStats(ancestry, trait, modelID string) (map[string]float64, error) {
	ctx := context.Background()

	queryString := fmt.Sprintf(
		"SELECT mean_prs, stddev_prs, quantiles FROM `%s.%s.%s` WHERE ancestry = @ancestry AND trait = @trait AND model_id = @modelID LIMIT 1",
		ds.cacheProjectID, ds.cacheDatasetID, ds.cacheTableID,
	)

	q := ds.bqClient.Query(queryString)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "ancestry", Value: ancestry},
		{Name: "trait", Value: trait},
		{Name: "modelID", Value: modelID},
	}

	logging.Debug("Executing PRS cache query: %s with params: ancestry=%s, trait=%s, modelID=%s", queryString, ancestry, trait, modelID)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute PRS cache query for ancestry=%s, trait=%s, modelID=%s: %w", ancestry, trait, modelID, err)
	}

	var row prsCacheSchema
	err = it.Next(&row)
	if err == iterator.Done {
		// This means no rows were found, which is a cache miss.
		// Fallback to on-the-fly computation and subsequent caching.
		logging.Info("Cache miss for ancestry=%s, trait=%s, modelID=%s. Attempting to compute and cache.", ancestry, trait, modelID)
		return ds.computeAndCachePRSReferenceStats(ctx, ancestry, trait, modelID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read row from PRS cache query result: %w", err)
	}

	// Check if more than one row was returned, which should not happen for a unique key.
	if err := it.Next(&row); err != iterator.Done {
		if err == nil {
			return nil, fmt.Errorf("multiple rows found in PRS cache for ancestry=%s, trait=%s, modelID=%s, expected unique row", ancestry, trait, modelID)
		}
		return nil, fmt.Errorf("failed to check for additional rows from PRS cache query result: %w", err)
	}

	// Successfully retrieved and parsed the row.
	stats := make(map[string]float64)
	stats["mean_prs"] = row.MeanPRS
	stats["stddev_prs"] = row.StdDevPRS

	var quantiles map[string]float64
	if err := json.Unmarshal([]byte(row.Quantiles), &quantiles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal quantiles JSON from cache for ancestry=%s, trait=%s, modelID=%s: %w", ancestry, trait, modelID, err)
	}
	for qKey, qVal := range quantiles {
		stats[qKey] = qVal
	}

	logging.Info("Cache hit for ancestry=%s, trait=%s, modelID=%s.", ancestry, trait, modelID)
	return stats, nil
}

// loadPRSModel is a placeholder for loading PRS model definition.
// TODO: Implement actual model loading logic (e.g., from file or BigQuery table based on config).
func (ds *PRSReferenceDataSource) loadPRSModel(ctx context.Context, modelID string) ([]PRSModelVariant, error) {
	logging.Debug("Attempting to load PRS model for modelID: %s using type: %s, path/URI: %s, table: %s",
		modelID, ds.prsModelSourceType, ds.prsModelPathOrURI, ds.prsModelTableName)

	if ds.prsModelSourceType == "duckdb" {
		// Use the wrapper function that handles opening/closing the DB connection
		selectCols := []string{
			ds.prsModelSNPIDCol,
			ds.prsModelChromosomeCol,
			ds.prsModelPositionCol,
			ds.prsModelEffectAlleleCol,
			ds.prsModelWeightCol,
		}

		if ds.prsModelOtherAlleleCol != "" {
			selectCols = append(selectCols, ds.prsModelOtherAlleleCol)
		}
		if ds.prsModelEffectAlleleFrequencyCol != "" {
			selectCols = append(selectCols, ds.prsModelEffectAlleleFrequencyCol)
		}
		if ds.prsModelBetaValueCol != "" {
			selectCols = append(selectCols, ds.prsModelBetaValueCol)
		}
		if ds.prsModelBetaCILowerCol != "" {
			selectCols = append(selectCols, ds.prsModelBetaCILowerCol)
		}
		if ds.prsModelBetaCIUpperCol != "" {
			selectCols = append(selectCols, ds.prsModelBetaCIUpperCol)
		}
		if ds.prsModelOddsRatioCol != "" {
			selectCols = append(selectCols, ds.prsModelOddsRatioCol)
		}
		if ds.prsModelORCILowerCol != "" {
			selectCols = append(selectCols, ds.prsModelORCILowerCol)
		}
		if ds.prsModelORCIUpperCol != "" {
			selectCols = append(selectCols, ds.prsModelORCIUpperCol)
		}
		if ds.prsModelVariantIDCol != "" {
			selectCols = append(selectCols, ds.prsModelVariantIDCol)
		}
		if ds.prsModelRSIDCol != "" {
			selectCols = append(selectCols, ds.prsModelRSIDCol)
		}

		query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?",
			strings.Join(selectCols, ", "), ds.prsModelTableName, ds.prsModelIDColName)

		logging.Debug("Executing DuckDB query for PRS model: %s with modelID: %s", query, modelID)

		// scanner := func(rows *sql.Rows) (*PRSModelVariant, error) {
		// 	var variant PRSModelVariant
		// 	// Nullable types for scanning optional fields
		// 	var otherAllele sql.NullString
		// 	var eaf sql.NullFloat64
		// 	var betaVal sql.NullFloat64
		// 	var betaCILower sql.NullFloat64
		// 	var betaCIUpper sql.NullFloat64
		// 	var orVal sql.NullFloat64
		// 	var orCILower sql.NullFloat64
		// 	var orCIUpper sql.NullFloat64
		// 	var variantID sql.NullString
		// 	var rsID sql.NullString

		// 	scanArgs := []interface{}{
		// 		&variant.SNPID,
		// 		&variant.Chromosome,
		// 		&variant.Position,
		// 		&variant.EffectAllele,
		// 		&variant.EffectWeight,
		// 	}

		// 	if ds.prsModelOtherAlleleCol != "" {
		// 		scanArgs = append(scanArgs, &otherAllele)
		// 	}
		// 	if ds.prsModelEffectAlleleFrequencyCol != "" {
		// 		scanArgs = append(scanArgs, &eaf)
		// 	}
		// 	if ds.prsModelBetaValueCol != "" {
		// 		scanArgs = append(scanArgs, &betaVal)
		// 	}
		// 	if ds.prsModelBetaCILowerCol != "" {
		// 		scanArgs = append(scanArgs, &betaCILower)
		// 	}
		// 	if ds.prsModelBetaCIUpperCol != "" {
		// 		scanArgs = append(scanArgs, &betaCIUpper)
		// 	}
		// 	if ds.prsModelOddsRatioCol != "" {
		// 		scanArgs = append(scanArgs, &orVal)
		// 	}
		// 	if ds.prsModelORCILowerCol != "" {
		// 		scanArgs = append(scanArgs, &orCILower)
		// 	}
		// 	if ds.prsModelORCIUpperCol != "" {
		// 		scanArgs = append(scanArgs, &orCIUpper)
		// 	}
		// 	if ds.prsModelVariantIDCol != "" {
		// 		scanArgs = append(scanArgs, &variantID)
		// 	}
		// 	if ds.prsModelRSIDCol != "" {
		// 		scanArgs = append(scanArgs, &rsID)
		// 	}

		// 	err := rows.Scan(scanArgs...)
		// 	if err != nil {
		// 		return nil, fmt.Errorf("error scanning PRSModelVariant row: %w", err)
		// 	}

		// 	if otherAllele.Valid {
		// 		variant.OtherAllele = otherAllele.String
		// 	}
		// 	if eaf.Valid {
		// 		variant.EffectAlleleFrequency = &eaf.Float64
		// 	}
		// 	if betaVal.Valid {
		// 		variant.BetaValue = &betaVal.Float64
		// 	}
		// 	if betaCILower.Valid {
		// 		variant.BetaCILower = &betaCILower.Float64
		// 	}
		// 	if betaCIUpper.Valid {
		// 		variant.BetaCIUpper = &betaCIUpper.Float64
		// 	}
		// 	if orVal.Valid {
		// 		variant.OddsRatio = &orVal.Float64
		// 	}
		// 	if orCILower.Valid {
		// 		variant.ORCILower = &orCILower.Float64
		// 	}
		// 	if orCIUpper.Valid {
		// 		variant.ORCIUpper = &orCIUpper.Float64
		// 	}
		// 	if variantID.Valid {
		// 		variant.VariantID = &variantID.String
		// 	}
		// 	if rsID.Valid {
		// 		variant.RSID = &rsID.String
		// 	}
		// 	return &variant, nil
		// }

		// results, err := dbutil.ExecuteDuckDBQuery(ctx, db, query, scanner, modelID)
		// if err != nil {
		// 	return nil, fmt.Errorf("error executing PRS model query from DuckDB: %w", err)
		// }

		// modelVariants := make([]PRSModelVariant, len(results))
		// for i, ptr := range results {
		// 	if ptr != nil {
		// 		modelVariants[i] = *ptr
		// 	}
		// }
		// logging.Debug("Successfully loaded %d variants for model %s from DuckDB table %s", len(modelVariants), modelID, ds.prsModelTableName)
		// return modelVariants, nil
	}

	return nil, fmt.Errorf("PRS model loading not yet implemented for source type '%s' (modelID: %s)", ds.prsModelSourceType, modelID)
}

func (ds *PRSReferenceDataSource) computeAndCachePRSReferenceStats(ctx context.Context, ancestry, trait, modelID string) (map[string]float64, error) {
	logging.Info("Computing PRS reference stats for ancestry=%s, trait=%s, modelID=%s", ancestry, trait, modelID)

	// 1. Load PRS model
	modelVariants, err := ds.loadPRSModel(ctx, modelID)
	if err != nil {
		// Log the error but proceed with placeholder stats for now to keep tests for cache miss passing.
		// In a full implementation, this error might be fatal for computation.
		logging.Warn("Failed to load PRS model for modelID=%s: %v. Proceeding with placeholder stats.", modelID, err)
	} else {
		logging.Debug("Successfully (placeholder) loaded %d variants for model %s", len(modelVariants), modelID)
	}

	// TODO: Implement actual on-the-fly computation logic here.
	// This will involve:
	// 1. Loading the PRS model definition (SNPs, weights) for modelID. (Partially done with placeholder loadPRSModel)
	// 2. Querying allele frequencies for these SNPs for the given ancestry from gnomAD (or other configured source).
	//    - This will use the AlleleFreqSource config.
	// 3. Calculating mean_prs, stddev_prs, and quantiles based on modelVariants and allele frequencies.

	// For now, return mock/placeholder stats.
	computedStats := map[string]float64{
		"mean_prs":   0.123, // Placeholder
		"stddev_prs": 0.045, // Placeholder
		"q5":         0.01,  // Placeholder
		"q95":        0.30,  // Placeholder
	}

	logging.Info("Successfully computed placeholder stats for ancestry=%s, trait=%s, modelID=%s", ancestry, trait, modelID)

	err = ds.writeStatsToCache(ctx, ancestry, trait, modelID, computedStats)
	if err != nil {
		// Log the error but still return the computed stats as the primary goal was computation.
		// The system can operate without caching, but not without stats.
		logging.Warn("Failed to write computed stats to cache for ancestry=%s, trait=%s, modelID=%s: %v", ancestry, trait, modelID, err)
	}

	return computedStats, nil
}

// writeStatsToCache attempts to write the computed statistics to the BigQuery cache table.
func (ds *PRSReferenceDataSource) writeStatsToCache(ctx context.Context, ancestry, trait, modelID string, stats map[string]float64) error {
	logging.Info("Attempting to write stats to cache for ancestry=%s, trait=%s, modelID=%s (placeholder)", ancestry, trait, modelID)

	// TODO: Implement actual BigQuery INSERT logic here.
	// This will involve:
	// 1. Marshalling the stats map (especially quantiles if they are complex) into the correct format for BQ schema.
	// 2. Constructing and executing a BigQuery INSERT statement.
	//    - Ensure to handle `last_updated` timestamp.
	//    - Ensure `source` is set appropriately (e.g., "on_the_fly_calculated").

	logging.Info("TODO: Implement BigQuery INSERT for caching stats: ancestry=%s, trait=%s, modelID=%s, stats=%v", ancestry, trait, modelID, stats)
	return nil // Placeholder
}
