# Agent Brief: Score Normalizer

## Purpose
Normalize the raw PRS against a reference population to yield a Z-score or percentile.

## Dependencies
- Polygenic Risk Score Calculator
- Reference population statistics (mean, std, or distribution)

## Structs
See [data_model.md](./data_model.md) for canonical struct definitions:
- `NormalizedPRS`

## Inputs
- `prs_result`: `PRSResult` from PRS Calculator
- `reference_stats`: Population mean/std or distribution

## Outputs
- `normalized_prs`: `NormalizedPRS` struct with raw, z, and percentile scores

## Consumed By
- `normalized_prs` → Trait-Specific Summary Generator, Output Formatter

## Special Notes
- Reference stats must be precomputed and available

## Required Tests
- Computes correct z-score and percentile given a `PRSResult` and reference stats.
- Handles edge cases (PRS at mean, min, max of reference).
- Fails gracefully if reference stats are missing or malformed.
- Returns `NormalizedPRS` with correct fields and types.
