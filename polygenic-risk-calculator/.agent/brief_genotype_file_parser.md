# Agent Brief: Genotype Input Handler

## Purpose
Parse user genotype files (AncestryDNA or 23andMe), extract SNP genotypes, validate requested SNPs, and report missing SNPs for downstream analysis.

## Dependencies
- Standard file I/O and text parsing libraries
- GWAS associations table (for validation)

## Structs
See [data_model.md](./data_model.md) for canonical struct definitions:
- `UserGenotype`
- `SNPRequest`
- `ValidatedSNP`

## Inputs
- `genotype_file`: Path to user genotype file (AncestryDNA or 23andMe)
- `requested_snps`: List of SNP rsids
- `associations_clean`: GWAS table reference

## Outputs
- `user_genotypes`: List of `UserGenotype`
- `validated_snps`: List of `ValidatedSNP` (present in both user data and GWAS)
- `snps_missing`: List of rsids not found in user data or GWAS

## Consumed By
- `validated_snps` → SNP Annotation Engine
- `snps_missing` → Output Formatter

## Special Notes
- Must autodetect file format
- Should handle missing/malformed lines gracefully
- Output must be ready for direct use by SNP Annotation Engine

## Required Tests
- Parses valid AncestryDNA and 23andMe files, producing correct `user_genotypes`.
- Correctly identifies and returns `validated_snps` present in both user data and GWAS.
- Correctly lists `snps_missing` for SNPs not found in user data or GWAS.
- Non `GACT` SNP's are considered missing.
- Handles malformed genotype lines gracefully (skips, warns, or errors as specified).
- Autodetects file format correctly.
- Handles empty files and files with only headers.
- Produces output objects in the correct format and types.
