# Brief: Polygenic Risk Score (PRS) Reference Statistics: Retrieval and On-the-Fly Computation

**Date**: 2025-06-03 (Updated)
**Status**: Design Enhancement

## Summary

This document outlines the enhanced strategy for managing Polygenic Risk Score (PRS) reference statistics within the `polygenic-risk-calculator`. The system will employ a dual approach:

1. **Attempt to retrieve pre-computed PRS statistics** (mean, standard deviation, min, max) from a user-configured BigQuery table within their own GCP project (e.g., `jerkytreats`). This table acts as a cache.
2. **If statistics are not found in the cache, compute them on-the-fly**. This involves:
    * Using a defined PRS model (SNPs and weights)
    * Querying a public allele frequency dataset (e.g., gnomAD v3.x or later)
    * Calculating the necessary statistics.
    * **Uploading the newly computed statistics to the user's BigQuery cache table** for future retrieval.

This enhancement aims to provide flexibility, reduce redundant computations, and allow users to manage their own curated reference statistics while also leveraging on-demand calculations.

Refer to the main `.agent/README.md` for overall project goals and `.agent/data_model.md` for details on key data structures.

## Details

### 1. User's BigQuery Cache Table for PRS Statistics

The primary source for PRS reference statistics will be a table in the user's GCP project (defined by `bq_billing_project` for writes, and specific `bq_project`, `bq_dataset`, `bq_table` for read/write operations on the cache).

**Expected Schema:**

| Column Name   | Type      | Description                                                      | Notes                        |
|--------------|-----------|------------------------------------------------------------------|------------------------------|
| ancestry     | STRING    | Identifier for the ancestral population (e.g., 'EUR', 'AFR').    | Primary Key component        |
| trait        | STRING    | Identifier for the phenotype or trait.                           | Primary Key component        |
| model_id     | STRING    | Identifier for the specific PRS model used.                      | Primary Key component        |
| mean_prs     | FLOAT     | The mean of the PRS distribution.                                | Not null                     |
| std_dev_prs  | FLOAT     | The standard deviation of the PRS distribution.                  | Not null                     |
| min_prs      | FLOAT     | The minimum value of the PRS distribution.                       | Nullable; may be estimated   |
| max_prs      | FLOAT     | The maximum value of the PRS distribution.                       | Nullable; may be estimated   |
| sample_size  | INTEGER   | Number of samples used to compute the statistics.                | Nullable; optional           |
| source       | STRING    | How the data was generated (e.g., 'pre_computed', 'on_the_fly_calculated'). | Not null           |
| notes        | STRING    | Free-text notes or provenance information.                       | Nullable; optional           |
| last_updated | TIMESTAMP | When the record was last updated.                                | Not null                     |

**Field rationale:**
* `ancestry`, `trait`, `model_id`: Together form the primary key for each record.
* `mean_prs`, `std_dev_prs`: Core statistics for normalization and interpretation.
* `min_prs`, `max_prs`: Useful for range checks and percentiles; may be estimated if not directly computable.

---

### 1a. Required gnomAD BigQuery Table Schema for Allele Frequencies

To compute PRS reference statistics on-the-fly, the system requires access to a public allele frequency dataset, such as gnomAD. The following schema elements are required to support robust and accurate PRS calculations:

| Column Name        | Type      | Description                                                         | Notes                        |
|--------------------|-----------|---------------------------------------------------------------------|------------------------------|
| chrom              | STRING    | Chromosome (e.g., '1', '2', 'X', 'Y')                              | Part of variant identifier   |
| pos                | INTEGER   | 1-based position on the chromosome                                 | Part of variant identifier   |
| ref                | STRING    | Reference allele                                                   | Part of variant identifier   |
| alt                | STRING    | Alternate allele                                                   | Part of variant identifier   |
| ancestry           | STRING    | Ancestral population label (e.g., 'EUR', 'AFR', 'EAS')            | Required for filtering       |
| allele_freq        | FLOAT     | Alternate allele frequency in the given ancestry                   | Required for PRS computation |
| n_samples          | INTEGER   | Number of samples for this ancestry and variant                    | Optional; for QC/context     |
| variant_id         | STRING    | (Optional) Unique variant identifier (e.g., '1:12345:A:G')         | May be constructed from fields|
| ...                | ...       | Other columns as present in gnomAD (e.g., info, filters, etc.)     | Not required for PRS         |

