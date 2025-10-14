package prs

import (
	"errors"
	"math"

	"phite.io/polygenic-risk-calculator/internal/logging"

	"phite.io/polygenic-risk-calculator/internal/model"
)

// NormalizedPRS represents the normalized PRS result.
type NormalizedPRS struct {
	RawScore   float64 `json:"raw_score"`
	ZScore     float64 `json:"z_score"`
	Percentile float64 `json:"percentile"`
}

// ReferenceStats holds reference population statistics for normalization.

// NormalizePRS normalizes a raw PRS score using reference stats.
// Returns NormalizedPRS and error if stats are missing or malformed.
func NormalizePRS(prs PRSResult, ref model.ReferenceStats) (NormalizedPRS, error) {
	logging.Info("Normalizing PRS score: raw=%v, ref_mean=%v, ref_std=%v", prs.PRSScore, ref.Mean, ref.Std)
	if ref.Std == 0 || math.IsNaN(ref.Mean) || math.IsNaN(ref.Std) {
		logging.Error("invalid reference stats for normalization: mean=%v, std=%v", ref.Mean, ref.Std)
		return NormalizedPRS{}, errors.New("invalid reference stats: std must be nonzero and values must not be NaN")
	}
	z := (prs.PRSScore - ref.Mean) / ref.Std
	percentile := 100 * normCdf(z)
	result := NormalizedPRS{
		RawScore:   prs.PRSScore,
		ZScore:     z,
		Percentile: percentile,
	}
	logging.Info("PRS normalization complete: z=%.4f, percentile=%.2f", z, percentile)
	return result, nil
}

// normCdf returns the cumulative distribution function for the standard normal distribution.
func normCdf(z float64) float64 {
	return 0.5 * (1 + math.Erf(z/math.Sqrt2))
}
