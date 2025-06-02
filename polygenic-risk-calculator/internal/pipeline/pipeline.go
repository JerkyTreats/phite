package pipeline

import (
	"context"
	"errors"

	"phite.io/polygenic-risk-calculator/internal/clientsets/bigquery"
	"phite.io/polygenic-risk-calculator/internal/genotype"
	"phite.io/polygenic-risk-calculator/internal/gwas"
	gwasdata "phite.io/polygenic-risk-calculator/internal/gwas"
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/output"
	"phite.io/polygenic-risk-calculator/internal/prs"
	reference "phite.io/polygenic-risk-calculator/internal/reference"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// PipelineInput defines all inputs required for the risk calculation pipeline.
type PipelineInput struct {
	GenotypeFile   string
	SNPs           []string
	GWASDB         string
	GWASTable      string
	ReferenceDB    string
	ReferenceTable string // reference stats table name (default: reference_panel)
	OutputFormat   string
	OutputPath     string
	Ancestry       string // optional
	Model          string // optional
}

// PipelineOutput defines the results of the pipeline execution.
type PipelineOutput struct {
	TraitSummaries []output.TraitSummary
	NormalizedPRS  map[string]prs.NormalizedPRS // per trait
	PRSResults     map[string]prs.PRSResult     // per trait
	SNPSMissing    []string
}

// Run executes the full polygenic risk pipeline.
func Run(input PipelineInput) (PipelineOutput, error) {
	logging.Info("Pipeline started with input: %+v", input)

	ctx := context.Background()
	if input.GenotypeFile == "" || input.GWASDB == "" || input.GWASTable == "" || len(input.SNPs) == 0 {
		logging.Error("Missing required pipeline input: %+v", input)
		return PipelineOutput{}, errors.New("missing required input")
	}
	refTable := input.ReferenceTable
	if refTable == "" {
		refTable = "reference_panel"
	}

	// Step 1: Fetch GWAS records
	logging.Info("Fetching GWAS records from DB: %s, table: %s", input.GWASDB, input.GWASTable)
	gwasRecords, err := gwas.FetchGWASRecordsWithTable(input.GWASDB, input.GWASTable, input.SNPs)
	if err != nil {
		logging.Error("Failed to fetch GWAS records: %v", err)
		return PipelineOutput{}, err
	}
	logging.Info("Fetched %d GWAS records", len(gwasRecords))
	// Step 2: Parse genotype data
	logging.Info("Parsing genotype data from file: %s", input.GenotypeFile)
	genoOut, err := genotype.ParseGenotypeData(genotype.ParseGenotypeDataInput{
		GenotypeFilePath: input.GenotypeFile,
		RequestedRSIDs:   input.SNPs,
		GWASData:         gwasRecords,
	})
	if err != nil {
		logging.Error("Failed to parse genotype data: %v", err)
		return PipelineOutput{}, err
	}
	logging.Info("Parsed genotype data: %d SNPs validated, %d SNPs missing", len(genoOut.ValidatedSNPs), len(genoOut.SNPsMissing))

	// Step 3: Annotate SNPs
	annotated := gwasdata.FetchAndAnnotateGWAS(gwasdata.GWASDataFetcherInput{
		ValidatedSNPs:     genoOut.ValidatedSNPs,
		AssociationsClean: gwasdata.MapToGWASList(gwasRecords),
	})
	_ = annotated // TODO: use in next pipeline step

	// Step 4: Aggregate unique traits from annotated SNPs
	traitSet := make(map[string]struct{})
	for _, snp := range annotated.AnnotatedSNPs {
		if snp.Trait != "" {
			traitSet[snp.Trait] = struct{}{}
		}
	}

	// Step 5: For each trait, perform PRS calculation, normalization, and summary stub
	normPRSs := make(map[string]prs.NormalizedPRS)
	prsResults := make(map[string]prs.PRSResult)
	summaries := make([]output.TraitSummary, 0, len(traitSet))
	for trait := range traitSet {
		logging.Info("Processing trait: %s", trait)
		// Filter SNPs for this trait
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
		// Load reference stats for this trait
		backend, err := bigquery.NewClient(ctx)
		if err != nil {
			logging.Error("Failed to create BigQuery client for trait %s: %v", trait, err)
			return PipelineOutput{}, err
		}
		defer backend.Close()
		refBackend := reference.NewReferenceStatsLoader(backend)
		defer refBackend.Close()
		refStats, _ := refBackend.GetReferenceStats(ctx, input.Ancestry, trait, input.Model)
		// Normalize PRS
		var norm prs.NormalizedPRS
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
			norm, _ = prs.NormalizePRS(prsResult, modelRef)
		}
		normPRSs[trait] = norm
		// Generate real trait summary using output.GenerateTraitSummaries
		logging.Info("Generating trait summary for trait: %s", trait)
		ts := output.GenerateTraitSummaries(traitSNPs, norm)
		summaries = append(summaries, ts...)
	}

	logging.Info("Pipeline completed successfully. Traits processed: %d", len(traitSet))
	return PipelineOutput{
		TraitSummaries: summaries,
		NormalizedPRS:  normPRSs,
		PRSResults:     prsResults,
		SNPSMissing:    genoOut.SNPsMissing,
	}, nil
}
