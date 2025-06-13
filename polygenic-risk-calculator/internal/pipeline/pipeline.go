package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"phite.io/polygenic-risk-calculator/internal/genotype"
	gwas "phite.io/polygenic-risk-calculator/internal/gwas"
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

// Run executes the full polygenic risk pipeline.
func Run(input PipelineInput) (PipelineOutput, error) {
	logging.Info("Pipeline started with input: %+v", input)

	ctx := context.Background()
	if input.GenotypeFile == "" || input.ReferenceTable == "" || len(input.SNPs) == 0 {
		logging.Error("Missing required pipeline input: %+v", input)
		return PipelineOutput{}, errors.New("missing required input")
	}

	gwasService := gwas.NewGWASService()

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

	// For each trait, perform PRS calculation, normalization, and summary stub
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
		// Load reference stats for this trait

		refService := reference.NewReferenceService()
		refStats, err := refService.GetReferenceStats(ctx, input.Ancestry, trait, input.Model)
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
	riskAllele, _ := record["risk_allele"].(string)
	beta, _ := record["effect_size"].(float64)
	trait, _ := record["trait"].(string)

	return model.GWASSNPRecord{
		RSID:       rsid,
		RiskAllele: riskAllele,
		Beta:       beta,
		Trait:      trait,
	}
}
