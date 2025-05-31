package gwas

import (
	"fmt"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/dbutil"
	"phite.io/polygenic-risk-calculator/internal/model"
)

// FetchGWASRecords loads GWAS SNP records for the given rsids from DuckDB.
// Returns a map from rsid to model.GWASSNPRecord.
func FetchGWASRecords(dbPath string, rsids []string) (map[string]model.GWASSNPRecord, error) {
	if len(rsids) == 0 {
		return map[string]model.GWASSNPRecord{}, nil
	}
	db, err := dbutil.OpenDuckDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	defer db.Close()

	// Validate GWAS table exists
	err = dbutil.ValidateTable(db, "gwas", []string{"rsid", "risk_allele", "beta", "trait"})
	if err != nil {
		return nil, fmt.Errorf("GWAS table validation failed: %w", err)
	}

	// Prepare query with IN clause
	placeholders := make([]string, len(rsids))
	args := make([]interface{}, len(rsids))
	for i, rsid := range rsids {
		placeholders[i] = "?"
		args[i] = rsid
	}
	query := `SELECT rsid, risk_allele, beta, trait FROM gwas WHERE rsid IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var records = make(map[string]model.GWASSNPRecord)
	for rows.Next() {
		var rec model.GWASSNPRecord
		if err := rows.Scan(&rec.RSID, &rec.RiskAllele, &rec.Beta, &rec.Trait); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		records[rec.RSID] = rec
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}
	return records, nil
}
