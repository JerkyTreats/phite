# PHITE Logging System Brief

## Objective
Implement a robust logging system for PHITE CLI and internal packages using the [`uber-go/zap`](https://github.com/uber-go/zap) logging library. The system must support configurable log levels via a user config file and provide clear instructions for INFO-level logging.

## Requirements

### 0. Centralized, Extensible Config System (via Viper)
- Use [`spf13/viper`](https://github.com/spf13/viper) as the configuration backend for PHITE.
- Implement a dedicated package (e.g., `internal/config`) that wraps viper for all configuration loading and access.
- The config system must:
  - Load settings from `~/.phite/config.json` (with a default path override for testing), leveraging viper's multi-format support (JSON, YAML, TOML, etc.).
  - Support arbitrary config keys and values—viper allows dynamic keys and nested config.
  - Provide a typed API for retrieving config values (e.g., `GetString(key)`, `GetInt(key)`, `GetBool(key)`, etc.), using viper's API.
  - Allow for easy addition of new config options as features expand.
  - Support environment variable overrides and default values via viper.
  - Expose a method to reload config at runtime (viper supports hot reload for some formats; design for this but optional for this brief).
  - Validate config schema and provide clear error messages for missing/invalid settings.
  - Be thoroughly unit tested (defaulting, error handling, value retrieval, reload logic).
- All modules (logging, CLI, internal packages, etc.) must access config exclusively via this package, never by reading files directly or using viper directly.
- Document the config system for developers, including how to add new settings and how to use it in feature code.

### 1. Core Logging Logic (via zap)
- Use [`uber-go/zap`](https://github.com/uber-go/zap) as the logging backend for all PHITE CLI and internal package logging.
- Centralize logger initialization and configuration in a dedicated package (e.g., `internal/logging`).
- All code (CLI and internal packages) must use this logger for output, not `fmt.Println` or the standard library `log` package.
- Use zap's SugaredLogger for developer ergonomics (printf-style logging) and structured logging for diagnostics/telemetry.
- Integrate logger configuration (e.g., log level) with the centralized config system (viper).
- Provide clear documentation and examples for INFO-level logging:

```go
import "phite.io/polygenic-risk-calculator/internal/logging"
logging.Info("Loaded GWAS summary statistics from %s", filePath)
```

### 2. Configurable Log Level
- Assume the existence of a user config file at `~/.phite/config.json`.
- The config file contains a key `log_level` (e.g., "INFO", "DEBUG", "ERROR").
- On startup, the logger must:
  - Read `log_level` from the config file (default to "INFO" if missing or invalid).
  - Set the global log level accordingly.
  - If the config file is missing, default to "INFO" and log a warning at startup.
- Changing the log level at runtime is not required for this brief.

### 3. INFO Level Logging Instructions
- All INFO-level events (e.g., start/finish of major steps, successful file loads, user actions, etc.) must use the logger's INFO method.
- Example usage:

```go
import "phite.io/polygenic-risk-calculator/internal/logging"
logging.Info("Loaded GWAS summary statistics from %s", filePath)
```

- Do not use `fmt.Println` or `log.Printf` for such messages.

### 4. Testability
- The logger initialization logic must be testable (e.g., allow injecting a config path for tests).
- Provide unit tests for:

---

## Test Requirements

### Config System (`internal/config`)
- **Functional tests:**
  - Loads config from default path (`~/.phite/config.json`).
  - Loads config from custom path (for tests/overrides).
  - Correctly parses and returns values for all supported types (string, int, bool, etc.).
  - Supports nested config keys and environment variable overrides.
- **Schema/validation tests:**
  - Handles missing config file gracefully (uses defaults, logs warning).
  - Handles invalid/malformed config file (returns error, logs warning).
  - Returns clear errors for missing/invalid required settings.
- **Edge-case tests:**
  - Handles empty config files.
  - Handles config files with extra/unexpected keys.
  - Supports reload (if implemented).
- **Integration tests:**
  - Works correctly with the logging system (log level is set from config).
- **Testability:**
  - All config-loading logic is injectable and mockable for testing.

### Logging System (`internal/logging`)
- **Functional tests:**
  - Logger initializes with correct log level from config.
  - Defaults to INFO level if config is missing or invalid.
  - Logs at all supported levels (INFO, DEBUG, ERROR, etc.)
  - Logs output is formatted as expected (human/JSON as configured).
  - SugaredLogger methods (e.g., `Infof`, `Errorf`) work as intended.
- **Edge-case tests:**
  - Handles missing/invalid log level in config gracefully.
  - Logs warning if config file is missing.
- **Integration tests:**
  - Logging output can be captured and asserted in tests.
  - Logging and config packages work together as expected (e.g., changing config changes log level on restart).
- **Testability:**
  - Logger initialization accepts injected config for tests.
  - No global state leaks between tests.
  - Correct log level parsing from config.
  - Defaulting to INFO on missing/invalid config.
  - Logging at INFO level.

---

## Deliverables
- `internal/logging/` package with logger initialization and usage API.
- Unit tests for the logging package.
- Example usage in CLI and at least one internal package.
- Brief documentation update for developers.

---

## References
- [uber-go/zap documentation](https://github.com/uber-go/zap)
- [spf13/viper documentation](https://github.com/spf13/viper)
- Example config file:

```json
{
  "log_level": "DEBUG"
}
```

- **Links to Broader Implementation Guidance:**
  - See [README.md](./README.md) for project-wide agent and implementation standards.
  - See [SNP File Input Brief](./brief-snps-file-input.md) for related data ingestion requirements.

---

## Logging Coverage Audit

The following files and functions currently lack required logging and must be updated to comply with PHITE logging standards. Logging should be introduced at all major process boundaries, error conditions, and user-meaningful events as described in the standards.

### CLI Entrypoint

- **cmd/risk-calculator/main.go**
  - `RunCLI`:
    - Add INFO logs for: CLI start, argument parsing, file loads, major pipeline stages (GWAS load, genotype parse, PRS calculation, output).
    - Add ERROR logs for: flag errors, missing/invalid files, failed loads/parses, output errors.
  - `main`:
    - Add INFO log for CLI invocation.

### SNP Input

- **internal/snps/snps_file_parser.go**
  - `ParseSNPsFromFile`:
    - INFO: File opened, format detected, SNPs parsed.
    - ERROR: Unsupported extension, file open errors, parse errors.
  - `parseJSON`, `parseCSV`, `parseTSV`, `parseDelimited`:
    - ERROR: Malformed input, empty rsids, null bytes.

### Genotype Parsing

- **internal/genotype/genotype_parser.go**
  - `ParseGenotypeData`:
    - INFO: File opened, format detected, SNPs validated.
    - ERROR: File open errors, unknown format, validation errors.

### GWAS Data

- **internal/gwas/gwas_data_fetcher.go**
  - `FetchAndAnnotateGWAS`:
    - INFO: Annotation started/completed.
    - ERROR: No GWAS association found (if actionable).

- **internal/gwas/gwas_duckdb_loader.go**
  - `FetchGWASRecords`:
    - INFO: DuckDB opened, query executed, records loaded.
    - ERROR: DB open errors, table validation, query/scan errors.

### Reference Stats

- **internal/reference/reference_stats_loader.go**
  - `LoadReferenceStatsFromDuckDB`:
    - INFO: Reference stats loaded, no stats found.
    - ERROR: DB open errors, table validation, scan errors.

### PRS Calculation

- **internal/prs/prs_calculator.go**
  - `CalculatePRS`:
    - INFO: PRS calculation started/completed.

- **internal/prs/score_normalizer.go**
  - `NormalizePRS`:
    - INFO: Normalization started/completed.
    - ERROR: Invalid stats, math errors.

### Output

- **internal/output/output_formatter.go**
  - `FormatOutput`:
    - INFO: Output started, format selected, file written.
    - ERROR: Unsupported format, file write errors, encoding errors.

- **internal/output/trait_summary_generator.go**
  - `GenerateTraitSummaries`:
    - INFO: Trait summaries generated.

### DuckDB Utilities

- **internal/dbutil/duckdb.go**
  - `OpenDuckDB`, `WithConnection`, `ValidateTable`, `IsTableEmpty`, `TableExists`, `ExecInTransaction`, `CloseDB`:
    - INFO: DB connections opened/closed, validations passed.
    - ERROR: Connection errors, table/column errors, transaction failures.

---

**Instructions:**
- For each function above, add INFO logs for successful major steps and ERROR logs for all error/exceptional conditions.
- Use the centralized logger (`internal/logging`), never `fmt.Println` or `log`.
- See the standards section for message style and examples.
