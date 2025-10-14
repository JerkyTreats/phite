# Agent Brief: Trait-Specific Summary Generator

## Purpose
Generate a summary for each trait, including number of risk alleles, effect-weighted contribution, and risk interpretation.

## Dependencies
- SNP Annotation Engine
- Polygenic Risk Score Calculator
- Score Normalizer

## Structs
See [data_model.md](./data_model.md) for canonical struct definitions:
- `TraitSummary`

## Inputs
- `annotated_snps`: List of `AnnotatedSNP` (with trait info)
- `normalized_prs`: `NormalizedPRS` from Score Normalizer

## Outputs
- `trait_summaries`: List of `TraitSummary`

## Consumed By
- `trait_summaries` → Output Formatter

## Special Notes
- Should support multiple traits if SNPs map to more than one

## Required Tests
- Aggregates per-SNP data into correct trait summaries (number of risk alleles, effect-weighted contribution).
- Assigns correct risk interpretation (low/moderate/high) based on normalized PRS.
- Handles multiple traits per user.
- Handles missing or ambiguous trait information gracefully.
- Returns `TraitSummary` objects with correct fields and types.
