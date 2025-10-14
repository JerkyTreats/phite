package prs

import (
	"math"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/logging"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

// Test #1: Closed-form population mean & variance
// Input: 3-SNP: p=(0.2,0.5,0.8), β=(0.1,-0.3,0.2)
// Expected Output: Mean=0.06±1e-12, Var=0.0610±1e-12

func TestMeanVarianceTheory_KnownHWEPopulation(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Enable validation for mathematical correctness testing
	config.Set("invariance.enable_validation", true)
	defer config.Set("invariance.enable_validation", false)

	// Known Hardy-Weinberg Equilibrium test case from the brief
	alleleFreqs := map[string]float64{
		"rs1": 0.2, // Low frequency
		"rs2": 0.5, // Medium frequency
		"rs3": 0.8, // High frequency
	}

	effectSizes := map[string]float64{
		"rs1": 0.1,  // Positive effect
		"rs2": -0.3, // Large negative effect
		"rs3": 0.2,  // Positive effect
	}

	// Hand-calculated theoretical values (absolute truth)
	// Population Mean: μ_pop = Σ_j(2*p_j*β_j)
	//   = 2*(0.2*0.1 + 0.5*(-0.3) + 0.8*0.2)
	//   = 2*(0.02 - 0.15 + 0.16)
	//   = 2*(0.03) = 0.06
	expectedMean := 2 * (0.2*0.1 + 0.5*(-0.3) + 0.8*0.2)

	// Population Variance: Var(PRS) = Σ_j(2*p_j*(1-p_j)*β_j²)
	//   = 2*(0.2*0.8*0.01 + 0.5*0.5*0.09 + 0.8*0.2*0.04)
	//   = 2*(0.0016 + 0.0225 + 0.0064)
	//   = 2*(0.0305) = 0.061
	expectedVariance := 2 * (0.2*0.8*0.01 + 0.5*0.5*0.09 + 0.8*0.2*0.04)
	expectedStd := math.Sqrt(expectedVariance)

	t.Logf("Theoretical expectations: Mean=%.12f, Variance=%.12f, Std=%.12f",
		expectedMean, expectedVariance, expectedStd)

	// Test computation using corrected formula
	refStats, err := reference_stats.Compute(alleleFreqs, effectSizes)
	if err != nil {
		t.Fatalf("Reference stats computation failed: %v", err)
	}

	// Validate against theoretical values with high precision
	const tolerance = 1e-12

	// Test population mean
	if math.Abs(refStats.Mean-expectedMean) > tolerance {
		t.Errorf("Population mean incorrect:\n"+
			"  Got:      %.15f\n"+
			"  Expected: %.15f\n"+
			"  Diff:     %e\n"+
			"  Theory:   μ_pop = Σ_j(2*p_j*β_j)",
			refStats.Mean, expectedMean, refStats.Mean-expectedMean)
	}

	// Test population variance (through standard deviation)
	computedVariance := refStats.Std * refStats.Std
	if math.Abs(computedVariance-expectedVariance) > tolerance {
		t.Errorf("Population variance incorrect:\n"+
			"  Got:      %.15f\n"+
			"  Expected: %.15f\n"+
			"  Diff:     %e\n"+
			"  Theory:   Var(PRS) = Σ_j(2*p_j*(1-p_j)*β_j²)",
			computedVariance, expectedVariance, computedVariance-expectedVariance)
	}

	// Test population standard deviation
	if math.Abs(refStats.Std-expectedStd) > tolerance {
		t.Errorf("Population standard deviation incorrect:\n"+
			"  Got:      %.15f\n"+
			"  Expected: %.15f\n"+
			"  Diff:     %e",
			refStats.Std, expectedStd, refStats.Std-expectedStd)
	}

	t.Logf("✓ Closed-form validation PASSED with precision ±%e", tolerance)
	t.Logf("  Population Mean: %.12f", refStats.Mean)
	t.Logf("  Population Std:  %.12f", refStats.Std)
	t.Logf("  Population Var:  %.12f", computedVariance)
}

func TestMeanVarianceTheory_SingleSNPValidation(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Test cases for individual SNP contributions
	testCases := []struct {
		name        string
		freq        float64
		beta        float64
		expectedVar float64
	}{
		{
			name:        "Low_frequency_SNP",
			freq:        0.05,
			beta:        0.5,
			expectedVar: 2 * 0.05 * 0.95 * 0.25, // 2*p*(1-p)*β²
		},
		{
			name:        "Medium_frequency_SNP",
			freq:        0.5,
			beta:        0.2,
			expectedVar: 2 * 0.5 * 0.5 * 0.04, // Maximum variance at p=0.5
		},
		{
			name:        "High_frequency_SNP",
			freq:        0.9,
			beta:        0.3,
			expectedVar: 2 * 0.9 * 0.1 * 0.09,
		},
		{
			name:        "Zero_effect_SNP",
			freq:        0.4,
			beta:        0.0,
			expectedVar: 0.0, // Zero variance for zero effect
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alleleFreqs := map[string]float64{"test_snp": tc.freq}
			effectSizes := map[string]float64{"test_snp": tc.beta}

			// Expected mean: 2*p*β
			expectedMean := 2 * tc.freq * tc.beta

			refStats, err := reference_stats.Compute(alleleFreqs, effectSizes)
			if err != nil {
				t.Fatalf("Single SNP computation failed: %v", err)
			}

			const tolerance = 1e-15

			// Validate mean
			if math.Abs(refStats.Mean-expectedMean) > tolerance {
				t.Errorf("Single SNP mean incorrect: got %v, expected %v", refStats.Mean, expectedMean)
			}

			// Validate variance
			computedVariance := refStats.Std * refStats.Std
			if math.Abs(computedVariance-tc.expectedVar) > tolerance {
				t.Errorf("Single SNP variance incorrect: got %v, expected %v", computedVariance, tc.expectedVar)
			}

			t.Logf("✓ %s: freq=%.2f, β=%.2f → Mean=%.6f, Var=%.6f",
				tc.name, tc.freq, tc.beta, refStats.Mean, computedVariance)
		})
	}
}

