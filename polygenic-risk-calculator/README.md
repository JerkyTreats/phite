# Polygenic Risk Calculator

The Polygenic Risk Calculator is a command-line tool for computing polygenic risk scores (PRS) from user genotype data and GWAS summary statistics. It is part of the PHITE project and is designed for researchers, clinicians, and developers working with genetic risk prediction.

## Features

- **Flexible Input:** Supports genotype files, GWAS summary statistics in DuckDB, and SNP selection via file or CLI.
- **Robust Pipeline:** Modular pipeline for genotype parsing, GWAS data loading, PRS calculation, normalization, and trait summary generation.
- **CLI-Driven:** All required parameters are configurable via command-line flags for reproducible, scriptable workflows.
- **Structured Output:** Supports multiple output formats and destinations.
- **Centralized Logging:** Consistent, user-meaningful logs for all major steps and errors.
- **Tested, Modular Codebase:** Built using TDD and idiomatic Go best practices.

## Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/your-org/polygenic-risk-calculator.git
   cd polygenic-risk-calculator
   ```
2. Build the CLI:
   ```sh
   go build -o risk-calculator ./cmd/risk-calculator
   ```

## Usage

Run the calculator with required arguments:

```sh
./risk-calculator \
  --genotype-file <genotype.vcf|txt|...> \
  --gwas-db <gwas.duckdb> \
  --snps <rsID1,rsID2,...> | --snps-file <snps.txt> \
  [--gwas-table <table>] \
  [--reference-table <table>] \
  [--output <output.json|csv|...>] \
  [--format <json|csv|...>]
```

### Required Arguments
- `--genotype-file`   : Path to the user genotype file
- `--gwas-db`         : Path to the GWAS summary statistics DuckDB database
- `--snps`            : Comma-separated list of SNP IDs (mutually exclusive with `--snps-file`)
- `--snps-file`       : File containing SNP IDs (one per line)

### Optional Arguments
- `--gwas-table`      : GWAS table name (default: first table in DB)
- `--reference-table` : Reference stats table name (default: `reference_panel`)
- `--output`          : Output file path (default: stdout)
- `--format`          : Output format (`json`, `csv`, etc.; default: `json`)

### Example

```sh
./risk-calculator \
  --genotype-file sample.vcf \
  --gwas-db gwas.duckdb \
  --snps-file snps.txt \
  --output results.json \
  --format json
```

## Data Requirements
- **Genotype File:** Supported formats include VCF and simple text (see documentation for details).
- **GWAS Database:** DuckDB format with summary statistics. Table schema must match the canonical data model (see `.agent/data_model.md`).
- **SNPs:** Provide either a list (`--snps`) or a file (`--snps-file`).
- **Reference Table:** Used for normalization; defaults to `reference_panel` if not specified.

## Output
- **PRS Results:** Polygenic risk scores for each trait.
- **Normalized PRS:** Scores normalized using reference panel statistics.
- **Trait Summaries:** Optional, depending on output format.
- **Missing SNPs:** List of SNPs missing from input or reference data.

Output is written to the specified file or stdout, in the selected format.

## Logging
- Logs are written at INFO and ERROR levels to standard error.
- All major steps and errors are logged with context.

## Testing
- All modules are tested using Go's built-in testing framework.
- Run tests with:
  ```sh
  go test ./...
  ```

## Project Structure
- `cmd/`      : CLI entrypoint
- `internal/` : Modular pipeline and core logic (genotype, gwas, prs, output, etc.)
- `.agent/`   : Developer docs, agent briefs, data models

## References
- [Canonical Data Model](.agent/data_model.md)
- [Developer Guide](.agent/README.md)

---
For questions or support, see the developer guide or contact the PHITE project maintainers.
