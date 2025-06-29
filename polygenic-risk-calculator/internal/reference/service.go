package reference

import (
	"context"
	"fmt"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/ancestry"
	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/db"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
	"phite.io/polygenic-risk-calculator/internal/logging"
	reference_cache "phite.io/polygenic-risk-calculator/internal/reference/cache"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
	"phite.io/polygenic-risk-calculator/internal/utils"

	"phite.io/polygenic-risk-calculator/internal/model"
)

// ReferenceService handles loading PRS models and allele frequencies using the repository pattern
type ReferenceService struct {
	gnomadDB        dbinterface.Repository
	modelDB         dbinterface.Repository
	ReferenceCache  reference_cache.Cache
	modelTable      string
	alleleFreqTable string
}

// ReferenceStatsRequest represents a request for reference statistics computation
type ReferenceStatsRequest struct {
	Ancestry *ancestry.Ancestry
	Trait    string
}

func init() {
	// Register required infrastructure constants for reference service
	config.RegisterRequiredKey(config.TableModelTableKey)      // Model table reference
	config.RegisterRequiredKey(config.TableAlleleFreqTableKey) // Allele frequency table reference
	config.RegisterRequiredKey(config.GCPBillingProjectKey)    // User's billing project for gnomAD queries
	config.RegisterRequiredKey(config.GCPCacheProjectKey)      // Cache storage project
	config.RegisterRequiredKey(config.BigQueryCacheDatasetKey) // Cache dataset name
}

// NewReferenceService creates a new reference service with dependency injection
// If gnomadDB or ReferenceCache are nil, they will be created using default configuration
func NewReferenceService(gnomadDB, modelDB dbinterface.Repository, ReferenceCache reference_cache.Cache) (*ReferenceService, error) {
	var err error

	// Create gnomAD repository if not provided
	if gnomadDB == nil {
		gnomadDB, err = db.GetRepository(context.Background(), "bq", map[string]string{
			"project_id":      config.GetString(config.GCPDataProjectKey),        // gnomAD data project
			"dataset_id":      config.GetString(config.BigQueryGnomadDatasetKey), // gnomAD dataset
			"billing_project": config.GetString(config.GCPBillingProjectKey),     // User's billing project
		})
		if err != nil {
			logging.Error("Failed to create gnomAD repository: %v", err)
			return nil, fmt.Errorf("failed to create gnomAD repository: %w", err)
		}
	}

	// Create model repository if not provided
	if modelDB == nil {
		modelDB, err = db.GetRepository(context.Background(), "duckdb", map[string]string{"path": config.GetString("gwas_db_path")})
		if err != nil {
			logging.Error("Failed to create model repository: %v", err)
			return nil, fmt.Errorf("failed to create model repository: %w", err)
		}
	}

	// Create cache if not provided
	if ReferenceCache == nil {
		cacheParams := map[string]string{
			"project_id":      config.GetString(config.GCPCacheProjectKey),      // Cache storage project
			"dataset_id":      config.GetString(config.BigQueryCacheDatasetKey), // Cache dataset
			"billing_project": config.GetString(config.GCPBillingProjectKey),    // User's billing project
		}
		ReferenceCache, err = reference_cache.NewRepositoryCache(nil, cacheParams)
		if err != nil {
			logging.Error("Failed to create cache repository: %v", err)
			return nil, fmt.Errorf("failed to create cache repository: %w", err)
		}
	}

	return &ReferenceService{
		gnomadDB:        gnomadDB,
		modelDB:         modelDB,
		ReferenceCache:  ReferenceCache,
		modelTable:      config.GetString(config.TableModelTableKey),
		alleleFreqTable: config.GetString(config.TableAlleleFreqTableKey),
	}, nil
}

