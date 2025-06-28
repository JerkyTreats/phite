package prs

import (
	"math"
	"testing"

	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

// Test cases from the brief: 3 SNPs with known theoretical values
// p=(0.2,0.5,0.8), β=(0.1,-0.3,0.2)
// Expected Mean: 2×(0.2×0.1 + 0.5×(-0.3) + 0.8×0.2) = 0.06
// Expected Variance: 2×(0.2×0.8×0.01 + 0.5×0.5×0.09 + 0.8×0.2×0.04) = 0.0610

func TestStatisticalCorrectness_PopulationParameters(t *testing.T) {
	// Test case from brief - known Hardy-Weinberg population
	alleleFreqs := map[string]float64{
		"rs1": 0.2, // p1 = 0.2
		"rs2": 0.5, // p2 = 0.5
		"rs3": 0.8, // p3 = 0.8
	}

	effectSizes := map[string]float64{
		"rs1": 0.1,  // β1 = 0.1
		"rs2": -0.3, // β2 = -0.3
		"rs3": 0.2,  // β3 = 0.2
	}

	// Theoretical population parameters (CORRECT formulas)
	expectedMean := 2 * (0.2*0.1 + 0.5*(-0.3) + 0.8*0.2)                 // = 0.06
	expectedVariance := 2 * (0.2*0.8*0.01 + 0.5*0.5*0.09 + 0.8*0.2*0.04) // = 0.0610
	expectedStd := math.Sqrt(expectedVariance)

	// Current implementation (WRONG - uses sample statistics)
	stats, err := reference_stats.Compute(alleleFreqs, effectSizes)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// These assertions will FAIL with current implementation
	// because it uses incorrect sample statistics formulas
	const tolerance = 1e-12

	if math.Abs(stats.Mean-expectedMean) > tolerance {
		t.Errorf("Population Mean incorrect: got %v, expected %v (diff: %v)",
			stats.Mean, expectedMean, math.Abs(stats.Mean-expectedMean))
	}

	if math.Abs(stats.Std-expectedStd) > tolerance {
		t.Errorf("Population Std incorrect: got %v, expected %v (diff: %v)",
			stats.Std, expectedStd, math.Abs(stats.Std-expectedStd))
	}
}

func TestStatisticalCorrectness_MathematicalFlaws(t *testing.T) {
	// This test documents the specific mathematical flaws in current implementation
	alleleFreqs := map[string]float64{"rs1": 0.3}
	effectSizes := map[string]float64{"rs1": 0.5}

	// Current implementation computes WRONG sample statistics:
	// mean = sum/n = (2*0.3*0.5)/1 = 0.3
	// variance = E[expected²] - E[expected]² = (0.3²)/1 - 0.3² = 0

	// CORRECT population parameters should be:
	// mean = 2*p*β = 2*0.3*0.5 = 0.3 (happens to match)
	// variance = 2*p*(1-p)*β² = 2*0.3*0.7*0.25 = 0.105

	correctMean := 2 * 0.3 * 0.5            // = 0.3
	correctVariance := 2 * 0.3 * 0.7 * 0.25 // = 0.105
	correctStd := math.Sqrt(correctVariance)

	stats, err := reference_stats.Compute(alleleFreqs, effectSizes)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// Mean might accidentally be correct due to single variant
	if math.Abs(stats.Mean-correctMean) > 1e-12 {
		t.Errorf("Mean: got %v, expected %v", stats.Mean, correctMean)
	}

	// Variance will be WRONG - current implementation gives ~0, should be 0.105
	if math.Abs(stats.Std-correctStd) > 1e-12 {
		t.Errorf("MATHEMATICAL FLAW DETECTED: Std got %v, expected %v", stats.Std, correctStd)
		t.Errorf("Current implementation uses sample statistics (E[x²]-E[x]²)")
		t.Errorf("Should use population variance: 2*p*(1-p)*β²")
	}
}

func TestStatisticalCorrectness_ZeroVarianceFlaw(t *testing.T) {
	// Current implementation will give variance ≈ 0 for single SNP
	// This exposes the E[x²] - E[x]² flaw
	alleleFreqs := map[string]float64{"rs1": 0.4}
	effectSizes := map[string]float64{"rs1": 0.8}

	correctVariance := 2 * 0.4 * 0.6 * 0.64 // = 0.3072

	stats, err := reference_stats.Compute(alleleFreqs, effectSizes)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// Current implementation gives variance ≈ 0 for single variant
	// This is mathematically impossible for HWE population
	if stats.Std < 0.1 {
		t.Errorf("CRITICAL FLAW: Variance near zero (%v) for single SNP", stats.Std*stats.Std)
		t.Errorf("Population variance should be %v", correctVariance)
		t.Errorf("Current formula: E[expected²] - E[expected]² ≈ 0 for single variant")
		t.Errorf("Correct formula: 2*p*(1-p)*β² = %v", correctVariance)
	}
}

func TestStatisticalCorrectness_EdgeCases(t *testing.T) {
	testCases := []struct {
		name    string
		freq    float64
		effect  float64
		expMean float64
		expVar  float64
	}{
		{"Extreme frequency", 0.01, 0.5, 0.01, 2 * 0.01 * 0.99 * 0.25},
		{"Zero effect", 0.5, 0.0, 0.0, 0.0},
		{"High frequency", 0.95, 0.3, 0.57, 2 * 0.95 * 0.05 * 0.09},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alleleFreqs := map[string]float64{"rs1": tc.freq}
			effectSizes := map[string]float64{"rs1": tc.effect}

			stats, err := reference_stats.Compute(alleleFreqs, effectSizes)
			if err != nil {
				t.Fatalf("Compute failed: %v", err)
			}

			if math.Abs(stats.Mean-tc.expMean) > 1e-12 {
				t.Errorf("Mean: got %v, expected %v", stats.Mean, tc.expMean)
			}

			expStd := math.Sqrt(tc.expVar)
			if math.Abs(stats.Std-expStd) > 1e-12 {
				t.Errorf("Std: got %v, expected %v", stats.Std, expStd)
			}
		})
	}
}
