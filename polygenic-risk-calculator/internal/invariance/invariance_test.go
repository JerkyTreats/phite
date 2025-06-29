package invariance

import (
	"math"
	"testing"

	"phite.io/polygenic-risk-calculator/internal/config"
)

// Test AssertValidProbability
func TestAssertValidProbability(t *testing.T) {
	testCases := []struct {
		name      string
		value     float64
		shouldErr bool
	}{
		{"Valid probability 0", 0.0, false},
		{"Valid probability 0.5", 0.5, false},
		{"Valid probability 1", 1.0, false},
		{"Invalid negative", -0.1, true},
		{"Invalid > 1", 1.1, true},
		{"Invalid NaN", math.NaN(), true},
		{"Invalid +Inf", math.Inf(1), true},
		{"Invalid -Inf", math.Inf(-1), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AssertValidProbability(tc.value, "unit test")
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for value %v", tc.value)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error for value %v: %v", tc.value, err)
			}
		})
	}
}

// Test AssertValidVariance
func TestAssertValidVariance(t *testing.T) {
	testCases := []struct {
		name      string
		variance  float64
		shouldErr bool
	}{
		{"Valid zero variance", 0.0, false},
		{"Valid positive variance", 1.5, false},
		{"Invalid negative variance", -0.1, true},
		{"Invalid NaN", math.NaN(), true},
		{"Invalid Inf", math.Inf(1), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AssertValidVariance(tc.variance, "unit test")
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for variance %v", tc.variance)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error for variance %v: %v", tc.variance, err)
			}
		})
	}
}

// Test AssertValidDosage
func TestAssertValidDosage(t *testing.T) {
	testCases := []struct {
		name      string
		dosage    int
		shouldErr bool
	}{
		{"Valid dosage 0", 0, false},
		{"Valid dosage 1", 1, false},
		{"Valid dosage 2", 2, false},
		{"Invalid negative", -1, true},
		{"Invalid > 2", 3, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AssertValidDosage(tc.dosage, "unit test")
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for dosage %v", tc.dosage)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error for dosage %v: %v", tc.dosage, err)
			}
		})
	}
}

// Test AssertValidBetaCoefficient
func TestAssertValidBetaCoefficient(t *testing.T) {
	testCases := []struct {
		name      string
		beta      float64
		shouldErr bool
	}{
		{"Valid beta", 0.5, false},
		{"Valid negative beta", -0.3, false},
		{"Valid zero beta", 0.0, false},
		{"Invalid NaN", math.NaN(), true},
		{"Invalid Inf", math.Inf(1), true},
		{"Large but valid beta", 9.9, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AssertValidBetaCoefficient(tc.beta, "unit test")
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for beta %v", tc.beta)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error for beta %v: %v", tc.beta, err)
			}
		})
	}
}

// Test AssertHardyWeinbergVariance
func TestAssertHardyWeinbergVariance(t *testing.T) {
	testCases := []struct {
		name             string
		frequency        float64
		effect           float64
		observedVariance float64
		shouldErr        bool
	}{
		{
			name:             "Correct HWE variance",
			frequency:        0.3,
			effect:           0.5,
			observedVariance: 2 * 0.3 * 0.7 * 0.25, // 2*p*(1-p)*β²
			shouldErr:        false,
		},
		{
			name:             "Zero frequency zero variance",
			frequency:        0.0,
			effect:           0.5,
			observedVariance: 0.0,
			shouldErr:        false,
		},
		{
			name:             "Incorrect variance",
			frequency:        0.3,
			effect:           0.5,
			observedVariance: 0.1,
			shouldErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AssertHardyWeinbergVariance(tc.frequency, tc.effect, tc.observedVariance, "unit test")
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for variance validation")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test AssertPopulationParameterConsistency
func TestAssertPopulationParameterConsistency(t *testing.T) {
	// Test case from the brief: p=(0.2,0.5,0.8), β=(0.1,-0.3,0.2)
	freqs := []float64{0.2, 0.5, 0.8}
	effects := []float64{0.1, -0.3, 0.2}

	expectedMean := 2 * (0.2*0.1 + 0.5*(-0.3) + 0.8*0.2)                 // = 0.06
	expectedVariance := 2 * (0.2*0.8*0.01 + 0.5*0.5*0.09 + 0.8*0.2*0.04) // = 0.0610

	testCases := []struct {
		name      string
		mean      float64
		variance  float64
		shouldErr bool
	}{
		{"Correct parameters", expectedMean, expectedVariance, false},
		{"Incorrect mean", 0.03, expectedVariance, true},
		{"Incorrect variance", expectedMean, 0.03, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AssertPopulationParameterConsistency(freqs, effects, tc.mean, tc.variance, "unit test")
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for parameter consistency check")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test AssertMonotonicity (renamed from AssertMonotonicityProperty)
func TestAssertMonotonicity(t *testing.T) {
	testCases := []struct {
		name       string
		values     []float64
		increasing bool
		shouldErr  bool
	}{
		{
			name:       "Monotonic increasing",
			values:     []float64{0.1, 0.5, 0.8, 0.95},
			increasing: true,
			shouldErr:  false,
		},
		{
			name:       "Non-monotonic",
			values:     []float64{0.1, 0.8, 0.5, 0.95},
			increasing: true,
			shouldErr:  true,
		},
		{
			name:       "Monotonic decreasing",
			values:     []float64{0.95, 0.8, 0.5, 0.1},
			increasing: false,
			shouldErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AssertMonotonicity(tc.values, tc.increasing, "unit test")
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for monotonicity check")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Benchmark tests for performance overhead
func BenchmarkAssertValidProbability(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = AssertValidProbability(0.5, "benchmark")
	}
}

func BenchmarkAssertValidDosage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = AssertValidDosage(1, "benchmark")
	}
}

// Test numerical stability assertions
func TestNumericalStability(t *testing.T) {
	testCases := []struct {
		name      string
		value     float64
		shouldErr bool
	}{
		{"Normal value", 1.5, false},
		{"Zero", 0.0, false},
		{"Small value", 1e-10, false},
		{"Large value", 1e50, false},
		{"NaN", math.NaN(), true},
		{"Positive infinity", math.Inf(1), true},
		{"Negative infinity", math.Inf(-1), true},
		{"Overflow boundary", 1e101, true},
		{"Underflow boundary", 1e-301, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Enable strict mode for overflow/underflow tests
			if tc.name == "Overflow boundary" || tc.name == "Underflow boundary" {
				config.SetForTest(StrictModeKey, true)
				defer config.SetForTest(StrictModeKey, false)
			}

			err := AssertNumericalStability(tc.value, "unit test")
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for value %v", tc.value)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error for value %v: %v", tc.value, err)
			}
		})
	}
}
