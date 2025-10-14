# Brief: CLI Options & Unified Parameter Resolution System

## Purpose
Establish a robust, testable, and extensible system for parsing, validating, and resolving all user-facing parameters—across CLI flags, environment variables, and config files—for the PHITE polygenic-risk-calculator CLI and agents.

---

## Motivation
- Decouple CLI flag parsing and validation from main application logic.
- Ensure all parameter precedence (CLI > env > config > default) is handled consistently and testably.
- Provide a single source of truth (`Options` struct) for all downstream pipeline components.
- Facilitate future extensions (subcommands, interactive prompts, advanced validation, etc).

---

## Responsibilities
- Define an `Options` struct representing all user-facing parameters (required and optional).
- Parse CLI flags into `Options`.
- For each parameter, if the CLI flag is unset, resolve from env/config using the centralized config system.
- Validate required parameters and enforce mutual exclusivity or dependencies.
- Document all supported flags, config keys, and env vars.
- Provide a function (e.g., `ParseOptions(args []string) (Options, error)`) for use in `main.go` and tests.
- Support unit and integration testing of all parameter resolution logic.

---

## Design Pattern
- **Separation of Concerns:** CLI parsing, config/env resolution, and application logic are cleanly separated.
- **Single Source of Truth:** The `Options` struct is the canonical representation of all runtime parameters.
- **Explicit Precedence:** For each parameter: CLI > env > config > hardcoded default.
- **Testability:** All resolution and validation logic is covered by tests.

---

## Example
```go
// options.go
package cli

type Options struct {
    GenotypeFile string
    SNPs         []string
    SNPsFile     string
    GWASDB       string
    GWASTable    string
    Output       string
    Format       string
    ReferenceDB  string
    // ...future fields
}

// ParseOptions parses CLI flags and resolves each parameter from CLI/env/config/default.
func ParseOptions(args []string) (Options, error) { /* ... */ }
```

---

## Required Tests

### 1. Unit Tests
- Parameter precedence: CLI > env > config > default for all fields (table-driven)
- Required flags: error if missing (e.g., --genotype-file, --snps or --snps-file)
- Mutual exclusivity: error if both --snps and --snps-file are set
- Edge cases: empty/invalid values, duplicate/conflicting flags
- Help/usage output: correct and complete
- Error messages: clear and actionable for all invalid input

### 2. Integration Tests
- Simulate full CLI invocation with various combinations of CLI args, env vars, and config file values
- Validate correct Options struct is produced and passed to downstream code
- Ensure no state leakage between tests (reset env/config between runs)

### 3. Coverage
- All code branches in ParseOptions and helpers must be covered
- Table-driven tests for permutations of CLI/env/config/default

---

## Acceptance Criteria
- All CLI/config/env logic is in a dedicated package (e.g., `internal/cli`).
- The `Options` struct is used throughout the pipeline; no direct flag/env/config access elsewhere.
- Precedence and validation logic are unit tested.
- Documentation and help text are generated from the options system.
- All required user-facing parameters are CLI-settable, with config/env fallback as documented.

---

## Files to Create/Update
- `internal/cli/options.go` (new)
- `cmd/risk-calculator/main.go` (refactor to use `Options`)
- `internal/config/config.go` (may add helpers)
- `.agent/brief_config_system.md` (reference this brief)
- `.agent/README.md` (reference this brief)
- Relevant tests

---

## References
- [Brief: Centralized Configuration System](./brief_config_system.md)
- [Brief: Reference Stats Loader — Configuration Compliance](./brief_reference_stats_loader_config.md)

---

## Out of Scope
- Implementation of config file parsing details (already handled by Viper)
- Changes to unrelated agent/component logic
