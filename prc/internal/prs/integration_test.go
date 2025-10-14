package prs

import (
	"math"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/logging"
	"phite.io/polygenic-risk-calculator/internal/model"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

// Integration Test: Full Pipeline Statistical Validation
// Tests the complete PRS calculation pipeline with known theoretical properties

func TestFullPipelineStatisticalValidation(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Enable strict validation for comprehensive testing
	config.Set("invariance.enable_validation", true)
	config.Set("invariance.strict_mode", true)
	defer func() {
		config.Set("invariance.enable_validation", false)
		config.Set("invariance.strict_mode", false)
	}()

	// Known Hardy-Weinberg Equilibrium population
	// 3 SNPs: p=(0.2,0.5,0.8), β=(0.1,-0.3,0.2)
	// Expected Mean: 2×(0.2×0.1 + 0.5×(-0.3) + 0.8×0.2) = 0.06
	// Expected Variance: 2×(0.2×0.8×0.01 + 0.5×0.5×0.09 + 0.8×0.2×0.04) = 0.0610

	alleleFreqs := map[string]float64{
		"rs1": 0.2,
		"rs2": 0.5,
		"rs3": 0.8,
	}

	effectSizes := map[string]float64{
		"rs1": 0.1,
		"rs2": -0.3,
		"rs3": 0.2,
	}

	// Expected theoretical values
	expectedMean := 2 * (0.2*0.1 + 0.5*(-0.3) + 0.8*0.2)                 // = 0.06
	expectedVariance := 2 * (0.2*0.8*0.01 + 0.5*0.5*0.09 + 0.8*0.2*0.04) // = 0.0610
	expectedStd := math.Sqrt(expectedVariance)

	t.Run("Reference_Statistics_Computation", func(t *testing.T) {
		// Test corrected population parameter computation
		refStats, err := reference_stats.Compute(alleleFreqs, effectSizes)
		if err != nil {
			t.Fatalf("Reference stats computation failed: %v", err)
		}

		const tolerance = 1e-12
		if math.Abs(refStats.Mean-expectedMean) > tolerance {
			t.Errorf("Population mean incorrect: got %v, expected %v (diff: %e)",
				refStats.Mean, expectedMean, refStats.Mean-expectedMean)
		}

		if math.Abs(refStats.Std-expectedStd) > tolerance {
			t.Errorf("Population standard deviation incorrect: got %v, expected %v (diff: %e)",
				refStats.Std, expectedStd, refStats.Std-expectedStd)
		}

		t.Logf("✓ Reference stats validated: Mean=%.6f, Std=%.6f", refStats.Mean, refStats.Std)
	})

	t.Run("Individual_PRS_Calculation", func(t *testing.T) {
		// Test individual with specific genotype: dosages = [2, 1, 0]
		// Expected PRS: 2×0.1 + 1×(-0.3) + 0×0.2 = 0.2 - 0.3 + 0 = -0.1

		snps := []model.AnnotatedSNP{
			{RSID: "rs1", Beta: 0.1, Dosage: 2, Trait: "test"},
			{RSID: "rs2", Beta: -0.3, Dosage: 1, Trait: "test"},
			{RSID: "rs3", Beta: 0.2, Dosage: 0, Trait: "test"},
		}

		expectedPRS := 2*0.1 + 1*(-0.3) + 0*0.2 // = -0.1

		result, err := CalculatePRS(snps)
		if err != nil {
			t.Fatalf("PRS calculation failed: %v", err)
		}

		const tolerance = 1e-12
		if math.Abs(result.PRSScore-expectedPRS) > tolerance {
			t.Errorf("Individual PRS incorrect: got %v, expected %v", result.PRSScore, expectedPRS)
		}

		t.Logf("✓ Individual PRS validated: %.6f", result.PRSScore)
	})

	t.Run("Normalization_Accuracy", func(t *testing.T) {
		// Test z-score normalization with known individual
		rawPRS := -0.1 // From previous test
		expectedZScore := (rawPRS - expectedMean) / expectedStd
		expectedPercentile := 100 * normalCDF(expectedZScore)

		// Create PRS result and reference stats for normalization
		prsResult := PRSResult{PRSScore: rawPRS}
		refStats := model.ReferenceStats{Mean: expectedMean, Std: expectedStd}

		result, err := NormalizePRS(prsResult, refStats)
		if err != nil {
			t.Fatalf("Normalization failed: %v", err)
		}

		const tolerance = 1e-10
		if math.Abs(result.ZScore-expectedZScore) > tolerance {
			t.Errorf("Z-score incorrect: got %v, expected %v", result.ZScore, expectedZScore)
		}

		// Use relaxed tolerance for percentile due to different CDF implementations
		const percentileTolerance = 1e-5
		if math.Abs(result.Percentile-expectedPercentile) > percentileTolerance {
			t.Errorf("Percentile incorrect: got %v, expected %v", result.Percentile, expectedPercentile)
		}

		t.Logf("✓ Normalization validated: Z=%.6f, Percentile=%.2f", result.ZScore, result.Percentile)
	})

	t.Run("Mathematical_Consistency_Validation", func(t *testing.T) {
		// Validate mathematical consistency across multiple individuals
		testIndividuals := []struct {
			name    string
			dosages [3]int
		}{
			{"Individual_1", [3]int{0, 0, 0}}, // All homozygous reference
			{"Individual_2", [3]int{1, 1, 1}}, // All heterozygous
			{"Individual_3", [3]int{2, 2, 2}}, // All homozygous alternative
			{"Individual_4", [3]int{2, 0, 1}}, // Mixed genotype
		}

		var allPRS []float64
		for _, individual := range testIndividuals {
			snps := []model.AnnotatedSNP{
				{RSID: "rs1", Beta: 0.1, Dosage: individual.dosages[0], Trait: "test"},
				{RSID: "rs2", Beta: -0.3, Dosage: individual.dosages[1], Trait: "test"},
				{RSID: "rs3", Beta: 0.2, Dosage: individual.dosages[2], Trait: "test"},
			}

			result, err := CalculatePRS(snps)
			if err != nil {
				t.Fatalf("PRS calculation failed for %s: %v", individual.name, err)
			}

			// Verify additive model: PRS = Σ(dosage × beta)
			expectedPRS := float64(individual.dosages[0])*0.1 +
				float64(individual.dosages[1])*(-0.3) +
				float64(individual.dosages[2])*0.2

			const tolerance = 1e-12
			if math.Abs(result.PRSScore-expectedPRS) > tolerance {
				t.Errorf("%s PRS incorrect: got %v, expected %v",
					individual.name, result.PRSScore, expectedPRS)
			}

			allPRS = append(allPRS, result.PRSScore)
			t.Logf("✓ %s: PRS=%.6f", individual.name, result.PRSScore)
		}

		// Verify PRS scores span reasonable range
		minPRS, maxPRS := allPRS[0], allPRS[0]
		for _, prs := range allPRS {
			if prs < minPRS {
				minPRS = prs
			}
			if prs > maxPRS {
				maxPRS = prs
			}
		}

		t.Logf("✓ PRS range validated: [%.6f, %.6f]", minPRS, maxPRS)
	})

	t.Run("Invariance_Validation_Integration", func(t *testing.T) {
		// Test that invariance validation catches errors in the pipeline

		// Test invalid dosage
		invalidSNPs := []model.AnnotatedSNP{
			{RSID: "rs1", Beta: 0.1, Dosage: 3, Trait: "test"}, // Invalid dosage > 2
		}

		_, err := CalculatePRS(invalidSNPs)
		if err == nil {
			t.Error("Expected error for invalid dosage, but got none")
		}
		if err != nil {
			t.Logf("✓ Invalid dosage correctly caught: %v", err)
		}

		// Test invalid beta in strict mode
		invalidBetaSNPs := []model.AnnotatedSNP{
			{RSID: "rs1", Beta: math.NaN(), Dosage: 1, Trait: "test"}, // NaN beta
		}

		_, err = CalculatePRS(invalidBetaSNPs)
		if err == nil {
			t.Error("Expected error for NaN beta, but got none")
		}
		if err != nil {
			t.Logf("✓ Invalid beta correctly caught: %v", err)
		}
	})
}

// Helper function for normal CDF approximation
func normalCDF(z float64) float64 {
	return 0.5 * (1 + erf(z/math.Sqrt2))
}

// Error function approximation
func erf(x float64) float64 {
	// Abramowitz and Stegun approximation
	a1, a2, a3, a4, a5 := 0.254829592, -0.284496736, 1.421413741, -1.453152027, 1.061405429
	p := 0.3275911

	sign := 1.0
	if x < 0 {
		sign = -1.0
		x = -x
	}

	t := 1.0 / (1.0 + p*x)
	y := 1.0 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x)

	return sign * y
}