**Field rationale:**
* `chrom`, `pos`, `ref`, `alt`: Together uniquely identify a variant and are needed to match PRS model SNPs to gnomAD records.
* `ancestry`: Enables ancestry-specific PRS computation by filtering for the relevant population.
* `allele_freq`: Core value needed to compute expected PRS distribution statistics.
* `n_samples`: Useful for quality control and weighting; optional.
* `variant_id`: May simplify joins if present, but can be constructed from the other fields if not.

> **Note:** The actual column names in gnomAD may differ (e.g., `contig` for chromosome, or population-specific frequency columns like `AF_nfe`, `AF_afr`, etc.). The application should be configurable to map these to the expected fields using config keys such as `allele_freq_source.variant_id_column_names`, `allele_freq_source.ancestry_column_name`, and `allele_freq_source.allele_freq_column_name`.

---

### 1b. gnomAD Dataset/Table Structure and Allele Frequency Extraction

The gnomAD public dataset is organized by chromosome. For example:
* **Dataset name:** `gnomAD`
* **Tables:** `genomes_chr1`, ..., `genomes_chr22`
  * **Note:** The table names may vary based on the version being used
  * **Implication:** Queries must select the correct table by chromosome (e.g., `gnomAD.genomes_chr1` for chromosome 1)

#### Table Structure Overview
* Each table contains records for genomic variants, with the following relevant fields:
  * `reference_name` (STRING): Chromosome (e.g., '1', '2', 'X')
  * `start_position` (INTEGER): 0-based start position
  * `end_position` (INTEGER): 0-based end position
  * `reference_bases` (STRING): Reference allele
  * `alternate_bases` (REPEATED RECORD): One record for each alternate allele, containing:
    * `alt` (STRING): Alternate base
    * `AF` (FLOAT): Overall alternate allele frequency
    * `AF_afr`, `AF_nfe`, `AF_eas`, etc. (FLOAT): Population-specific alternate allele frequencies
    * Additional ancestry* and sex-specific fields (e.g., `AF_afr_male`, `AF_nfe_female`)

#### Extraction of Ancestry-Specific Allele Frequency
* To obtain the allele frequency for a specific ancestry (e.g., African, European), extract the relevant `AF_<ancestry>` field from the `alternate_bases` record for the desired alternate allele.
* The ancestry label (e.g., 'AFR', 'NFE', 'EAS') must be mapped to the correct column (e.g., `AF_afr`, `AF_nfe`, `AF_eas`).
* The application should allow configuration of this mapping via `allele_freq_source.ancestry_column_name` or similar config keys.

#### Example Mapping for PRS Calculation
Suppose you need the allele frequency for a SNP on chromosome 1, position 123456, ref 'A', alt 'G', for the 'NFE' (Non-Finnish European) population:
* Query table: `gnomAD.genomes_chr1`
* Match fields: `reference_name = '1'`, `start_position = 123456`, `reference_bases = 'A'`, `alternate_bases.alt = 'G'`
* Extract: `alternate_bases.AF_nfe` as the allele frequency for 'NFE'

#### Configuration Requirements: Ancestry Mapping

* The mapping from logical ancestry labels (e.g., 'EUR', 'AFR', 'NFE') to gnomAD column names (e.g., 'AF_nfe', 'AF_afr') **must be configurable**.
* No hardcoded ancestry-to-column mapping is allowed in implementation code. All mappings must be provided via configuration.
* The configuration key for this mapping should be `allele_freq_source.ancestry_mapping` (or similar), and must be registered as a required config key in the configuration system.
* The application must use this mapping to extract the correct population-specific allele frequency column for the requested ancestry at runtime.

