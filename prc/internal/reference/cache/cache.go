package reference_cache

import (
	"context"
	"fmt"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/ancestry"
	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/db"
	dbinterface "phite.io/polygenic-risk-calculator/internal/db/interface"
	"phite.io/polygenic-risk-calculator/internal/logging"
	reference_stats "phite.io/polygenic-risk-calculator/internal/reference/stats"
)

// Domain-specific configuration keys for cache
const (
	BatchSizeKey = "cache.batch_size" // Cache batch operation size
)

func init() {
	// Register required infrastructure keys (table reference)
	config.RegisterRequiredKey(config.TableCacheTableKey)
	// Domain-specific keys with defaults don't need to be required
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

// CacheEntry represents a cache entry for batch operations.
type CacheEntry struct {
	Request StatsRequest
	Stats   *reference_stats.ReferenceStats
}

// Cache defines the interface for storing and retrieving reference statistics.
type Cache interface {
	Get(ctx context.Context, req StatsRequest) (*reference_stats.ReferenceStats, error)
	Store(ctx context.Context, req StatsRequest, stats *reference_stats.ReferenceStats) error
	// Batch operations
	GetBatch(ctx context.Context, reqs []StatsRequest) (map[string]*reference_stats.ReferenceStats, error)
	StoreBatch(ctx context.Context, entries []CacheEntry) error
}

// RepositoryCache implements both Cache and ReferenceStatsBackend using DBRepository.
type RepositoryCache struct {
	Repo      dbinterface.Repository
	TableID   string
	datasetID string // Add dataset ID to build fully qualified table names
	projectID string // Add project ID for full qualification
}

// NewRepositoryCache creates a new cache with dependency injection
// If repo is nil, it will be created using provided params or default configuration
func NewRepositoryCache(repo dbinterface.Repository, params ...map[string]string) (*RepositoryCache, error) {
	var err error
	var datasetID, projectID string

	// Create repository if not provided
	if repo == nil {
		if len(params) > 0 && params[0] != nil {
			// Use provided parameters
			repo, err = db.GetRepository(context.Background(), "bq", params[0])
			datasetID = params[0]["dataset_id"] // Capture dataset ID from params
			projectID = params[0]["project_id"] // Capture project ID from params
		} else {
			// Use default configuration
			repo, err = db.GetRepository(context.Background(), "bq")
			datasetID = config.GetString(config.BigQueryCacheDatasetKey) // Use config fallback
			projectID = config.GetString(config.GCPCacheProjectKey)      // Use config fallback
		}
		if err != nil {
			logging.Error("Failed to create RepositoryCache: %v", err)
			return nil, fmt.Errorf("failed to create RepositoryCache: %w", err)
		}
	} else {
		// If repository is provided, try to get dataset from params or config
		if len(params) > 0 && params[0] != nil {
			if params[0]["dataset_id"] != "" {
				datasetID = params[0]["dataset_id"]
			}
			if params[0]["project_id"] != "" {
				projectID = params[0]["project_id"]
			}
		}
		// Fall back to config if not provided in params
		if datasetID == "" {
			datasetID = config.GetString(config.BigQueryCacheDatasetKey)
		}
		if projectID == "" {
			projectID = config.GetString(config.GCPCacheProjectKey)
		}
	}

	return &RepositoryCache{
		Repo:      repo,
		TableID:   config.GetString(config.TableCacheTableKey),
		datasetID: datasetID,
		projectID: projectID,
	}, nil
}

// GetFullyQualifiedTableName returns the properly qualified table name (`project.dataset.table`)
// with backticks to prevent SQL injection and parsing issues.
// Returns an error if any component is missing.
func (c *RepositoryCache) GetFullyQualifiedTableName() (string, error) {
	if c.projectID == "" {
		return "", fmt.Errorf("project ID is required for BigQuery cache operations, got empty value")
	}
	if c.datasetID == "" {
		return "", fmt.Errorf("dataset ID is required for BigQuery cache operations, got empty value")
	}
	if c.TableID == "" {
		return "", fmt.Errorf("table ID is required for BigQuery cache operations, got empty value")
	}

	fqTable := fmt.Sprintf("`%s`.`%s`.`%s`", c.projectID, c.datasetID, c.TableID)
	logging.Debug("Fully qualified table name: %s", fqTable)
	return fqTable, nil
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
	fullyQualifiedTable, err := c.GetFullyQualifiedTableName()
	if err != nil {
		return nil, fmt.Errorf("failed to build fully qualified table name: %w", err)
	}

	queryString := fmt.Sprintf(
		"SELECT mean, std, min, max, ancestry, trait, model FROM %s WHERE ancestry = ? AND trait = ? AND model = ? LIMIT 1",
		fullyQualifiedTable,
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

// GetBatch retrieves multiple reference statistics from the repository in a single query.
func (c *RepositoryCache) GetBatch(ctx context.Context, reqs []StatsRequest) (map[string]*reference_stats.ReferenceStats, error) {
	logging.Info("Getting batch of %d reference stats", len(reqs))
	if len(reqs) == 0 {
		return make(map[string]*reference_stats.ReferenceStats), nil
	}

	fullyQualifiedTable, err := c.GetFullyQualifiedTableName()
	logging.Debug("Fully qualified table name: %s", fullyQualifiedTable)
	if err != nil {
		return nil, fmt.Errorf("failed to build fully qualified table name: %w", err)
	}

	// Build batch query with OR clause for optimal performance
	var conditions []string
	var args []interface{}

	for _, req := range reqs {
		conditions = append(conditions, "(ancestry = ? AND trait = ? AND model = ?)")
		args = append(args, req.Ancestry, req.Trait, req.ModelID)
	}

	queryString := fmt.Sprintf(
		"SELECT mean, std, min, max, ancestry, trait, model FROM %s WHERE %s",
		fullyQualifiedTable,
		strings.Join(conditions, " OR "),
	)

	logging.Debug("Executing batch cache query for %d requests", len(reqs))

	results, err := c.Repo.Query(ctx, queryString, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch cache query: %w", err)
	}

	// Convert results to map keyed by "ancestry|trait|model"
	statsMap := make(map[string]*reference_stats.ReferenceStats)
	for _, row := range results {
		stats := &reference_stats.ReferenceStats{
			Mean:     row["mean"].(float64),
			Std:      row["std"].(float64),
			Min:      row["min"].(float64),
			Max:      row["max"].(float64),
			Ancestry: row["ancestry"].(string),
			Trait:    row["trait"].(string),
			Model:    row["model"].(string),
		}

		// Validate the stats before adding to map
		if err := stats.Validate(); err != nil {
			logging.Warn("Invalid reference stats from batch cache query: %v", err)
			continue
		}

		key := fmt.Sprintf("%s|%s|%s", stats.Ancestry, stats.Trait, stats.Model)
		statsMap[key] = stats
	}

	logging.Debug("Retrieved %d stats from batch cache query", len(statsMap))
	return statsMap, nil
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

// StoreBatch stores multiple reference statistics to the repository in a single operation.
func (c *RepositoryCache) StoreBatch(ctx context.Context, entries []CacheEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Get batch size from config (default to 100 if not set)
	batchSize := config.GetInt(BatchSizeKey)
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	// Process entries in batches for optimal BigQuery performance
	for i := 0; i < len(entries); i += batchSize {
		end := i + batchSize
		if end > len(entries) {
			end = len(entries)
		}

		batch := entries[i:end]
		if err := c.storeBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to store batch %d-%d: %w", i, end-1, err)
		}
	}

	return nil
}

// storeBatch performs the actual batch storage operation.
func (c *RepositoryCache) storeBatch(ctx context.Context, entries []CacheEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Prepare rows for batch insert
	rows := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		// Validate stats before storing
		if err := entry.Stats.Validate(); err != nil {
			return fmt.Errorf("invalid reference stats for batch storage: %w", err)
		}

		row := map[string]interface{}{
			"mean":     entry.Stats.Mean,
			"std":      entry.Stats.Std,
			"min":      entry.Stats.Min,
			"max":      entry.Stats.Max,
			"ancestry": entry.Request.Ancestry,
			"trait":    entry.Request.Trait,
			"model":    entry.Request.ModelID,
		}
		rows = append(rows, row)
	}

	// Execute batch insert
	if err := c.Repo.Insert(ctx, c.TableID, rows); err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	logging.Debug("Stored %d stats in batch cache operation", len(entries))
	return nil
}
