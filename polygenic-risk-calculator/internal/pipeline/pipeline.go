package pipeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"phite.io/polygenic-risk-calculator/internal/ancestry"
	"phite.io/polygenic-risk-calculator/internal/genotype"
	gwas "phite.io/polygenic-risk-calculator/internal/gwas"
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
	Errors         []error
}

// PipelineRequirements holds all data requirements identified during analysis phase
type PipelineRequirements struct {
	TraitSet      map[string]struct{}
	CacheKeys     []reference_cache.StatsRequest
	StatsRequests []reference.ReferenceStatsRequest
	AncestryObj   *ancestry.Ancestry
}

// BulkDataContext holds all data retrieved in bulk operations
type BulkDataContext struct {
	AlleleFrequencies map[string]map[string]float64              // trait -> variant -> freq
	CachedStats       map[string]*reference_stats.ReferenceStats // cache hits
	ComputedStats     map[string]*reference_stats.ReferenceStats // computed for cache misses
	PRSModels         map[string]*model.PRSModel                 // trait -> model
	TraitSNPs         map[string][]model.AnnotatedSNP            // trait -> SNPs
	Errors            []error
}

// ProcessingResults holds all computed results before final output
type ProcessingResults struct {
	TraitSummaries []output.TraitSummary
	NormalizedPRS  map[string]prs.NormalizedPRS
	PRSResults     map[string]prs.PRSResult
	CacheEntries   []reference_cache.CacheEntry // for bulk storage
	Errors         []error
}

// BulkOperationContext holds all cache operations that need to be executed in bulk.
type BulkOperationContext struct {
	CacheRequests []reference_cache.StatsRequest
	CacheEntries  []reference_cache.CacheEntry
	StatsRequests []reference.ReferenceStatsRequest
}

