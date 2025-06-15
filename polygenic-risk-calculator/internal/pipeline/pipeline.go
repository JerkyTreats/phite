package pipeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"phite.io/polygenic-risk-calculator/internal/ancestry"
	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/genotype"
	gwas "phite.io/polygenic-risk-calculator/internal/gwas"
	gwasdata "phite.io/polygenic-risk-calculator/internal/gwas"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/output"
	"phite.io/polygenic-risk-calculator/internal/prs"
	reference "phite.io/polygenic-risk-calculator/internal/reference"
	reference_cache "phite.io/polygenic-risk-calculator/internal/reference/cache"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

// PipelineInput defines all inputs required for the risk calculation pipeline.
type PipelineInput struct {
	GenotypeFile   string
	SNPs           []string
	ReferenceTable string // reference stats table name (default: reference_panel)
	OutputFormat   string
	OutputPath     string
	Config         *viper.Viper // Add config parameter
}

// PipelineOutput defines the results of the pipeline execution.
type PipelineOutput struct {
	TraitSummaries []output.TraitSummary
	NormalizedPRS  map[string]prs.NormalizedPRS // per trait
	PRSResults     map[string]prs.PRSResult     // per trait
	SNPSMissing    []string
}

// BulkOperationContext holds all cache operations that need to be executed in bulk.
type BulkOperationContext struct {
	CacheRequests []reference_cache.StatsRequest
	CacheEntries  []reference_cache.CacheEntry
	StatsRequests []reference.ReferenceStatsRequest // Added for batch reference stats computation
}

