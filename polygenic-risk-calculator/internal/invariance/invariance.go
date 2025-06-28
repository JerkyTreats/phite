package invariance

import (
	"fmt"
	"math"

	"phite.io/polygenic-risk-calculator/internal/config"
)

// Configuration keys for invariance validation
const (
	// EnableInvarianceValidationKey controls whether to enable runtime invariance validation
	EnableInvarianceValidationKey = "invariance.enable_validation"
	// StrictValidationModeKey enables strict validation mode with additional checks
	StrictValidationModeKey = "invariance.strict_mode"
)

// init registers required configuration keys
func init() {
	// Set reasonable defaults for invariance validation
	// Production systems should override these in their config files
	if !config.HasKey(EnableInvarianceValidationKey) {
		// Default to enabled for safety in development, production configs should set explicitly
		config.RegisterRequiredKey(EnableInvarianceValidationKey)
	}
	if !config.HasKey(StrictValidationModeKey) {
		// Default to disabled for performance
		// Production configs can enable for critical validation scenarios
	}
}

// IsValidationEnabled returns whether invariance validation is currently enabled
func IsValidationEnabled() bool {
	return config.GetBool(EnableInvarianceValidationKey)
}

// IsStrictModeEnabled returns whether strict validation mode is currently enabled
func IsStrictModeEnabled() bool {
	return config.GetBool(StrictValidationModeKey)
}

// InvariantViolationError represents a violation of a mathematical or statistical invariant
type InvariantViolationError struct {
	Type    string
	Message string
	Context string
	Value   interface{}
}

func (e *InvariantViolationError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("invariant violation [%s] in %s: %s (value: %v)",
			e.Type, e.Context, e.Message, e.Value)
	}
	return fmt.Sprintf("invariant violation [%s]: %s (value: %v)",
		e.Type, e.Message, e.Value)
}

// Core mathematical invariance assertions

// AssertValidProbability ensures a value is a valid probability [0,1]
func AssertValidProbability(p float64, context string) error {
	if math.IsNaN(p) || math.IsInf(p, 0) {
		return &InvariantViolationError{
			Type:    "probability",
			Message: "probability must be finite",
			Context: context,
			Value:   p,
		}
	}
	if p < 0 || p > 1 {
		return &InvariantViolationError{
			Type:    "probability",
			Message: "probability must be in range [0,1]",
			Context: context,
			Value:   p,
		}
	}
	return nil
}

// AssertValidVariance ensures a variance is non-negative and finite
func AssertValidVariance(variance float64, context string) error {
	if math.IsNaN(variance) || math.IsInf(variance, 0) {
		return &InvariantViolationError{
			Type:    "variance",
			Message: "variance must be finite",
			Context: context,
			Value:   variance,
		}
	}
	if variance < 0 {
		return &InvariantViolationError{
			Type:    "variance",
			Message: "variance must be non-negative",
			Context: context,
			Value:   variance,
		}
	}
	return nil
}

// AssertValidDosage ensures a genotype dosage is valid [0,2] for diploid organisms
func AssertValidDosage(dosage int, context string) error {
	if dosage < 0 || dosage > 2 {
		return &InvariantViolationError{
			Type:    "dosage",
			Message: "dosage must be in range [0,2] for diploid genotypes",
			Context: context,
			Value:   dosage,
		}
	}
	return nil
}

// AssertValidBetaCoefficient ensures a beta coefficient is reasonable
func AssertValidBetaCoefficient(beta float64, context string) error {
	if math.IsNaN(beta) || math.IsInf(beta, 0) {
		return &InvariantViolationError{
			Type:    "beta_coefficient",
			Message: "beta coefficient must be finite",
			Context: context,
			Value:   beta,
		}
	}

	// Strict mode: additional checks for reasonable beta values
	if IsStrictModeEnabled() {
		if math.Abs(beta) > 10.0 {
			return &InvariantViolationError{
				Type:    "beta_coefficient",
				Message: "beta coefficient unusually large (>10) in strict mode",
				Context: context,
				Value:   beta,
			}
		}
	}

	return nil
}

