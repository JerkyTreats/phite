# Feature Brief: Compute Polygenic Risk Scores

## Overview
Create a Python script to compute polygenic risk scores (PRS) for each trait cluster. For each cluster, calculate PRS as the sum of effect sizes multiplied by user genotype values.

## Requirements
- Input DataFrames must match the structure of grouped SNP and user genotype tables (see Data Model Spec for full column list).
- Output DataFrame must include: `topic`, `topic_uri`, `PRS`, and optionally `ci_95_text`, `contributing_snps`.
- For additional context on required columns and types, review `data_model_spec.md` in the `.agent` folder.
- Use `pandas` for DataFrame operations.
- Enforce strict input validation for grouped SNP and genotype data, including required columns.
- Define expected behavior for missing or ambiguous data (e.g., skip, warn, or fail).
- Accept grouped SNP data and user genotype data as parameters.
- For each cluster, compute:  
  \( PRS = \sum_i (\beta_i \cdot g_i) \)  
  where \( \beta_i \) is GWAS effect size, \( g_i \) is user genotype (0, 1, 2).
- Output a summary DataFrame with PRS per cluster.
- Validate input structure and required columns; handle errors gracefully.
- All processing must remain local; no network calls or external data transmission.

## Inputs
- Grouped SNP data (DataFrame or dictionary).
- User genotype data (DataFrame).

## Outputs
- DataFrame summarizing PRS per trait cluster.

## Privacy & Validation
- Validate input types, structure, and required columns.
- Provide clear error messages for invalid input.
- Log or report on any data that is ignored or ambiguous.
- Include test cases or validation for edge cases during development.
- Do not transmit, upload, or expose any data.
- Validate input types, structure, and required columns.
- Provide clear error messages for invalid input.

## Directory
- Place script in `risk-scoring/scripts/polygenic_score.py`.

---

## Unit Tests
- Test PRS calculation for a known genotype and effect size.
- Test output DataFrame includes `topic`, `topic_uri`, `PRS`.
- Test missing or ambiguous genotype values handled as specified (skip, warn, or fail).
- Test output for empty input is empty DataFrame.
- Test error raised if required columns missing in input.
