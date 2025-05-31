package prs

import "phite.io/polygenic-risk-calculator/internal/model"

// Canonical structs from data_model.md

type SNPContribution struct {
	Rsid         string
	Dosage       int
	Beta         float64
	Contribution float64
}

type PRSResult struct {
	PRSScore float64
	Details  []SNPContribution
}

// CalculatePRS computes the polygenic risk score for a set of SNPs.
func CalculatePRS(snps []model.AnnotatedSNP) PRSResult {
	var total float64
	contributions := make([]SNPContribution, 0, len(snps))
	for _, snp := range snps {
		contribution := float64(snp.Dosage) * snp.Beta
		contributions = append(contributions, SNPContribution{
			Rsid:         snp.RSID,
			Dosage:       snp.Dosage,
			Beta:         snp.Beta,
			Contribution: contribution,
		})
		total += contribution
	}
	return PRSResult{
		PRSScore: total,
		Details:  contributions,
	}
}
