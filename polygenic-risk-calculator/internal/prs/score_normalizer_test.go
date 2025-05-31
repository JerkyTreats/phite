package prs

import "phite.io/polygenic-risk-calculator/internal/model"

import (
	"testing"
	"math"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

func TestNormalizePRS(t *testing.T) {
	logging.SetSilentLoggingForTest()
	tests := []struct {
		name           string
		prsScore       float64
		ref            model.ReferenceStats
		expectsError   bool
		expectedZ      float64
		expectedPct    float64
	}{
		{
			name:         "PRS at mean",
			prsScore:     1.0,
			ref:          model.ReferenceStats{Mean: 1.0, Std: 0.5, Min: 0.0, Max: 2.0},
			expectedZ:    0.0,
			expectedPct:  50.0,
		},
		{
			name:         "PRS at min",
			prsScore:     0.0,
			ref:          model.ReferenceStats{Mean: 1.0, Std: 0.5, Min: 0.0, Max: 2.0},
			expectedZ:    -2.0,
			expectedPct:  2.28, // approx percentile for z=-2
		},
		{
			name:         "PRS at max",
			prsScore:     2.0,
			ref:          model.ReferenceStats{Mean: 1.0, Std: 0.5, Min: 0.0, Max: 2.0},
			expectedZ:    2.0,
			expectedPct:  97.72, // approx percentile for z=2
		},
		{
			name:         "Malformed stats (zero std)",
			prsScore:     1.0,
			ref:          model.ReferenceStats{Mean: 1.0, Std: 0.0, Min: 0.0, Max: 2.0},
			expectsError: true,
		},
		{
			name:         "Malformed stats (missing mean)",
			prsScore:     1.0,
			ref:          model.ReferenceStats{Std: 0.5, Min: 0.0, Max: 2.0},
			expectsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prs := PRSResult{PRSScore: tt.prsScore}
			norm, err := NormalizePRS(prs, tt.ref)
			if tt.expectsError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if math.Abs(norm.ZScore-tt.expectedZ) > 0.01 {
				t.Errorf("z-score: got %v, want %v", norm.ZScore, tt.expectedZ)
			}
			if math.Abs(norm.Percentile-tt.expectedPct) > 0.1 {
				t.Errorf("percentile: got %v, want %v", norm.Percentile, tt.expectedPct)
			}
		})
	}
}
