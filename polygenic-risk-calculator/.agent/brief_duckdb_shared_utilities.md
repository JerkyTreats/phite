# Agent Brief: DuckDB Shared Utilities and Coordination

## Purpose
Establish a set of shared utilities, patterns, and standards for all agents/components interacting with DuckDB databases. This ensures robust, maintainable, and consistent data access, error handling, and schema validation across the pipeline.

## Responsibilities
- Provide centralized DuckDB connection/session management.
- Offer schema validation utilities for required tables and columns.
- Standardize error types and reporting for DuckDB operations.
- Enable configuration and dependency injection of DuckDB connections.
- Support unified logging and (optional) metrics for DuckDB queries and errors.
- Coordinate schema versioning and migration checks.

## Inputs
- DuckDB file path (from CLI or config)
- Table and column requirements from consuming agents
- Logger/metrics interfaces (optional)

## Outputs
- Shared connection/session objects
- Schema validation errors or status
- Standardized error objects for downstream handling
- Logs and metrics (if configured)

## Consumed By
- gwas_duckdb_loader
- reference_stats_loader
- Any future DuckDB-based agents

## Required Tests
- Connection open/close lifecycle
- Table and column presence/validation
- Error propagation and reporting
- Logging and metrics output (if enabled)

## Example Usage

```go
// Open connection
conn, err := dbutil.OpenDuckDB(path)
if err != nil { /* handle error */ }

defer conn.Close()

// Validate schema
if err := dbutil.ValidateTable(conn, "reference_stats", []string{"mean","std","min","max"}); err != nil {
    // handle schema error
}

// Use connection in loader
ref, err := reference.LoadReferenceStatsFromConn(conn, ancestry, trait, model)
```

## Notes
- All agents should use these utilities for DuckDB access and validation.
- Schema definitions should reside in a canonical location (e.g., `schema/` directory) and be referenced by both utilities and migration/creation briefs.
