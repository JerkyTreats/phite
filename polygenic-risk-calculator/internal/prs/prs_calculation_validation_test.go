package prs

import (
	"fmt"
	"math"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/model"
)

// Individual PRS Calculation Validation Tests
// These tests verify that individual PRS calculations are mathematically correct

func TestPRSCalculation_IndividualAccuracy(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Test individual with known genotypes
	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Genotype: "AA", RiskAllele: "A", Beta: 0.1, Dosage: 2, Trait: "trait1"},
		{RSID: "rs2", Genotype: "AG", RiskAllele: "G", Beta: -0.3, Dosage: 1, Trait: "trait2"},
		{RSID: "rs3", Genotype: "CC", RiskAllele: "T", Beta: 0.2, Dosage: 0, Trait: "trait3"},
	}

	// Expected PRS: 2*0.1 + 1*(-0.3) + 0*0.2 = 0.2 - 0.3 + 0 = -0.1
	expectedPRS := 2*0.1 + 1*(-0.3) + 0*0.2

	result, err := CalculatePRS(snps)
	if err != nil {
		t.Fatalf("CalculatePRS returned error: %v", err)
	}

	const tolerance = 1e-12
	if math.Abs(result.PRSScore-expectedPRS) > tolerance {
		t.Errorf("Individual PRS incorrect: got %v, expected %v", result.PRSScore, expectedPRS)
	}

	// Verify individual contributions
	expectedContributions := []float64{0.2, -0.3, 0.0}
	if len(result.Details) != len(expectedContributions) {
		t.Fatalf("Wrong number of contributions: got %d, expected %d", len(result.Details), len(expectedContributions))
	}

	for i, detail := range result.Details {
		if math.Abs(detail.Contribution-expectedContributions[i]) > tolerance {
			t.Errorf("Contribution %d incorrect: got %v, expected %v", i, detail.Contribution, expectedContributions[i])
		}
	}
}

func TestPRSCalculation_DosageValidation(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Test all possible dosages (0, 1, 2)
	testCases := []struct {
		name     string
		dosage   int
		beta     float64
		expected float64
	}{
		{"Homozygous reference", 0, 0.5, 0.0},
		{"Heterozygous", 1, 0.5, 0.5},
		{"Homozygous alternative", 2, 0.5, 1.0},
		{"Negative effect", 1, -0.3, -0.3},
		{"Zero effect", 2, 0.0, 0.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			snps := []model.AnnotatedSNP{
				{RSID: "rs1", Beta: tc.beta, Dosage: tc.dosage, Trait: "test"},
			}

			result, err := CalculatePRS(snps)
			if err != nil {
				t.Fatalf("CalculatePRS returned error: %v", err)
			}

			const tolerance = 1e-12
			if math.Abs(result.PRSScore-tc.expected) > tolerance {
				t.Errorf("PRS incorrect: got %v, expected %v", result.PRSScore, tc.expected)
			}
		})
	}
}

func TestPRSCalculation_AdditiveModel(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Test that PRS follows additive model: PRS_i = Σ_j β_j · G_{ij}
	snps := []model.AnnotatedSNP{
		{RSID: "rs1", Beta: 0.2, Dosage: 1, Trait: "trait1"},
		{RSID: "rs2", Beta: -0.1, Dosage: 2, Trait: "trait2"},
		{RSID: "rs3", Beta: 0.3, Dosage: 0, Trait: "trait3"},
		{RSID: "rs4", Beta: 0.15, Dosage: 1, Trait: "trait4"},
		{RSID: "rs5", Beta: -0.25, Dosage: 2, Trait: "trait5"},
	}

	// Manual calculation
	expectedPRS := 1*0.2 + 2*(-0.1) + 0*0.3 + 1*0.15 + 2*(-0.25)
	// = 0.2 - 0.2 + 0 + 0.15 - 0.5 = -0.35

	result, err := CalculatePRS(snps)
	if err != nil {
		t.Fatalf("CalculatePRS returned error: %v", err)
	}

	const tolerance = 1e-12
	if math.Abs(result.PRSScore-expectedPRS) > tolerance {
		t.Errorf("Additive model violation: got %v, expected %v", result.PRSScore, expectedPRS)
	}

	// Verify sum of individual contributions equals total
	var sumContributions float64
	for _, detail := range result.Details {
		sumContributions += detail.Contribution
	}

	if math.Abs(sumContributions-result.PRSScore) > tolerance {
		t.Errorf("Contribution sum mismatch: sum=%v, total=%v", sumContributions, result.PRSScore)
	}
}

