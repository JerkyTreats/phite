package reference

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	bigqueryclientset "phite.io/polygenic-risk-calculator/internal/clientsets/bigquery"
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

// ReferenceStatsLoader implements ReferenceStatsBackend using a BigQuery clientset.
type ReferenceStatsLoader struct {
	client *bigqueryclientset.Client
}

// NewReferenceStatsLoader returns a new ReferenceStatsLoader using the provided clientset.
func NewReferenceStatsLoader(client *bigqueryclientset.Client) *ReferenceStatsLoader {
	return &ReferenceStatsLoader{client: client}
}

// GetReferenceStats queries BigQuery for reference stats for the given ancestry, trait, and model.
func (r *ReferenceStatsLoader) GetReferenceStats(ctx context.Context, ancestry, trait, model string) (*ReferenceStats, error) {
	logging.Info("Querying reference stats from BigQuery: ancestry=%s, trait=%s, model=%s", ancestry, trait, model)

	sqlQuery := fmt.Sprintf("SELECT mean, std, min, max, ancestry, trait, model FROM `%s.%s.%s` WHERE ancestry = @ancestry AND trait = @trait AND model = @model LIMIT 1",
		r.client.ProjectID, r.client.Dataset, r.client.Table)
	logging.Debug("Executing BigQuery SQL: %s", sqlQuery)
	q := r.client.BigQuery().Query(sqlQuery)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "ancestry", Value: ancestry},
		{Name: "trait", Value: trait},
		{Name: "model", Value: model},
	}
	it, err := q.Read(ctx)
	if err != nil {
		logging.Error("BigQuery query failed for ancestry=%s, trait=%s, model=%s: %v", ancestry, trait, model, err)
		return nil, fmt.Errorf("BigQuery query failed: %w", err)
	}
	var stats ReferenceStats
	if err := it.Next(&stats); err != nil {
		if err.Error() == "iterator.Done" {
			logging.Info("No reference stats found for ancestry=%s, trait=%s, model=%s", ancestry, trait, model)
			return nil, nil // No matching stats found
		}
		logging.Error("Failed to scan BigQuery result for ancestry=%s, trait=%s, model=%s: %v", ancestry, trait, model, err)
		return nil, fmt.Errorf("failed to scan result: %w", err)
	}
	logging.Info("Loaded reference stats for ancestry=%s, trait=%s, model=%s: mean=%.3f, std=%.3f", ancestry, trait, model, stats.Mean, stats.Std)
	return &stats, nil
}

// Close releases resources held by the clientset.
func (r *ReferenceStatsLoader) Close() error {
	logging.Info("Closing ReferenceStatsLoader BigQuery client")
	return r.client.Close()
}
