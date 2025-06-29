package prs

import (
	"fmt"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/invariance"
	"phite.io/polygenic-risk-calculator/internal/model"
)

// Test InvariantValidator (PRS-specific)
func TestInvariantValidator(t *testing.T) {
	// Setup configuration for non-strict mode
	config.ResetForTest()
	config.Set(invariance.EnableValidationKey, true)
	config.Set(invariance.StrictModeKey, false)

	validator := NewInvariantValidator(false) // Non-strict mode

	// Setup configuration for strict mode
	config.Set(invariance.StrictModeKey, true)
	strictValidator := NewInvariantValidator(true) // Strict mode

	// Test ValidateAnnotatedSNP
	validSNP := model.AnnotatedSNP{
		RSID:       "rs123",
		Genotype:   "AG",
		RiskAllele: "A",
		Beta:       0.3,
		Dosage:     1,
		Trait:      "height",
	}

	// Reset to non-strict for valid SNP test
	config.Set(invariance.StrictModeKey, false)
	err := validator.ValidateAnnotatedSNP(validSNP, "unit test")
	if err != nil {
		t.Errorf("Valid SNP should pass validation: %v", err)
	}

	// Test with invalid dosage
	invalidSNP := validSNP
	invalidSNP.Dosage = 3 // Invalid dosage
	err = validator.ValidateAnnotatedSNP(invalidSNP, "unit test")
	if err == nil {
		t.Errorf("Invalid SNP should fail validation")
	}

	// Test strict mode with very large beta coefficient
	largeBetaSNP := validSNP
	largeBetaSNP.Beta = 100.0 // Very large beta

	// Non-strict mode should allow large beta
	config.Set(invariance.StrictModeKey, false)
	err = validator.ValidateAnnotatedSNP(largeBetaSNP, "unit test")
	if err != nil {
		t.Errorf("Non-strict mode should allow large beta")
	}

	// Strict mode should reject very large beta
	config.Set(invariance.StrictModeKey, true)
	err = strictValidator.ValidateAnnotatedSNP(largeBetaSNP, "unit test")
	if err == nil {
		t.Errorf("Strict mode should reject very large beta")
	}
}

// Test ValidatePRSCalculation
func TestValidatePRSCalculation(t *testing.T) {
	validator := NewInvariantValidator(false)

	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Beta: 0.1, Dosage: 2, Trait: "test"},
		{RSID: "rs2", Beta: -0.3, Dosage: 1, Trait: "test"},
		{RSID: "rs3", Beta: 0.2, Dosage: 0, Trait: "test"},
	}

	// Correct PRS calculation
	expectedPRS := 2*0.1 + 1*(-0.3) + 0*0.2 // = -0.1
	err := validator.ValidatePRSCalculation(snps, expectedPRS, "unit test")
	if err != nil {
		t.Errorf("Valid PRS calculation should pass: %v", err)
	}

	// Incorrect PRS calculation
	incorrectPRS := 0.5
	err = validator.ValidatePRSCalculation(snps, incorrectPRS, "unit test")
	if err == nil {
		t.Errorf("Incorrect PRS calculation should fail")
	}
}

// Test ValidateNormalizationParameters
func TestValidateNormalizationParameters(t *testing.T) {
	validator := NewInvariantValidator(false)

	// Test valid normalization parameters
	err := validator.ValidateNormalizationParameters(0.5, 0.2, "unit test")
	if err != nil {
		t.Errorf("Valid normalization parameters should pass: %v", err)
	}

	// Test with negative standard deviation
	err = validator.ValidateNormalizationParameters(0.5, -0.2, "unit test")
	if err == nil {
		t.Errorf("Negative standard deviation should fail")
	}

	// Test with zero standard deviation
	err = validator.ValidateNormalizationParameters(0.5, 0.0, "unit test")
	if err == nil {
		t.Errorf("Zero standard deviation should fail")
	}
}

