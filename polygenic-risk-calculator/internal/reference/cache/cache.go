package reference_cache

import (
	"context"
	"fmt"

	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
	"phite.io/polygenic-risk-calculator/internal/logging"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

// ReferenceStatsBackend defines the interface for any backend that can provide reference stats.
type ReferenceStatsBackend interface {
	GetReferenceStats(ctx context.Context, ancestry, trait, model string) (*reference_stats.ReferenceStats, error)
	Close() error
}

// StatsRequest represents a request for reference statistics.
type StatsRequest struct {
	Ancestry string
	Trait    string
	ModelID  string
}

// Cache defines the interface for storing and retrieving reference statistics.
type Cache interface {
	Get(ctx context.Context, req StatsRequest) (*reference_stats.ReferenceStats, error)
	Store(ctx context.Context, req StatsRequest, stats *reference_stats.ReferenceStats) error
}

// RepositoryCache implements both Cache and ReferenceStatsBackend using DBRepository.
type RepositoryCache struct {
	repo    dbinterface.Repository
	tableID string
}

// NewRepositoryCache creates a new cache using DBRepository.
func NewRepositoryCache(repo dbinterface.Repository, tableID string) *RepositoryCache {
	return &RepositoryCache{
		repo:    repo,
		tableID: tableID,
	}
}

// GetReferenceStats implements ReferenceStatsBackend interface.
func (c *RepositoryCache) GetReferenceStats(ctx context.Context, ancestry, trait, model string) (*reference_stats.ReferenceStats, error) {
	return c.Get(ctx, StatsRequest{
		Ancestry: ancestry,
		Trait:    trait,
		ModelID:  model,
	})
}

// Get retrieves reference statistics from the repository.
func (c *RepositoryCache) Get(ctx context.Context, req StatsRequest) (*reference_stats.ReferenceStats, error) {
	queryString := fmt.Sprintf(
		"SELECT mean, std, min, max, ancestry, trait, model FROM %s WHERE ancestry = ? AND trait = ? AND model = ? LIMIT 1",
		c.tableID,
	)

	logging.Debug("Executing cache query: %s with params: ancestry=%s, trait=%s, modelID=%s",
		queryString, req.Ancestry, req.Trait, req.ModelID)

	results, err := c.repo.Query(ctx, queryString, req.Ancestry, req.Trait, req.ModelID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cache query: %w", err)
	}

	if len(results) == 0 {
		return nil, nil // Cache miss
	}

	if len(results) > 1 {
		return nil, fmt.Errorf("multiple rows found in cache for ancestry=%s, trait=%s, modelID=%s",
			req.Ancestry, req.Trait, req.ModelID)
	}

	row := results[0]
	stats := &reference_stats.ReferenceStats{
		Mean:     row["mean"].(float64),
		Std:      row["std"].(float64),
		Min:      row["min"].(float64),
		Max:      row["max"].(float64),
		Ancestry: row["ancestry"].(string),
		Trait:    row["trait"].(string),
		Model:    row["model"].(string),
	}

	// Validate the stats before returning
	if err := stats.Validate(); err != nil {
		return nil, fmt.Errorf("invalid reference stats from cache: %w", err)
	}

	return stats, nil
}

// Store saves reference statistics to the repository.
func (c *RepositoryCache) Store(ctx context.Context, req StatsRequest, stats *reference_stats.ReferenceStats) error {
	// Validate stats before storing
	if err := stats.Validate(); err != nil {
		return fmt.Errorf("invalid reference stats for storage: %w", err)
	}

	row := map[string]interface{}{
		"mean":     stats.Mean,
		"std":      stats.Std,
		"min":      stats.Min,
		"max":      stats.Max,
		"ancestry": req.Ancestry,
		"trait":    req.Trait,
		"model":    req.ModelID,
	}

	if err := c.repo.Insert(ctx, c.tableID, []map[string]interface{}{row}); err != nil {
		return fmt.Errorf("failed to store stats in cache: %w", err)
	}

	return nil
}

// Close implements ReferenceStatsBackend interface.
func (c *RepositoryCache) Close() error {
	logging.Info("Closing RepositoryCache")
	return nil // No resources to clean up with DBRepository
}