// AssertHardyWeinbergVariance validates variance under Hardy-Weinberg equilibrium
func AssertHardyWeinbergVariance(freq, beta, variance float64, context string) error {
	if err := AssertValidProbability(freq, context+" frequency"); err != nil {
		return err
	}
	if err := AssertValidBetaCoefficient(beta, context+" beta"); err != nil {
		return err
	}

	expectedVariance := 2 * freq * (1 - freq) * beta * beta
	if err := AssertValidVariance(expectedVariance, context+" expected variance"); err != nil {
		return err
	}

	const tolerance = 1e-12
	if math.Abs(variance-expectedVariance) > tolerance {
		return &InvariantViolationError{
			Type:    "hardy_weinberg_variance",
			Message: fmt.Sprintf("variance mismatch: got %v, expected %v (HWE)", variance, expectedVariance),
			Context: context,
			Value:   variance,
		}
	}

	return nil
}

// AssertPopulationParameterConsistency validates population mean and variance
func AssertPopulationParameterConsistency(frequencies, effects []float64, mean, variance float64, context string) error {
	if len(frequencies) != len(effects) {
		return &InvariantViolationError{
			Type:    "parameter_consistency",
			Message: fmt.Sprintf("frequencies and effects length mismatch: %d vs %d", len(frequencies), len(effects)),
			Context: context,
			Value:   fmt.Sprintf("freq_len=%d, effect_len=%d", len(frequencies), len(effects)),
		}
	}

	var expectedMean, expectedVariance float64
	for i := range frequencies {
		freq, effect := frequencies[i], effects[i]

		if err := AssertValidProbability(freq, fmt.Sprintf("%s variant %d", context, i)); err != nil {
			return err
		}
		if err := AssertValidBetaCoefficient(effect, fmt.Sprintf("%s variant %d", context, i)); err != nil {
			return err
		}

		expectedMean += 2 * freq * effect
		expectedVariance += 2 * freq * (1 - freq) * effect * effect
	}

	const tolerance = 1e-12
	if math.Abs(mean-expectedMean) > tolerance {
		return &InvariantViolationError{
			Type:    "population_mean",
			Message: fmt.Sprintf("population mean mismatch: got %v, expected %v", mean, expectedMean),
			Context: context,
			Value:   mean,
		}
	}

	if math.Abs(variance-expectedVariance) > tolerance {
		return &InvariantViolationError{
			Type:    "population_variance",
			Message: fmt.Sprintf("population variance mismatch: got %v, expected %v", variance, expectedVariance),
			Context: context,
			Value:   variance,
		}
	}

	return nil
}

// AssertMonotonicity ensures a sequence is monotonically increasing or decreasing
func AssertMonotonicity(values []float64, increasing bool, context string) error {
	for i := 1; i < len(values); i++ {
		if increasing && values[i] < values[i-1] {
			return &InvariantViolationError{
				Type:    "monotonicity",
				Message: fmt.Sprintf("sequence not monotonically increasing at index %d: %v >= %v", i, values[i-1], values[i]),
				Context: context,
				Value:   fmt.Sprintf("[%d]=%v, [%d]=%v", i-1, values[i-1], i, values[i]),
			}
		}
		if !increasing && values[i] > values[i-1] {
			return &InvariantViolationError{
				Type:    "monotonicity",
				Message: fmt.Sprintf("sequence not monotonically decreasing at index %d: %v <= %v", i, values[i-1], values[i]),
				Context: context,
				Value:   fmt.Sprintf("[%d]=%v, [%d]=%v", i-1, values[i-1], i, values[i]),
			}
		}
	}
	return nil
}

// AssertNumericalStability validates numerical computation stability
func AssertNumericalStability(value float64, context string) error {
	if math.IsNaN(value) {
		return &InvariantViolationError{
			Type:    "numerical_stability",
			Message: "computation resulted in NaN",
			Context: context,
			Value:   value,
		}
	}
	if math.IsInf(value, 0) {
		return &InvariantViolationError{
			Type:    "numerical_stability",
			Message: "computation resulted in infinity",
			Context: context,
			Value:   value,
		}
	}

	// Strict mode: additional numerical stability checks
	if IsStrictModeEnabled() {
		if math.Abs(value) > 1e12 {
			return &InvariantViolationError{
				Type:    "numerical_stability",
				Message: "computation resulted in extremely large value (>1e12) in strict mode",
				Context: context,
				Value:   value,
			}
		}
		// Check for underflow (extremely small non-zero values)
		if value != 0 && math.Abs(value) < 1e-300 {
			return &InvariantViolationError{
				Type:    "numerical_stability",
				Message: "computation resulted in extremely small value (<1e-300) in strict mode",
				Context: context,
				Value:   value,
			}
		}
	}

	return nil
}