// Run executes the polygenic risk pipeline in bulk operations.
// This implementation uses a 4-phase approach:
// Phase 1: Requirements Analysis - Pre-analyze all data needs
// Phase 2: Bulk Data Retrieval - Execute minimal BigQuery operations
// Phase 3: In-Memory Processing - Process all traits using cached data
// Phase 4: Bulk Storage - Store all results in single operation
func Run(input PipelineInput, refService ...*reference.ReferenceService) (PipelineOutput, error) {
	logging.Info("Starting pipeline: %+v", input)

	ctx := context.Background()
	if input.GenotypeFile == "" || input.ReferenceTable == "" || len(input.SNPs) == 0 {
		logging.Error("Missing required pipeline input: %+v", input)
		return PipelineOutput{}, errors.New("missing required input")
	}

	// Use provided reference service or create default
	var rs *reference.ReferenceService
	var err error
	if len(refService) > 0 && refService[0] != nil {
		rs = refService[0]
	} else {
		rs, err = reference.NewReferenceService(nil, nil, nil)
		if err != nil {
			return PipelineOutput{}, fmt.Errorf("failed to create reference service: %w", err)
		}
	}

	// ==================== PHASE 1: REQUIREMENTS ANALYSIS ====================
	logging.Info("Phase 1: Analyzing all pipeline requirements...")
	requirements, genoOut, annotated, err := analyzeAllRequirements(ctx, input)
	if err != nil {
		logging.Error("Phase 1 failed - Requirements analysis error: %v", err)
		return PipelineOutput{}, fmt.Errorf("requirements analysis failed: %w", err)
	}
	logging.Info("Phase 1 complete: %d traits, %d cache requests, %d stats requests",
		len(requirements.TraitSet), len(requirements.CacheKeys), len(requirements.StatsRequests))

	// ==================== PHASE 2: BULK DATA RETRIEVAL ====================
	logging.Info("Phase 2: Executing bulk data retrieval operations...")
	bulkData, err := retrieveAllDataBulk(ctx, requirements, &annotated, rs)
	if err != nil {
		logging.Error("Phase 2 failed - Bulk data retrieval error: %v", err)
		return PipelineOutput{}, fmt.Errorf("bulk data retrieval failed: %w", err)
	}
	logging.Info("Phase 2 complete: Retrieved data for %d traits with %d cache hits, %d computed stats",
		len(requirements.TraitSet), len(bulkData.CachedStats), len(bulkData.ComputedStats))
	if len(bulkData.Errors) > 0 {
		logging.Warn("Phase 2 encountered %d errors.", len(bulkData.Errors))
		for _, e := range bulkData.Errors {
			logging.Error("- %v", e)
		}
	}

	// ==================== PHASE 3: IN-MEMORY PROCESSING ====================
	logging.Info("Phase 3: Processing all traits in-memory...")
	results, err := processAllTraitsInMemory(requirements, bulkData)
	if err != nil {
		logging.Error("Phase 3 failed - In-memory processing error: %v", err)
		return PipelineOutput{}, fmt.Errorf("in-memory processing failed: %w", err)
	}
	logging.Info("Phase 3 complete: Processed %d traits, generated %d summaries",
		len(requirements.TraitSet), len(results.TraitSummaries))
	if len(results.Errors) > 0 {
		logging.Warn("Phase 3 finished with %d total errors.", len(results.Errors))
	}

	// ==================== PHASE 4: BULK STORAGE ====================
	if len(results.Errors) >= 10 {
		logging.Error("Error cap reached. Aborting before Phase 4 (Bulk Storage).")
	} else {
		logging.Info("Phase 4: Executing bulk storage operations...")
		err = storeBulkResults(ctx, results, rs)
		if err != nil {
			logging.Error("Phase 4 failed - Bulk storage error: %v", err)
			return PipelineOutput{}, fmt.Errorf("bulk storage failed: %w", err)
		}
		logging.Info("Phase 4 complete: Stored %d cache entries", len(results.CacheEntries))
	}

	logging.Info("Optimized pipeline completed successfully. Total traits processed: %d", len(requirements.TraitSet))

	return PipelineOutput{
		TraitSummaries: results.TraitSummaries,
		NormalizedPRS:  results.NormalizedPRS,
		PRSResults:     results.PRSResults,
		SNPSMissing:    genoOut.SNPsMissing,
		Errors:         results.Errors,
	}, nil
}

// analyzeAllRequirements performs comprehensive analysis of all pipeline data requirements
func analyzeAllRequirements(ctx context.Context, input PipelineInput) (*PipelineRequirements, genotype.ParseGenotypeDataOutput, gwas.GWASDataFetcherOutput, error) {
	// Initialize ancestry from configuration
	ancestryObj, err := ancestry.NewFromConfig()
	if err != nil {
		return nil, genotype.ParseGenotypeDataOutput{}, gwas.GWASDataFetcherOutput{}, fmt.Errorf("failed to initialize ancestry: %w", err)
	}

	// Fetch GWAS data
	gwasService := gwas.NewGWASService()
	if gwasService == nil {
		return nil, genotype.ParseGenotypeDataOutput{}, gwas.GWASDataFetcherOutput{}, errors.New("failed to initialize GWAS service")
	}

	gwasRecords, err := gwasService.FetchGWASRecords(ctx, input.SNPs)
	if err != nil {
		return nil, genotype.ParseGenotypeDataOutput{}, gwas.GWASDataFetcherOutput{}, fmt.Errorf("failed to fetch GWAS records: %w", err)
	}

	gwasMap := make(map[string]model.GWASSNPRecord)
	for _, record := range gwasRecords {
		gwasMap[record.RSID] = record
	}

	// Parse genotype data
	genoOut, err := genotype.ParseGenotypeData(genotype.ParseGenotypeDataInput{
		GenotypeFilePath: input.GenotypeFile,
		RequestedRSIDs:   input.SNPs,
		GWASData:         gwasMap,
	})
	if err != nil {
		return nil, genotype.ParseGenotypeDataOutput{}, gwas.GWASDataFetcherOutput{}, fmt.Errorf("failed to parse genotype data: %w", err)
	}

	// Annotate GWAS data
	annotated := gwas.FetchAndAnnotateGWAS(gwas.GWASDataFetcherInput{
		ValidatedSNPs:     genoOut.ValidatedSNPs,
		AssociationsClean: gwas.MapToGWASList(gwasMap),
	})

	// Identify all traits
	traitSet := make(map[string]struct{})
	for _, snp := range annotated.AnnotatedSNPs {
		if snp.Trait != "" {
			traitSet[snp.Trait] = struct{}{}
		}
	}

	// Build cache requests for all traits
	cacheKeys := make([]reference_cache.StatsRequest, 0, len(traitSet))
	for trait := range traitSet {
		cacheKeys = append(cacheKeys, reference_cache.StatsRequest{
			Ancestry: ancestryObj.Code(),
			Trait:    trait,
			ModelID:  trait,
		})
	}

	requirements := &PipelineRequirements{
		TraitSet:    traitSet,
		CacheKeys:   cacheKeys,
		AncestryObj: ancestryObj,
	}

	return requirements, genoOut, annotated, nil
}

