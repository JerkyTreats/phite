# Polygenic Risk Calculator — Agent Contributor Guide

Welcome to the PHITE polygenic risk calculator project. This document provides common instructions and standards for all agent/component implementers.

## Progress

> **NOTE:** When you begin work on a new agent/component brief, or complete its implementation, please update the Progress table below to reflect its current status (To Do, In Progress, Complete). This ensures all contributors have an accurate view of project progress.

| Brief | Status | Description |
|-------|--------|-------------|
| DuckDB Shared Utilities           | Complete   |  |
| Entrypoint                        | Complete   | CLI pipeline unified, type-safe, and all conversions removed |
| Genotype File Parser              | Complete   |  |
| GWAS Data Fetcher                 | Complete   |  |
| GWAS DuckDB Loader                | Complete   |  |
| Output Formatter                  | Complete   |  |
| PRS Calculator                    | Complete   |  |
| Reference Stats Loader            | Complete   |  |
| Score Normalizer                  | Complete   |  |
| Shared Model Package              | Complete   | Canonical models used throughout pipeline, type unification enforced |
| Trait Summary Generator           | Complete   | |

**Legend:**
- **To Do**: No substantive implementation found.
- **In Progress**: Partial implementation; core functionality not yet complete.
- **Complete**: Fully implemented and tested, matches brief requirements.

## General Principles
- **Test Driven Development (TDD):**
  - All code must be developed using TDD, following the red-green-refactor cycle.
  - Write failing tests first, then implement code to pass, then refactor for clarity and maintainability.
- **High Quality Code:**
  - Code must be idiomatic Go, clear, maintainable, and robust.
  - Avoid duplication (DRY), but not at the expense of readability or simplicity.
  - All public functions, structs, and packages must have clear, concise comments.
  - Use descriptive variable and function names; avoid unnecessary abbreviations.
- **Best Practices:**
  - Use Go modules for dependency management.
  - Prefer composition over inheritance.
  - Handle errors explicitly and gracefully.
  - Use context.Context where appropriate for cancellation and deadlines.
  - Validate all external input and handle malformed data safely.

## External Data & Parameter Loading

See also: [Brief: Centralized Configuration System](./brief_config_system.md)


All required external data (such as files, databases, or key parameters) must be configurable via CLI flags. CLI flags should be the primary method for users to specify required inputs. Configuration files and environment variables may be used as fallbacks or for advanced/automated deployments, but must not be the only way to set required data.

**Pattern:**
- If a parameter is required for successful operation (e.g., input file, GWAS DB path), it MUST be settable via a CLI flag.
- CLI flags take precedence over config and environment variables.
- Config files and environment variables are for defaults, automation, and advanced use.
- Hardcoded defaults are allowed only as a last resort for developer convenience.
- The CLI should error early and clearly if required data is missing.

**Example:**
```go
// CLI flag (required): --gwas-db
// Config key: gwas_db_path
// Env var: GWAS_DUCKDB
```

## Logging Standards

All logging in PHITE must use the centralized logger (`internal/logging`). Do not use `fmt.Println` or the standard library `log` for application output.

### INFO Level
- **Purpose:** Record high-level, user-meaningful events and successful operations.
- **When to log:**
  - Start and completion of major processing steps (e.g., file loaded, analysis started/finished).
  - Successful external resource loads (e.g., config, data, models).
  - User actions (e.g., command-line invocation, parameter parsing).
- **How to log:** Use clear, concise, and structured messages. Include relevant context (file name, operation, user input).
- **Examples:**

  ```go
  logging.Info("Loaded GWAS summary statistics from %s", filePath)
  logging.Info("PRS calculation completed for %d samples", sampleCount)
  logging.Info("User selected output format: %s", format)
  ```

### ERROR Level
- **Purpose:** Record failures, exceptions, and conditions requiring user or operator intervention.
- **When to log:**
  - Any operation that fails and is not handled by retry or fallback.
  - Invalid or missing input data/configuration.
  - External system or dependency failures.
- **How to log:** Clearly state what failed and why. Include error details and context.
- **Examples:**

  ```go
  logging.Error("Failed to load config from %s: %v", path, err)
  logging.Error("PRS calculation aborted: missing required SNP data")
  logging.Error("Could not write output file %s: %v", outPath, err)
  ```

### DEBUG Level
- **Purpose:** For developer diagnostics and deep troubleshooting.
- **Standard:** Use case-by-case; not required for ordinary operation.

#### Additional Guidance
- **Never log sensitive data** (e.g., user genotypes, private information).
- **Keep messages actionable**—they should help users or developers understand what happened and what to do next.
- **INFO and ERROR logs** should be sufficient for most operational monitoring and user support.

## Commenting Style
- Use GoDoc conventions for all exported symbols:
  - Start comments with the name of the item being described.
  - Be concise but informative.
  - Example:

  ```go
  // CalculatePRS computes the polygenic risk score for a set of SNPs.
  func CalculatePRS(snps []AnnotatedSNP) float64 { ... }
  ```

- Inline comments should clarify intent, not restate code.

## Folder and File Structure
- **Project Root:** All paths and organization are relative to `polygenic-risk-calculator/`, which is the project root.
- **Source code:** Place all implementation code in well-named subdirectories within `polygenic-risk-calculator/`, following common Go project layout. For example:
  - `cmd/` — Entrypoints (e.g., main.go)
  - `internal/` — Private packages not intended for external use
  - `pkg/` — Public packages (if any)
  - Domain-specific folders: `genotype/`, `gwas/`, `prs/`, `output/`, etc.
- **Agent briefs:** All agent briefs must reside in `.agent/` within the project root.
- **Data model:** The canonical data model is in `.agent/data_model.md`.
- **Tests:**
  - **Go-specific:** For Go code, all test files must be placed alongside the code they test, following standard Go idioms.
    - Test files must be named with the `_test.go` suffix (e.g., `genotype_parser_test.go`).
    - Tests should be in the same package or in a `_test` package within the same directory as the implementation.
    - Example:
      - Source: `genotype/genotype_parser.go`
      - Test: `genotype/genotype_parser_test.go`
  - For other languages, tests should be placed in a top-level `tests/` directory, mirroring the source structure if needed.
- **Keep the codebase organized and modular.**

## Test File Naming and Structure
- Every exported function or method must have corresponding tests.
- Use table-driven tests where appropriate.
- Use descriptive test names (e.g., `TestCalculatePRS_ValidInput`).
- Place all test helpers or fixtures in `tests/helpers/`.

## Code Quality
- Code must:
  - Pass all tests and linters.
  - Avoid global state unless justified.
  - Be easy to extend and maintain.
  - Be DRY within reason; do not over-abstract.

## References
- See `.agent/data_model.md` for canonical data model and input schemas.
- See all agent briefs in `.agent/` for component responsibilities, interfaces, and test requirements.

## Summary Checklist
- [ ] Use TDD and red-green-refactor
- [ ] Follow Go idioms and best practices
- [ ] Comment all exported symbols and clarify intent
- [ ] Organize code and tests as described
- [ ] Keep code DRY, clear, and high quality
- [ ] Reference the data model and briefs as the source of truth

For any questions, consult the agent briefs or contact the project maintainers.
