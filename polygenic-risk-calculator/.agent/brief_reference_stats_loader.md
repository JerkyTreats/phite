# Agent Brief: PRS Reference Stats Loader (Optional)

> **Note:** The DuckDB database used by this loader is provided by the external shared GWAS database (`../gwas/gwas.duckdb`), which is managed separately. For data engineering and schema creation, see briefs in `gwas/.agent/`.

## Purpose
Provide a mechanism for loading population-level reference statistics (mean, std, min, max) from a DuckDB database used for normalizing polygenic risk scores. Reference data is **optional**—if not provided, only the raw PRS will be output.

## Responsibilities
- Use DuckDB Shared Utilities for connection/session management and schema validation.
- Connect to a DuckDB database at a specified file path.
- Query the reference stats table for mean, std, min, max (and optionally ancestry, trait, model).
- If reference stats are not provided, allow the pipeline to proceed with raw PRS output only.
- Validate the presence and format of reference stats if supplied.
- Provide reference stats as Go structs to downstream components.
- Support extensibility for additional statistics if needed.

## Inputs
- DuckDB file path (from CLI argument `--gwas-db` or `--reference-db`)
- (Optional) Ancestry, trait, or model identifier to select appropriate reference stats.

## Outputs
- `referenceStats` struct (mean, std, min, max, ancestry, trait, model) if available.
- Nil or empty if not available.

## Consumed By
- PRS Score Normalizer
- Entrypoint (for orchestration)

## Related Briefs
- [DuckDB Shared Utilities and Coordination](brief_duckdb_shared_utilities.md)

## Required Tests
- Loads valid reference stats file/config if provided.
- Fails gracefully with clear errors on malformed files.
- Allows pipeline to proceed with raw PRS if no reference stats are provided.
- Provides correct Go structs for downstream use.

## Required DuckDB Table Schema

The loader expects a DuckDB table named `reference_stats` with the following required columns. This schema must be present in the database for the loader to function correctly. Schema creation is handled in a separate brief.

| Column    | Type    | Description                                   |
|-----------|---------|-----------------------------------------------|
| mean      | DOUBLE  | Mean PRS score in the reference population    |
| std       | DOUBLE  | Standard deviation of PRS in the population   |
| min       | DOUBLE  | Minimum PRS score in the population           |
| max       | DOUBLE  | Maximum PRS score in the population           |
| ancestry  | TEXT    | (Optional) Ancestry group for these stats     |
| trait     | TEXT    | (Optional) Trait or phenotype name            |
| model     | TEXT    | (Optional) PRS model identifier/version       |

> **Note:** This schema definition should be referenced as input when writing the schema creation brief for the reference stats table.

## Example Usage
```go
ref, err := reference.LoadReferenceStatsFromDuckDB(dbPath, ancestry, trait, model)
if err != nil {
    // handle error, or proceed with raw PRS only
}
if ref != nil {
    norm, err := prs.NormalizePRS(prsResult, *ref)
    // ...
} else {
    // Output raw PRS only
}
```
