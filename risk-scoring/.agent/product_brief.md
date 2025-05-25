# PHITE Genomic Risk Scoring – Product Brief: Data and Module Linkages

## Overview

The PHITE risk scoring pipeline is composed of five modular, locally-executed scripts. Each module consumes the outputs of the previous stage, ensuring a robust, privacy-preserving flow from raw genotype input to final risk reporting. All data remains local at every step, and each module validates its inputs and outputs for integrity and compliance.

---

## 1. Load User Genotype File

- **Input:**  
  - Path to user genotype file (`.txt` or `.csv`), e.g., AncestryDNA or 23andMe format.
- **Output:**  
  - Pandas DataFrame with columns: `rsid`, `genotype`.

**Linkage:**  
The output DataFrame of user SNP genotypes is passed as input to the GWAS query module, providing the list of `rsid` values to search for in the GWAS catalog.

---

## 2. Query Local GWAS Catalog

- **Input:**  
  - Path to GWAS association file (`.duckdb` or `.parquet`).
  - List of `rsid` values (from the previous module’s DataFrame).
- **Output:**  
  - Pandas DataFrame of GWAS associations filtered to only those matching the user’s `rsid` values.

**Linkage:**  
This filtered GWAS association DataFrame is used as input for ontology grouping, ensuring only relevant SNPs are processed downstream. The rsid list may be large (up to 600k), so this module is optimized for efficiency.

---

## 3. Group SNPs by Ontology

- **Input:**  
  - GWAS association data (filtered DataFrame from previous module).
  - Path(s) to trait ontology mapping file(s).
- **Output:**  
  - DataFrame or dictionary grouping SNPs by higher-level trait clusters (e.g., “Cardiovascular disease”).

**Linkage:**  
This grouped structure provides the mapping from SNPs to trait clusters, which is essential for polygenic risk calculation by cluster in the next module.

---

## 4. Compute Polygenic Risk Scores

- **Input:**  
  - Grouped SNP data (from ontology grouping).
  - User genotype data (original DataFrame from module 1).
- **Output:**  
  - DataFrame summarizing polygenic risk scores (PRS) per trait cluster.

**Linkage:**  
The PRS summary DataFrame is the core quantitative result, feeding directly into the reporting module for user-facing interpretation.

---

## 5. Generate Risk Report

- **Input:**  
  - PRS summary data (from previous module).
  - Trait grouping data (from ontology grouping).
- **Output:**  
  - Markdown or HTML report, saved locally, presenting risk scores, trait clusters, matched SNPs, effect directions, and confidence intervals (if available).

**Linkage:**  
This report is the final product, integrating all prior outputs into a user-readable summary, strictly generated and stored locally to maintain privacy.

---

## Data and Privacy Flow

- **All data remains local** at every step; no external transmission or API calls are permitted.
- **Each module validates** its inputs and outputs, ensuring file existence, format correctness, and required columns.
- **Clear error handling** is required at every stage to prevent propagation of malformed or incomplete data.

---

## Directory and File Structure

- Scripts are placed in `risk-scoring/scripts/`.
- Feature briefs and documentation are in `risk-scoring/.agent/`.
- All intermediate and final outputs are written to local directories as specified in each module’s brief.

---

## Summary Table: Module Linkages

| Module                | Consumes Input(s) From           | Produces Output For         |
|-----------------------|----------------------------------|----------------------------|
| Load User Genotype    | User file                        | Query GWAS                 |
| Query GWAS            | Genotype DataFrame, GWAS file    | Ontology Grouping          |
| Ontology Grouping     | Filtered GWAS Data, Mapping file | Polygenic Score, Reporting |
| Polygenic Score       | Grouped SNPs, Genotype Data      | Reporting                  |
| Report Generator      | PRS, Grouping Data               | User (final report)        |

---

This linkage ensures a seamless, privacy-first, and auditable data flow from raw input to actionable genomic risk insights.
