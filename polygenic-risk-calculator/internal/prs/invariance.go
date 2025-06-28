package prs

import (
	"fmt"

	"phite.io/polygenic-risk-calculator/internal/invariance"
	"phite.io/polygenic-risk-calculator/internal/model"
)

// ValidateInputSNPs validates all input SNPs for PRS calculation
// Returns error if validation fails or nil if validation passes/is disabled
func ValidateInputSNPs(snps []model.AnnotatedSNP) error {
	// Check if validation is enabled
	if !invariance.IsValidationEnabled() {
		return nil
	}

	// Validate each SNP individually
	for i, snp := range snps {
		context := fmt.Sprintf("input SNP %d/%d (%s)", i+1, len(snps), snp.RSID)

		// Validate dosage
		if err := invariance.AssertValidDosage(snp.Dosage, context); err != nil {
			return fmt.Errorf("invalid input SNP %s: %w", snp.RSID, err)
		}

		// Validate beta coefficient
		if err := invariance.AssertValidBetaCoefficient(snp.Beta, context); err != nil {
			return fmt.Errorf("invalid input SNP %s: %w", snp.RSID, err)
		}

		// Validate contribution numerical stability
		contribution := float64(snp.Dosage) * snp.Beta
		if err := invariance.AssertNumericalStability(contribution, context+" contribution"); err != nil {
			return fmt.Errorf("invalid input SNP %s: %w", snp.RSID, err)
		}
	}

	return nil
}

// ValidatePRSResult validates the calculated PRS result
// Returns error if validation fails or nil if validation passes/is disabled
func ValidatePRSResult(snps []model.AnnotatedSNP, result PRSResult) error {
	// Check if validation is enabled
	if !invariance.IsValidationEnabled() {
		return nil
	}

	context := fmt.Sprintf("PRS result (%d SNPs)", len(snps))

	// Validate final PRS score numerical stability
	if err := invariance.AssertNumericalStability(result.PRSScore, context); err != nil {
		return fmt.Errorf("invalid PRS result: %w", err)
	}

	// Validate PRS calculation accuracy - verify it equals sum of contributions
	var expectedScore float64
	for _, snp := range snps {
		expectedScore += float64(snp.Dosage) * snp.Beta
	}

	const tolerance = 1e-12
	if abs(result.PRSScore-expectedScore) > tolerance {
		return fmt.Errorf("PRS calculation error: got %v, expected %v", result.PRSScore, expectedScore)
	}

	return nil
}

// ValidateVariantContribution validates an individual variant's contribution during calculation
// Returns error if validation fails or nil if validation passes/is disabled
func ValidateVariantContribution(rsid string, dosage int, beta, contribution float64) error {
	// Check if validation is enabled
	if !invariance.IsValidationEnabled() {
		return nil
	}

	context := fmt.Sprintf("variant %s", rsid)

	// Validate dosage
	if err := invariance.AssertValidDosage(dosage, context); err != nil {
		return fmt.Errorf("invalid variant %s: %w", rsid, err)
	}

	// Validate beta coefficient
	if err := invariance.AssertValidBetaCoefficient(beta, context); err != nil {
		return fmt.Errorf("invalid variant %s: %w", rsid, err)
	}

	// Validate contribution calculation
	expectedContribution := float64(dosage) * beta
	const tolerance = 1e-12
	if abs(contribution-expectedContribution) > tolerance {
		return fmt.Errorf("variant %s contribution error: got %v, expected %v", rsid, contribution, expectedContribution)
	}

	// Validate numerical stability
	if err := invariance.AssertNumericalStability(contribution, context+" contribution"); err != nil {
		return fmt.Errorf("invalid variant %s: %w", rsid, err)
	}

	return nil
}

// ValidatePRSBounds validates that a PRS score falls within theoretical bounds
// Returns error if validation fails or nil if validation passes/is disabled
func ValidatePRSBounds(prsScore, minBound, maxBound float64) error {
	// Check if validation is enabled
	if !invariance.IsValidationEnabled() {
		return nil
	}

	context := "PRS bounds validation"

	// Validate the bounds themselves
	if err := invariance.AssertNumericalStability(minBound, context+" min bound"); err != nil {
		return fmt.Errorf("invalid bounds: %w", err)
	}
	if err := invariance.AssertNumericalStability(maxBound, context+" max bound"); err != nil {
		return fmt.Errorf("invalid bounds: %w", err)
	}

	// Validate bounds ordering
	if minBound > maxBound {
		return fmt.Errorf("invalid bounds: min %v > max %v", minBound, maxBound)
	}

	// Validate PRS score is within bounds
	if prsScore < minBound || prsScore > maxBound {
		return fmt.Errorf("PRS score %v outside bounds [%v, %v]", prsScore, minBound, maxBound)
	}

	return nil
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Legacy support - these types and functions are kept for backward compatibility with existing tests

// InvariantValidator provides PRS-specific invariance validation
type InvariantValidator struct {
	strictMode bool
}

// NewInvariantValidator creates a new PRS invariant validator
func NewInvariantValidator(strictMode bool) *InvariantValidator {
	return &InvariantValidator{
		strictMode: strictMode,
	}
}

// ValidateAnnotatedSNP validates a single annotated SNP for PRS calculation
func (v *InvariantValidator) ValidateAnnotatedSNP(snp model.AnnotatedSNP, context string) error {
	// Use the new self-contained validation
	return ValidateInputSNPs([]model.AnnotatedSNP{snp})
}

// ValidateVariantContribution validates individual variant contribution during PRS calculation
func (v *InvariantValidator) ValidateVariantContribution(rsid string, dosage int, beta, contribution float64, context string) error {
	// Use the new self-contained validation
	return ValidateVariantContribution(rsid, dosage, beta, contribution)
}

// ValidatePRSCalculation validates the overall PRS calculation
func (v *InvariantValidator) ValidatePRSCalculation(snps []model.AnnotatedSNP, prsScore float64, context string) error {
	// Create a minimal result for validation
	result := PRSResult{PRSScore: prsScore}
	return ValidatePRSResult(snps, result)
}

// ValidatePRSBounds validates that a PRS score falls within theoretical bounds
func (v *InvariantValidator) ValidatePRSBounds(prsScore, minBound, maxBound float64, context string) error {
	// Use the new self-contained validation
	return ValidatePRSBounds(prsScore, minBound, maxBound)
}

// ValidateNormalizationParameters validates parameters used for PRS normalization
func (v *InvariantValidator) ValidateNormalizationParameters(mean, std float64, context string) error {
	if !invariance.IsValidationEnabled() {
		return nil
	}

	// Validate population mean
	if err := invariance.AssertNumericalStability(mean, context+" population mean"); err != nil {
		return err
	}

	// Validate standard deviation (must be positive and finite)
	if err := invariance.AssertNumericalStability(std, context+" population std"); err != nil {
		return err
	}

	if std <= 0 {
		return fmt.Errorf("standard deviation must be positive: %v", std)
	}

	return nil
}

// ValidatePopulationModel validates population-level statistical parameters
func (v *InvariantValidator) ValidatePopulationModel(frequencies, effects []float64, mean, variance float64, context string) error {
	if !invariance.IsValidationEnabled() {
		return nil
	}

	// Use core invariance validation for population parameter consistency
	return invariance.AssertPopulationParameterConsistency(frequencies, effects, mean, variance, context)
}