// LoadModel loads a PRS model from the configured table for a specific trait
func (s *ReferenceService) LoadModel(ctx context.Context, trait string) (*model.PRSModel, error) {
	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE trait = ?",
		s.modelTable,
	)

	logging.Info("Loading PRS model for trait: %s", trait)
	rows, err := s.modelDB.Query(ctx, query, trait)
	if err != nil {
		return nil, fmt.Errorf("failed to query model for trait %s: %w", trait, err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no variants found for trait: %s", trait)
	}

	var variants []model.Variant
	for _, row := range rows {
		variant, err := s.convertRowToVariant(row)
		if err != nil {
			return nil, fmt.Errorf("failed to convert row to variant: %w", err)
		}
		variants = append(variants, variant)
	}

	prsModel := &model.PRSModel{
		ID:       trait, // Use trait as the model identifier
		Variants: variants,
	}

	if err := prsModel.Validate(); err != nil {
		return nil, fmt.Errorf("invalid PRS model: %w", err)
	}

	logging.Info("Successfully loaded PRS model with %d variants", len(variants))
	return prsModel, nil
}

// GetAlleleFrequenciesForTraits retrieves allele frequencies for variants across multiple traits in a single BigQuery operation
// This method optimizes costs by batching all variant queries together instead of making separate queries per trait
func (s *ReferenceService) GetAlleleFrequenciesForTraits(ctx context.Context, traitVariants map[string][]model.Variant, ancestry *ancestry.Ancestry) (map[string]map[string]float64, error) {
	if len(traitVariants) == 0 {
		return map[string]map[string]float64{}, nil
	}

	// Collect all unique variants across all traits to avoid duplicates in the query
	uniqueVariants := make(map[string]model.Variant)
	for trait, variants := range traitVariants {
		logging.Debug("Collecting %d variants for trait %s", len(variants), trait)
		for _, v := range variants {
			uniqueVariants[v.ID] = v
		}
	}

	if len(uniqueVariants) == 0 {
		logging.Info("No variants found across all traits")
		return map[string]map[string]float64{}, nil
	}

	// Get all columns needed for this ancestry's precedence logic
	columns := ancestry.ColumnPrecedence()
	selectCols := append([]string{"chrom", "pos", "ref", "alt"}, columns...)

	// Build consolidated variant filters for all unique variants
	var filters []string
	var args []interface{}
	for _, v := range uniqueVariants {
		if v.Chromosome == "" || v.Position == 0 || v.Ref == "" || v.Alt == "" {
			logging.Warn("cannot build filter for variant %s, missing chrom/pos/ref/alt", v.ID)
			continue
		}
		filters = append(filters, "(chrom = ? AND pos = ? AND ref = ? AND alt = ?)")
		args = append(args, v.Chromosome, v.Position, v.Ref, v.Alt)
	}

	if len(filters) == 0 {
		logging.Info("No variants with sufficient information for allele frequency lookup")
		return map[string]map[string]float64{}, nil
	}

	// Build and execute single consolidated query for all variants
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		strings.Join(selectCols, ", "),
		s.alleleFreqTable,
		strings.Join(filters, " OR "),
	)

	logging.Info("Querying allele frequencies for %d unique variants across %d traits with ancestry %s",
		len(uniqueVariants), len(traitVariants), ancestry.Code())
	rows, err := s.gnomadDB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query allele frequencies: %w", err)
	}

	// Process consolidated results and build frequency map
	allFreqs := make(map[string]float64)
	for _, row := range rows {
		// Let ancestry object select the best frequency from available columns
		freq, usedCol, err := ancestry.SelectFrequency(row)
		if err != nil {
			// Skip variants with no frequency data available
			logging.Debug("No frequency data available for variant in row: %v", err)
			continue
		}

		chrom := utils.ToString(row["chrom"])
		pos := utils.ToInt64(row["pos"])
		ref := utils.ToString(row["ref"])
		alt := utils.ToString(row["alt"])

		variantID := model.FormatVariantID(chrom, pos, ref, alt)
		allFreqs[variantID] = freq

		// Log which column was used for debugging
		logging.Debug("Used column %s for variant %s (frequency: %f)", usedCol, variantID, freq)
	}

	// Partition consolidated results back to per-trait format
	result := make(map[string]map[string]float64)
	for trait, variants := range traitVariants {
		traitFreqs := make(map[string]float64)
		for _, v := range variants {
			if freq, found := allFreqs[v.ID]; found {
				traitFreqs[v.ID] = freq
			}
		}
		result[trait] = traitFreqs
		logging.Debug("Partitioned %d variant frequencies for trait %s", len(traitFreqs), trait)
	}

	logging.Info("Retrieved allele frequencies for %d unique variants across %d traits using ancestry %s",
		len(allFreqs), len(traitVariants), ancestry.Code())
	return result, nil
}