// Test ValidatePopulationModel
func TestValidatePopulationModel(t *testing.T) {
	validator := NewInvariantValidator(false)

	// Test case from the brief: p=(0.2,0.5,0.8), β=(0.1,-0.3,0.2)
	frequencies := []float64{0.2, 0.5, 0.8}
	effects := []float64{0.1, -0.3, 0.2}

	expectedMean := 2 * (0.2*0.1 + 0.5*(-0.3) + 0.8*0.2)                 // = 0.06
	expectedVariance := 2 * (0.2*0.8*0.01 + 0.5*0.5*0.09 + 0.8*0.2*0.04) // = 0.0610

	err := validator.ValidatePopulationModel(frequencies, effects, expectedMean, expectedVariance, "unit test")
	if err != nil {
		t.Errorf("Valid population model should pass: %v", err)
	}

	// Test with incorrect mean
	err = validator.ValidatePopulationModel(frequencies, effects, 0.03, expectedVariance, "unit test")
	if err == nil {
		t.Errorf("Incorrect population mean should fail")
	}

	// Test with mismatched array lengths
	wrongFreqs := []float64{0.2, 0.5} // One less than effects
	err = validator.ValidatePopulationModel(wrongFreqs, effects, expectedMean, expectedVariance, "unit test")
	if err == nil {
		t.Errorf("Mismatched array lengths should fail")
	}
}

// Test ValidateVariantContribution
func TestValidateVariantContribution(t *testing.T) {
	validator := NewInvariantValidator(false)

	// Valid contribution
	err := validator.ValidateVariantContribution("rs123", 1, 0.3, 0.3, "unit test")
	if err != nil {
		t.Errorf("Valid variant contribution should pass: %v", err)
	}

	// Invalid contribution calculation
	err = validator.ValidateVariantContribution("rs123", 1, 0.3, 0.5, "unit test")
	if err == nil {
		t.Errorf("Incorrect contribution calculation should fail")
	}

	// Invalid dosage
	err = validator.ValidateVariantContribution("rs123", 3, 0.3, 0.9, "unit test")
	if err == nil {
		t.Errorf("Invalid dosage should fail")
	}
}

// Test ValidatePRSBounds
func TestValidatePRSBounds(t *testing.T) {
	validator := NewInvariantValidator(false)

	// Valid PRS within bounds
	err := validator.ValidatePRSBounds(1.5, -2.0, 3.0, "unit test")
	if err != nil {
		t.Errorf("Valid PRS bounds should pass: %v", err)
	}

	// PRS outside bounds (too high)
	err = validator.ValidatePRSBounds(4.0, -2.0, 3.0, "unit test")
	if err == nil {
		t.Errorf("PRS above upper bound should fail")
	}

	// PRS outside bounds (too low)
	err = validator.ValidatePRSBounds(-3.0, -2.0, 3.0, "unit test")
	if err == nil {
		t.Errorf("PRS below lower bound should fail")
	}

	// Invalid bounds (min > max)
	err = validator.ValidatePRSBounds(1.0, 3.0, -2.0, "unit test")
	if err == nil {
		t.Errorf("Invalid bounds ordering should fail")
	}
}

// Benchmark InvariantValidator performance
func BenchmarkInvariantValidator(b *testing.B) {
	validator := NewInvariantValidator(false)

	snp := model.AnnotatedSNP{
		RSID:       "rs123",
		Genotype:   "AG",
		RiskAllele: "A",
		Beta:       0.3,
		Dosage:     1,
		Trait:      "height",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateAnnotatedSNP(snp, "benchmark")
	}
}

// Benchmark large-scale PRS validation
func BenchmarkLargeScalePRSValidation(b *testing.B) {
	validator := NewInvariantValidator(false)

	// Create 1000 SNPs for large-scale test
	snps := make([]model.AnnotatedSNP, 1000)
	var totalPRS float64
	for i := 0; i < 1000; i++ {
		beta := 0.1 * float64(i%10) // Vary beta values
		dosage := i % 3             // Vary dosage 0, 1, 2
		snps[i] = model.AnnotatedSNP{
			RSID:   fmt.Sprintf("rs%d", i),
			Beta:   beta,
			Dosage: dosage,
			Trait:  "test",
		}
		totalPRS += float64(dosage) * beta
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidatePRSCalculation(snps, totalPRS, "benchmark")
	}
}
