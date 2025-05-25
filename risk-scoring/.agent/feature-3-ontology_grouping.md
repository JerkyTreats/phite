# Feature Brief: Group SNPs by Ontology

## Overview
Create a Python script to group SNP associations by ontology or trait cluster. The script should join GWAS associations to trait ontology mappings using provided mapping files and output a grouped DataFrame or dictionary.

## Requirements
- Input DataFrame must match the structure of `associations_clean` and trait ontology mapping tables (see Data Model Spec for full column list).
- Output must include trait/topic columns as specified in the Data Model Spec.
- For additional context on required columns and types, review `data_model_spec.md` in the `.agent` folder.
- Explicitly require mapping file(s) to include schema: at minimum, columns such as `rsid`, `trait_uri`, `ontology_cluster`.
- Provide example mapping file row(s) in documentation.
- Validate mapping file structure on load and raise clear errors if expectations are not met.
- Handle ambiguous or missing mappings by logging and/or warning, and define expected behavior (e.g., skip or fail).
- Accept GWAS association data and trait mapping file paths as parameters.
- Join on trait/ontology identifiers as specified in mapping files.
- Output a DataFrame or dictionary grouping SNPs under higher-level trait clusters.
- Validate mapping file format and required columns; handle errors gracefully.
- All processing must remain local; no network calls or external data transmission.

## Inputs
- GWAS association data (DataFrame).
- Path(s) to trait ontology mapping file(s).

## Outputs
- DataFrame or dictionary grouping SNPs by trait cluster.

## Privacy & Validation
- Validate mapping file existence, format, and required columns before processing.
- Provide clear error messages for invalid input or ambiguous mappings.
- Log or report on any data that is ignored or ambiguous.
- Include test cases or validation for edge cases during development.
- Do not transmit, upload, or expose any data.
- Validate mapping file existence, format, and required columns.
- Provide clear error messages for invalid input.

## Directory
- Place script in `risk-scoring/scripts/ontology_grouping.py`.