// convertRowToVariant converts a database row to a Variant
func (s *ReferenceService) convertRowToVariant(row map[string]interface{}) (model.Variant, error) {
	// Required fields
	rsid := utils.ToString(row["rsid"])

	effectWeight := utils.ToFloat64(row["beta"])
	if effectWeight == 0 {
		// This check can be problematic if an effect weight is genuinely 0.
		// For now, we assume it indicates a missing value.
		return model.Variant{}, fmt.Errorf("missing or invalid effect weight")
	}

	effectAllele := utils.ToString(row["risk_allele"])
	if effectAllele == "" {
		return model.Variant{}, fmt.Errorf("missing or invalid effect allele")
	}

	// Optional fields
	otherAllele := utils.ToString(row["other_allele"])
	var effectFreq *float64
	if val := utils.ToFloat64(row["risk_allele_freq"]); val != 0 {
		effectFreq = &val
	}

	// Get components for variant ID
	chrom := utils.ToString(row["chr"])
	pos := utils.ToInt64(row["chr_pos"])
	ref := utils.ToString(row["ref"])
	alt := utils.ToString(row["alt"])

	if chrom == "" || pos == 0 {
		return model.Variant{}, fmt.Errorf("missing chromosome or position for rsid %s", rsid)
	}

	// If ref or alt are missing, we cannot form a fully qualified variant ID for gnomAD lookup.
	// For now, we allow this and will use the rsid as the ID, but this may fail in downstream functions.
	variantID := rsid
	if ref != "" && alt != "" {
		variantID = model.FormatVariantID(chrom, pos, ref, alt)
	}

	// To handle duplicate variants from different studies, append study_id to the variant ID.
	studyID := utils.ToString(row["study_id"])
	if studyID != "" {
		variantID = fmt.Sprintf("%s_%s", variantID, studyID)
	}

	variant := model.Variant{
		ID:           variantID,
		Chromosome:   chrom,
		Position:     pos,
		Ref:          ref,
		Alt:          alt,
		EffectWeight: effectWeight,
		EffectAllele: effectAllele,
		OtherAllele:  otherAllele,
		EffectFreq:   effectFreq,
	}

	if rsid != "" {
		variant.RSID = &rsid
	}

	return variant, nil
}