// retrieveAllDataBulk executes all required BigQuery operations in minimal bulk calls
func retrieveAllDataBulk(ctx context.Context, requirements *PipelineRequirements, annotated *gwas.GWASDataFetcherOutput, refService *reference.ReferenceService) (*BulkDataContext, error) {
	// BULK OPERATION 1: Single bulk cache lookup for all traits
	logging.Info("Executing bulk cache lookup for %d traits", len(requirements.CacheKeys))
	cacheResults, err := refService.ReferenceCache.GetBatch(ctx, requirements.CacheKeys)
	if err != nil {
		return nil, fmt.Errorf("bulk cache lookup failed: %w", err)
	}

	// Identify cache misses and prepare for bulk stats computation
	cacheMisses := make([]string, 0)
	ancestryCode := requirements.AncestryObj.Code()

	for trait := range requirements.TraitSet {
		key := fmt.Sprintf("%s|%s|%s", ancestryCode, trait, trait)
		if _, found := cacheResults[key]; !found {
			cacheMisses = append(cacheMisses, trait)
		}
	}

	computedStats := make(map[string]*reference_stats.ReferenceStats)
	var allErrors []error

	if len(cacheMisses) > 0 {
		// BULK OPERATION 3: Single bulk reference stats computation for all cache misses
		logging.Info("Computing reference stats for %d cache misses", len(cacheMisses))

		statsRequests := make([]reference.ReferenceStatsRequest, 0, len(cacheMisses))
		for _, trait := range cacheMisses {
			statsRequests = append(statsRequests, reference.ReferenceStatsRequest{
				Ancestry: requirements.AncestryObj,
				Trait:    trait,
			})
		}

		// This returns
		bulkStats, errs := refService.GetReferenceStatsBatch(ctx, statsRequests)
		if len(errs) > 0 {
			logging.Warn("Encountered %d errors during bulk reference stats computation", len(errs))
			allErrors = append(allErrors, errs...)
			if len(allErrors) >= 10 {
				return nil, fmt.Errorf("error cap of 10 reached, aborting pipeline")
			}
		}

		// Process bulk stats results
		for _, trait := range cacheMisses {
			key := fmt.Sprintf("%s|%s|%s", ancestryCode, trait, trait)
			if stats, found := bulkStats[key]; found {
				computedStats[trait] = stats
			} else {
				logging.Warn("No reference stats computed for trait %s, likely due to a processing error.", trait)
			}
		}
	}

	// Organize trait SNPs for processing
	traitSNPs := make(map[string][]model.AnnotatedSNP)
	for trait := range requirements.TraitSet {
		traitSNPs[trait] = make([]model.AnnotatedSNP, 0)
	}

	for _, snp := range annotated.AnnotatedSNPs {
		if snp.Trait != "" {
			if _, exists := traitSNPs[snp.Trait]; exists {
				traitSNPs[snp.Trait] = append(traitSNPs[snp.Trait], snp)
			}
		}
	}

	return &BulkDataContext{
		CachedStats:   cacheResults,
		ComputedStats: computedStats,
		TraitSNPs:     traitSNPs,
		Errors:        allErrors,
	}, nil
}

