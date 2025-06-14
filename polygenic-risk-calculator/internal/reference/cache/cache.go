package reference_cache

import (
	"context"
	"fmt"

	"phite.io/polygenic-risk-calculator/internal/ancestry"
	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/db"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
	"phite.io/polygenic-risk-calculator/internal/logging"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

const (
	// TableIDKey is the configuration key for the reference stats table ID
	TableIDKey = "reference_stats.table_id"
)

func init() {
	// Register required configuration keys
	config.RegisterRequiredKey(TableIDKey)
}

// ReferenceStatsBackend defines the interface for any backend that can provide reference stats.
type ReferenceStatsBackend interface {
	GetReferenceStats(ctx context.Context, ancestry *ancestry.Ancestry, trait, model string) (*reference_stats.ReferenceStats, error)
	Close() error
}

// StatsRequest represents a request for reference statistics.
type StatsRequest struct {
	Ancestry string // Still use string internally for cache storage
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
	Repo    dbinterface.Repository
	TableID string
}

// NewRepositoryCache creates a new cache using DBRepository.
// params is optional - if provided, will be passed to the repository constructor
func NewRepositoryCache(params ...map[string]string) (*RepositoryCache, error) {
	var repo dbinterface.Repository
	var err error

	if len(params) > 0 && params[0] != nil {
		// Use provided parameters
		repo, err = db.GetRepository(context.Background(), "bq", params[0])
	} else {
		// Use default configuration
		repo, err = db.GetRepository(context.Background(), "bq")
	}

	if err != nil {
		logging.Error("Failed to create RepositoryCache: %v", err)
		return nil, fmt.Errorf("failed to create RepositoryCache: %w", err)
	}
	return &RepositoryCache{
		Repo:    repo,
		TableID: config.GetString(TableIDKey),
	}, nil
}

// GetReferenceStats implements ReferenceStatsBackend interface with ancestry objects.
func (c *RepositoryCache) GetReferenceStats(ctx context.Context, ancestry *ancestry.Ancestry, trait, model string) (*reference_stats.ReferenceStats, error) {
	// Convert ancestry object to code for cache operations
	ancestryCode := ancestry.Code()
	return c.Get(ctx, StatsRequest{
		Ancestry: ancestryCode,
		Trait:    trait,
		ModelID:  model,
	})
}

// Get retrieves reference statistics from the repository.
func (c *RepositoryCache) Get(ctx context.Context, req StatsRequest) (*reference_stats.ReferenceStats, error) {
	queryString := fmt.Sprintf(
		"SELECT mean, std, min, max, ancestry, trait, model FROM %s WHERE ancestry = ? AND trait = ? AND model = ? LIMIT 1",
		c.TableID,
	)

	logging.Debug("Executing cache query: %s with params: ancestry=%s, trait=%s, modelID=%s",
		queryString, req.Ancestry, req.Trait, req.ModelID)

	results, err := c.Repo.Query(ctx, queryString, req.Ancestry, req.Trait, req.ModelID)
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

	if err := c.Repo.Insert(ctx, c.TableID, []map[string]interface{}{row}); err != nil {
		return fmt.Errorf("failed to store stats in cache: %w", err)
	}

	logging.Debug("Stored stats in cache for ancestry=%s, trait=%s, model=%s", req.Ancestry, req.Trait, req.ModelID)
	return nil
}