// GetReferenceStatsBatch retrieves reference statistics for multiple traits in a single operation.
// This method optimizes costs by loading all required models, and then making a single
// consolidated query to get allele frequencies for all variants across all models.
func (s *ReferenceService) GetReferenceStatsBatch(ctx context.Context, requests []ReferenceStatsRequest) (map[string]*reference_stats.ReferenceStats, error) {
	logging.Info("Getting reference stats in batch for %d traits", len(requests))

	if len(requests) == 0 {
		return make(map[string]*reference_stats.ReferenceStats), nil
	}

	ancestryObj := requests[0].Ancestry
	traitModels := make(map[string]*model.PRSModel)
	allTraitVariants := make(map[string][]model.Variant)

	// Step 1: Load the model for each trait and collect all variants.
	for _, req := range requests {
		if _, ok := traitModels[req.Trait]; ok {
			continue // Already loaded this model
		}
		prsModel, err := s.LoadModel(ctx, req.Trait)
		if err != nil {
			return nil, fmt.Errorf("failed to load PRS model for trait %s: %w", req.Trait, err)
		}
		traitModels[req.Trait] = prsModel
		allTraitVariants[req.Trait] = prsModel.Variants
	}

	// Step 2: Get allele frequencies for all variants across all models in a single bulk query.
	alleleFrequencies, err := s.GetAlleleFrequenciesForTraits(ctx, allTraitVariants, ancestryObj)
	if err != nil {
		return nil, fmt.Errorf("failed to get allele frequencies: %w", err)
	}

	// Step 3: Compute stats for each trait using its specific model and frequencies.
	results := make(map[string]*reference_stats.ReferenceStats)
	for _, req := range requests {
		prsModel := traitModels[req.Trait]
		traitFreqs := alleleFrequencies[req.Trait]

		stats, err := reference_stats.Compute(traitFreqs, prsModel.GetEffectSizes())
		if err != nil {
			return nil, fmt.Errorf("failed to compute stats for trait %s: %w", req.Trait, err)
		}

		stats.Ancestry = ancestryObj.Code()
		stats.Trait = req.Trait
		stats.Model = req.Trait // The model is identified by the trait

		key := fmt.Sprintf("%s|%s|%s", stats.Ancestry, stats.Trait, stats.Model)
		results[key] = stats
	}

	logging.Info("Successfully computed reference stats for %d traits in batch", len(results))
	return results, nil
}

// GetReferenceStats retrieves PRS reference statistics for a given ancestry, trait, and model ID.
// It first attempts to fetch from the cache, then falls back to on-the-fly computation if needed.
func (s *ReferenceService) GetReferenceStats(ctx context.Context, ancestry *ancestry.Ancestry, trait string) (*reference_stats.ReferenceStats, error) {
	// Use ancestry code for cache operations
	ancestryCode := ancestry.Code()

	// Use trait as the model identifier for cache key
	stats, err := s.ReferenceCache.Get(ctx, reference_cache.StatsRequest{
		Ancestry: ancestryCode,
		Trait:    trait,
		ModelID:  trait,
	})
	if err != nil {
		// This can happen for cache misses, log and continue
		logging.Debug("Cache miss for trait %s: %v", trait, err)
	}
	if stats != nil {
		return stats, nil
	}

	// Cache miss, compute on the fly
	return s.computeAndCacheStats(ctx, ancestry, trait)
}

// computeAndCacheStats computes PRS statistics on the fly and caches the result.
func (s *ReferenceService) computeAndCacheStats(ctx context.Context, ancestry *ancestry.Ancestry, trait string) (*reference_stats.ReferenceStats, error) {
	// Load the PRS model
	prsModel, err := s.LoadModel(ctx, trait)
	if err != nil {
		return nil, fmt.Errorf("failed to load PRS model: %w", err)
	}

	// Get allele frequencies for this model's variants
	traitVariants := map[string][]model.Variant{trait: prsModel.Variants}
	alleleFrequencies, err := s.GetAlleleFrequenciesForTraits(ctx, traitVariants, ancestry)
	if err != nil {
		return nil, fmt.Errorf("failed to get allele frequencies for trait %s: %w", trait, err)
	}

	// Compute stats
	stats, err := reference_stats.Compute(alleleFrequencies[trait], prsModel.GetEffectSizes())
	if err != nil {
		return nil, fmt.Errorf("failed to compute stats for trait %s: %w", trait, err)
	}

	ancestryCode := ancestry.Code()
	stats.Ancestry = ancestryCode
	stats.Trait = trait
	stats.Model = trait // Use trait as the model identifier

	// Cache the result using ancestry code
	if err := s.ReferenceCache.Store(ctx, reference_cache.StatsRequest{
		Ancestry: ancestryCode,
		Trait:    trait,
		ModelID:  trait,
	}, stats); err != nil {
		logging.Warn("Failed to cache computed stats: %v", err)
	}

	return stats, nil
}