// Run executes the full polygenic risk pipeline using optimized bulk operations.
func Run(input PipelineInput) (PipelineOutput, error) {
	logging.Info("Pipeline started with bulk operations: %+v", input)

	ctx := context.Background()
	if input.GenotypeFile == "" || input.ReferenceTable == "" || len(input.SNPs) == 0 {
		logging.Error("Missing required pipeline input: %+v", input)
		return PipelineOutput{}, errors.New("missing required input")
	}

	// Initialize ancestry from configuration
	ancestryObj, err := ancestry.NewFromConfig()
	if err != nil {
		logging.Error("Failed to initialize ancestry from configuration: %v", err)
		return PipelineOutput{}, fmt.Errorf("failed to initialize ancestry: %w", err)
	}
	logging.Info("Initialized ancestry: %s (%s)", ancestryObj.Code(), ancestryObj.Description())

	gwasService := gwas.NewGWASService()
	if gwasService == nil {
		logging.Error("Failed to initialize GWAS service")
		return PipelineOutput{}, errors.New("failed to initialize GWAS service")
	}

	gwasRecords, err := gwasService.FetchGWASRecords(ctx, input.SNPs)
	if err != nil {
		logging.Error("Failed to fetch GWAS records: %v", err)
		return PipelineOutput{}, err
	}

	// Convert GWAS records to the expected format
	gwasMap := make(map[string]model.GWASSNPRecord)
	for _, record := range gwasRecords {
		gwasMap[record.RSID] = record
	}

	gwasList := gwasdata.MapToGWASList(gwasMap)

	logging.Info("Fetched %d GWAS records", len(gwasMap))
	logging.Info("Using GWAS list with %d records", len(gwasList))
	logging.Info("Parsing genotype data from file: %s", input.GenotypeFile)

	genoOut, err := genotype.ParseGenotypeData(genotype.ParseGenotypeDataInput{
		GenotypeFilePath: input.GenotypeFile,
		RequestedRSIDs:   input.SNPs,
		GWASData:         gwasMap,
	})
	if err != nil {
		logging.Error("Failed to parse genotype data: %v", err)
		return PipelineOutput{}, err
	}

	logging.Info("Parsed genotype data: %d SNPs validated, %d SNPs missing", len(genoOut.ValidatedSNPs), len(genoOut.SNPsMissing))

	annotated := gwasdata.FetchAndAnnotateGWAS(gwasdata.GWASDataFetcherInput{
		ValidatedSNPs:     genoOut.ValidatedSNPs,
		AssociationsClean: gwasdata.MapToGWASList(gwasMap),
	})

	traitSet := make(map[string]struct{})
	for _, snp := range annotated.AnnotatedSNPs {
		if snp.Trait != "" {
			traitSet[snp.Trait] = struct{}{}
		}
	}

	logging.Info("Processing %d traits with bulk operations", len(traitSet))

	modelId := config.GetString("reference.model")
	refService := reference.NewReferenceService()

	// Create cache instance for batch operations
	cache, err := reference_cache.NewRepositoryCache()
	if err != nil {
		logging.Error("Failed to create cache for bulk operations: %v", err)
		return PipelineOutput{}, fmt.Errorf("failed to create cache: %w", err)
	}

	// Phase 1: Pre-collect all cache requirements
	bulkCtx := &BulkOperationContext{
		CacheRequests: make([]reference_cache.StatsRequest, 0, len(traitSet)),
		CacheEntries:  make([]reference_cache.CacheEntry, 0),
	}

	ancestryCode := ancestryObj.Code()
	for trait := range traitSet {
		bulkCtx.CacheRequests = append(bulkCtx.CacheRequests, reference_cache.StatsRequest{
			Ancestry: ancestryCode,
			Trait:    trait,
			ModelID:  modelId,
		})
	}

	// Phase 2: Bulk cache lookup
	logging.Info("Executing bulk cache lookup for %d traits", len(bulkCtx.CacheRequests))
	cacheResults, err := cache.GetBatch(ctx, bulkCtx.CacheRequests)
	if err != nil {
		logging.Error("Failed to execute bulk cache lookup: %v", err)
		return PipelineOutput{}, fmt.Errorf("failed to execute bulk cache lookup: %w", err)
	}

	// Phase 3: Identify cache misses and compute stats with bulk operations
	cacheMisses := make([]string, 0)
	computedStats := make(map[string]*reference_stats.ReferenceStats)

	for trait := range traitSet {
		key := fmt.Sprintf("%s|%s|%s", ancestryCode, trait, modelId)
		if _, found := cacheResults[key]; !found {
			cacheMisses = append(cacheMisses, trait)
		}
	}

	if len(cacheMisses) > 0 {
		logging.Info("Computing reference stats for %d cache misses using bulk operations", len(cacheMisses))

		// Build batch reference stats requests for all cache misses
		statsRequests := make([]reference.ReferenceStatsRequest, 0, len(cacheMisses))
		for _, trait := range cacheMisses {
			statsRequests = append(statsRequests, reference.ReferenceStatsRequest{
				Ancestry: ancestryObj,
				Trait:    trait,
				ModelID:  modelId,
			})
		}

		// Single bulk computation for all reference stats
		logging.Info("Executing bulk reference stats computation for %d traits", len(cacheMisses))
		bulkStats, err := refService.GetReferenceStatsBatch(ctx, statsRequests)
		if err != nil {
			logging.Error("Failed to compute bulk reference stats: %v", err)
			return PipelineOutput{}, fmt.Errorf("failed to compute bulk reference stats: %w", err)
		}

		// Process bulk results and cache them
		for _, trait := range cacheMisses {
			key := fmt.Sprintf("%s|%s|%s", ancestryCode, trait, modelId)
			if stats, found := bulkStats[key]; found {
				computedStats[trait] = stats

				// Cache the result
				if err := cache.Store(ctx, reference_cache.StatsRequest{
					Ancestry: ancestryCode,
					Trait:    trait,
					ModelID:  modelId,
				}, stats); err != nil {
					logging.Warn("Failed to cache computed stats for trait %s: %v", trait, err)
					// Don't return error, as we still have valid stats
				}

				// Add to bulk cache entries for consistency tracking
				bulkCtx.CacheEntries = append(bulkCtx.CacheEntries, reference_cache.CacheEntry{
					Request: reference_cache.StatsRequest{
						Ancestry: ancestryCode,
						Trait:    trait,
						ModelID:  modelId,
					},
					Stats: stats,
				})
			} else {
				logging.Error("No reference stats computed for trait %s", trait)
				return PipelineOutput{}, fmt.Errorf("no reference stats computed for trait %s", trait)
			}
		}

		logging.Info("Successfully computed and cached reference stats for %d cache misses using bulk operations", len(cacheMisses))
	}

	// Phase 4: Process traits using cached and computed data
	normPRSs := make(map[string]prs.NormalizedPRS)
	prsResults := make(map[string]prs.PRSResult)
	summaries := make([]output.TraitSummary, 0, len(traitSet))

	for trait := range traitSet {
		logging.Info("Processing trait: %s", trait)
		traitSNPs := make([]model.AnnotatedSNP, 0)
		for _, snp := range annotated.AnnotatedSNPs {
			if snp.Trait == trait {
				traitSNPs = append(traitSNPs, snp)
			}
		}

		// Calculate PRS
		logging.Info("Calculating PRS for trait: %s", trait)
		prsResult := prs.CalculatePRS(traitSNPs)
		prsResults[trait] = prsResult

		// Get reference stats (from cache or computed)
		var refStats *reference_stats.ReferenceStats
		key := fmt.Sprintf("%s|%s|%s", ancestryCode, trait, modelId)

		if cachedStats, found := cacheResults[key]; found {
			refStats = cachedStats
		} else if computedStats[trait] != nil {
			refStats = computedStats[trait]
		} else {
			logging.Error("No reference stats available for trait %s", trait)
			return PipelineOutput{}, fmt.Errorf("no reference stats available for trait %s", trait)
		}

		// Normalize PRS
		if refStats != nil {
			logging.Info("Normalizing PRS for trait: %s", trait)
			modelRef := model.ReferenceStats{
				Mean:     refStats.Mean,
				Std:      refStats.Std,
				Min:      refStats.Min,
				Max:      refStats.Max,
				Ancestry: refStats.Ancestry,
				Trait:    refStats.Trait,
				Model:    refStats.Model,
			}
			norm, err := prs.NormalizePRS(prsResult, modelRef)
			if err != nil {
				logging.Error("Failed to normalize PRS for trait %s: %v", trait, err)
				return PipelineOutput{}, fmt.Errorf("failed to normalize PRS for trait %s: %w", trait, err)
			}
			normPRSs[trait] = norm
		}

		// Generate trait summary
		logging.Info("Generating trait summary for trait: %s", trait)
		ts := output.GenerateTraitSummaries(traitSNPs, normPRSs[trait])
		summaries = append(summaries, ts...)
	}

	// Phase 5: Bulk cache storage optimization will be implemented in future phases
	// Currently GetReferenceStats already handles individual caching

	logging.Info("Pipeline completed successfully with bulk operations. Traits processed: %d", len(traitSet))
	return PipelineOutput{
		TraitSummaries: summaries,
		NormalizedPRS:  normPRSs,
		PRSResults:     prsResults,
		SNPSMissing:    genoOut.SNPsMissing,
	}, nil
}
