# Reference Stats Data Schema Overview

> **Deprecation Notice (June 2025):**
> 
> All local reference panel workflows (sample metadata and VCF ingestion into DuckDB) are now deprecated. Users should use the Google BigQuery public gnomAD datasets (`bigquery-public-data.gnomad`) for all reference data queries. Local scripts and schemas for reference panel ingestion are no longer maintained or recommended for use. This change enables scalable, cloud-native analysis without the need for large local downloads or storage.

This document provides a holistic overview of the reference statistics (reference_stats) data schema and requirements for the PHITE GWAS/PRS system. It is intended to complement the detailed implementation briefs in this directory by describing the overall feature, its motivation, and how the various components fit together.

## Purpose

The reference_stats system enables ancestry-aware normalization and polygenic risk score (PRS) calculation by providing:
- A standardized reference panel of samples (with ancestry, sex, and sample IDs)
- Access to high-quality, genome-aligned variant data (VCFs)
- The infrastructure to support reproducible, automated, and idempotent ingestion of reference data into the GWAS database

## Key Requirements

- **Genome Build:** All reference data must be aligned to GRCh38 (hg38) using the HGDP + 1000 Genomes (gnomAD v3.1.2) resource.
- **Reproducibility:** The entire reference panel and stats setup must be fully automated and reproducible from scratch via a single entry point (`build_db.sh`).
- **Idempotency:** Scripts should not re-download or re-insert data if it already exists.
- **Provenance:** Input files and script versions must be tracked for reproducibility.
- **Integration:** All scripts and data ingestion steps must be invoked by `build_db.sh`.

## Data Schema Components

### 1. Reference Panel Table (`reference_panel`)
- **Purpose:** Stores the full, unfiltered sample metadata for all reference samples, enabling flexible ancestry or population selection at query time.
- **Schema (example, all columns from gnomAD v3.1.2 sample panel):**
  - `sample_id TEXT`
  - `sex TEXT`
  - `pop TEXT` (population)
  - `super_pop TEXT` (super-population/ancestry)
  - `platform TEXT`
  - ... (all other columns present in the raw metadata TSV)
- **Source:** Full gnomAD v3.1.2 sample panel metadata (see [reference_stats_brief.md](./reference_stats_brief.md))
- **Note:** Filtering for ancestry or other subsets should be performed via SQL query, not at ingestion.

### 2. Reference VCFs
- **Purpose:** Provide variant site data for all reference samples, enabling downstream PRS and allele frequency calculations
- **Format:** bgzipped, tabix-indexed VCFs for chromosomes 1–22
- **Source:** gnomAD v3.1.2 (see [vcf_download_hgdp_gnomad_grch38.md](./vcf_download_hgdp_gnomad_grch38.md))
- **Note:** VCFs are stored on disk and not loaded into DuckDB; they are used for downstream analysis and stats calculation.

### 3. PRS Reference Statistics (Future/Optional)
- **Purpose:** Store allele frequencies, LD scores, and other reference stats needed for PRS normalization
- **Schema:** To be defined as part of PRS pipeline expansion. Should be created in SQL and referenced by `build_db.sh`.

## Workflow Summary

1. **Database Schema Creation:** All required tables (including `reference_panel`) are created by SQL scripts invoked by `build_db.sh`.
2. **Reference Data Ingestion:**
   - Sample panel metadata is downloaded, filtered for target ancestry, and loaded into DuckDB.
   - VCFs are batch-downloaded and stored for downstream use.
3. **Reproducibility:** The entire process can be rerun at any time for a clean, fully-populated, PRS-ready database.

## References to Implementation Briefs

- [reference_stats_brief.md](./reference_stats_brief.md): Main workflow and requirements for reference panel/sample ingestion
- [vcf_download_hgdp_gnomad_grch38.md](./vcf_download_hgdp_gnomad_grch38.md): Detailed VCF download script and notes for GRCh38

## Future Expansion

- Additional reference statistics (e.g., allele frequencies, LD scores) can be added as new tables and ingestion scripts, following the same reproducibility and integration standards.

---

For implementation details, see the individual briefs in this directory. For overall reproducibility and integration, refer to this README and the project-level `.agent/README.md`.
