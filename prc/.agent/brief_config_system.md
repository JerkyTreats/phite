# Brief: Centralized Configuration System

## Objective
Establish a single, robust, and extensible configuration system for the PHITE polygenic-risk-calculator, supporting CLI flags, environment variables, and config files with clear precedence and best practices.

---

## Original Spec
- All configuration must go through a centralized config package (`internal/config`).
- The config system must support loading from:
  - CLI flags (highest precedence, for user-facing/required parameters)
  - Environment variables (for automation/advanced use)
  - Config files (for persistent/project or user-level defaults)
- Precedence order: CLI > env > config file > hardcoded default.
- The config package must provide typed getters (e.g., `GetString`, `GetInt`).
- All config keys and environment variables must be documented.
- The config system must support hot reload (optional, for advanced use).
- All required user-facing parameters must be settable via CLI flag.
- The system must error early and clearly if required configuration is missing.
- All config access in the codebase must go through this package; no direct env or file reads elsewhere.

---

## Required Changes (from reference stats loader audit)
- Ensure the reference stats DB path is settable via CLI flag (`--reference-db`), with env/config fallback.
- (Optional) Add CLI/config/env support for reference ancestry, trait, and model.
- Update documentation to reference this brief and the config system pattern.

---

## CLI Flag and Config Key Mapping

| CLI Flag         | Config Key        | Env Var            | Notes                                                      |
|------------------|------------------|--------------------|------------------------------------------------------------|
| `--genotype-file`| *(none)*         | *(none)*           | CLI only (required user input)                             |
| `--snps`         | *(none)*         | *(none)*           | CLI only (required user input)                             |
| `--snps-file`    | *(none)*         | *(none)*           | CLI only (required user input)                             |
| `--gwas-db`      | `gwas_db_path`   | `GWAS_DUCKDB`      | CLI > env > config > default                               |
| `--gwas-table`   | `gwas_table`     | `GWAS_TABLE`       | CLI > env > config > default                               |
| `--output`       | *(none)*         | *(none)*           | CLI only                                                   |
| `--format`       | *(none)*         | *(none)*           | CLI only                                                   |
| `--reference-db` | `reference_db`   | `REFERENCE_DUCKDB` | **Should be added** (see Reference Stats Loader brief)      |

**Policy:**
- Only parameters that benefit from automation or persistent defaults should have config/env keys.
- Required user-facing files/parameters must always be CLI-settable.
- All config keys and env vars must be documented and consistent across CLI, config, and code.

---

## Acceptance Criteria
- All configuration access is centralized in `internal/config`.
- CLI flags are used for all required user-facing parameters.
- Environment variables and config files are used for defaults/advanced use.
- Documentation and code comments reference this brief and the config system pattern.
- All config keys and env vars are documented in one place.

---

## References
- [Agent Contributor Guide — Configuration Section](./README.md#external-data--parameter-loading)
- [Brief: Reference Stats Loader — Configuration Compliance](./brief_reference_stats_loader_config.md)
- [Original Logging System Brief] (add link if available)

---

## Files to Update
- `internal/config/config.go` (ensure all config access is through this package)
- `cmd/risk-calculator/main.go` (thread CLI flags through config)
- `.agent/README.md` (add reference to this brief)
- All relevant module briefs

---

## Out of Scope
- Implementation of config file parsing details (already handled by Viper)
- Changes to unrelated agent/component logic
