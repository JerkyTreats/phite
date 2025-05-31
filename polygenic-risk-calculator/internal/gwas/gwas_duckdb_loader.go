package gwas

import (
	"os"
	"strings"

	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/logging"

	"phite.io/polygenic-risk-calculator/internal/dbutil"
	"phite.io/polygenic-risk-calculator/internal/model"
)

// FetchGWASRecords loads GWAS SNP records for the given rsids from DuckDB.
// The DuckDB file path and table name can be set via environment variables (GWAS_DUCKDB, GWAS_TABLE)
// or config keys (gwas_db_path, gwas_table). Returns a map from rsid to model.GWASSNPRecord.
// FetchGWASRecordsWithTable loads GWAS SNP records for the given rsids from DuckDB, using the specified table.
// The dbPath and table are required. Returns a map from rsid to model.GWASSNPRecord.
func FetchGWASRecordsWithTable(dbPath, table string, rsids []string) (map[string]model.GWASSNPRecord, error) {
	// Allow override from config/env
	if dbPath == "" {
		dbPath = os.Getenv("GWAS_DUCKDB")
		if dbPath == "" {
			dbPath = config.GetString("gwas_db_path")
			if dbPath == "" {
				dbPath = "gwas/gwas.duckdb" // project default
			}
		}
	}
	logging.Info("Opening GWAS database: %s (table: %s)", dbPath, table)
	if len(rsids) == 0 {
		return map[string]model.GWASSNPRecord{}, nil
	}
	db, err := dbutil.OpenDuckDB(dbPath)
	if err != nil {
		logging.Error("Failed to open GWAS DuckDB at %s: %v", dbPath, err)
		return nil, err
	}
	defer db.Close()

	// Validate GWAS table exists
	// Validate GWAS table exists
	err = dbutil.ValidateTable(db, table, []string{"rsid", "risk_allele", "beta", "trait"})
	if err != nil {
		logging.Error("GWAS table validation failed: table=%s, err=%v", table, err)
		return nil, err
	}

	// Prepare query with IN clause
	placeholders := make([]string, len(rsids))
	args := make([]interface{}, len(rsids))
	for i, rsid := range rsids {
		placeholders[i] = "?"
		args[i] = rsid
	}
	query := `SELECT rsid, risk_allele, beta, trait FROM ` + table + ` WHERE rsid IN (` + strings.Join(placeholders, ",") + ")"
	logging.Info("Executing GWAS query for %d SNPs", len(rsids))
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
			logging.Error("Failed to scan GWAS row: %v", err)
			return nil, err
		}
		records[rec.RSID] = rec
	}
	if err := rows.Err(); err != nil {
		logging.Error("row iteration failed: %v", err)
		return nil, err
	}
	logging.Info("Loaded %d GWAS records from DuckDB", len(records))
	logging.Info("Loaded %d GWAS records from DuckDB", len(records))
	return records, nil
}

// FetchGWASRecords loads GWAS SNP records for the given rsids from DuckDB.
// The DuckDB file path can be set via CLI, environment variables (GWAS_DUCKDB), or config keys (gwas_db_path).
// The table name can be set via CLI, env (GWAS_TABLE), or config (gwas_table). Returns a map from rsid to model.GWASSNPRecord.
func FetchGWASRecords(dbPath string, rsids []string) (map[string]model.GWASSNPRecord, error) {
	if dbPath == "" {
		dbPath = os.Getenv("GWAS_DUCKDB")
		if dbPath == "" {
			dbPath = config.GetString("gwas_db_path")
			if dbPath == "" {
				dbPath = "gwas/gwas.duckdb" // project default
			}
		}
	}
	table := os.Getenv("GWAS_TABLE")
	if table == "" {
		table = config.GetString("gwas_table")
		if table == "" {
			table = "associations_clean"
		}
	}
	return FetchGWASRecordsWithTable(dbPath, table, rsids)
}
