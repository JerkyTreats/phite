# BigQuery Reference Panel

## Objective
Leverage Google BigQuery as a scalable backend for population reference statistics (e.g., gnomAD) to support PRS normalization and ancestry-aware queries in the PHITE pipeline.

- Use the public [gnomAD](https://gnomad.broadinstitute.org/downloads) datasets available in BigQuery (see [marketplace link](https://console.cloud.google.com/marketplace/product/broad-institute/gnomad?project=jerkytreats)).
- Enable efficient, cost-effective, and up-to-date access to population allele frequencies and summary stats without local ingestion of massive VCFs (>5TB).

## Reasoning
- gnomAD VCFs are too large for local processing; BigQuery provides a sharded, indexed, and annotation-parsed version suitable for on-demand queries.
- The BigQuery free tier (1TB/mo) is sufficient for most research and clinical use cases.

## Requirements & Responsibilities
- **Implement a BigQuery reference stats backend** as a Go package (suggested: `internal/reference/bq_reference.go`).
- **Clientset Abstraction:**
  - Implement a `BigQueryClient` (or `BigQueryClientset`) struct that encapsulates all configuration, authentication, connection, and query logic for BigQuery.
  - All BigQuery operations (e.g., `GetReferenceStats`, `Ping`, etc.) must be methods on this struct.
  - The struct should be initialized with all required config (project, dataset, table, credentials).
  - Provide an interface for mocking/testing.
  - Example interface:
    ```go
    type BigQueryClient struct {
        // Fields: project, dataset, table, credentials, etc.
    }
    func NewBigQueryClient(cfg BigQueryConfig) (*BigQueryClient, error)
    func (c *BigQueryClient) GetReferenceStats(ctx context.Context, ancestry, trait, model string) (*ReferenceStats, error)
    ```
- **Configuration:**
  - **All configuration for BigQuery (project, dataset, table, credentials, etc.) must be handled through the centralized config system in `config.go`. No implementation may create its own config loader, environment variable parser, or CLI flag system.**
  - All connection parameters must be settable via the mechanisms defined in `config.go` (typically CLI flags and config files), with config/env fallback as specified there.
  - Fail early and clearly if required config is missing (see agent README and `config.go` for CLI/config precedence).
- **Authentication:**
  - Support Google Application Default Credentials and explicit service account JSON key.
  - Clearly document credential loading order and error out if not found.
- **Expected Table Schema:**
  - Must support querying mean, std, min, max, ancestry, trait, and model (see `reference_stats` DuckDB schema for canonical columns).
  - If the gnomAD table schema differs, map/transform columns as needed in code.
- **Query Patterns:**
  - Efficiently query per-ancestry, per-trait, and per-model stats using indexed columns.
  - Support filtering by chromosome, region, or variant if needed for downstream use.
- **Error Handling:**
  - Surface BigQuery errors with actionable messages (e.g., auth, quota, malformed queries).
  - Allow pipeline to proceed with raw PRS if reference stats are unavailable.
- **Extensibility:**
  - Design for future support of additional summary statistics or alternative reference panels.
- **Testing:**
  - Use TDD: write failing tests for config, connection, query, and error scenarios.
  - Mock BigQuery for unit tests; integration tests should require explicit opt-in (env var or CLI flag).
  - Test all CLI/config permutations and error paths.

## Consumed By
- PRS Score Normalizer
- Reference Stats Loader (as a backend option)
- Main pipeline orchestrator (`internal/pipeline/pipeline.go`)—this file must be updated/uploaded whenever the loader or backend interface changes.

## Performance Notes
- Typical BigQuery reference stats query (per trait/ancestry/model): 0.5–3 seconds.
- Returned object size: ~100–200 bytes per trait (negligible for memory/transport).

## Related Briefs
- [Reference Stats Loader](brief_reference_stats_loader.md)
- [DuckDB Shared Utilities](brief_duckdb_shared_utilities.md)

## Example Usage (Go)
```go
stats, err := reference.LoadReferenceStatsFromBigQuery(ctx, bqConfig, ancestry, trait, model)
if err != nil {
    // handle error or proceed with raw PRS
}
```

## Checklist
- [ ] Implements BigQuery backend as described
- [ ] Implements a `BigQueryClient` (Clientset abstraction) per requirements
- [ ] CLI/config/env for all connection params
- [ ] Adheres to error handling and test requirements
- [ ] Reference stats schema matches canonical columns or is mapped
- [ ] TDD: all code covered by tests