// processAllTraitsInMemory processes all traits using pre-loaded bulk data
func processAllTraitsInMemory(requirements *PipelineRequirements, bulkData *BulkDataContext) (*ProcessingResults, error) {
	normPRSs := make(map[string]prs.NormalizedPRS)
	prsResults := make(map[string]prs.PRSResult)
	summaries := make([]output.TraitSummary, 0, len(requirements.TraitSet))
	cacheEntries := make([]reference_cache.CacheEntry, 0)
	pipelineErrors := make([]error, 0)

	ancestryCode := requirements.AncestryObj.Code()

	// Process each trait using pre-loaded data
	for trait := range requirements.TraitSet {
		logging.Info("Processing trait: %s", trait)

		traitSNPs := bulkData.TraitSNPs[trait]
		if len(traitSNPs) == 0 {
			logging.Warn("No SNPs found for trait %s, skipping", trait)
			continue
		}

		// Calculate PRS using pre-loaded data
		prsResult, err := prs.CalculatePRS(traitSNPs)
		if err != nil {
			err = fmt.Errorf("failed to calculate PRS for trait %s: %w", trait, err)
			pipelineErrors = append(pipelineErrors, err)
			logging.Error(err.Error())
			continue
		}
		prsResults[trait] = prsResult

		// Get reference stats (from cache or computed)
		var refStats *reference_stats.ReferenceStats
		key := fmt.Sprintf("%s|%s|%s", ancestryCode, trait, trait)

		if cachedStats, found := bulkData.CachedStats[key]; found {
			refStats = cachedStats
		} else if computedStats, found := bulkData.ComputedStats[trait]; found {
			refStats = computedStats

			// Prepare cache entry for bulk storage
			cacheEntries = append(cacheEntries, reference_cache.CacheEntry{
				Request: reference_cache.StatsRequest{
					Ancestry: ancestryCode,
					Trait:    trait,
					ModelID:  trait,
				},
				Stats: refStats,
			})
		} else {
			logging.Warn("No reference stats available for trait %s, skipping processing. Error likely occurred in Phase 2.", trait)
			continue
		}

		// Normalize PRS using pre-loaded reference stats
		if refStats != nil {
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
				err = fmt.Errorf("failed to normalize PRS for trait %s: %w", trait, err)
				pipelineErrors = append(pipelineErrors, err)
				logging.Error(err.Error())
				continue
			}
			normPRSs[trait] = norm
		}

		// Generate trait summary only if normalization was successful
		if norm, ok := normPRSs[trait]; ok {
			ts := output.GenerateTraitSummaries(traitSNPs, norm)
			summaries = append(summaries, ts...)
		}
	}

	allErrors := append(bulkData.Errors, pipelineErrors...)

	return &ProcessingResults{
		TraitSummaries: summaries,
		NormalizedPRS:  normPRSs,
		PRSResults:     prsResults,
		CacheEntries:   cacheEntries,
		Errors:         allErrors,
	}, nil
}

// storeBulkResults executes all storage operations in bulk
func storeBulkResults(ctx context.Context, results *ProcessingResults, refService *reference.ReferenceService) error {
	if len(results.CacheEntries) == 0 {
		logging.Info("No cache entries to store (all cache hits)")
		return nil
	}

	// BULK OPERATION 4: Single bulk cache storage for all computed stats
	logging.Info("Executing bulk cache storage for %d entries", len(results.CacheEntries))
	err := refService.ReferenceCache.StoreBatch(ctx, results.CacheEntries)
	if err != nil {
		return fmt.Errorf("bulk cache storage failed: %w", err)
	}

	return nil
}
