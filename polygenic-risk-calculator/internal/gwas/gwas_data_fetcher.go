package gwas

import (
	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/model"
)

type GWASDataFetcherInput struct {
	ValidatedSNPs     []model.ValidatedSNP
	AssociationsClean []model.GWASSNPRecord
}

type GWASDataFetcherOutput struct {
	AnnotatedSNPs []model.AnnotatedSNP
	GWASRecords   []model.GWASSNPRecord
}

// FetchAndAnnotateGWAS fetches GWAS associations for validated SNPs and annotates them with risk allele, effect size, and computed dosage.
func FetchAndAnnotateGWAS(input GWASDataFetcherInput) GWASDataFetcherOutput {
	logging.Info("Starting GWAS annotation for %d SNPs", len(input.ValidatedSNPs))

	var result GWASDataFetcherOutput
	for _, snp := range input.ValidatedSNPs {
		found := false
		for _, assoc := range input.AssociationsClean {
			if assoc.RSID == snp.RSID {
				found = true
				dosage := computeDosage(snp.Genotype, assoc.RiskAllele)
				annotated := model.AnnotatedSNP{
					RSID:       snp.RSID,
					Genotype:   snp.Genotype,
					RiskAllele: assoc.RiskAllele,
					Beta:       assoc.Beta,
					Dosage:     dosage,
					Trait:      assoc.Trait,
				}
				result.AnnotatedSNPs = append(result.AnnotatedSNPs, annotated)
				result.GWASRecords = append(result.GWASRecords, assoc)
			}
		}
		if !found {
			logging.Info("No GWAS association found for SNP: %s", snp.RSID)
		}
	}
	logging.Info("GWAS annotation complete: %d SNPs annotated", len(result.AnnotatedSNPs))
	return result
}

// computeDosage calculates the count of the risk allele in the genotype string. Ambiguous/missing genotypes yield 0.
func computeDosage(genotype string, riskAllele string) int {
	if len(genotype) != 2 {
		return 0
	}
	if riskAllele == "" || genotype == "NN" || genotype == "--" {
		return 0
	}
	count := 0
	for i := 0; i < 2; i++ {
		if string(genotype[i]) == riskAllele {
			count++
		}
	}
	return count
}
