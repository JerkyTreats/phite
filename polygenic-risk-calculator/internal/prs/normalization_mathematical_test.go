package prs

import (
	"fmt"
	"math"
	"testing"

	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

// PRS Normalization Mathematical Tests
// These tests validate Z-score and percentile calculations

func TestNormalization_ZScoreAccuracy(t *testing.T) {
	// Test Z-score calculation: Z = (PRS - μ) / σ
	refStats := &reference_stats.ReferenceStats{
		Mean: 0.5,
		Std:  0.2,
		Min:  0.0,
		Max:  1.0,
	}

	testCases := []struct {
		name      string
		rawPRS    float64
		expectedZ float64
		expectedP float64
	}{
		{
			name:      "Mean PRS",
			rawPRS:    0.5,
			expectedZ: 0.0,
			expectedP: 0.5,
		},
		{
			name:      "One std above mean",
			rawPRS:    0.7,
			expectedZ: 1.0,
			expectedP: 0.8413447460685429, // Φ(1)
		},
		{
			name:      "One std below mean",
			rawPRS:    0.3,
			expectedZ: -1.0,
			expectedP: 0.15865525393145705, // Φ(-1)
		},
		{
			name:      "Two std above mean",
			rawPRS:    0.9,
			expectedZ: 2.0,
			expectedP: 0.9772498680518208, // Φ(2)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate Z-score manually
			zScore := (tc.rawPRS - refStats.Mean) / refStats.Std

			const tolerance = 1e-12
			if math.Abs(zScore-tc.expectedZ) > tolerance {
				t.Errorf("Z-score incorrect: got %v, expected %v", zScore, tc.expectedZ)
			}

			// Test percentile calculation
			percentile, err := refStats.NormalizePRS(tc.rawPRS)
			if err != nil {
				t.Fatalf("NormalizePRS failed: %v", err)
			}

			const pTolerance = 1e-14
			if math.Abs(percentile-tc.expectedP) > pTolerance {
				t.Errorf("Percentile incorrect: got %v, expected %v", percentile, tc.expectedP)
			}
		})
	}
}

func TestNormalization_PercentileFormula(t *testing.T) {
	// Test percentile formula: Φ(z) = 0.5 * (1 + erf(z/√2))
	refStats := &reference_stats.ReferenceStats{
		Mean: 0.0,
		Std:  1.0,
		Min:  -5.0,
		Max:  5.0,
	}

	testCases := []struct {
		name     string
		zScore   float64
		expected float64
	}{
		{"Standard normal mean", 0.0, 0.5},
		{"z = 1", 1.0, 0.8413447460685429},
		{"z = -1", -1.0, 0.15865525393145705},
		{"z = 2", 2.0, 0.9772498680518208},
		{"z = -2", -2.0, 0.022750131948179195},
		{"z = 3", 3.0, 0.9986501019683699},
		{"z = -3", -3.0, 0.0013498980316301035},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rawPRS := tc.zScore // Since mean=0, std=1

			percentile, err := refStats.NormalizePRS(rawPRS)
			if err != nil {
				t.Fatalf("NormalizePRS failed: %v", err)
			}

			const tolerance = 1e-14
			if math.Abs(percentile-tc.expected) > tolerance {
				t.Errorf("Percentile formula incorrect: got %v, expected %v", percentile, tc.expected)
			}

			// Verify formula manually
			erfValue := math.Erf(tc.zScore / math.Sqrt(2))
			manualPercentile := 0.5 * (1 + erfValue)

			if math.Abs(manualPercentile-tc.expected) > tolerance {
				t.Errorf("Manual percentile calculation incorrect: got %v, expected %v", manualPercentile, tc.expected)
			}
		})
	}
}

func TestNormalization_SymmetryProperty(t *testing.T) {
	// Test that Φ(-z) = 1 - Φ(z)
	refStats := &reference_stats.ReferenceStats{
		Mean: 0.0,
		Std:  1.0,
		Min:  -5.0,
		Max:  5.0,
	}

	zValues := []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0}

	for _, z := range zValues {
		t.Run(fmt.Sprintf("z=%v", z), func(t *testing.T) {
			pPos, err := refStats.NormalizePRS(z)
			if err != nil {
				t.Fatalf("NormalizePRS(+z) failed: %v", err)
			}

			pNeg, err := refStats.NormalizePRS(-z)
			if err != nil {
				t.Fatalf("NormalizePRS(-z) failed: %v", err)
			}

			// Verify symmetry: Φ(-z) + Φ(z) = 1
			sum := pPos + pNeg
			const tolerance = 1e-14
			if math.Abs(sum-1.0) > tolerance {
				t.Errorf("Symmetry violation: Φ(%v) + Φ(%v) = %v, expected 1.0", z, -z, sum)
			}
		})
	}
}

