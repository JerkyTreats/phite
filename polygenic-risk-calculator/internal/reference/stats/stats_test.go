package reference_stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReferenceStats_Validate(t *testing.T) {
	tests := []struct {
		name    string
		stats   ReferenceStats
		wantErr bool
	}{
		{
			name: "valid stats",
			stats: ReferenceStats{
				Mean:     0.5,
				Std:      1.0,
				Min:      0.0,
				Max:      1.0,
				Ancestry: "EUR",
				Trait:    "Height",
				Model:    "test_model",
			},
			wantErr: false,
		},
		{
			name: "negative std dev",
			stats: ReferenceStats{
				Mean:     0.5,
				Std:      -1.0,
				Min:      0.0,
				Max:      1.0,
				Ancestry: "EUR",
				Trait:    "Height",
				Model:    "test_model",
			},
			wantErr: true,
		},
		{
			name: "zero std dev",
			stats: ReferenceStats{
				Mean:     0.5,
				Std:      0.0,
				Min:      0.0,
				Max:      1.0,
				Ancestry: "EUR",
				Trait:    "Height",
				Model:    "test_model",
			},
			wantErr: true,
		},
		{
			name: "min greater than max",
			stats: ReferenceStats{
				Mean:     0.5,
				Std:      1.0,
				Min:      1.0,
				Max:      0.0,
				Ancestry: "EUR",
				Trait:    "Height",
				Model:    "test_model",
			},
			wantErr: true,
		},
		{
			name: "mean below min",
			stats: ReferenceStats{
				Mean:     -0.1,
				Std:      1.0,
				Min:      0.0,
				Max:      1.0,
				Ancestry: "EUR",
				Trait:    "Height",
				Model:    "test_model",
			},
			wantErr: true,
		},
		{
			name: "mean above max",
			stats: ReferenceStats{
				Mean:     1.1,
				Std:      1.0,
				Min:      0.0,
				Max:      1.0,
				Ancestry: "EUR",
				Trait:    "Height",
				Model:    "test_model",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stats.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReferenceStats_NormalizePRS(t *testing.T) {
	validStats := ReferenceStats{
		Mean:     0.5,
		Std:      1.0,
		Min:      0.0,
		Max:      1.0,
		Ancestry: "EUR",
		Trait:    "Height",
		Model:    "test_model",
	}

	tests := []struct {
		name      string
		stats     ReferenceStats
		rawPRS    float64
		wantErr   bool
		checkFunc func(t *testing.T, percentile float64)
	}{
		{
			name:    "mean value",
			stats:   validStats,
			rawPRS:  0.5,
			wantErr: false,
			checkFunc: func(t *testing.T, percentile float64) {
				assert.InDelta(t, 0.5, percentile, 0.01)
			},
		},
		{
			name:    "one std dev above mean",
			stats:   validStats,
			rawPRS:  1.5,
			wantErr: false,
			checkFunc: func(t *testing.T, percentile float64) {
				assert.InDelta(t, 0.841, percentile, 0.01)
			},
		},
		{
			name:    "one std dev below mean",
			stats:   validStats,
			rawPRS:  -0.5,
			wantErr: false,
			checkFunc: func(t *testing.T, percentile float64) {
				assert.InDelta(t, 0.159, percentile, 0.01)
			},
		},
		{
			name:    "invalid stats",
			stats:   ReferenceStats{Std: -1.0},
			rawPRS:  0.5,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			percentile, err := tt.stats.NormalizePRS(tt.rawPRS)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.checkFunc(t, percentile)
			}
		})
	}
}

func TestCompute(t *testing.T) {
	tests := []struct {
		name        string
		alleleFreqs map[string]float64
		effectSizes map[string]float64
		wantErr     bool
		checkFunc   func(t *testing.T, stats *ReferenceStats)
	}{
		{
			name: "valid computation",
			alleleFreqs: map[string]float64{
				"rs1": 0.5,
				"rs2": 0.3,
				"rs3": 0.7,
			},
			effectSizes: map[string]float64{
				"rs1": 0.1,
				"rs2": -0.2,
				"rs3": 0.3,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, stats *ReferenceStats) {
				assert.Greater(t, stats.Std, 0.0)
				assert.LessOrEqual(t, stats.Min, stats.Max)
				assert.GreaterOrEqual(t, stats.Mean, stats.Min)
				assert.LessOrEqual(t, stats.Mean, stats.Max)
			},
		},
		{
			name:        "empty allele frequencies",
			alleleFreqs: map[string]float64{},
			effectSizes: map[string]float64{
				"rs1": 0.1,
			},
			wantErr: true,
		},
		{
			name: "empty effect sizes",
			alleleFreqs: map[string]float64{
				"rs1": 0.5,
			},
			effectSizes: map[string]float64{},
			wantErr:     true,
		},
		{
			name: "no matching variants",
			alleleFreqs: map[string]float64{
				"rs1": 0.5,
			},
			effectSizes: map[string]float64{
				"rs2": 0.1,
			},
			wantErr: true,
		},
		{
			name: "extreme values (zero variance)",
			alleleFreqs: map[string]float64{
				"rs1": 0.0,
				"rs2": 1.0,
			},
			effectSizes: map[string]float64{
				"rs1": -1.0,
				"rs2": 1.0,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, stats *ReferenceStats) {
				// With frequencies 0.0 and 1.0, variance should be 0 (no genetic variation)
				// This is mathematically correct under Hardy-Weinberg equilibrium
				assert.Equal(t, 0.0, stats.Std)
				assert.LessOrEqual(t, stats.Min, stats.Max)
				assert.GreaterOrEqual(t, stats.Mean, stats.Min)
				assert.LessOrEqual(t, stats.Mean, stats.Max)
				// Mean should be 2*0.0*(-1.0) + 2*1.0*1.0 = 2.0
				assert.Equal(t, 2.0, stats.Mean)
			},
		},
		{
			name: "non-zero variance case",
			alleleFreqs: map[string]float64{
				"rs1": 0.3,
				"rs2": 0.6,
			},
			effectSizes: map[string]float64{
				"rs1": 0.5,
				"rs2": -0.4,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, stats *ReferenceStats) {
				// With intermediate frequencies, variance should be > 0
				assert.Greater(t, stats.Std, 0.0)
				assert.LessOrEqual(t, stats.Min, stats.Max)
				assert.GreaterOrEqual(t, stats.Mean, stats.Min)
				assert.LessOrEqual(t, stats.Mean, stats.Max)
				// Expected values with corrected formulas:
				// Mean = 2*(0.3*0.5 + 0.6*(-0.4)) = 2*(0.15 - 0.24) = -0.18
				// Variance = 2*(0.3*0.7*0.25 + 0.6*0.4*0.16) = 2*(0.0525 + 0.0384) = 0.1818
				assert.InDelta(t, -0.18, stats.Mean, 1e-10)
				assert.InDelta(t, 0.4264, stats.Std, 1e-4) // sqrt(0.1818)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := Compute(tt.alleleFreqs, tt.effectSizes)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
				tt.checkFunc(t, stats)
			}
		})
	}
}
