# Agent Brief: SNP Annotation Engine

## Purpose
Fetch GWAS association data for validated SNPs and annotate user genotypes with risk allele, effect size, and computed dosage for each SNP.

## Dependencies
- Genotype Input Handler
- GWAS associations table

## Structs
See [data_model.md](./data_model.md) for canonical struct definitions:
- `GWASSNPRecord`
- `AnnotatedSNP`

## Inputs
- `validated_snps`: List of `ValidatedSNP` from Genotype Input Handler
- `associations_clean`: GWAS table reference

## Outputs
- `annotated_snps`: List of `AnnotatedSNP` (with effect size, risk allele, and computed dosage)
- `gwas_records`: List of `GWASSNPRecord` (if needed for downstream)

## Consumed By
- `annotated_snps` → Polygenic Risk Score Calculator, Trait-Specific Summary Generator

## Special Notes
- Must handle multiple associations per SNP if present
- Should handle ambiguous or missing genotype calls
- Output must be ready for direct use by Polygenic Risk Score Calculator

## Required Tests
- Correctly fetches GWAS data for all `validated_snps`.
- Annotates each SNP with the correct risk allele, effect size, and computes correct dosage.
- Handles multiple associations per SNP (returns all or selects as specified).
- Handles ambiguous or missing genotype calls.
- Produces `AnnotatedSNP` objects with correct fields and types.
- Fails gracefully if GWAS data is missing for a SNP.
