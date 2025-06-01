package reference

import (
	"strings"

	"phite.io/polygenic-risk-calculator/internal/logging"

	"phite.io/polygenic-risk-calculator/internal/dbutil"
)

type ReferenceStats struct {
	Mean     float64
	Std      float64
	Min      float64
	Max      float64
	Ancestry string
	Trait    string
	Model    string
}

// LoadReferenceStatsFromDuckDB loads reference stats for given ancestry, trait, and model from DuckDB.
func LoadReferenceStatsFromDuckDB(dbPath, ancestry, trait, model string) (*ReferenceStats, error) {
	logging.Debug("Loading reference stats for ancestry=%s, trait=%s, model=%s", ancestry, trait, model)
	logging.Info("Opening DuckDB at %s", dbPath)
	db, err := dbutil.OpenDuckDB(dbPath)
	if err != nil {
		logging.Error("failed to open DuckDB: %v", err)
		return nil, err
	}
	defer db.Close()

	err = dbutil.ValidateTable(db, "reference_stats", []string{"mean", "std", "min", "max", "ancestry", "trait", "model"})
	if err != nil {
		logging.Error("reference_stats table validation failed: %v", err)
		return nil, err
	}

	query := `SELECT mean, std, min, max, ancestry, trait, model FROM reference_stats WHERE ancestry = ? AND trait = ? AND model = ? LIMIT 1`
	row := db.QueryRow(query, ancestry, trait, model)

	var stats ReferenceStats
	err = row.Scan(&stats.Mean, &stats.Std, &stats.Min, &stats.Max, &stats.Ancestry, &stats.Trait, &stats.Model)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			logging.Info("No reference stats found for ancestry=%s, trait=%s, model=%s", ancestry, trait, model)
			return nil, nil // No matching stats found
		}
		logging.Error("failed to scan stats: %v", err)
		return nil, err
	}
	logging.Info("Loaded reference stats: ancestry=%s, trait=%s, model=%s", stats.Ancestry, stats.Trait, stats.Model)
	return &stats, nil
}

// LoadDefaultReferenceStats loads reference stats for the default ancestry, trait, and model.
// Defaults: ancestry="EUR", trait="height", model="v1"
func LoadDefaultReferenceStats(dbPath string) (*ReferenceStats, error) {
	const (
		defaultAncestry = "EUR"
		defaultTrait    = "height"
		defaultModel    = "v1"
	)
	return LoadReferenceStatsFromDuckDB(dbPath, defaultAncestry, defaultTrait, defaultModel)
}
