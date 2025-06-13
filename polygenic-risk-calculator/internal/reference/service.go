package reference

import (
	"context"
	"fmt"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/db"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
	"phite.io/polygenic-risk-calculator/internal/logging"
	reference_cache "phite.io/polygenic-risk-calculator/internal/reference/cache"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
	"phite.io/polygenic-risk-calculator/internal/utils"

	model "phite.io/polygenic-risk-calculator/internal/reference/model"
)

func init() {
	// Register required configuration keys for reference service
	config.RegisterRequiredKey("reference.model_table")
	config.RegisterRequiredKey("reference.allele_freq_table")
	config.RegisterRequiredKey("reference.column_mapping")
	config.RegisterRequiredKey("reference.ancestry_mapping")
}

// ReferenceService handles loading PRS models and allele frequencies using the repository pattern
type ReferenceService struct {
	gnomadDB        dbinterface.Repository
	referenceCache  reference_cache.Cache
	modelTable      string
	alleleFreqTable string
	columnMapping   map[string]string
	ancestryMapping map[string]string
}

// NewReferenceService creates a new reference service
func NewReferenceService() *ReferenceService {
	gnomadDB, err := db.GetRepository(context.Background(), "bq")
	if err != nil {
		logging.Error("Failed to create ReferenceService: %v", err)
		return nil
	}
	referenceCache, err := reference_cache.NewRepositoryCache()
	if err != nil {
		logging.Error("Failed to create ReferenceService: %v", err)
		return nil
	}

	return &ReferenceService{
		gnomadDB:        gnomadDB,
		referenceCache:  referenceCache,
		modelTable:      config.GetString("reference.model_table"),
		alleleFreqTable: config.GetString("reference.allele_freq_table"),
		columnMapping:   config.GetStringMapString("reference.column_mapping"),
		ancestryMapping: config.GetStringMapString("reference.ancestry_mapping"),
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

// GetAlleleFrequencies retrieves allele frequencies for the given variants and ancestry
func (s *ReferenceService) GetAlleleFrequencies(ctx context.Context, variants []model.Variant, ancestry string) (map[string]float64, error) {
	if len(variants) == 0 {
		return map[string]float64{}, nil
	}

	// Get the ancestry-specific column name
	ancestryCol, ok := s.ancestryMapping[ancestry]
	if !ok {
		return nil, fmt.Errorf("unsupported ancestry: %s", ancestry)
	}

	// Build variant filters
	var filters []string
	var args []interface{}
	for _, v := range variants {
		chrom, pos, ref, alt, err := model.ParseVariantID(v.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid variant ID %s: %w", v.ID, err)
		}

		filters = append(filters, "(chrom = ? AND pos = ? AND ref = ? AND alt = ?)")
		args = append(args, chrom, pos, ref, alt)
	}

	// Build and execute query
	query := fmt.Sprintf(
		"SELECT chrom, pos, ref, alt, %s as freq FROM %s WHERE %s",
		ancestryCol,
		s.alleleFreqTable,
		strings.Join(filters, " OR "),
	)

	logging.Info("Querying allele frequencies for %d variants", len(variants))
	rows, err := s.gnomadDB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query allele frequencies: %w", err)
	}

	freqs := make(map[string]float64)
	for _, row := range rows {
		chrom := utils.ToString(row["chrom"])
		pos := utils.ToInt64(row["pos"])
		ref := utils.ToString(row["ref"])
		alt := utils.ToString(row["alt"])
		freq := utils.ToFloat64(row["freq"])

		variantID := model.FormatVariantID(chrom, pos, ref, alt)
		freqs[variantID] = freq
	}

	logging.Info("Retrieved allele frequencies for %d variants", len(freqs))
	return freqs, nil
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
func (s *ReferenceService) GetReferenceStats(ctx context.Context, ancestry, trait, modelID string) (*reference_stats.ReferenceStats, error) {
	// Try to get from cache first
	stats, err := s.referenceCache.Get(ctx, reference_cache.StatsRequest{
		Ancestry: ancestry,
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
func (s *ReferenceService) computeAndCacheStats(ctx context.Context, ancestry, trait, modelID string) (*reference_stats.ReferenceStats, error) {
	// Load the PRS model
	model, err := s.LoadModel(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load PRS model: %w", err)
	}

	// Get allele frequencies
	freqs, err := s.GetAlleleFrequencies(ctx, model.Variants, ancestry)
	if err != nil {
		return nil, fmt.Errorf("failed to get allele frequencies: %w", err)
	}

	// Compute statistics
	stats, err := reference_stats.Compute(freqs, model.GetEffectSizes())
	if err != nil {
		return nil, fmt.Errorf("failed to compute PRS statistics: %w", err)
	}

	// Add metadata
	stats.Ancestry = ancestry
	stats.Trait = trait
	stats.Model = modelID

	// Cache the result
	if err := s.referenceCache.Store(ctx, reference_cache.StatsRequest{
		Ancestry: ancestry,
		Trait:    trait,
		ModelID:  modelID,
	}, stats); err != nil {
		logging.Warn("Failed to cache computed stats: %v", err)
		// Don't return error, as we still have valid stats
	}

	return stats, nil
}
