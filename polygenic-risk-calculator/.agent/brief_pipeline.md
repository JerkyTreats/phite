# Agent Brief: Pipeline Orchestration

## Purpose
Coordinate the end-to-end workflow for polygenic risk score analysis, integrating all domain agents (genotype parser, GWAS loader, SNP annotator, PRS calculator, reference stats loader, score normalizer, trait summary generator, and output formatter) into a unified, testable, and modular pipeline.

## Responsibilities
- Orchestrate the following steps:
  1. **GWAS Record Fetching:** Load GWAS association records for requested SNPs from DuckDB.
  2. **Genotype Parsing:** Parse and validate user genotype data from supported file formats.
  3. **SNP Annotation:** Annotate validated SNPs with GWAS effect sizes, risk alleles, and traits.
  4. **Trait Aggregation:** Identify all unique traits present in the annotated SNPs.
  5. **Per-Trait Scoring:** For each trait:
     - Filter annotated SNPs for the trait.
     - Calculate the PRS.
     - Load trait-specific reference statistics for normalization.
     - Normalize the PRS using the appropriate reference stats.
     - Generate a trait summary (risk alleles, effect-weighted contribution, risk level).
  6. **Output Formatting:** Structure and serialize results for reporting or downstream consumption.
- Handle errors and missing data gracefully at each step.
- Expose a single entrypoint for the CLI and other consumers.

## Inputs
- User genotype file path
- List of SNPs (direct or file-based)
- GWAS DuckDB path and table
- Reference stats DuckDB path
- Output configuration (format, file path)
- Optional: ancestry, model, and other parameters

## Outputs
- Trait-specific PRS results and normalized scores
- Trait summaries
- List of missing SNPs
- Structured output for reporting

## Consumed By
- CLI entrypoint (`main.go`)
- Potentially other interfaces (API, batch jobs)

## Required Tests
- End-to-end tests covering the full pipeline for single and multi-trait scenarios
- Unit tests for error handling at each stage
- Table-driven tests for trait-specific scoring and normalization
- Tests for correct output structure and error reporting

## Special Notes
- The pipeline must be modular, with each stage implemented as a composable agent/component.
- All configuration must be passed explicitly; no hidden global state.
- Pipeline orchestration logic should reside in `internal/pipeline/`.
- Reference the canonical data model in `.agent/data_model.md` for all input/output types.

---

**Rationale:**
This agent ensures that the PHITE polygenic risk calculator remains maintainable, extensible, and robust as new traits, input types, or scoring models are introduced. Centralizing orchestration in a dedicated pipeline agent enables clear separation of concerns, easier testing, and reliable multi-trait support.
