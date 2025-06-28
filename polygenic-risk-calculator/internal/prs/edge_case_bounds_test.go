package prs

import (
	"math"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/model"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

// Test #2: Invariants under extreme inputs
// Input: Single SNP, p=1, β=0.5; β=0, p=0.3; p→0
// Expected Output: Mean/Var analytically zero/finite, Var≥0

func TestEdgeCaseBounds_ExtremeFrequencies(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Enable validation for edge case testing
	config.SetForTest("invariance.enable_validation", true)
	defer config.SetForTest("invariance.enable_validation", false)

	testCases := []struct {
		name         string
		freq         float64
		beta         float64
		expectedMean float64
		expectedVar  float64
		description  string
	}{
		{
			name:         "Fixed_allele_p_equals_1",
			freq:         1.0,
			beta:         0.5,
			expectedMean: 2 * 1.0 * 0.5, // = 1.0
			expectedVar:  0.0,           // No variance when p=1 (fixed allele)
			description:  "Fixed allele (p=1) should have zero variance",
		},
		{
			name:         "Absent_allele_p_equals_0",
			freq:         0.0,
			beta:         0.3,
			expectedMean: 0.0, // No contribution when allele absent
			expectedVar:  0.0, // No variance when p=0
			description:  "Absent allele (p=0) should contribute nothing",
		},
		{
			name:         "Zero_effect_nonzero_frequency",
			freq:         0.3,
			beta:         0.0,
			expectedMean: 0.0, // No effect regardless of frequency
			expectedVar:  0.0, // No variance when β=0
			description:  "Zero effect should contribute nothing regardless of frequency",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alleleFreqs := map[string]float64{"test_snp": tc.freq}
			effectSizes := map[string]float64{"test_snp": tc.beta}

			refStats, err := reference_stats.Compute(alleleFreqs, effectSizes)
			if err != nil {
				t.Fatalf("Edge case computation failed: %v", err)
			}

			const tolerance = 1e-15

			// Validate mean
			if math.Abs(refStats.Mean-tc.expectedMean) > tolerance {
				t.Errorf("%s: Mean incorrect, got %.15f, expected %.15f",
					tc.description, refStats.Mean, tc.expectedMean)
			}

			// Validate variance
			computedVariance := refStats.Std * refStats.Std
			if math.Abs(computedVariance-tc.expectedVar) > tolerance {
				t.Errorf("%s: Variance incorrect, got %.15f, expected %.15f",
					tc.description, computedVariance, tc.expectedVar)
			}

			// Critical invariant: Variance must be non-negative
			if computedVariance < -tolerance {
				t.Errorf("%s: CRITICAL INVARIANT VIOLATION - Negative variance: %.15f",
					tc.description, computedVariance)
			}

			t.Logf("✓ %s: freq=%.2e, β=%.2f → Mean=%.6e, Var=%.6e",
				tc.name, tc.freq, tc.beta, refStats.Mean, computedVariance)
		})
	}
}

func TestEdgeCaseBounds_PRSCalculationExtreme(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Test PRS calculation with extreme dosages and effects
	testCases := []struct {
		name        string
		snps        []model.AnnotatedSNP
		expectedPRS float64
		description string
	}{
		{
			name: "All_homozygous_reference",
			snps: []model.AnnotatedSNP{
				{RSID: "rs1", Beta: 0.5, Dosage: 0, Trait: "test"},
				{RSID: "rs2", Beta: -0.3, Dosage: 0, Trait: "test"},
			},
			expectedPRS: 0.0, // All dosages are 0
			description: "All homozygous reference should give PRS=0",
		},
		{
			name: "All_homozygous_alternative",
			snps: []model.AnnotatedSNP{
				{RSID: "rs1", Beta: 0.1, Dosage: 2, Trait: "test"},
				{RSID: "rs2", Beta: 0.2, Dosage: 2, Trait: "test"},
			},
			expectedPRS: 2 * (0.1 + 0.2), // 2 * 0.3 = 0.6
			description: "All homozygous alternative with positive effects",
		},
	}

	// Disable strict mode for extreme numerical testing
	config.SetForTest("invariance.strict_mode", false)
	defer config.SetForTest("invariance.strict_mode", false)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CalculatePRS(tc.snps)
			if err != nil {
				t.Errorf("%s: Unexpected error: %v", tc.description, err)
			}

			const tolerance = 1e-12
			if math.Abs(result.PRSScore-tc.expectedPRS) > tolerance {
				t.Errorf("%s: PRS incorrect, got %.15f, expected %.15f",
					tc.description, result.PRSScore, tc.expectedPRS)
			}

			t.Logf("✓ %s: PRS=%.6f", tc.name, result.PRSScore)
		})
	}
}
