package gwas

import (
	"os"
	"phite.io/polygenic-risk-calculator/internal/config"
	"phite.io/polygenic-risk-calculator/internal/dbutil"
)

// ValidateGWASDBAndTable checks that the DuckDB file exists, can be opened, and contains the required table/columns.
func ValidateGWASDBAndTable(dbPath, table string) error {
	if dbPath == "" {
		dbPath = os.Getenv("GWAS_DUCKDB")
		if dbPath == "" {
			dbPath = config.GetString("gwas_db_path")
			if dbPath == "" {
				dbPath = "gwas/gwas.duckdb" // project default
			}
		}
	}
	db, err := dbutil.OpenDuckDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	if table == "" {
		table = os.Getenv("GWAS_TABLE")
		if table == "" {
			table = config.GetString("gwas_table")
			if table == "" {
				table = "associations_clean"
			}
		}
	}
	return dbutil.ValidateTable(db, table, []string{"rsid", "risk_allele", "beta", "trait"})
}
