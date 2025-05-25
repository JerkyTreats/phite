# Feature Brief: Query Local GWAS Catalog

## Overview
Create a Python script to load GWAS association data from a local DuckDB or Parquet file. The script should filter the data to retain only rows where `rsid` matches those in a provided list, and output the filtered GWAS data as a DataFrame.

## Requirements
- Input DataFrame must have column: `rsid` (see Data Model Spec for details).
- Output DataFrame must match the structure of `associations_clean` (see Data Model Spec for full column list).
- For additional context on required columns and types, review `data_model_spec.md` in the `.agent` folder.
- Use efficient libraries for large-scale filtering (e.g., `pandas`, `duckdb`).
- Benchmark and optimize filtering operations (e.g., vectorized queries, indexed lookups) for large rsid lists.
- Document performance expectations and provide guidance for tuning or scaling if needed.
- Accept GWAS file path and list of `rsid` values as parameters. Note that the list of `rsid` values can be large (up to 600k values) and the script should be optimized to handle this.
- Support both DuckDB and Parquet file formats.
- Output a DataFrame containing only matching SNPs.
- Validate file existence, format, and required columns; handle errors gracefully.
- All processing must remain local; no network calls or external data transmission.

## Inputs
- Path to GWAS association file (`.duckdb` or `.parquet`).
- List of `rsid` values to filter.

## Outputs
- Pandas DataFrame containing filtered GWAS associations.

## Privacy & Validation
- Validate file type, existence, and required columns before processing.
- Provide clear error messages for invalid input.
- Enforce strict input validation at the start of the module.
- Standardize error messages and handling (raise exceptions with actionable messages).
- Do not transmit, upload, or expose any data.
- Validate file type, existence, and required columns before processing.
- Provide clear error messages for invalid input.

## Directory
- Place script in `risk-scoring/scripts/query_gwas.py`.

---

## Unit Tests
- Test filtering returns only matching `rsid` values.
- Test output DataFrame matches `associations_clean` schema.
- Test empty input list returns empty DataFrame.
- Test invalid input DataFrame (missing `rsid`) raises error.
- Test performance on large input (mock or use sample data).
- Test error handling for missing GWAS source file.
