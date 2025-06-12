package reference

import (
	"context"
	"fmt"

	"phite.io/polygenic-risk-calculator/internal/db"
	"phite.io/polygenic-risk-calculator/internal/logging"
)

// ReferenceStatsBackend defines the interface for any backend that can provide reference stats.
type ReferenceStatsBackend interface {
	GetReferenceStats(ctx context.Context, ancestry, trait, model string) (*ReferenceStats, error)
	Close() error
}

// ReferenceStats matches the canonical struct used throughout the pipeline.
type ReferenceStats struct {
	Mean     float64
	Std      float64
	Min      float64
	Max      float64
	Ancestry string
	Trait    string
	Model    string
}

// ReferenceStatsLoader implements ReferenceStatsBackend using a DBRepository.
type ReferenceStatsLoader struct {
	repo db.DBRepository
}

// NewReferenceStatsLoader returns a new ReferenceStatsLoader using the provided repository.
func NewReferenceStatsLoader(repo db.DBRepository) *ReferenceStatsLoader {
	return &ReferenceStatsLoader{repo: repo}
}

// GetReferenceStats queries the database for reference stats for the given ancestry, trait, and model.
func (r *ReferenceStatsLoader) GetReferenceStats(ctx context.Context, ancestry, trait, model string) (*ReferenceStats, error) {
	logging.Info("Querying reference stats: ancestry=%s, trait=%s, model=%s", ancestry, trait, model)

	query := "SELECT mean, std, min, max, ancestry, trait, model FROM reference_panel WHERE ancestry = ? AND trait = ? AND model = ? LIMIT 1"
	results, err := r.repo.Query(ctx, query, ancestry, trait, model)
	if err != nil {
		logging.Error("Query failed for ancestry=%s, trait=%s, model=%s: %v", ancestry, trait, model, err)
		return nil, fmt.Errorf("query failed: %w", err)
	}

	if len(results) == 0 {
		logging.Info("No reference stats found for ancestry=%s, trait=%s, model=%s", ancestry, trait, model)
		return nil, nil // No matching stats found
	}

	record := results[0]
	stats := &ReferenceStats{
		Mean:     record["mean"].(float64),
		Std:      record["std"].(float64),
		Min:      record["min"].(float64),
		Max:      record["max"].(float64),
		Ancestry: record["ancestry"].(string),
		Trait:    record["trait"].(string),
		Model:    record["model"].(string),
	}

	logging.Info("Loaded reference stats for ancestry=%s, trait=%s, model=%s: mean=%.3f, std=%.3f", ancestry, trait, model, stats.Mean, stats.Std)
	return stats, nil
}

// Close releases resources held by the repository.
func (r *ReferenceStatsLoader) Close() error {
	logging.Info("Closing ReferenceStatsLoader")
	return nil // No resources to clean up with DBRepository
}
