package prs

// Canonical structs from data_model.md

type AnnotatedSNP struct {
	Rsid       string
	Genotype   string
	RiskAllele string
	Beta       float64
	Dosage     int
	Trait      string // optional
}

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
func CalculatePRS(snps []AnnotatedSNP) PRSResult {
	var total float64
	contributions := make([]SNPContribution, 0, len(snps))
	for _, snp := range snps {
		contribution := float64(snp.Dosage) * snp.Beta
		contributions = append(contributions, SNPContribution{
			Rsid:         snp.Rsid,
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
