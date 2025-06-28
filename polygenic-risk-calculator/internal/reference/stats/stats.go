package reference_stats

import (
	"fmt"
	"math"

	"phite.io/polygenic-risk-calculator/internal/model"
)

// ReferenceStats wraps model.ReferenceStats and provides domain-specific helpers.
// It has identical underlying structure, enabling direct field access.
type ReferenceStats model.ReferenceStats

// Validate checks if the statistics are valid.
func (s *ReferenceStats) Validate() error {
	if s.Std <= 0 {
		return fmt.Errorf("standard deviation must be positive, got %f", s.Std)
	}
	if s.Min > s.Max {
		return fmt.Errorf("min (%f) cannot be greater than max (%f)", s.Min, s.Max)
	}
	if s.Mean < s.Min || s.Mean > s.Max {
		return fmt.Errorf("mean (%f) must be between min (%f) and max (%f)", s.Mean, s.Min, s.Max)
	}
	return nil
}

// NormalizePRS converts a raw PRS score to a normalized percentile.
func (s *ReferenceStats) NormalizePRS(rawPRS float64) (float64, error) {
	if err := s.Validate(); err != nil {
		return 0, fmt.Errorf("invalid reference stats: %w", err)
	}
	zScore := (rawPRS - s.Mean) / s.Std
	percentile := 0.5 * (1 + math.Erf(zScore/math.Sqrt(2)))
	return percentile, nil
}

// Compute calculates PRS statistics from allele frequencies and effect sizes using
// CORRECT population parameter formulas (not sample statistics).
//
// Population Mean: μ_pop = Σ_j(2*p_j*β_j)
// Population Variance: Var(PRS) = Σ_j(2*p_j*(1-p_j)*β_j²)
//
// This implements the industry-standard 2025 PRS methodology with proper
// Hardy-Weinberg equilibrium assumptions.
func Compute(alleleFreqs map[string]float64, effectSizes map[string]float64) (*ReferenceStats, error) {
	if len(alleleFreqs) == 0 || len(effectSizes) == 0 {
		return nil, fmt.Errorf("empty allele frequencies or effect sizes")
	}

	var populationMean float64
	var populationVariance float64
	var minContribution float64 = math.MaxFloat64
	var maxContribution float64 = -math.MaxFloat64
	var validVariants int

	// Calculate population parameters using CORRECT formulas
	for variant, freq := range alleleFreqs {
		effect, ok := effectSizes[variant]
		if !ok {
			continue // Skip variants without effect sizes
		}

		// Validate frequency bounds [0, 1]
		if freq < 0 || freq > 1 {
			return nil, fmt.Errorf("invalid allele frequency %f for variant %s: must be in [0,1]", freq, variant)
		}

		validVariants++

		// Population mean: μ_pop = Σ_j(2*p_j*β_j)
		contributionToMean := 2 * freq * effect
		populationMean += contributionToMean

		// Population variance: Var(PRS) = Σ_j(2*p_j*(1-p_j)*β_j²)
		// This is the Hardy-Weinberg equilibrium variance formula
		contributionToVariance := 2 * freq * (1 - freq) * effect * effect
		populationVariance += contributionToVariance

		// Track min/max individual contributions for bounds
		if contributionToMean < minContribution {
			minContribution = contributionToMean
		}
		if contributionToMean > maxContribution {
			maxContribution = contributionToMean
		}
	}

	if validVariants == 0 {
		return nil, fmt.Errorf("no matching variants found between allele frequencies and effect sizes")
	}

	// Population standard deviation
	populationStd := math.Sqrt(populationVariance)

	// Validate mathematical consistency
	if populationVariance < 0 {
		return nil, fmt.Errorf("negative population variance %f: mathematical error", populationVariance)
	}

	// For single variant case, ensure we don't have zero variance (unless mathematically expected)
	if validVariants == 1 && populationVariance == 0 {
		// Zero variance is mathematically correct in these cases:
		// 1. Effect size is zero (β = 0)
		// 2. Allele frequency is 0 or 1 (fixed allele, no segregation)
		var isValidZeroVariance bool
		for variant, freq := range alleleFreqs {
			if effect, ok := effectSizes[variant]; ok {
				// Valid cases for zero variance
				if math.Abs(effect) <= 1e-15 || freq == 0.0 || freq == 1.0 {
					isValidZeroVariance = true
					break
				}
			}
		}
		if !isValidZeroVariance {
			return nil, fmt.Errorf("zero variance detected with non-zero effect and segregating allele: mathematical error in population variance calculation")
		}
	}

	// Estimate reasonable bounds for population distribution
	// For a population following HWE, most individuals will be within ±3σ of the mean
	estimatedMin := populationMean - 3*populationStd
	estimatedMax := populationMean + 3*populationStd

	return &ReferenceStats{
		Mean: populationMean,
		Std:  populationStd,
		Min:  estimatedMin,
		Max:  estimatedMax,
	}, nil
}
