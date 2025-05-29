# Agent Brief: Entrypoint (main.go)

## Purpose
Coordinate the execution of all components of the polygenic risk calculator, handling command-line arguments, orchestrating the pipeline, and managing overall input/output.

## Dependencies
- All functional modules/agents:
  - Genotype Input Handler
  - SNP Annotation Engine
  - Polygenic Risk Score Calculator
  - Score Normalizer
  - Trait-Specific Summary Generator
  - Output Formatter

## Inputs
- Command-line arguments:
  - `--genotype-file <path>`: Path to user genotype file (AncestryDNA or 23andMe)
  - `--snps <rsid,...>`: Comma-separated list of SNP rsids
  - `--output <path>`: (Optional) Output file path
  - `--format <json|csv>`: (Optional) Output format
  - Any other configuration as specified in README

## Outputs
- Writes formatted results to file or stdout, as specified by arguments

## Responsibilities
- Parse and validate command-line arguments
- Load and validate all required input files
- Orchestrate the execution of all pipeline components in correct order
- Handle and report errors gracefully
- Ensure all outputs conform to the data model
- Exit with appropriate status code

## Consumed By
- End user (CLI)
- Downstream automation or scripting (via output file/stdout)

## Required Tests
- Handles all required and optional command-line arguments correctly
- Fails gracefully with helpful error messages for missing/invalid arguments
- Successfully runs the full pipeline with valid inputs, producing correct output
- Handles all error and edge cases (missing files, malformed input, etc.)
- Produces output in the correct format (JSON/CSV) and location
- Exits with correct status code on success or failure