func TestNormalization_ExtremeCases(t *testing.T) {
	refStats := &reference_stats.ReferenceStats{
		Mean: 0.0,
		Std:  1.0,
		Min:  -10.0,
		Max:  10.0,
	}

	testCases := []struct {
		name        string
		rawPRS      float64
		expectedMin float64
		expectedMax float64
	}{
		{"Very negative", -5.0, 0.0, 1e-6},
		{"Very positive", 5.0, 1.0 - 1e-6, 1.0},
		{"Extremely negative", -10.0, 0.0, 1e-20},
		{"Extremely positive", 10.0, 1.0 - 1e-20, 1.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			percentile, err := refStats.NormalizePRS(tc.rawPRS)
			if err != nil {
				t.Fatalf("NormalizePRS failed: %v", err)
			}

			if percentile < tc.expectedMin || percentile > tc.expectedMax {
				t.Errorf("Extreme case percentile out of range: got %v, expected [%v, %v]",
					percentile, tc.expectedMin, tc.expectedMax)
			}

			// Verify bounds
			if percentile < 0.0 || percentile > 1.0 {
				t.Errorf("Percentile out of [0,1] bounds: %v", percentile)
			}
		})
	}
}

func TestNormalization_NumericalStability(t *testing.T) {
	// Test numerical stability with various parameter combinations
	testCases := []struct {
		name string
		mean float64
		std  float64
		prs  float64
	}{
		{"Small std", 0.0, 1e-6, 1e-7},
		{"Large std", 0.0, 1e6, 1e5},
		{"Large mean", 1e6, 1.0, 1e6 + 1.0},
		{"Small mean", -1e6, 1.0, -1e6 + 1.0},
		{"Very small values", 1e-10, 1e-10, 2e-10},
		{"Very large values", 1e10, 1e9, 1.5e10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			refStats := &reference_stats.ReferenceStats{
				Mean: tc.mean,
				Std:  tc.std,
				Min:  tc.mean - 5*tc.std,
				Max:  tc.mean + 5*tc.std,
			}

			percentile, err := refStats.NormalizePRS(tc.prs)
			if err != nil {
				t.Fatalf("NormalizePRS failed: %v", err)
			}

			// Basic sanity checks
			if math.IsNaN(percentile) || math.IsInf(percentile, 0) {
				t.Errorf("Percentile is NaN or Inf: %v", percentile)
			}

			if percentile < 0.0 || percentile > 1.0 {
				t.Errorf("Percentile out of bounds: %v", percentile)
			}
		})
	}
}

func TestNormalization_InvalidStats(t *testing.T) {
	// Test error handling for invalid reference statistics
	invalidStats := []reference_stats.ReferenceStats{
		{Mean: 0.0, Std: 0.0, Min: -1.0, Max: 1.0},  // Zero std
		{Mean: 0.0, Std: -1.0, Min: -1.0, Max: 1.0}, // Negative std
		{Mean: 0.0, Std: 1.0, Min: 1.0, Max: -1.0},  // Min > Max
		{Mean: 2.0, Std: 1.0, Min: -1.0, Max: 1.0},  // Mean outside [min, max]
	}

	for i, stats := range invalidStats {
		t.Run(fmt.Sprintf("Invalid case %d", i), func(t *testing.T) {
			_, err := stats.NormalizePRS(0.0)
			if err == nil {
				t.Errorf("Expected error for invalid stats, got nil")
			}
		})
	}
}

func TestNormalization_MonotonicityProperty(t *testing.T) {
	// Test that percentiles are monotonically increasing with PRS
	refStats := &reference_stats.ReferenceStats{
		Mean: 0.0,
		Std:  1.0,
		Min:  -5.0,
		Max:  5.0,
	}

	prsValues := []float64{-3.0, -2.0, -1.0, 0.0, 1.0, 2.0, 3.0}
	var percentiles []float64

	for _, prs := range prsValues {
		p, err := refStats.NormalizePRS(prs)
		if err != nil {
			t.Fatalf("NormalizePRS failed for %v: %v", prs, err)
		}
		percentiles = append(percentiles, p)
	}

	// Verify monotonicity
	for i := 1; i < len(percentiles); i++ {
		if percentiles[i] <= percentiles[i-1] {
			t.Errorf("Monotonicity violation: percentile[%d]=%v <= percentile[%d]=%v",
				i, percentiles[i], i-1, percentiles[i-1])
		}
	}
}