func TestPRSCalculation_LargeScale(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Test with many SNPs to verify numerical stability
	const numSNPs = 1000
	snps := make([]model.AnnotatedSNP, numSNPs)
	var expectedPRS float64

	for i := 0; i < numSNPs; i++ {
		dosage := i % 3                 // Cycle through 0, 1, 2
		beta := float64(i%100) / 1000.0 // Small effects: 0.000 to 0.099
		snps[i] = model.AnnotatedSNP{
			RSID:   fmt.Sprintf("rs%d", i),
			Beta:   beta,
			Dosage: dosage,
			Trait:  "test",
		}
		expectedPRS += float64(dosage) * beta
	}

	result, err := CalculatePRS(snps)
	if err != nil {
		t.Fatalf("CalculatePRS returned error: %v", err)
	}

	const tolerance = 1e-10 // Slightly relaxed for numerical accumulation
	if math.Abs(result.PRSScore-expectedPRS) > tolerance {
		t.Errorf("Large scale PRS incorrect: got %v, expected %v", result.PRSScore, expectedPRS)
	}

	if len(result.Details) != numSNPs {
		t.Errorf("Wrong number of details: got %d, expected %d", len(result.Details), numSNPs)
	}
}

func TestPRSCalculation_ExtremeCases(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Disable strict mode for extreme numerical testing
	setupTestConfig(true, false) // Enable validation but disable strict mode

	testCases := []struct {
		name     string
		beta     float64
		dosage   int
		expected float64
	}{
		{"Very small effect", 1e-10, 2, 2e-10},
		{"Very large effect", 1e10, 1, 1e10},
		{"Negative large effect", -1e5, 2, -2e5},
		{"Zero effect", 0.0, 2, 0.0},
		{"Zero dosage", 1.0, 0, 0.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			snps := []model.AnnotatedSNP{
				{RSID: "rs1", Beta: tc.beta, Dosage: tc.dosage, Trait: "test"},
			}

			result, err := CalculatePRS(snps)
			if err != nil {
				t.Fatalf("CalculatePRS returned error: %v", err)
			}

			// Use relative tolerance for very large numbers
			tolerance := math.Max(1e-12, math.Abs(tc.expected)*1e-12)
			if math.Abs(result.PRSScore-tc.expected) > tolerance {
				t.Errorf("Extreme case PRS incorrect: got %v, expected %v", result.PRSScore, tc.expected)
			}

			// Check for numerical issues
			if math.IsNaN(result.PRSScore) || math.IsInf(result.PRSScore, 0) {
				t.Errorf("PRS is NaN or Inf: %v", result.PRSScore)
			}
		})
	}
}

func TestPRSCalculation_PrecisionBoundaries(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Test precision boundaries with many small contributions
	const numSNPs = 10000
	snps := make([]model.AnnotatedSNP, numSNPs)

	// Each SNP contributes a very small amount
	smallEffect := 1e-8
	for i := 0; i < numSNPs; i++ {
		snps[i] = model.AnnotatedSNP{
			RSID:   fmt.Sprintf("rs%d", i),
			Beta:   smallEffect,
			Dosage: 1,
			Trait:  "test",
		}
	}

	expectedPRS := float64(numSNPs) * smallEffect

	result, err := CalculatePRS(snps)
	if err != nil {
		t.Fatalf("CalculatePRS returned error: %v", err)
	}

	// Use relative tolerance for accumulated small effects
	tolerance := math.Max(1e-10, math.Abs(expectedPRS)*1e-10)
	if math.Abs(result.PRSScore-expectedPRS) > tolerance {
		t.Errorf("Precision boundary PRS incorrect: got %v, expected %v", result.PRSScore, expectedPRS)
	}

	// Verify no precision loss in accumulation
	if math.IsNaN(result.PRSScore) || math.IsInf(result.PRSScore, 0) {
		t.Errorf("Precision loss: PRS is NaN or Inf: %v", result.PRSScore)
	}
}
