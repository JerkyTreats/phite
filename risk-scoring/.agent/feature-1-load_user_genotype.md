# Feature Brief: Load User Genotype File

## Overview
Create a Python script to load a user genotype file in `.txt` or `.csv` format (AncestryDNA or 23andMe format). The script must parse the file and output a DataFrame containing columns: `rsid`, `genotype`.

## Requirements
- Output DataFrame must have columns: `rsid`, `genotype` (see Data Model Spec for details).
- For additional context on required columns and types, review `data_model_spec.md` in the `.agent` folder.
- Use `pandas` for DataFrame operations.
- Enforce strict input validation and clear error handling for file parsing and required columns.
- Accept input file path as a parameter.
- Support `.txt` and `.csv` formats from major consumer genomics providers.
- Output a DataFrame with columns: `rsid`, `genotype`.
- Validate file format and required columns; handle malformed files gracefully.
- All processing must remain local; no network calls or external data transmission.

## Inputs
- Path to user genotype file (`.txt` or `.csv`).

## Outputs
- Pandas DataFrame with columns: `rsid`, `genotype`.

## Privacy & Validation
- Validate file type and content before processing.
- Provide clear error messages for invalid input.
- Standardize error handling and validation across modules.
- Do not transmit, upload, or expose any user data.
- Validate file type and content before processing.
- Provide clear error messages for invalid input.

## Directory
- Place script in `risk-scoring/scripts/load_user_genotype.py`.

---

## Unit Tests
- Test valid AncestryDNA/23andMe file loads correctly.
- Test output DataFrame has columns `rsid`, `genotype` only.
- Test missing required columns raises clear error.
- Test invalid file type (e.g., `.xlsx`) raises error.
- Test malformed genotype file (e.g., wrong delimiter, missing values) handled gracefully.
- Test no network calls are made (mock network libraries if needed).