func TestMeanVarianceTheory_AdditivityProperty(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Test that population parameters are additive across SNPs
	// Individual SNPs
	snp1 := map[string]float64{"rs1": 0.3}
	beta1 := map[string]float64{"rs1": 0.1}

	snp2 := map[string]float64{"rs2": 0.7}
	beta2 := map[string]float64{"rs2": -0.2}

	// Combined SNPs
	combined := map[string]float64{"rs1": 0.3, "rs2": 0.7}
	betaCombined := map[string]float64{"rs1": 0.1, "rs2": -0.2}

	// Compute individual statistics
	stats1, err := reference_stats.Compute(snp1, beta1)
	if err != nil {
		t.Fatalf("SNP1 computation failed: %v", err)
	}

	stats2, err := reference_stats.Compute(snp2, beta2)
	if err != nil {
		t.Fatalf("SNP2 computation failed: %v", err)
	}

	// Compute combined statistics
	statsCombined, err := reference_stats.Compute(combined, betaCombined)
	if err != nil {
		t.Fatalf("Combined computation failed: %v", err)
	}

	// Test additivity of means
	expectedCombinedMean := stats1.Mean + stats2.Mean
	const tolerance = 1e-15

	if math.Abs(statsCombined.Mean-expectedCombinedMean) > tolerance {
		t.Errorf("Mean additivity violated:\n"+
			"  Combined: %.15f\n"+
			"  Sum:      %.15f\n"+
			"  Diff:     %e",
			statsCombined.Mean, expectedCombinedMean, statsCombined.Mean-expectedCombinedMean)
	}

	// Test additivity of variances
	expectedCombinedVariance := stats1.Std*stats1.Std + stats2.Std*stats2.Std
	combinedVariance := statsCombined.Std * statsCombined.Std

	if math.Abs(combinedVariance-expectedCombinedVariance) > tolerance {
		t.Errorf("Variance additivity violated:\n"+
			"  Combined: %.15f\n"+
			"  Sum:      %.15f\n"+
			"  Diff:     %e",
			combinedVariance, expectedCombinedVariance, combinedVariance-expectedCombinedVariance)
	}

	t.Logf("✓ Additivity property validated:")
	t.Logf("  Individual means: %.6f + %.6f = %.6f", stats1.Mean, stats2.Mean, expectedCombinedMean)
	t.Logf("  Combined mean:    %.6f", statsCombined.Mean)
	t.Logf("  Individual vars:  %.6f + %.6f = %.6f", stats1.Std*stats1.Std, stats2.Std*stats2.Std, expectedCombinedVariance)
	t.Logf("  Combined var:     %.6f", combinedVariance)
}
