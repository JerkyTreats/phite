# Agent Brief: Output Formatter

## Purpose
Format the final results as JSON or CSV and write to file or stdout.

## Dependencies
- All above components

## Inputs
See [data_model.md](./data_model.md) for canonical struct definitions.
- `normalized_prs`: `NormalizedPRS` struct (from Score Normalizer)
- `prs_result`: `PRSResult` struct (from Polygenic Risk Score Calculator)
- `trait_summaries`: List of `TraitSummary` (from Trait-Specific Summary Generator)
- `snps_missing`: List of rsids (from Genotype Input Handler)
- Output format and file path

## Outputs
- Formatted output file or stdout

## Special Notes
- Should validate output structure against the defined data model

## Required Tests
- Serializes all output objects (`NormalizedPRS`, `PRSResult`, `TraitSummary`, `snps_missing`) to JSON and CSV as specified.
- Handles output to file and stdout.
- Validates output structure matches data model.
- Handles missing or empty input objects gracefully.
- Produces human-readable and machine-parseable output.
