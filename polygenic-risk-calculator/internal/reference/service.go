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

func init() {
	// Register required configuration keys for reference service
	config.RegisterRequiredKey("reference.model_table")
	config.RegisterRequiredKey("reference.allele_freq_table")
	config.RegisterRequiredKey("reference.column_mapping")

	// User GCP configuration for billing
	config.RegisterRequiredKey("user.gcp_project") // For billing project

	// Cache configuration for user's private BigQuery dataset
	config.RegisterRequiredKey("cache.gcp_project") // Cache storage project
	config.RegisterRequiredKey("cache.dataset")     // Cache dataset name
}

// ReferenceService handles loading PRS models and allele frequencies using the repository pattern
type ReferenceService struct {
	gnomadDB        dbinterface.Repository
	referenceCache  reference_cache.Cache
	modelTable      string
	alleleFreqTable string
	columnMapping   map[string]string
}

// NewReferenceService creates a new reference service
func NewReferenceService() *ReferenceService {
	// gnomAD public data repository (read-only)
	gnomadDB, err := db.GetRepository(context.Background(), "bq", map[string]string{
		"project_id":      "bigquery-public-data",
		"dataset_id":      "gnomad",
		"billing_project": config.GetString("user.gcp_project"),
	})
	if err != nil {
		logging.Error("Failed to create gnomAD repository: %v", err)
		return nil
	}

	// Cache repository (read-write to user's project)
	cacheParams := map[string]string{
		"project_id":      config.GetString("cache.gcp_project"),
		"dataset_id":      config.GetString("cache.dataset"),
		"billing_project": config.GetString("user.gcp_project"),
	}
	referenceCache, err := reference_cache.NewRepositoryCache(cacheParams)
	if err != nil {
		logging.Error("Failed to create cache repository: %v", err)
		return nil
	}

	return &ReferenceService{
		gnomadDB:        gnomadDB,
		referenceCache:  referenceCache,
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
	}
}

// LoadModel loads a PRS model from the configured table
func (s *ReferenceService) LoadModel(ctx context.Context, modelID string) (*model.PRSModel, error) {
	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE %s = ?",
		s.modelTable,
		s.columnMapping["model_id"],
	)

	logging.Info("Loading PRS model: %s", modelID)
	rows, err := s.gnomadDB.Query(ctx, query, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query model: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no variants found for model: %s", modelID)
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
		ID:       modelID,
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
		chrom, pos, ref, alt, err := model.ParseVariantID(v.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid variant ID %s: %w", v.ID, err)
		}

		filters = append(filters, "(chrom = ? AND pos = ? AND ref = ? AND alt = ?)")
		args = append(args, chrom, pos, ref, alt)
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
	id := utils.ToString(row[s.columnMapping["id"]])
	if id == "" {
		return model.Variant{}, fmt.Errorf("missing or invalid ID")
	}

	effectWeight := utils.ToFloat64(row[s.columnMapping["effect_weight"]])
	if effectWeight == 0 {
		return model.Variant{}, fmt.Errorf("missing or invalid effect weight")
	}

	effectAllele := utils.ToString(row[s.columnMapping["effect_allele"]])
	if effectAllele == "" {
		return model.Variant{}, fmt.Errorf("missing or invalid effect allele")
	}

	// Optional fields
	otherAllele := utils.ToString(row[s.columnMapping["other_allele"]])

	var effectFreq *float64
	if val := utils.ToFloat64(row[s.columnMapping["effect_freq"]]); val != 0 {
		effectFreq = &val
	}

	// Parse variant ID to get chromosome and position
	chrom, pos, _, _, err := model.ParseVariantID(id)
	if err != nil {
		return model.Variant{}, fmt.Errorf("invalid variant ID: %w", err)
	}

	return model.Variant{
		ID:           id,
		Chromosome:   chrom,
		Position:     pos,
		EffectAllele: effectAllele,
		OtherAllele:  otherAllele,
		EffectWeight: effectWeight,
		EffectFreq:   effectFreq,
	}, nil
}

// GetReferenceStats retrieves PRS reference statistics for a given ancestry, trait, and model ID.
// It first attempts to fetch from the cache, then falls back to on-the-fly computation if needed.
func (s *ReferenceService) GetReferenceStats(ctx context.Context, ancestry *ancestry.Ancestry, trait, modelID string) (*reference_stats.ReferenceStats, error) {
	// Use ancestry code for cache operations
	ancestryCode := ancestry.Code()

	// Try to get from cache first
	stats, err := s.referenceCache.Get(ctx, reference_cache.StatsRequest{
		Ancestry: ancestryCode,
		Trait:    trait,
		ModelID:  modelID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get stats from cache: %w", err)
	}
	if stats != nil {
		return stats, nil
	}

	// Cache miss, compute on the fly
	return s.computeAndCacheStats(ctx, ancestry, trait, modelID)
}

// computeAndCacheStats computes PRS statistics on the fly and caches the result.
func (s *ReferenceService) computeAndCacheStats(ctx context.Context, ancestry *ancestry.Ancestry, trait, modelID string) (*reference_stats.ReferenceStats, error) {
	// Load the PRS model
	prsModel, err := s.LoadModel(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load PRS model: %w", err)
	}

	// Get allele frequencies using ancestry object
	freqs, err := s.GetAlleleFrequenciesForTraits(ctx, map[string][]model.Variant{
		trait: prsModel.Variants,
	}, ancestry)
	if err != nil {
		return nil, fmt.Errorf("failed to get allele frequencies: %w", err)
	}

	// Compute statistics
	stats, err := reference_stats.Compute(freqs[trait], prsModel.GetEffectSizes())
	if err != nil {
		return nil, fmt.Errorf("failed to compute PRS statistics: %w", err)
	}

	// Add metadata using ancestry code
	ancestryCode := ancestry.Code()
	stats.Ancestry = ancestryCode
	stats.Trait = trait
	stats.Model = modelID

	// Cache the result using ancestry code
	if err := s.referenceCache.Store(ctx, reference_cache.StatsRequest{
		Ancestry: ancestryCode,
		Trait:    trait,
		ModelID:  modelID,
	}, stats); err != nil {
		logging.Warn("Failed to cache computed stats: %v", err)
		// Don't return error, as we still have valid stats
	}

	return stats, nil
}
