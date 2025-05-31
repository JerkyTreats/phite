package output

// GenerateTraitSummaries aggregates SNPs by trait and produces a summary for each trait.
// It assigns risk levels based on normalized PRS percentile: <20 = low, <80 = moderate, >=80 = high.
// Missing or empty trait names are grouped as "unknown".
import (
	"phite.io/polygenic-risk-calculator/internal/model"
	"phite.io/polygenic-risk-calculator/internal/prs"
)

// GenerateTraitSummaries aggregates SNPs by trait and produces a summary for each trait.
// It assigns risk levels based on normalized PRS percentile: <20 = low, <80 = moderate, >=80 = high.
// Missing or empty trait names are grouped as "unknown".
func GenerateTraitSummaries(snps []model.AnnotatedSNP, norm prs.NormalizedPRS) []TraitSummary {
	if len(snps) == 0 {
		return nil
	}
	traitMap := make(map[string]*TraitSummary)
	for _, snp := range snps {
		trait := snp.Trait
		if trait == "" {
			trait = "unknown"
		}
		ts, ok := traitMap[trait]
		if !ok {
			ts = &TraitSummary{Trait: trait}
			traitMap[trait] = ts
		}
		ts.NumRiskAlleles += snp.Dosage
		ts.EffectWeightedContribution += float64(snp.Dosage) * snp.Beta
	}
	// Assign risk level based on normalized PRS percentile
	riskLevel := "moderate"
	if norm.Percentile < 20 {
		riskLevel = "low"
	} else if norm.Percentile >= 80 {
		riskLevel = "high"
	}
	// Copy to slice
	summaries := make([]TraitSummary, 0, len(traitMap))
	for _, ts := range traitMap {
		ts.RiskLevel = riskLevel
		summaries = append(summaries, *ts)
	}
	return summaries
}
