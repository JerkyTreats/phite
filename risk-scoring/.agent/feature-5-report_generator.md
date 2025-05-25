# Feature Brief: Generate Risk Report

## Overview
Create a Python script to generate a summary report from PRS and trait group data. The script should render a sorted list of trait clusters by risk score, including per-trait matched SNPs, effect directions, and confidence intervals (if present), and output the report in Markdown or HTML format.

## Requirements
- Input DataFrames must match the structure of PRS summary and trait grouping tables (see Data Model Spec for full column list).
- For additional context on required columns and types, review `data_model_spec.md` in the `.agent` folder.
- Specify the output directory and filename pattern for reports (e.g., `risk-scoring/reports/report_<timestamp>.md`).
- Use `jinja2` and `markdown` for rendering reports.
- Enforce strict input validation for PRS and grouping data.
- Define expected behavior for missing fields or data (e.g., skip, warn, or fail).
- Accept PRS data and grouping data as parameters.
- Render a sorted summary of trait clusters by risk score.
- Include per-trait matched SNPs, effect directions, and confidence intervals (if available).
- Output the report in Markdown or HTML format and save locally.
- Validate input structure and required fields; handle errors gracefully.
- All processing must remain local; no network calls or external data transmission.

## Inputs
- PRS summary data (DataFrame).
- Trait grouping data (DataFrame or dictionary).

## Outputs
- Markdown or HTML report saved to local directory.

## Privacy & Validation
- Validate input types and required fields.
- Provide clear error messages for invalid input.
- Log or report on any data that is ignored or ambiguous.
- Include test cases or validation for edge cases during development.
- Do not transmit, upload, or expose any data.
- Validate input types and required fields.
- Provide clear error messages for invalid input.

## Directory
- Place script in `risk-scoring/scripts/report_generator.py`.
