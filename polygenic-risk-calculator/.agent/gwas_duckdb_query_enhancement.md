# Specification: Enhanced DuckDB Querying for PRS Model Loading and GWAS Data Access

**Version:** 1.0
**Date:** 2025-06-07

## 1. Objective

To implement a flexible and reusable system for querying DuckDB databases within the `polygenic-risk-calculator` project. This enhancement will primarily support the `PRSReferenceDataSource` in loading PRS model definitions directly from a DuckDB instance (e.g., the `associations_clean` table in `gwas.duckdb`) and aims to provide a generic querying capability for future needs involving DuckDB.

## 2. Background & Motivation

The `PRSReferenceDataSource` requires functionality to load PRS model variants (SNPID, Chromosome, Position, EffectAllele, OtherAllele, EffectWeight) when a cache miss occurs for pre-computed reference statistics. The source for these models is a DuckDB database containing GWAS association data.

Currently, `internal/gwas/gwas_duckdb_loader.go` provides specific functions to fetch GWAS records by a list of RSIDs. While useful, this is not directly applicable for loading a complete PRS model, which requires:
- Querying by a model identifier (e.g., `study_id`).
- Selecting a dynamic set of columns based on configuration.
- Mapping results to the `reference.PRSModelVariant` struct.

Implementing a more generic DuckDB query executor will prevent proliferation of specialized fetch functions, promote code reuse, and make the system more adaptable to future, varied querying requirements against DuckDB.

## 3. Proposed Solution

### 3.1. Generic DuckDB Query Executor

A new generic function, `ExecuteDuckDBQuery`, will be introduced, likely within the `internal/dbutil` package.

```go
// In internal/dbutil/dbutil.go (or new internal/dbutil/executor.go)
package dbutil

import (
	"database/sql"
	// ... other necessary imports (logging, fmt)
)

// RowScanner defines an interface for scanning a single sql.Row or sql.Rows.Next()
// into a target data structure.
type RowScanner[T any] func(rows *sql.Rows) (T, error)

// ExecuteDuckDBQuery provides a generic way to query a DuckDB database.
func ExecuteDuckDBQuery[T any](
	dbPath string,
	query string,
	args []interface{},
	scanner RowScanner[T],
) ([]T, error) {
	// Implementation will handle:
	// 1. Opening the DuckDB connection (using existing dbutil.OpenDuckDB).
	// 2. Executing the provided query with arguments.
	// 3. Iterating through rows and using the provided scanner to map each row.
	// 4. Closing the connection and handling errors.
}
```

### 3.2. `PRSReferenceDataSource.loadPRSModel` Integration

The `loadPRSModel` method in `internal/reference/prs_reference_data_source.go` (for `prs_model_source.type = "duckdb"`) will utilize `dbutil.ExecuteDuckDBQuery`.
It will be responsible for:
- Constructing the SQL query string dynamically:
  - `SELECT` clause: Based on configured column names (`prs_model_source.snp_id_column_name`, `prs_model_source.chromosome_column_name`, etc.).
  - `FROM` clause: Targeting the `associations_clean` table (or a configured table name).
  - `WHERE` clause: Using the configured `prs_model_source.model_id_column_name` (defaulting to `"study_id"`) to filter by the `modelID` passed to `loadPRSModel`.
- Providing query arguments (the `modelID` value).
- Implementing a `RowScanner[reference.PRSModelVariant]` function to map query results to `reference.PRSModelVariant` structs, handling type conversions and optional fields (like `OtherAllele`).

### 3.3. New Configuration Key

An **optional** configuration key will be added:
- `config.PRSModelSourceModelIDColKey = "prs_model_source.model_id_column_name"`

This key allows users to specify which column in the DuckDB table identifies the PRS model. If not set, `PRSReferenceDataSource` will default to using `"study_id"`.

### 3.4. Potential Refactoring (Future Consideration)

Existing functions in `internal/gwas/gwas_duckdb_loader.go` (e.g., `FetchGWASRecordsWithTable`) could be refactored to use `dbutil.ExecuteDuckDBQuery` to reduce code duplication.

## 4. Key Components & Files to Modify

- **New/Modified Functionality:**
  - `internal/dbutil/dbutil.go` (or `internal/dbutil/executor.go`): For `ExecuteDuckDBQuery` and `RowScanner` type definition.
  - `internal/config/config.go`: To define `PRSModelSourceModelIDColKey`.
  - `internal/reference/prs_reference_data_source.go`: To implement DuckDB loading logic in `loadPRSModel` using the generic executor and new config key.
- **Tests:**
  - `internal/dbutil/dbutil_test.go` (or `executor_test.go`): Unit tests for `ExecuteDuckDBQuery`.
  - `internal/reference/prs_reference_data_source_test.go`: Unit tests for `loadPRSModel` (DuckDB case), including setup of in-memory DuckDB or mocking.
  - `internal/config/config_test.go`: Tests for the new optional configuration key if it impacts validation or specific getter logic (likely minimal if just read with a default).

## 5. Dependencies & Related Specifications

- This specification builds upon requirements outlined for `PRSReferenceDataSource` on-the-fly computation, detailed in `.agent/prs_reference_data_source.md`.

## 6. Relevant Existing Files for Context

- `internal/gwas/gwas_duckdb_loader.go`: Shows current DuckDB interaction patterns.
- `internal/gwas/gwas_data_fetcher.go`: Illustrates how GWAS data is consumed downstream.
- `internal/pipeline/pipeline.go`: Shows the overall data flow and how different components, including `PRSReferenceDataSource` and GWAS data fetching, are orchestrated.
- `polygenic-risk-calculator/.agent/data_model.md`: Defines data structures like `model.GWASSNPRecord` and `reference.PRSModelVariant` (implicitly, as it's defined in `prs_reference_data_source.go`).
- `PHITE/gwas/sql/create_table_associations_clean.sql`: Defines the schema of the target `associations_clean` table in DuckDB.

## 7. Required Tests

- **Unit Tests for `dbutil.ExecuteDuckDBQuery`:**
  - Test successful query execution and data mapping for various data types.
  - Test handling of empty results.
  - Test error handling (DB connection errors, query errors, scanner errors).
- **Unit Tests for `PRSReferenceDataSource.loadPRSModel` (DuckDB case):**
  - Test successful loading and mapping of PRS model variants from a mock/in-memory DuckDB.
  - Test with all configured columns present.
  - Test with optional `other_allele_column_name` configured and not configured.
  - Test correct use of `model_id_column_name` (both default `"study_id"` and a custom configured value).
  - Test error handling (DB errors, table/column not found, data type mismatches during scan).
  - Test scenario where no variants are found for a given `modelID`.

## 8. Implementation Instructions

Refer to the general agent development guidelines in `.agent/README.md` for coding standards, testing practices, and PR procedures.

**Implementation Steps Overview:**
1. Define `RowScanner` type and implement `dbutil.ExecuteDuckDBQuery` in `internal/dbutil`.
2. Add `PRSModelSourceModelIDColKey` to `internal/config/config.go`.
3. Update `PRSReferenceDataSource` struct and `NewPRSReferenceDataSource` in `internal/reference/prs_reference_data_source.go` to handle the new optional config key.
4. Implement the DuckDB-specific logic within `PRSReferenceDataSource.loadPRSModel` using `dbutil.ExecuteDuckDBQuery` and a custom `RowScanner[reference.PRSModelVariant]`.
5. Write comprehensive unit tests for all new and modified components.
