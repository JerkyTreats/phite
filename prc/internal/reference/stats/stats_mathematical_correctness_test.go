package reference_stats

import (
	"math"
	"testing"
)

// Hardy-Weinberg Equilibrium Tests
// These tests validate population parameters against known theoretical values

func TestHWE_TheoreticalPopulationMean(t *testing.T) {
	// Test case from brief: 3 SNPs with p=(0.2,0.5,0.8), β=(0.1,-0.3,0.2)
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

	// Theoretical population mean: μ_pop = Σ_j(2*p_j*β_j)
	expectedMean := 2 * (0.2*0.1 + 0.5*(-0.3) + 0.8*0.2)
	// = 2*(0.02 + (-0.15) + 0.16) = 2*0.03 = 0.06

	stats, err := Compute(alleleFreqs, effectSizes)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	const tolerance = 1e-12
	if math.Abs(stats.Mean-expectedMean) > tolerance {
		t.Errorf("Population mean incorrect: got %v, expected %v", stats.Mean, expectedMean)
		t.Errorf("Formula should be: μ_pop = Σ_j(2*p_j*β_j)")
		t.Errorf("Current implementation uses sample statistics (WRONG)")
	}
}

func TestHWE_TheoreticalPopulationVariance(t *testing.T) {
	// Same test case: p=(0.2,0.5,0.8), β=(0.1,-0.3,0.2)
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

	// Theoretical population variance: Var(PRS) = Σ_j(2*p_j*(1-p_j)*β_j²)
	expectedVariance := 2 * (0.2*0.8*0.01 + 0.5*0.5*0.09 + 0.8*0.2*0.04)
	// = 2*(0.0016 + 0.0225 + 0.0064) = 2*0.0305 = 0.0610
	expectedStd := math.Sqrt(expectedVariance)

	stats, err := Compute(alleleFreqs, effectSizes)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	const tolerance = 1e-12
	if math.Abs(stats.Std-expectedStd) > tolerance {
		t.Errorf("Population variance incorrect: got %v², expected %v²", stats.Std, expectedStd)
		t.Errorf("Formula should be: Var(PRS) = Σ_j(2*p_j*(1-p_j)*β_j²)")
		t.Errorf("Current implementation uses E[x²]-E[x]² (WRONG)")
	}
}

func TestHWE_SingleSNPVariance(t *testing.T) {
	// Single SNP test to expose the variance calculation flaw
	testCases := []struct {
		name   string
		freq   float64
		effect float64
	}{
		{"Low frequency", 0.1, 0.5},
		{"Medium frequency", 0.5, 0.3},
		{"High frequency", 0.9, 0.2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alleleFreqs := map[string]float64{"rs1": tc.freq}
			effectSizes := map[string]float64{"rs1": tc.effect}

			// Theoretical variance for single SNP: 2*p*(1-p)*β²
			expectedVariance := 2 * tc.freq * (1 - tc.freq) * tc.effect * tc.effect
			expectedStd := math.Sqrt(expectedVariance)

			stats, err := Compute(alleleFreqs, effectSizes)
			if err != nil {
				t.Fatalf("Compute failed: %v", err)
			}

			const tolerance = 1e-12
			if math.Abs(stats.Std-expectedStd) > tolerance {
				t.Errorf("Single SNP variance incorrect: got %v, expected %v", stats.Std, expectedStd)
				t.Errorf("Current implementation likely gives variance ≈ 0 for single variant")
				t.Errorf("This is mathematically impossible under HWE")
			}
		})
	}
}

func TestHWE_PopulationParametersVsSampleStatistics(t *testing.T) {
	// This test demonstrates the difference between population parameters
	// and sample statistics - current implementation is wrong
	alleleFreqs := map[string]float64{
		"rs1": 0.3,
		"rs2": 0.7,
	}
	effectSizes := map[string]float64{
		"rs1": 0.4,
		"rs2": -0.2,
	}

	// CORRECT population parameters:
	popMean := 2 * (0.3*0.4 + 0.7*(-0.2))       // = 2*(0.12 - 0.14) = -0.04
	popVar := 2 * (0.3*0.7*0.16 + 0.7*0.3*0.04) // = 2*(0.0336 + 0.0084) = 0.084
	popStd := math.Sqrt(popVar)

	stats, err := Compute(alleleFreqs, effectSizes)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	// Current implementation uses sample statistics (divides by n)
	// This is fundamentally wrong for population parameters
	const tolerance = 1e-12
	if math.Abs(stats.Mean-popMean) > tolerance {
		t.Errorf("Population mean: got %v, expected %v", stats.Mean, popMean)
	}
	if math.Abs(stats.Std-popStd) > tolerance {
		t.Errorf("Population std: got %v, expected %v", stats.Std, popStd)
		t.Errorf("Current implementation uses wrong formula for population variance")
	}
}

func TestHWE_ExtremeCases(t *testing.T) {
	testCases := []struct {
		name    string
		freq    float64
		effect  float64
		expMean float64
		expVar  float64
	}{
		{
			name:    "Frequency near 0",
			freq:    0.001,
			effect:  1.0,
			expMean: 2 * 0.001 * 1.0,
			expVar:  2 * 0.001 * 0.999 * 1.0,
		},
		{
			name:    "Frequency near 1",
			freq:    0.999,
			effect:  0.5,
			expMean: 2 * 0.999 * 0.5,
			expVar:  2 * 0.999 * 0.001 * 0.25,
		},
		{
			name:    "Zero effect",
			freq:    0.5,
			effect:  0.0,
			expMean: 0.0,
			expVar:  0.0,
		},
		{
			name:    "Large effect",
			freq:    0.4,
			effect:  2.0,
			expMean: 2 * 0.4 * 2.0,
			expVar:  2 * 0.4 * 0.6 * 4.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alleleFreqs := map[string]float64{"rs1": tc.freq}
			effectSizes := map[string]float64{"rs1": tc.effect}

			stats, err := Compute(alleleFreqs, effectSizes)
			if err != nil {
				t.Fatalf("Compute failed: %v", err)
			}

			const tolerance = 1e-12
			if math.Abs(stats.Mean-tc.expMean) > tolerance {
				t.Errorf("Mean: got %v, expected %v", stats.Mean, tc.expMean)
			}

			expStd := math.Sqrt(tc.expVar)
			if math.Abs(stats.Std-expStd) > tolerance {
				t.Errorf("Std: got %v, expected %v", stats.Std, expStd)
			}
		})
	}
}

func TestHWE_NumericalStability(t *testing.T) {
	// Test numerical stability with very small and very large values
	testCases := []struct {
		name   string
		freq   float64
		effect float64
	}{
		{"Very small effect", 0.5, 1e-10},
		{"Very large effect", 0.5, 1e10},
		{"Very small frequency", 1e-6, 0.5},
		{"Very large frequency", 1 - 1e-6, 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alleleFreqs := map[string]float64{"rs1": tc.freq}
			effectSizes := map[string]float64{"rs1": tc.effect}

			stats, err := Compute(alleleFreqs, effectSizes)
			if err != nil {
				t.Fatalf("Compute failed: %v", err)
			}

			// Basic sanity checks
			if math.IsNaN(stats.Mean) || math.IsInf(stats.Mean, 0) {
				t.Errorf("Mean is NaN or Inf: %v", stats.Mean)
			}
			if math.IsNaN(stats.Std) || math.IsInf(stats.Std, 0) {
				t.Errorf("Std is NaN or Inf: %v", stats.Std)
			}
			if stats.Std < 0 {
				t.Errorf("Std is negative: %v", stats.Std)
			}
		})
	}
}
