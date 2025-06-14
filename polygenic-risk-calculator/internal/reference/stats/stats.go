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

// Compute calculates PRS statistics from allele frequencies and effect sizes.
// This is used for on-the-fly computation when cache misses occur.
func Compute(alleleFreqs map[string]float64, effectSizes map[string]float64) (*ReferenceStats, error) {
	if len(alleleFreqs) == 0 || len(effectSizes) == 0 {
		return nil, fmt.Errorf("empty allele frequencies or effect sizes")
	}

	var sum float64
	var sumSq float64
	var min float64 = math.MaxFloat64
	var max float64 = -math.MaxFloat64
	var validVariants int

	// Calculate mean and standard deviation
	for variant, freq := range alleleFreqs {
		effect, ok := effectSizes[variant]
		if !ok {
			continue // Skip variants without effect sizes
		}

		validVariants++

		// Expected value for this variant
		expected := 2 * freq * effect
		sum += expected
		sumSq += expected * expected

		// Update min/max
		if expected < min {
			min = expected
		}
		if expected > max {
			max = expected
		}
	}

	if validVariants == 0 {
		return nil, fmt.Errorf("no matching variants found between allele frequencies and effect sizes")
	}

	n := float64(validVariants)
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	std := math.Sqrt(variance)

	return &ReferenceStats{
		Mean: mean,
		Std:  std,
		Min:  min,
		Max:  max,
	}, nil
}
