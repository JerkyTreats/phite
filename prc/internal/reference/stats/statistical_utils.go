package reference_stats

import (
	"fmt"
	"math"
)

// HardyWeinbergValidator provides utilities for validating Hardy-Weinberg equilibrium assumptions
type HardyWeinbergValidator struct{}

// ValidateAlleleFrequency checks if an allele frequency is within valid bounds [0, 1]
func (hwv *HardyWeinbergValidator) ValidateAlleleFrequency(freq float64, variant string) error {
	if math.IsNaN(freq) || math.IsInf(freq, 0) {
		return fmt.Errorf("allele frequency for %s is NaN or Inf: %v", variant, freq)
	}
	if freq < 0 || freq > 1 {
		return fmt.Errorf("allele frequency for %s out of bounds [0,1]: %v", variant, freq)
	}
	return nil
}

// ValidateEffectSize checks if an effect size is numerically valid
func (hwv *HardyWeinbergValidator) ValidateEffectSize(effect float64, variant string) error {
	if math.IsNaN(effect) || math.IsInf(effect, 0) {
		return fmt.Errorf("effect size for %s is NaN or Inf: %v", variant, effect)
	}
	// Note: effect sizes can be any real number (positive, negative, or zero)
	return nil
}

// CalculateExpectedVariance computes the theoretical Hardy-Weinberg variance for a single SNP
func (hwv *HardyWeinbergValidator) CalculateExpectedVariance(freq, effect float64) float64 {
	// Hardy-Weinberg variance: 2*p*(1-p)*β²
	return 2 * freq * (1 - freq) * effect * effect
}

// CalculateExpectedMean computes the theoretical Hardy-Weinberg mean for a single SNP
func (hwv *HardyWeinbergValidator) CalculateExpectedMean(freq, effect float64) float64 {
	// Hardy-Weinberg mean: 2*p*β
	return 2 * freq * effect
}

// NumericalStabilityChecker provides utilities for ensuring numerical stability
type NumericalStabilityChecker struct{}

// CheckPopulationParameters validates that computed population parameters are numerically stable
func (nsc *NumericalStabilityChecker) CheckPopulationParameters(mean, variance float64, numVariants int) error {
	// Check for NaN or Inf
	if math.IsNaN(mean) || math.IsInf(mean, 0) {
		return fmt.Errorf("population mean is NaN or Inf: %v", mean)
	}
	if math.IsNaN(variance) || math.IsInf(variance, 0) {
		return fmt.Errorf("population variance is NaN or Inf: %v", variance)
	}

	// Variance must be non-negative
	if variance < 0 {
		return fmt.Errorf("negative population variance: %v", variance)
	}

	// Check for underflow/overflow with very small/large numbers
	if math.Abs(mean) > 1e100 {
		return fmt.Errorf("population mean too large (possible overflow): %v", mean)
	}
	if variance > 1e200 {
		return fmt.Errorf("population variance too large (possible overflow): %v", variance)
	}

	// For numerical stability, very small variances should be handled carefully
	if variance > 0 && variance < 1e-300 {
		return fmt.Errorf("population variance too small (possible underflow): %v", variance)
	}

	return nil
}

// CheckAccumulation validates numerical accuracy during sum accumulation
func (nsc *NumericalStabilityChecker) CheckAccumulation(sum, newValue float64, iteration int) error {
	newSum := sum + newValue

	// Check for overflow
	if math.IsInf(newSum, 0) && !math.IsInf(sum, 0) && !math.IsInf(newValue, 0) {
		return fmt.Errorf("numerical overflow during accumulation at iteration %d: sum=%v, adding=%v", iteration, sum, newValue)
	}

	// Check for loss of precision with very different magnitudes
	if math.Abs(sum) > 0 && math.Abs(newValue) > 0 {
		ratio := math.Abs(newValue) / math.Abs(sum)
		if ratio < 1e-15 {
			// This is just a warning, not an error - very small contributions might be expected
			// In a production system, you might want to log this
		}
	}

	return nil
}

// PopulationParameterCalculator encapsulates the correct population parameter calculations
type PopulationParameterCalculator struct {
	validator        *HardyWeinbergValidator
	stabilityChecker *NumericalStabilityChecker
}

// NewPopulationParameterCalculator creates a new calculator with validation
func NewPopulationParameterCalculator() *PopulationParameterCalculator {
	return &PopulationParameterCalculator{
		validator:        &HardyWeinbergValidator{},
		stabilityChecker: &NumericalStabilityChecker{},
	}
}

// ComputeValidated calculates population parameters with comprehensive validation
func (ppc *PopulationParameterCalculator) ComputeValidated(alleleFreqs map[string]float64, effectSizes map[string]float64) (*ReferenceStats, error) {
	if len(alleleFreqs) == 0 || len(effectSizes) == 0 {
		return nil, fmt.Errorf("empty input: alleleFreqs=%d, effectSizes=%d", len(alleleFreqs), len(effectSizes))
	}

	var populationMean float64
	var populationVariance float64
	var validVariants int

	// Validate all inputs first
	for variant, freq := range alleleFreqs {
		if err := ppc.validator.ValidateAlleleFrequency(freq, variant); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		if effect, ok := effectSizes[variant]; ok {
			if err := ppc.validator.ValidateEffectSize(effect, variant); err != nil {
				return nil, fmt.Errorf("validation failed: %w", err)
			}
		}
	}

	// Calculate population parameters with numerical stability checks
	iteration := 0
	for variant, freq := range alleleFreqs {
		effect, ok := effectSizes[variant]
		if !ok {
			continue
		}

		iteration++
		validVariants++

		// Calculate contributions
		meanContribution := 2 * freq * effect
		varianceContribution := 2 * freq * (1 - freq) * effect * effect

		// Check numerical stability during accumulation
		if err := ppc.stabilityChecker.CheckAccumulation(populationMean, meanContribution, iteration); err != nil {
			return nil, fmt.Errorf("numerical instability in mean calculation: %w", err)
		}
		if err := ppc.stabilityChecker.CheckAccumulation(populationVariance, varianceContribution, iteration); err != nil {
			return nil, fmt.Errorf("numerical instability in variance calculation: %w", err)
		}

		// Accumulate
		populationMean += meanContribution
		populationVariance += varianceContribution
	}

	if validVariants == 0 {
		return nil, fmt.Errorf("no valid variants found")
	}

	// Final validation of computed parameters
	if err := ppc.stabilityChecker.CheckPopulationParameters(populationMean, populationVariance, validVariants); err != nil {
		return nil, fmt.Errorf("final validation failed: %w", err)
	}

	populationStd := math.Sqrt(populationVariance)

	// Estimate population bounds (±3σ covers ~99.7% of normal distribution)
	estimatedMin := populationMean - 3*populationStd
	estimatedMax := populationMean + 3*populationStd

	return &ReferenceStats{
		Mean: populationMean,
		Std:  populationStd,
		Min:  estimatedMin,
		Max:  estimatedMax,
	}, nil
}
