package pipeline

import (
	"errors"
	"phite.io/polygenic-risk-calculator/internal/output"
	"phite.io/polygenic-risk-calculator/internal/prs"
	"phite.io/polygenic-risk-calculator/internal/gwas"
	"phite.io/polygenic-risk-calculator/internal/genotype"
	gwasdata "phite.io/polygenic-risk-calculator/internal/gwas"
	reference "phite.io/polygenic-risk-calculator/internal/reference"
	"phite.io/polygenic-risk-calculator/internal/model"
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
	if input.GenotypeFile == "" || input.GWASDB == "" || input.GWASTable == "" || len(input.SNPs) == 0 {
		return PipelineOutput{}, errors.New("missing required input")
	}
	refTable := input.ReferenceTable
	if refTable == "" {
		refTable = "reference_panel"
	}

	// Step 1: Fetch GWAS records
	gwasRecords, err := gwas.FetchGWASRecordsWithTable(input.GWASDB, input.GWASTable, input.SNPs)
	if err != nil {
		return PipelineOutput{}, err
	}
	// Step 2: Parse genotype data
	genoOut, err := genotype.ParseGenotypeData(genotype.ParseGenotypeDataInput{
		GenotypeFilePath: input.GenotypeFile,
		RequestedRSIDs:   input.SNPs,
		GWASData:         gwasRecords,
	})
	if err != nil {
		return PipelineOutput{}, err
	}
	_ = genoOut // TODO: use in next pipeline step

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
		// Filter SNPs for this trait
		traitSNPs := make([]model.AnnotatedSNP, 0)
		for _, snp := range annotated.AnnotatedSNPs {
			if snp.Trait == trait {
				traitSNPs = append(traitSNPs, snp)
			}
		}
		// Calculate PRS
		prsResult := prs.CalculatePRS(traitSNPs)
		prsResults[trait] = prsResult
		// Load reference stats for this trait
		refStats, _ := reference.LoadReferenceStatsFromDuckDB(input.ReferenceDB, refTable, input.Ancestry, trait, input.Model)
		// Normalize PRS
		var norm prs.NormalizedPRS
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
			norm, _ = prs.NormalizePRS(prsResult, modelRef)
		}
		normPRSs[trait] = norm
		// Generate real trait summary using output.GenerateTraitSummaries
		ts := output.GenerateTraitSummaries(traitSNPs, norm)
		summaries = append(summaries, ts...)
	}

	return PipelineOutput{
		TraitSummaries: summaries,
		NormalizedPRS:  normPRSs,
		PRSResults:     prsResults,
		SNPSMissing:    genoOut.SNPsMissing,
	}, nil
}
