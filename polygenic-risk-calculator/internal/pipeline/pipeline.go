package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"

	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
	"phite.io/polygenic-risk-calculator/internal/genotype"
	gwasdata "phite.io/polygenic-risk-calculator/internal/gwas"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/output"
	"phite.io/polygenic-risk-calculator/internal/prs"
	reference "phite.io/polygenic-risk-calculator/internal/reference"
)

// PipelineInput defines all inputs required for the risk calculation pipeline.
type PipelineInput struct {
	GenotypeFile   string
	SNPs           []string
	GWASRepository dbinterface.Repository
	GWASTable      string
	RefRepository  dbinterface.Repository
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
	if input.GenotypeFile == "" || input.GWASRepository == nil || input.GWASTable == "" || len(input.SNPs) == 0 {
		logging.Error("Missing required pipeline input: %+v", input)
		return PipelineOutput{}, errors.New("missing required input")
	}
	refTable := input.ReferenceTable
	if refTable == "" {
		refTable = "reference_panel"
	}

	// Step 1: Fetch GWAS records using repository
	logging.Info("Fetching GWAS records from table: %s", input.GWASTable)
	query := fmt.Sprintf("SELECT * FROM %s WHERE rsid IN (%s)", input.GWASTable, buildPlaceholders(len(input.SNPs)))
	args := make([]interface{}, len(input.SNPs))
	for i, snp := range input.SNPs {
		args[i] = snp
	}

	gwasRecords, err := input.GWASRepository.Query(ctx, query, args...)
	if err != nil {
		logging.Error("Failed to fetch GWAS records: %v", err)
		return PipelineOutput{}, err
	}

	// Convert query results to GWAS records
	gwasMap := make(map[string]model.GWASSNPRecord)
	for _, record := range gwasRecords {
		gwasRecord := convertToGWASRecord(record)
		gwasMap[gwasRecord.RSID] = gwasRecord
	}

	logging.Info("Fetched %d GWAS records", len(gwasMap))

	// Step 2: Parse genotype data
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

	// Step 3: Annotate SNPs
	annotated := gwasdata.FetchAndAnnotateGWAS(gwasdata.GWASDataFetcherInput{
		ValidatedSNPs:     genoOut.ValidatedSNPs,
		AssociationsClean: gwasdata.MapToGWASList(gwasMap),
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
		if input.RefRepository != nil {
			refBackend := reference.NewReferenceStatsLoader(input.RefRepository)
			defer refBackend.Close()
			refStats, err := refBackend.GetReferenceStats(ctx, input.Ancestry, trait, input.Model)
			if err != nil {
				logging.Error("Failed to get reference stats for trait %s (ancestry: %s, model: %s): %v", trait, input.Ancestry, input.Model, err)
				return PipelineOutput{}, fmt.Errorf("failed to get reference stats for trait %s: %w", trait, err)
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
		}
		// Generate real trait summary using output.GenerateTraitSummaries
		logging.Info("Generating trait summary for trait: %s", trait)
		ts := output.GenerateTraitSummaries(traitSNPs, normPRSs[trait])
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

// Helper function to build SQL placeholders
func buildPlaceholders(count int) string {
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ",")
}

// Helper function to convert map to GWASSNPRecord
func convertToGWASRecord(record map[string]interface{}) model.GWASSNPRecord {
	// Get values with type assertions, defaulting to zero values if missing
	rsid, _ := record["rsid"].(string)
	riskAllele, _ := record["effect_allele"].(string)
	beta, _ := record["effect_size"].(float64)
	trait, _ := record["trait"].(string)

	return model.GWASSNPRecord{
		RSID:       rsid,
		RiskAllele: riskAllele,
		Beta:       beta,
		Trait:      trait,
	}
}
