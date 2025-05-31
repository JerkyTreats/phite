package reference

import (
	"fmt"
	"strings"

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
	db, err := dbutil.OpenDuckDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	defer db.Close()

	err = dbutil.ValidateTable(db, "reference_stats", []string{"mean", "std", "min", "max", "ancestry", "trait", "model"})
	if err != nil {
		return nil, fmt.Errorf("reference_stats table validation failed: %w", err)
	}

	query := `SELECT mean, std, min, max, ancestry, trait, model FROM reference_stats WHERE ancestry = ? AND trait = ? AND model = ? LIMIT 1`
	row := db.QueryRow(query, ancestry, trait, model)

	var stats ReferenceStats
	err = row.Scan(&stats.Mean, &stats.Std, &stats.Min, &stats.Max, &stats.Ancestry, &stats.Trait, &stats.Model)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil // No matching stats found
		}
		return nil, fmt.Errorf("failed to scan stats: %w", err)
	}
	return &stats, nil
}
