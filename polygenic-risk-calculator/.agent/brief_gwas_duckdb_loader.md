# Agent Brief: GWAS DuckDB Loader

> **Note:** The DuckDB database used by this loader is provided by the external shared GWAS database (`../gwas/gwas.duckdb`), which is managed separately. For data engineering and schema creation, see briefs in `gwas/.agent/`.

## Purpose
Provide an interface for efficiently loading GWAS association records from a DuckDB database for use in the polygenic risk calculator pipeline.

## Responsibilities
- Use DuckDB Shared Utilities for connection/session management and schema validation.
- Connect to the DuckDB database at the configured path.
- Query the configured table for GWAS association records for the requested SNPs.
- Validate that the table exists and contains required columns (`rsid`, `risk_allele`, `beta`, `trait`) before querying.
- Return records as Go structs for downstream use.
- Handle errors gracefully and provide clear, actionable error messages (including if the table or columns are missing).
- Support extensibility for additional GWAS fields if needed.
- Support efficient batch queries for large SNP lists.

## Inputs
- DuckDB file path (configurable: CLI argument `--gwas-db`, env `GWAS_DUCKDB`, config key `gwas_db_path`, defaults to `gwas/gwas.duckdb`)
- Table name (configurable: env `GWAS_TABLE`, config key `gwas_table`, defaults to `associations_clean`)
- List of SNP rsids to fetch

## Outputs
- `[]GWASSNPRecord` and/or `map[string]GWASSNPRecord` for use in genotype validation and GWAS annotation

## Consumed By
- Genotype Input Handler (for SNP validation)
- GWAS Data Fetcher (for annotation)

## Required Tests
- Loads GWAS records for valid rsids from DuckDB.
- Handles missing rsids gracefully (returns empty or partial results as appropriate).
- Fails gracefully with clear errors on missing or malformed DuckDB files.
- Returns correct Go structs for downstream use.
- Efficiently handles large SNP lists (batch queries).
- Returns clear, actionable errors if the configured table does not exist or is missing required columns.

## Example Usage

```go
records, err := gwasduckdb.FetchGWASRecords(dbPath, rsidList)
if err != nil {
    // handle error
}
// records is []GWASSNPRecord or map[string]GWASSNPRecord
```
