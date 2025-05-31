package gwas

import (
	"phite.io/polygenic-risk-calculator/internal/logging"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/dbutil"
	"phite.io/polygenic-risk-calculator/internal/model"
)

// FetchGWASRecords loads GWAS SNP records for the given rsids from DuckDB.
// Returns a map from rsid to model.GWASSNPRecord.
func FetchGWASRecords(dbPath string, rsids []string) (map[string]model.GWASSNPRecord, error) {
	logging.Info("Opening DuckDB at %s", dbPath)
	if len(rsids) == 0 {
		return map[string]model.GWASSNPRecord{}, nil
	}
	db, err := dbutil.OpenDuckDB(dbPath)
	if err != nil {
		logging.Error("failed to open DuckDB: %v", err)
		return nil, err
	}
	defer db.Close()

	// Validate GWAS table exists
	err = dbutil.ValidateTable(db, "gwas", []string{"rsid", "risk_allele", "beta", "trait"})
	if err != nil {
		logging.Error("GWAS table validation failed: %v", err)
		return nil, err
	}

	// Prepare query with IN clause
	placeholders := make([]string, len(rsids))
	args := make([]interface{}, len(rsids))
	for i, rsid := range rsids {
		placeholders[i] = "?"
		args[i] = rsid
	}
	query := `SELECT rsid, risk_allele, beta, trait FROM gwas WHERE rsid IN (` + strings.Join(placeholders, ",") + ")"
	logging.Info("Executing GWAS query for %d rsids", len(rsids))
	rows, err := db.Query(query, args...)
	if err != nil {
		logging.Error("GWAS query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var records = make(map[string]model.GWASSNPRecord)
	for rows.Next() {
		var rec model.GWASSNPRecord
		if err := rows.Scan(&rec.RSID, &rec.RiskAllele, &rec.Beta, &rec.Trait); err != nil {
			logging.Error("row scan failed: %v", err)
			return nil, err
		}
		records[rec.RSID] = rec
	}
	if err := rows.Err(); err != nil {
		logging.Error("row iteration failed: %v", err)
		return nil, err
	}
	logging.Info("Loaded %d GWAS records from DuckDB", len(records))
	return records, nil
}
