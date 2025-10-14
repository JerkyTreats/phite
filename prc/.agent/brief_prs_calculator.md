# Agent Brief: Polygenic Risk Score Calculator

## Purpose
Compute the raw polygenic risk score using the formula: PRS = Σ (genotype dosage × effect size)

## Dependencies
- SNP Annotation Engine

## Structs
See [data_model.md](./data_model.md) for canonical struct definitions:
- `SNPContribution`
- `PRSResult`

## Inputs
- `annotated_snps`: List of `AnnotatedSNP` (from SNP Annotation Engine)

## Outputs
- `prs_result`: `PRSResult` struct containing the raw score and per-SNP contributions

## Consumed By
- `prs_result` → Score Normalizer, Trait-Specific Summary Generator, Output Formatter

## Special Notes
- Should handle negative (protective) effect sizes
- Exclude or adjust for missing SNPs

## Required Tests
- Computes PRS as Σ (dosage × effect size) for a set of `AnnotatedSNP`s.
- Handles negative (protective) effect sizes correctly.
- Excludes or adjusts for missing SNPs as specified.
- Returns a `PRSResult` with correct `prs_score` and detailed per-SNP contributions.
- Handles empty input (returns score 0 or as specified).