##### Example YAML Configuration
```yaml
allele_freq_source:
  gcp_project_id: bigquery-public-data
  dataset_id_pattern: gnomAD
  table_id_pattern: genomes_chr{chrom}
  ancestry_column_name: AF_nfe  # For Non-Finnish European
  variant_id_column_names: [reference_name, start_position, reference_bases, alternate_bases.alt]
```

> **Note:** The application must support flexible mapping to accommodate differences between gnomAD releases and ancestry naming conventions. Always consult the table schema for the exact field names.

* `sample_size`: Provides context for the reliability of the statistics (optional but recommended).
* `source`: Tracks whether the data was pre-computed or generated on-the-fly.
* `notes`: Allows for annotation of special cases, data provenance, or caveats.
* `last_updated`: Enables cache freshness checks and auditability.

> **Note:** Additional columns can be added as needed for project-specific requirements, but the above are required for full compatibility with the PRS reference statistics system.

### 2. On-the-Fly Computation and Caching

If a query for a specific `ancestry`, `trait`, and `model_id` yields no results from the user's cache table:

* The system will attempt to compute `mean_prs` and `std_dev_prs`.
  * **Required Inputs for Computation**:
    1. **PRS Model**: A definition of SNPs and their respective weights for the given `model_id`.
    2. **Allele Frequencies**: Access to a comprehensive allele frequency dataset, like gnomAD, filtered for the specified `ancestry`.
  * **Min/Max Handling**:
    * Direct computation of true min/max PRS from allele frequencies is complex.
    * Initially, on-the-fly computation might populate `min_prs` and `max_prs` as NULL or use an estimation (e.g., mean +/* 3\*std_dev_prs). Users can update these later if more precise values are known.
* **Caching**: Once computed, the statistics (including current `last_updated` timestamp) will be inserted into the user's BigQuery cache table.

### 3. Expanded Configuration Requirements

The application's configuration (`config.yaml` or environment variables) will need to be expanded:

* **User Cache Table (Primary PRS Stats Source)**:
  * `prs_stats_cache.gcp_project_id`: GCP Project ID where the user's cache table resides (e.g., "jerkytreats").
  * `prs_stats_cache.dataset_id`: BigQuery Dataset ID for the cache table.
  * `prs_stats_cache.table_id`: BigQuery Table ID for the cache table.
  * (The `bq_billing_project` will be used for quota/billing of these operations).
* **Allele Frequency Data Source (for on-the-fly computation)**:
  * `allele_freq_source.type`: (String, e.g., "bigquery_gnomad")
  * `allele_freq_source.gcp_project_id`: (String, e.g., "bigquery-public-data")
  * `allele_freq_source.dataset_id_pattern`: (String, e.g., "gnomAD")
  * `allele_freq_source.table_id_pattern`: (String, e.g., "genomes_v{version}") * to accommodate different gnomAD versions.
  * `allele_freq_source.ancestry_column_name`: (String, e.g., " população" or "ancestry")
  * `allele_freq_source.allele_freq_column_name`: (String, e.g., "AF")
  * `allele_freq_source.variant_id_column_names`: (List of Strings, e.g., ["chrom", "pos", "ref", "alt"])
* **PRS Model Definition Source**:
  * `prs_model_source.type`: (String, e.g., "file", "bigquery_table")
  * `prs_model_source.path_or_table_uri`: (String) Path to model file or BQ table URI.
  * `prs_model_source.snp_column_name`, `prs_model_source.effect_allele_column_name`, `prs_model_source.weight_column_name`.

### 4. Setup and Permissions

Users will need to:
1. **Create the PRS Statistics Cache Table**: In their specified GCP project and dataset, create the BigQuery table (e.g., `prs_stats_cache.table_id`) with the schema defined in section 1.
2. **Grant Permissions**: Ensure the service account or user credentials used by the `polygenic-risk-calculator` have:
    * `BigQuery Data Viewer` and `BigQuery User` roles on the allele frequency source (e.g., gnomAD public datasets).
    * `BigQuery Data Editor` and `BigQuery User` roles on their own PRS statistics cache table/dataset to allow reads and writes.
3. **Configure the Application**: Populate all new configuration sections accurately.

This enhanced design provides a more robust and user-friendly approach to managing essential reference data for PRS calculations.
