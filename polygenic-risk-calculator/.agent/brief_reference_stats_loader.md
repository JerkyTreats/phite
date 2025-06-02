# Agent Brief: PRS Reference Stats Loader

> **Note:** BigQuery is the only supported backend for population reference statistics. For backend implementation and configuration details, see the [BigQuery Reference Panel brief](bigquery.md).

## Purpose
Provide a unified, production-ready interface for loading population-level reference statistics (mean, std, min, max, ancestry, trait, model) required for PRS normalization. Reference stats are **required** for meaningful interpretation; outputting raw PRS is a degraded/fallback mode.

## Separation of Concerns & Overlap
- **This loader** is the orchestration/interface layer: it defines the API for retrieving reference stats and integrates with the rest of the PRS pipeline.
- **The BigQuery backend** (see [bigquery.md](bigquery.md)) encapsulates all connection, authentication, configuration, and query logic for interacting with the gnomAD reference data in BigQuery.
- **Boundary:** The loader should not duplicate or subsume backend logic (e.g., SQL construction, credential handling). Instead, it delegates to the BigQuery Clientset for all data access.
- **Cross-reference:** Any changes to schema, query patterns, or authentication must be implemented in the BigQuery backend and surfaced via the loader interface.

## Responsibilities
- Expose a single, robust interface for retrieving reference stats from BigQuery.
- **All configuration must be routed through the centralized config system as defined in `config.go`. No implementation may create its own config loader, environment variable parser, or CLI flag system.**
- Validate input parameters (ancestry, trait, model, etc.) and propagate errors from the backend.
- Handle orchestration logic: e.g., fallback to raw PRS if stats unavailable, log/report errors, and ensure downstream components receive the correct struct or error.
- Do **not** implement or maintain local database logic (e.g., DuckDB) for reference stats.
- Reference and depend on the BigQuery Clientset as defined in [bigquery.md](bigquery.md).

## Inputs
- BigQuery config (must be provided via the centralized config system in `config.go`; see [bigquery.md](bigquery.md) for required fields)
- ancestry, trait, or model identifier (as needed for filtering)

## Outputs
- `ReferenceStats` struct (mean, std, min, max, ancestry, trait, model) if found
- Nil/empty if not available (with clear error/warning)

## Consumed By
- PRS Score Normalizer (for normalization)
- Entrypoint (for orchestration)
- Main pipeline orchestrator (`internal/pipeline/pipeline.go`)

## Integration
- The main pipeline in `internal/pipeline/pipeline.go` is the primary consumer of the reference stats loader. This file must be updated and uploaded whenever the loader interface or output struct changes.
- See [bigquery.md](bigquery.md) for backend implementation details.

## Performance Notes
- A BigQuery reference stats query (per trait/ancestry/model) is expected to take 0.5–3 seconds, depending on BQ load and partitioning.
- The returned `ReferenceStats` object is small (~100–200 bytes per trait).

## Interfaces
```go
// Loads reference stats from BigQuery via the backend Clientset.
func LoadReferenceStats(ctx context.Context, config ReferenceStatsConfig, ancestry, trait, model string) (*ReferenceStats, error)
```
- `ReferenceStatsConfig` must encapsulate all BigQuery connection parameters; see [bigquery.md](bigquery.md) for required fields.
- All backend logic is delegated to the BigQuery Clientset (see [bigquery.md](bigquery.md)).

## Error Handling
- Fail early and clearly on missing/malformed config or schema
- Surface BigQuery-specific errors with actionable messages
- If stats are not found, return nil and allow pipeline to continue (with warning)

## Extensibility
- Loader must remain agnostic to backend implementation details; all changes to schema, query, or authentication must be isolated to the BigQuery backend.
- Design for future support of additional summary statistics or alternative cloud-native backends (reference all such changes in both briefs).

## See Also
- [BigQuery Reference Panel brief](bigquery.md) — for backend implementation, schema, and configuration

## Required Tests
- Loads valid reference stats from DuckDB (and BQ, when implemented)
- Fails gracefully and logs clear errors on malformed/missing files or config
- Allows pipeline to proceed with raw PRS if no stats are found
- Provides correct Go structs for downstream use
- Covers all CLI/config permutations and error paths

## Required Table Schema
The backend must provide a table named `reference_stats` (or equivalent) with the following columns:

| Column    | Type    | Description                                   |
|-----------|---------|-----------------------------------------------|
| mean      | DOUBLE  | Mean PRS score in the reference population    |
| std       | DOUBLE  | Standard deviation of PRS in the population   |
| min       | DOUBLE  | Minimum PRS score in the population           |
| max       | DOUBLE  | Maximum PRS score in the population           |
| ancestry  | TEXT    | (Optional) Ancestry group for these stats     |
| trait     | TEXT    | (Optional) Trait or phenotype name            |
| model     | TEXT    | (Optional) PRS model identifier/version       |

> **Note:** Map or transform columns as needed if backend schema differs.

## Example Usage
```go
ref, err := reference.LoadReferenceStats(ctx, config, ancestry, trait, model)
if err != nil {
    // handle error or proceed with raw PRS only
}
if ref != nil {
    norm, err := prs.NormalizePRS(prsResult, *ref)
    // ...
} else {
    // Output raw PRS only
}
```

## Related Briefs
- [BigQuery Reference Panel](bigquery.md)
- [DuckDB Shared Utilities](brief_duckdb_shared_utilities.md)
- [Data Model](../data_model.md)

## Checklist
- [ ] Supports DuckDB and future BigQuery
- [ ] CLI/config/env for all connection params
- [ ] Adheres to error handling and test requirements
- [ ] Table schema matches canonical columns or is mapped
- [ ] TDD: all code covered by tests
