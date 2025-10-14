# Polygenic Risk Calculator

A command-line tool for computing polygenic risk scores (PRS) from user genotype data and GWAS summary statistics. Part of the PHITE project.

## Features

- **Genotype Parsing**: Supports AncestryDNA and 23andMe format files
- **GWAS Integration**: Works with DuckDB databases containing GWAS summary statistics
- **Flexible SNP Selection**: Specify SNPs via comma-separated list or file
- **PRS Calculation**: Computes polygenic risk scores with normalization
- **Multiple Output Formats**: JSON and CSV output support
- **Comprehensive Logging**: Structured logging for all operations

## Installation

```sh
git clone <repository-url>
cd polygenic-risk-calculator
go build -o risk-calculator ./cmd/risk-calculator
```

## Usage

### Basic Command

```sh
./risk-calculator \
  --genotype-file <genotype.txt> \
  --gwas-db <gwas.duckdb> \
  --snps <rsID1,rsID2,...> \
  [--output <results.json>] \
  [--format <json|csv>]
```

### Required Arguments

- `--genotype-file`: Path to genotype file (AncestryDNA/23andMe format)
- `--gwas-db`: Path to GWAS summary statistics DuckDB database
- `--snps`: Comma-separated list of SNP IDs (or use `--snps-file`)

### Optional Arguments

- `--snps-file`: File containing SNP IDs (one per line, alternative to `--snps`)
- `--gwas-table`: GWAS table name (default: first table in database)
- `--reference-table`: Reference stats table name (default: `reference_panel`)
- `--output`: Output file path (default: stdout)
- `--format`: Output format (`json` or `csv`, default: `json`)

### Example

```sh
./risk-calculator \
  --genotype-file sample_genotype.txt \
  --gwas-db gwas_summary.duckdb \
  --snps rs190214723,rs3131972,rs12562034 \
  --output results.json \
  --format json
```

## Data Requirements

### Genotype File Format
Tab-delimited file with header row:
```
rsid	chromosome	position	allele1	allele2
rs190214723	1	693625	T	T
rs3131972	1	752721	G	G
```

### GWAS Database
DuckDB format with required columns:
- `rsid`: SNP identifier
- `chromosome`: Chromosome number
- `position`: Genomic position
- `risk_allele`: Risk allele
- `beta`: Effect size
- `trait`: Trait name

## Output

The tool outputs:
- **Raw PRS Score**: Unnormalized polygenic risk score
- **Normalized PRS**: Z-score and percentile relative to reference population
- **Trait Summaries**: Risk level assessment and SNP contribution details
- **Missing SNPs**: List of SNPs not found in input or reference data

## Development

### Project Structure
- `cmd/risk-calculator/`: CLI entrypoint
- `internal/`: Core implementation modules
- `.agent/`: Development documentation and specifications

### Testing
```sh
go test ./...
```

### Configuration
The tool supports configuration via:
- Command-line flags (highest precedence)
- Environment variables
- Configuration files

## References

- [Data Model Specification](.agent/data_model.md)
- [Developer Guide](.agent/README.md)

For questions or support, see the developer documentation or contact the PHITE project maintainers.