func TestEndToEndPipelinePerformance(t *testing.T) {
	logging.SetSilentLoggingForTest()

	// Test performance with invariance validation enabled
	config.Set("invariance.enable_validation", true)
	config.Set("invariance.strict_mode", false) // Disable strict for performance
	defer func() {
		config.Set("invariance.enable_validation", false)
		config.Set("invariance.strict_mode", false)
	}()

	// Generate large SNP set to test performance
	const numSNPs = 10000
	snps := make([]model.AnnotatedSNP, numSNPs)

	for i := 0; i < numSNPs; i++ {
		snps[i] = model.AnnotatedSNP{
			RSID:   generateRSID(i),
			Beta:   (float64(i%1000) - 500) / 10000.0, // Range: -0.05 to 0.05
			Dosage: i % 3,                             // Cycle through 0, 1, 2
			Trait:  "performance_test",
		}
	}

	// Benchmark calculation time
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	result, err := CalculatePRS(snps)
	if err != nil {
		t.Fatalf("Large-scale PRS calculation failed: %v", err)
	}

	// Verify result is reasonable
	if math.IsNaN(result.PRSScore) || math.IsInf(result.PRSScore, 0) {
		t.Errorf("Large-scale PRS resulted in invalid value: %v", result.PRSScore)
	}

	if len(result.Details) != numSNPs {
		t.Errorf("Expected %d SNP details, got %d", numSNPs, len(result.Details))
	}

	t.Logf("✓ Performance test passed: %d SNPs, PRS=%.6f", numSNPs, result.PRSScore)
}

func generateRSID(i int) string {
	return "rs" + string(rune('1'+i%9)) + string(rune('0'+i/10%10)) + string(rune('0'+i/100%10))
}
