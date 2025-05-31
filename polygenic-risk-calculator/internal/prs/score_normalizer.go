package prs

import (
	"errors"
	"math"
)

// NormalizedPRS represents the normalized PRS result.
type NormalizedPRS struct {
	RawScore   float64 `json:"raw_score"`
	ZScore     float64 `json:"z_score"`
	Percentile float64 `json:"percentile"`
}

// referenceStats holds reference population statistics for normalization.
type referenceStats struct {
	Mean float64
	Std  float64
	Min  float64
	Max  float64
}

// NormalizePRS normalizes a raw PRS score using reference stats.
// Returns NormalizedPRS and error if stats are missing or malformed.
func NormalizePRS(prs PRSResult, ref referenceStats) (NormalizedPRS, error) {
	if ref.Std == 0 || ref.Mean == 0 || math.IsNaN(ref.Mean) || math.IsNaN(ref.Std) {
		return NormalizedPRS{}, errors.New("invalid reference stats: mean and std must be nonzero and present")
	}
	z := (prs.PRSScore - ref.Mean) / ref.Std
	percentile := 100 * normCdf(z)
	return NormalizedPRS{
		RawScore:   prs.PRSScore,
		ZScore:     z,
		Percentile: percentile,
	}, nil
}

// normCdf returns the cumulative distribution function for the standard normal distribution.
func normCdf(z float64) float64 {
	return 0.5 * (1 + math.Erf(z/math.Sqrt2))
}
